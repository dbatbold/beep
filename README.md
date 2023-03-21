beep
====

Beep is a sound utility initially developed for alerting the end of a long running command execution.
Now beep can also play a MIDI file/URL or text music score with natural voices. To play demo music, run:<br>

```
$ go install github.com/dbatbold/beep/cmd/beep@latest
$ ~/go/bin/beep           # play beep sound
$ ~/go/bin/beep -vd       # download natural voice files
$ ~/go/bin/beep -m demo1  # play demo music by Mozart
$ ~/go/bin/beep -m demo2  # play Passacaglia - Handel Halvorsen
```

[Play demo #1 with piano voice&nbsp; ▶](http://bmrust.com/dl/beep/demo-mozart-k33b-piano.mp3)<br>
[Play demo #2 with piano voice&nbsp; ▶](http://bmrust.com/dl/beep/passacaglia-handel-halvorsen-piano.mp3)<br>

Library Usage
=============

```go
package main

import (
    "bufio"
    "strings"
    "log"
    "github.com/dbatbold/beep"
)

func main() {
    music := beep.NewMusic("") // output can be a file "music.wav"
    volume := 100

    if err := beep.OpenSoundDevice("default"); err != nil {
        log.Fatal(err)
    }
    if err := beep.InitSoundDevice(); err != nil {
        log.Fatal(err)
    }
    beep.PrintSheet = true
    defer beep.CloseSoundDevice()

    musicScore := `
        VP SA8 SR9
        A9HRDE cc DScszs|DEc DQzDE[|cc DScszs|DEc DQz DE[|vv DSvcsc|DEvs ]v|cc DScszs|VN
        A3HLDE [n z,    |cHRq HLz, |[n z,    |cHRq HLz,  |sl z,    |]m   pb|z, ]m    |
        
        A9HRDE cz [c|ss DSsz]z|DEs] ps|DSsz][ z][p|DEpDQ[ [|VN
        A3HLDE [n ov|]m [n    |  pb ic|  n,   lHRq|HLnc DQ[|
    `

    reader := bufio.NewReader(strings.NewReader(musicScore))
    go music.Play(reader, volume)
    music.Wait()
    beep.FlushSoundBuffer()
}

```

Building from Source
====================
```
On Linux:
 $ apt-get install golang libasound2-dev  # for Debian and Ubuntu
 $ apk add alsa-lib-dev                   # for Alpine linux
 $ go build ./cmd/beep
 $ cp beep /usr/local/bin/  # as root

On Windows: Requires git (git-scm.com), MinGW and Go compiler (golang.org)
 Run Git Bash,
 $ export GOPATH=$PWD/beep
 $ go get -u -v -d github.com/dbatbold/beep/cmd/beep
 $ cd beep
 $ ./build.bat
 $ cp bin/beep.exe /c/Windows/System32
```
[![Build Status](http://travis-ci.org/dbatbold/beep.svg?branch=master)](http://travis-ci.org/dbatbold/beep)

Usage
=====

```
beep [options]
  -c=1: beep count
  -d="default": audio device, Linux example: hw:0,0
  -f=523.25: frequency in Hertz (1-22050)
  -h: print help
  -l: beep per line from stdin
  -m: play music from sheet file, reads stdin if no arguments given (see beep notation)
  -p=0: print demo music sheet by number (1-2)
  -t=250: beep time duration in millisecond (1-60000)
  -v=100: volume (1-100)
  -b: send bell to PC speaker
  -q: quiet stdout while playing music
  -n: print notes while playing music
  -o=file: output music waveform to a WAV file. Use '-' for stdout
  -w: start beep web server
  -a ip:port: web server address (default 127.0.0.1:4444)
  -vd [name ..]: download voice files, if no names given, downloads all voices
  -mp=file: play a MIDI file
  -mu=URL: play a MIDI file from URL
  -mn=file: parses MIDI file and print notes
  -play=notes: play notes from command argument
  -battery: monitor battery and alert low charge level
```
Beep notation
=============

To play music with beep, music score needs to be converted into text. Beep uses its own music notation called beep notation. All music octaves are divided into computer keyboard keys similar to piano key layout. All 88 piano notes can be played as following.
```
$ beep -play 'H0,l.HLq2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l.HRq2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l.H7q2w3er5t6y7ui'
```
[Play all piano notes &nbsp; ▶](http://bmrust.com/dl/beep/piano-notes.mp3)

Uppercase "H" letter is a control key that changes the current octave. Lowercase letters are notes. Using all control keys shown below are used to convert musical notation.
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
 Lines start with '#' are ignored.

Control keys:

 Rest:
 RW     - whole rest
 RH     - half rest
 RQ     - quarter rest
 RE     - eighth rest
 RS     - sixteenth rest
 RT     - thirty-second rest
 RI     - sixty-fourth rest

 Durations:
 DW     - whole note
 DH     - half note
 DQ     - quarter note
 DE     - eighth note
 DS     - sixteenth note
 DT     - thirty-second note
 DI     - sixty-fourth note
 DD     - dotted note (adds half duration)

 Octave:
 H0     - octave 0 keys
 HL     - octave 1, 2, 3 (left hand keys)
 HR     - octave 4, 5, 6 (right hand keys)
 H7     - octave 7, 8 keys

 Tempo:
 T#     - where # is 0-9, default is 4 (1 unit speeds up/down by 4%)

 Sustain:
 SA#    - attack level, where # is 0-9, default is 8
 SD#    - decay level, 0-9, default 4
 SS#    - sustain level, 0-9, default 4
 SR#    - release level, 0-9, default 9

 Voice:
 VD     - Computer generated default voice
 VP     - Piano voice
 VV     - Violin voice (WIP)
 VN     - If a line ends with 'VN', the next line will be played 
          harmony with the line.

 Chord:
 C#     - Play next # notes as a chord, where # is 2-9. 
          For example C major chord is "C3qet"

 Amplitude:
 A#     - Changes current amplitude, where # is 1-9, default is 9

 Measures:
 |      - bar, ignored
 ' '    - space, ignored
 Tab    - tab, ignored

 Comments:
 #      - a line comment
 ##     - start or end of a block comment
```
Demo music #1
=============
```
# Mozart K33b
VP SA8 SR9
A9HRDE cc DScszs|DEc DQzDE[|cc DScszs|DEc DQz DE[|vv DSvcsc|DEvs ]v|cc DScszs|VN
A3HLDE [n z,    |cHRq HLz, |[n z,    |cHRq HLz,  |sl z,    |]m   pb|z, ]m    |

A9HRDE cz [c|ss DSsz]z|DEs] ps|DSsz][ z][p|DEpDQ[ [|VN
A3HLDE [n ov|]m [n    |  pb ic|  n,   lHRq|HLnc DQ[|

A9HRDE cc DScszs|DEc DQzDE[|cc DScszs|DEc DQz DE[|vv DSvcsc|DEvs ]v|cc DScszs|VN
A3HLDE [n z,    |cHRq HLz, |[n z,    |cHRq HLz,  |sl z,    |]m   pb|z, ]m    |

A9HRDE cz [c|ss DSsz]z|DEs] ps|DSsz][ z][p|DEpDQ[ [|VN
A3HLDE [n ov|]m [n    |  pb ic|  n,   lHRq|HLnc DQ[|

A9HRDS DERE] DS][p[ |][p[ ][p[  |DE] DQp DEi|REc DScszs|cszs |cszs|DEcDQzDE[|REv DSvcsc|DEvs ]v|VN
A3HLDE DEcHRq HLvHRw|HLbHRe HLvw|cHRq   HLic|[n  ]m    |z,   |]m  |zn   z,  |sl  [,    |z. DQp |

A9HRDE REc DScszs|DEcz [c|REs DSsz]z|DEs] ps|DSsz][ z][p|DE[DSitDQr|VN
A3HLDE z,  ]m    |[n   ov|]m  [n    |pb   ic|nz     sc  |DQn      [|

A9HRDS DERE] DS][p[ |][p[ ][p[  |DE] DQp DEi|REc DScszs|cszs |cszs|DEcDQzDE[|REv DSvcsc|DEvs ]v|VN
A3HLDE DEcHRq HLvHRw|HLbHRe HLvw|cHRq   HLic|[n  ]m    |z,   |]m  |zn   z,  |sl  [,    |z. DQp |

A9HRDE REc DScszs|DEcz [c|REs DSsz]z|DEs] ps|DSsz][ z][p|DE[DSitDQrRQ|VN
A3HLDE z,  ]m    |[n   ov|]m  [n    |pb   ic|nz     sc  |DQn      [RQ|
```
[Play with default voice&nbsp; ▶](http://bmrust.com/dl/beep/demo-mozart-k33b.mp3)<br>
[Play with natural piano voice&nbsp; ▶](http://bmrust.com/dl/beep/demo-mozart-k33b-piano.mp3)<br>
[View music score](http://imslp.org/images/1/15/TN-Mozart%2C_Wofgang_Amadeus-NMA_09_27_Band_02_I_01_KV_33b.jpg)

Demo music #2
=============
```
# Passacaglia - Handel Halvorsen
VP T5 SA9 SD9 SS9 SR9
# DQ - 130
A6HRDE RERERERE RERERERE|RERERERE RERERERE|icxc   zc]c  |[cpc   ocic    |VN
A4HLDE z,HReq   yqeq    |HLz,HReq yqeq    |HLz,HReq yqeq|HLov,n HRwHLn,n|
# 5
A6HRDE uxzx     ]x[x        |pxox ixux    |yz]z      [zpz        |oziz uzyz    |DQz DEa= DQDDa DEz|VN
A4HLDE pmHRwHL. HRtHL.HRwHL.|icmb HRqHLbmb|[nmHRqHL, HRrHL,HRqHL,|ov,[ HRwHLn,n|pb  .k   HReHLk .k|
# 10
A6HRDW z          |DEcioi   pi[i|]izi   xici    |xuiu     oupu        |[u]u zuxu    |VN
A4HLDE z,HReq yqeq|HLz,HReq yqeq|HLov,n HRwHLn,n|]mHRwHL. HRtHL.HRwHL.|icmb HRqHLbmb|
# 15
A6HRDE zyuy     iyoy        |py[y ]yzy    |DQz DEa= DQDDa DEz|DWz        |DEcH7qHR.H7q  HR,H7qHRmH7q|VN
A4HLDE pbHRqHL, HRrHL,HRqHL,|ov,n HRwHLn,n|pb  .k   HReHLk .k|z,HReq yqeq|z,HReq yqeq               |
# 20   8
A6HRDE nH7qHRbH7q HRvH7qHRcH7q|HRx.,.   m.n.        |b.v. c.x.    |z,m,      n,b,        |v,c, x,z,    |VN
A4HLDE ov,n       HRwHLn,n    |pmHRwHL. HRtHL.HRwHL.|icmb HRqHLbmb|[nmHRqHL, HRrHL,HRqHL,|ov,[ HRwHLn,n|
# 25   8
A6HRDQ , DEkj DQDDk DE,|DW,        |DEH7qHRcvc bcnc|mc,c   .cH7qHRc|.xcx     vxbx        |VN
A4HLDE pb.k   HReHLk .k|z,HReq yqeq|HLz,HReq   yqeq|HLov,n HRwHLn,n|]mHRwHL. HRtHL.HRwHL.|
# 30   8
A6HRDE nxmx ,x.x    |.zxz     czvz        |bznz mz,z    |DQ, DEkj DQDDk DE,|DW,        |VN
A4HLDE icmb HRqHLbmb|pbHRqHL, HRrHL,HRqHL,|ov,n HRwHLn,n|pb  .k   HReHLk .k|z,HReq yqeq|
# 35
A6HRDE iici   xizi|]i[i   pioi    |uuxu     zu]u        |[upu ouiu    |yyzy     ]y[y        |VN
A4HLDE z,HReq yqeq|HLov,n HRwHLn,n|pbHRqHL, HRrHL,HRqHL,|icmb HRqHLbmb|pbHRqHL, HRrHL,HRqHL,|
# 40                                               # 8va---------------------------
A6HRDE pyoy iyuy    |DQy DE65 DQDD6 DEy|DWy        |DEcH7qHRx. z,]m|[npb   ovic    |VN
A4HLDE ov,n HRwHLn,n|pb.k     HReHLk .k|z,HReq yqeq|HLz,HReq   yqeq|HLov,n HRwHLn,n|
# 45   8
A6HRDE x.z,   ]m[n|pbov   icux    |yzt]     r[ep        |woqi uHL.HRyHL,|HRDQy DE65 DQDD6 DEy|VN
A4HLDE z,HReq yqeq|HLov,n HRwHLn,n|]mHRwHL. HRtHL.HRwHL.|ob,n ,n,n      |pb.k       HReHLk .k|
# 50
A6HRDW y          |DEicxc   zc]c|[cpc   ocic    |uxzx     ]x[x        |pxox ixux    |VN
A4HLDE z,HReq yqeq|HLz,HReq yqeq|HLov,n HRwHLn,n|pmHRwHL. HRtHL.HRwHL.|icmb HRqHLbmb|
# 55
A6HRDE yz]z      [zpz        |oziz uzyz    |DQz DEa= DQDDa DEz|DWz        |DEcioi   pi[i|VN
A4HLDE [nmHRqHL, HRrHL,HRqHL,|ov,[ HRwHLn,n|pb  .k   HReHLk .k|z,HReq yqeq|HLz,HReq yqeq|
# 60
A6HRDE ]izi xici    |xuiu     oupu        |[u]u zuxu    |zyuy     iyoy        |py[y ]yzy    |VN
A4HLDE ov,n HRwHLn,n|]mHRwHL. HRtHL.HRwHL.|icmb HRqHLbmb|pbHRqHL, HRrHL,HRqHL,|ov,n HRwHLn,n|
# 65
A6HRDQ z DEa= DQDDa DEz  |DWz        |DEicxc zc]c|[cpc ocic|uxzx  ]x[x|VN
A4HLDE pb  .k   HReHLk .k|z,HReq yqeq|HLDWC2,z   |C2vo     |C2m]      |
# 70
T3
A6HRDE pxox ixux|yz]z [zpz|oziz uzyz|DQz DEa= DQDDa DEz|DWz |VN
A4HLDW C2ci     |C2n[     |C2vo     |C2bp              |C2,z|
```
[Play with default voice&nbsp; ▶](http://bmrust.com/dl/beep/passacaglia-handel-halvorsen.mp3)<br>
[Play with natural piano voice&nbsp; ▶](http://bmrust.com/dl/beep/passacaglia-handel-halvorsen-piano.mp3)<br>
[View music score](https://azmusicfest.org/app/uploads/Passacaglia-Handel-Halvorsen-Pianistos-2.pdf)

Natural Voices
==============

Beep uses computer generated voices, if no voice files exist.
A voice file is a ZIP file that contains sound samples of all notes that the
music instrument can play. By downloading and placing voice files in specific locations can improve music sound quality.

To download voice files, run:<br>
```
$ beep -vd  # downloads all voice files
$ beep -vd piano # piano only
$ beep -vd piano violin # piano and voice files
```
Voice files can also be downloaded manually. Move the files to location below after
downloading:

**Voice files:**<br>
 Piano voice: [piano.zip](http://bmrust.com/dl/beep/voices/piano.zip) (13MB)<br>
 Violin voice: [violin.zip](http://bmrust.com/dl/beep/voices/piano.zip) (6.9MB)<br>

**Voice file location:**<br>
 Windows: ```C:\Users\{username}\_beep\voices\``` <br>
 Linux: ```/home/{username}/.beep/voices/```

Web Interface
=============

Playing a music sheet from command line can be slow to start, because voice
files are loaded at startup for every time running beep.
Beep has a built-in web server for playing and storing music sheets.
The web server loads voice files only once and uses them for all playback.
To start the web interface, run:<br><br>
Linux:<br>
```
$ beep -w
```
Windows:<br>
```
C:\>beep -w
```
then open your browser and navigate to ```http://localhost:4444```. If the web interface
needs to be accessible from other computers, run:
```
$ beep -w :4444
```

**Screenshot:**<br>
![alt tag](http://bmrust.com/dl/beep/beep-web.png?1)

Usage Examples
==============
```
 $ cp -vr directory target; beep
 $ ffmpeg -i video.mp4 -vn -acodec libmp3lame sound.mp3; beep -t 3000
 
 # use '&' symbol instead of ';' on Windows
 C:\>dir /s \windows\*.cpl & beep
 
 # alarm for download completion
 $ curl -O http://host.com/bigfile.tgz; beep -c 4 -f 1000
 
 # monitor battery and alarm low level
 $ beep -battery
 
 # beep for every text file found under home
 $ find ~ -name '*.txt' | beep -l
 
 # set an alarm for 1 hour from now
 $ sh -c 'sleep 3600; beep -t 300 -c 6' &
 
 # play all piano notes
 $ echo "DEH0,l.\
         HLq2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l.\
         HRq2w3er5t6y7ui9o0p[=]azsxcfvgbnjmk,l.\
         H7q2w3er5t6y7ui" | beep -m
 
 # play demo music by Mozart
 $ beep -m demo
 C:\>beep -m demo

 # start beep web server and serve pages locally
 $ beep -w
 C:\>beep -w  (or click Start then Run and type: beep -w)
 
 # start beep web server with open access
 $ beep -w :4444
 
 # dump music waveform to a WAV file
 $ beep -m -o music.wav demo 
 
 # pipe to MP3 encoder
 $ beep -m -o - demo | lame - music.mp3
 
 # play misic sheet from files
 $ beep -m sheet.txt
 $ beep -m sheet1.txt sheet2.txt demo
 C:\>beep -m sheet.txt

 # play music sheet from URL
 $ beep -url 'http://bmrust.com/dl/beep/k333-1.txt'
 $ beep -url 'http://bmrust.com/dl/beep/passacaglia-handel-halvorsen.txt'

 # generate 528Hz sine wave for 60 seconds (wine glass frequency)
 $ beep -f 528 -t 60000
 
 # middle C note
 $ beep -f 261.625565 -t 1500
 
 # play a MIDI file
 $ beep -mp music.mid
 
 # play a MIDI file from URL
 $ beep -mu 'https://www.guitarist.com/media/midi/classical/bach/jesu.mid'
 
 # print notes with keyboard from MIDI file
 $ beep -mn music.mid
```
