vtsplit
=======

To post-process vtuber 歌枠 downloaded with youtube-dl, split it into individual songs.
Dependency: ffmpeg

Example Usage:
Search for comment under the stream, for the starting time and name of individual songs.  Copy it
into config.txt.  Also add other metadata (see example config1.txt).

Then run
```
youtube-dl -x --audio-format mp3 https://www.youtube.com/watch?v=N1fx5LqxUMY&t=2116s
go build
./vtsplit config.txt
```
