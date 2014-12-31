beep
====

Simple Go program useful for alerting end of a long running command execution.
```
Compiling:
$ apt-get install libasound2-dev
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

Usage examples:

$ cp -vr directory target; beep
$ ffmpeg -i video.mp4 -vn -acodec libmp3lame sound.mp3; beep -t 3 -f 0.076

# alarm for download completion
$ curl -O http://host.com/bigfile.tgz; beep -c 4

# beep for every text file found under home
$ find ~ -name '*.txt' | beep -l

# set an alarm for 1 hour from now
$ sh -c 'sleep 60; beep -t 3 -c 6' &
```
