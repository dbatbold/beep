beep
====

A simple Go program that is useful for alerting the end of a long running command execution.
It can also play 38-key piano music via pipe. To play a demo music, run: ```$ beep -p | beep -m```
```
Compiling:
$ apt-get install golang libasound2-dev  # for Debian and Ubuntu
$ go build beep.go
$ strip beep  # optional
$ cp beep /usr/bin/  # as root

Usage: beep [options]
  -c=1: count
  -d="default": audio device (hw:0,0)
  -f=0.088: frequency
  -h: help
  -l: beep per line via pipe input
  -m: play music notes via pipe input
  -p: print demo music by Mozart
  -t=1: time duration (1-100)
  -v=100: volume (0-100)

Piano Key Map:
  | | | | | | | | | | | | | | | | | | | | | | |
  |2|3| |5|6|7| |9|0| |=|a|s| |f|g| |j|k|l| |'|
 | | | | | | | | | | | | | | | | | | | | | | |
 |q|w|e|r|t|y|u|i|o|p|[|]|z|x|c|v|b|n|m|,|.|/|

 ' ' - whole rest
 ':' - half rest
 '!' - quarter rest
 'others' - whole rest

Demo Music Mozart K33b:
 c c cszsc z [!
 c c cszsc z [!
 v v vcscv s ] v!
 c c cszsc z [ c!
 s s sz]zs ] p!
 s sz][z][pp:i!y!rr

Usage Examples:

$ cp -vr directory target; beep
$ ffmpeg -i video.mp4 -vn -acodec libmp3lame sound.mp3; beep -t 3 -f 0.076

# alarm for download completion
$ curl -O http://host.com/bigfile.tgz; beep -c 4

# beep for every text file found under home
$ find ~ -name '*.txt' | beep -l

# set an alarm for 1 hour from now
$ sh -c 'sleep 3600; beep -t 3 -c 6' &

# play all music notes
# echo "q2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l./'" | beep -m

# play Mozart K33b
$ beep -p | beep -m
```
