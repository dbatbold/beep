beep
====

A Go program that is useful for alerting the end of a long running command execution.
Beep can also play piano music sheet. To play a demo music, run: ```$ beep -p | beep -m```

Listen to a demo: [demo-mozart-k33b.mp3](http://angiud.com/beep/demo-mozart-k33b.mp3)
Compiling
=========
```
On Linux:
 $ apt-get install golang libasound2-dev  # for Debian and Ubuntu
 $ git clone http://github.com/dbatbold/beep
 $ cd beep
 $ go build
 $ strip beep  # optional
 $ cp beep /usr/bin/  # as root

On Windows: (requires MinGW, Go compiler from golang.org)
 C:\> cd beep
 C:\beep> build.bat
 C:\beep> copy beep.exe \windows\system32
```
Prebuilt binaries
===============
 Windows: [beep.exe](http://angiud.com/beep/binary/windows/beep.exe) &nbsp; ```MD5: a2ca00a989de1e4046e990d2392bb170```<br>
 Linux 64-bit: [beep](http://angiud.com/beep/binary/linux/beep) &nbsp; ```MD5: 96ded4a378eef7c9f8c23dc350c13cdc```
Usage
=====
```
beep [options]
  -c=1: beep count
  -d="default": audio device, Linux example: hw:0,0
  -f=523.25: frequency in Hertz (1-22050)
  -h: print help
  -l: beep per line from stdin
  -m: play music notes from stdin (see beep notation)
  -p: print the demo music by Mozart
  -t=1: beep time duration in millisecond (1-600000)
  -v=100: volume (1-100)
  -b: send bell to PC speaker
  -q: quiet stdout while playing music
  -n: print notes while playing music
  -o=file: output music waveform to a WAV file. Use '-' for stdout
```
Beep notation
=============
```
Piano key map:
  | | | | | | | | | | | | | | | | | | | | | | 
  |2|3| |5|6|7| |9|0| |=|a|s| |f|g| |j|k|l| |
 | | | | | | | | | | | | | | | | | | | | | | 
 |q|w|e|r|t|y|u|i|o|p|[|]|z|x|c|v|b|n|m|,|.|

 q - middle C (261.6 hertz)

 Left and right hand keys are same. Uppercase 
 letters are control keys. Lowercase letters
 are music notes. Space bar is current duration
 rest. Spaces after first space are ignored.

Control keys:

 Rest:
 RW     - whole rest
 RH     - half rest
 RQ     - quarter rest
 RE     - eighth rest
 RS     - sixteenth rest
 RT     - thirty-second rest

 Space  - half of current duration rest

 Durations:
 DW     - whole note
 DH     - half note
 DQ     - quarter note
 DE     - eighth note
 DS     - sixteenth note
 DT     - thirty-second note

 Octave:
 HL     - switch to left hand keys
 HR     - switch to right hand keys
 HF     - switch to far right keys (last octave)

 Tempo:
 T#     - where # is 0-9, default is 4

 Clef:
 CB     - G and F clef partition (Base). If line ends
          with 'CB', the next line will be played as base.

 Measures:
 |      - bar (ignored)
```
Demo Music
==========
```
# Mozart K33b
HRDEc c DSc s z s |DEc DQz DE[ CB
HLDE[ n   z   ,      c HRq HLz ,

HRDEc c DSc s z s |DEc DQz DE[ CB
HLDE[ n   z   ,      c HRq HLz ,

HRDEv v DSv c s c |DEv s ] v CB
HLDEs l   z   ,      ] m p b

HRDEc c DSc s z s |DEc z [ c CB
HLDEz ,   ]   m      [ n o v 

HRDEs s DSs z ] z |DEs ] p s CB
HLDE] m   [   n      p b i c 

HRDSs z ] [ z ] [ p |DE[ DSi y DQr CB
HLDEn   z   s   c      n   c   DQ[ 
```
Usage Examples
==============
```
 $ cp -vr directory target; beep
 $ ffmpeg -i video.mp4 -vn -acodec libmp3lame sound.mp3; beep -t 3000
 
 # use '&' symbol instead of ';' on Windows
 C:\>dir /s \windows\*.cpl & beep
 
 # alarm for download completion
 $ curl -O http://host.com/bigfile.tgz; beep -c 4 -f 1000
 
 # beep for every text file found under home
 $ find ~ -name '*.txt' | beep -l
 
 # set an alarm for 1 hour from now
 $ sh -c 'sleep 3600; beep -t 300 -c 6' &
 
 # play all piano notes
 $ echo "DEHLq2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l.\
         HRq2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l.\
         HFq2w3er5t6y7u" | beep -m
 
 # play Mozart K33b
 $ beep -p | beep -m
 C:\>beep -p | beep -m
 
 # dump music waveform to a WAV file
 $ beep -p | beep -m -o music.wav
 
 # pipe to MP3 encoder
 $ beep -p | beep -m -o - | lame - music.mp3
 
 # play misic sheet from a file
 $ beep -m < sheet.txt
 C:\>beep -m < sheet.txt

 # generate 500Hz sine wave for 10 seconds
 $ beep -f 500 -t 10000
```
