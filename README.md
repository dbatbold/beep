beep
====

Simple Go program useful for alerting end of a long running command execution.
```
Compiling:
$ go build beep.go
$ cp beep /usr/bin/

Usage:
  -c=1: count
  -f=0.055: frequency
  -h: help

Examples:
$ cp -vr directory target; beep
$ ffmpeg -i video.mp4 -vn -acodec libmp3lame sound.mp3; beep -c 10 -f 0.046
```
