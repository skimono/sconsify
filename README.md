A spotify console app
---------------------

[![Join the chat at https://gitter.im/fabiofalci/sconsify](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/fabiofalci/sconsify?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

A very early stage of a spotify console application.

Requirements: [Libspotify SDK](https://developer.spotify.com/technologies/libspotify/) & [PortAudio](http://www.portaudio.com/) & Spotify user subscribed to the Premium tier of the Spotify Service ([Libspotify SDK terms of use](https://developer.spotify.com/developer-terms-of-use/)).


Installation
------------

* Download current version [0.4.0](https://github.com/fabiofalci/sconsify/releases) 

* Install dependencies:

`Archlinux`

	$ pacman -S portaudio
	$ yaourt -S libspotify

`Ubuntu` & `debian`

	$ curl http://apt.mopidy.com/mopidy.gpg | sudo apt-key add - && sudo curl -o /etc/apt/sources.list.d/mopidy.list http://apt.mopidy.com/mopidy.list
	$ sudo apt-get update && sudo apt-get install -y libportaudio2 libspotify12 --no-install-recommends 

`OSX`

Install [brew, the missing package manager for OS X](http://brew.sh/) and

	$ brew tap homebrew/binary
	$ brew install portaudio
	$ brew install libspotify
	$ cd /usr/local/opt/libspotify/lib/
	$ ln -s libspotify.dylib libspotify

* Run `./sconsify`

![alt tag](https://raw.githubusercontent.com/wiki/fabiofalci/sconsify/sconsify.png)

Modes
-----

There are 2 modes: 

* `Console user interface` mode: it presents a text user interface with playlists and tracks.

* `No user interface` mode: it doesn't present user interface and just suffle tracks.


Parameters
----------

* `-username=""`: Spotify username. If not present username will be asked.

* Password will be asked. To not be asked you can set an environment variable with your password `export SCONSIFY_PASSWORD=password`. Be aware your password will be exposed as plain text.

* `-ui=true/false`: Run Sconsify with Console User Interface. If false then no User Interface will be presented and it'll only shuffle tracks.

* `-playlists=""`: Select just some playlists to play. Comma separated list.


No UI Parameters
----------------

* `-noui-repeat-on=true/false`: Play your playlist and repeat it after the last track.

* `-noui-silent=true/false`: Silent mode when no UI is used.

* `-noui-shuffle=true/false`: Shuffle tracks or follow playlist order.


UI mode keyboard 
----------------

* &larr; &darr; &uarr; &rarr; for navigation.

* `space` or `enter`: play selected track.

* `>`: play next track.

* `p`: pause.

* `/`: open a search field.

Search fields: `album, artist or track`. 

```
    album:help
    artist:the beatles
    track:let it be
```

Aliases `al` = `album`, `ar` = `artist`, `tr` = `track`:

```
    al:help
    ar:the beatles
    tr:let it be
```

* `s`: shuffle tracks from current playlist. Press again to go back to normal mode.

* `S`: shuffle tracks from all playlists. Press again to go back to normal mode.

* `u`: queue selected track to play next.

* `dd`: delete selected element (playlist, track) from the UI (it doesn't save the change to spotify playlist).

* `D`: delete all tracks from the queue if the focus is on the queue.

* `PageUp` `PageDown` `Home` `End`. 

* `Control C` or `q`: exit.

Vi navigation style:

* `h` `j` `k` `l` for navigation.

* `Nj` and `Nk` where N is a number: repeat the command N times.

* `gg`: go to first element. 

* `G`: go to last element.

* `Ngg` and `NG` where N is a number: go to element at position N. 

* Temporary playlist. Type `c` in the queue view, type a name and then a temporary playlist will appear containing all songs in the queue view.


No UI mode keyboard 
-------------------

* `>`: play next track.

* `Control C`: exit.

Interprocess commands
--------------------

Sconsify starts a server for interprocess commands using `sconsify -command <command>`. Available commands: `replay, play_pause, next`. 

[i3](http://i3wm.org/) bindings for multimedia keys:

```
    bindsym XF86AudioPrev exec sconsify -command replay
    bindsym XF86AudioPlay exec sconsify -command play_pause
    bindsym XF86AudioNext exec sconsify -command next
```

`macOS`: TODO show how to create using Automator/Service/Keyboard

sconsifyrc
----------

Similar to [.ackrc](http://beyondgrep.com/documentation/) you can define default parameters in `~/.sconsify/sconsifyrc`:

	-username=your-username
	-noui-silent=true 
	-noui-repeat-on=false


How to build
---------------------------------

Install go (same version from Dockerfile), [glide](https://glide.sh/) and get a Spotify application key and copy as a byte array to `/sconsify/spotify/spotify_key_array.key`.

	var key = []byte{
	    0x02, 0xA2, ...
	    ...
	    0xA1}

* osx only: `brew install pkgconfig`

* `make build`

When building for OSX you may face an issue where it doesn't get your application key. Just retry the build that eventually it will get the key.
