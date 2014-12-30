beep
====

Simple Go program useful for alerting end of a long running command execution.
```
Compiling:
$ go build beep.go
$ strip beep
$ sudo cp beep /usr/bin/

Usage: beep [options]
  -c=1: count
  -d="default": audio device (hw:0,0)
  -f=0.088: frequency
  -h: help
  -l: beep per line via pipe input
  -t=1: time duration (1-100)
  -v=100: volume (0-100)

Examples:
$ cp -vr directory target; beep
$ curl -O http://host.com/bigfile.tgz; beep -c 4
$ ffmpeg -i video.mp4 -vn -acodec libmp3lame sound.mp3; beep -t 3 -f 0.076
$ find ~ -name '*.txt' | beep -l
```
