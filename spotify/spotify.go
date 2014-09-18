// A lot of pieces copied from the awesome library github.com/op/go-libspotify by Örjan Persson
package spotify

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"code.google.com/p/portaudio-go/portaudio"
	"github.com/fabiofalci/sconsify/events"
	"github.com/mitchellh/go-homedir"
	sp "github.com/op/go-libspotify/spotify"
)

type audio struct {
	format sp.AudioFormat
	frames []byte
}

type audio2 struct {
	format sp.AudioFormat
	frames []int16
}

type portAudio struct {
	buffer chan *audio
}

type Spotify struct {
	currentTrack  *sp.Track
	paused        bool
	cacheLocation string
	events        *events.Events
	pa            *portAudio
	session       *sp.Session
	appKey        *[]byte
}

func Initialise(username *string, pass *[]byte, events *events.Events) {
	if err := initialiseSpotify(username, pass, events); err != nil {
		fmt.Printf("Error: %v\n", err)
		events.Shutdown()
	}
}

func initialiseSpotify(username *string, pass *[]byte, events *events.Events) error {
	spotify := &Spotify{events: events}
	err := spotify.initKey()
	if err != nil {
		return err
	}
	spotify.initAudio()
	defer portaudio.Terminate()

	err = spotify.initCache()
	if err != nil {
		return err
	}

	spotify.initSession()

	err = spotify.login(username, pass)
	if err != nil {
		return err
	}

	err = spotify.checkIfLoggedIn()
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) login(username *string, pass *[]byte) error {
	credentials := sp.Credentials{Username: *username, Password: string(*pass)}
	if err := spotify.session.Login(credentials, false); err != nil {
		return err
	}

	err := <-spotify.session.LoginUpdates()
	if err != nil {
		return err
	}

	return nil
}

func (spotify *Spotify) initSession() error {
	var err error
	spotify.session, err = sp.NewSession(&sp.Config{
		ApplicationKey:   *spotify.appKey,
		ApplicationName:  "sconsify",
		CacheLocation:    spotify.cacheLocation,
		SettingsLocation: spotify.cacheLocation,
		AudioConsumer:    spotify.pa,
	})

	if err != nil {
		return err
	}
	return nil
}

func (spotify *Spotify) initKey() error {
	var err error
	spotify.appKey, err = getKey()
	if err != nil {
		return err
	}
	return nil
}

func newPortAudio() *portAudio {
	return &portAudio{buffer: make(chan *audio, 8)}
}

func (spotify *Spotify) initAudio() {
	portaudio.Initialize()

	spotify.pa = newPortAudio()
}

func (spotify *Spotify) initCache() error {
	spotify.initCacheLocation()
	if spotify.cacheLocation == "" {
		return errors.New("Cannot find cache dir")
	}

	spotify.deleteCache()
	return nil
}

func (spotify *Spotify) initCacheLocation() {
	dir, err := homedir.Dir()
	if err == nil {
		dir, err = homedir.Expand(dir)
		if err == nil && dir != "" {
			spotify.cacheLocation = dir + "/.sconsify/cache/"
		}
	}
}

func (spotify *Spotify) shutdownSpotify() {
	spotify.session.Logout()
	spotify.deleteCache()
	spotify.events.Shutdown()
}

func (spotify *Spotify) deleteCache() {
	if strings.HasSuffix(spotify.cacheLocation, "/.sconsify/cache/") {
		os.RemoveAll(spotify.cacheLocation)
	}
}

func (spotify *Spotify) checkIfLoggedIn() error {
	if spotify.waitForConnectionStateUpdates() {
		spotify.finishInitialisation()
	} else {
		spotify.events.NewPlaylist(nil)
		return errors.New("Could not login")
	}
	return nil
}

func (spotify *Spotify) waitForConnectionStateUpdates() bool {
	timeout := make(chan bool)
	go func() {
		time.Sleep(9 * time.Second)
		timeout <- true
	}()
	loggedIn := false
	running := true
	for running {
		select {
		case <-spotify.session.ConnectionStateUpdates():
			if spotify.isLoggedIn() {
				running = false
				loggedIn = true
			}
		case <-timeout:
			running = false
		}
	}
	return loggedIn
}

func (spotify *Spotify) isLoggedIn() bool {
	return spotify.session.ConnectionState() == sp.ConnectionStateLoggedIn
}

func (spotify *Spotify) finishInitialisation() {
	playlists := make(map[string]*sp.Playlist)
	allPlaylists, _ := spotify.session.Playlists()
	allPlaylists.Wait()
	for i := 0; i < allPlaylists.Playlists(); i++ {
		playlist := allPlaylists.Playlist(i)
		playlist.Wait()

		if allPlaylists.PlaylistType(i) == sp.PlaylistTypePlaylist {
			playlists[playlist.Name()] = playlist
		}
	}

	spotify.events.NewPlaylist(&playlists)

	go spotify.pa.player()

	go func() {
		for {
			select {
			case <-spotify.session.EndOfTrackUpdates():
				spotify.events.NextPlay <- true
			}
		}
	}()
	for {
		select {
		case track := <-spotify.events.ToPlay:
			spotify.play(track)
		case <-spotify.events.WaitForPause():
			spotify.pause()
		case <-spotify.events.WaitForShutdown():
			spotify.shutdownSpotify()
		}
	}
}

func (spotify *Spotify) pause() {
	if spotify.isPausedOrPlaying() {
		if spotify.paused {
			spotify.playCurrentTrack()
		} else {
			spotify.pauseCurrentTrack()
		}
	}
}

func (spotify *Spotify) playCurrentTrack() {
	spotify.play(spotify.currentTrack)
	spotify.paused = false
}

func (spotify *Spotify) pauseCurrentTrack() {
	player := spotify.session.Player()
	player.Pause()
	spotify.updateStatus("Paused", spotify.currentTrack)
	spotify.paused = true
}

func (spotify *Spotify) isPausedOrPlaying() bool {
	return spotify.currentTrack != nil
}

func (spotify *Spotify) play(track *sp.Track) {
	if !spotify.isTrackAvailable(track) {
		spotify.events.SetStatus("Not available")
		return
	}
	player := spotify.session.Player()
	if err := player.Load(track); err != nil {
		log.Fatal(err)
	}
	player.Play()

	spotify.updateStatus("Playing", track)
}

func (spotify *Spotify) isTrackAvailable(track *sp.Track) bool {
	return track.Availability() == sp.TrackAvailabilityAvailable
}

func (spotify *Spotify) updateStatus(status string, track *sp.Track) {
	spotify.currentTrack = track
	artist := track.Artist(0)
	artist.Wait()
	spotify.events.SetStatus(fmt.Sprintf("%v: %v - %v [%v]", status, artist.Name(), spotify.currentTrack.Name(), spotify.currentTrack.Duration().String()))
}

func (pa *portAudio) player() {
	out := make([]int16, 2048*2)

	stream, err := portaudio.OpenDefaultStream(
		0,
		2,     // audio.format.Channels,
		44100, // float64(audio.format.SampleRate),
		len(out),
		&out,
	)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	stream.Start()
	defer stream.Stop()

	for {
		// Decode the incoming data which is expected to be 2 channels and
		// delivered as int16 in []byte, hence we need to convert it.

		select {
		case audio := <-pa.buffer:
			if len(audio.frames) != 2048*2*2 {
				// panic("unexpected")
				// don't know if it's a panic or track just ended
				break
			}

			j := 0
			for i := 0; i < len(audio.frames); i += 2 {
				out[j] = int16(audio.frames[i]) | int16(audio.frames[i+1])<<8
				j++
			}

			stream.Write()
		}
	}
}

func (pa *portAudio) WriteAudio(format sp.AudioFormat, frames []byte) int {
	audio := &audio{format, frames}

	if len(frames) == 0 {
		return 0
	}

	select {
	case pa.buffer <- audio:
		return len(frames)
	default:
		return 0
	}
}
