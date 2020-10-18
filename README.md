beep
====

Beep is a sound library and utility for alerting end of a long running command execution.
Beep can also play a MIDI file/URL or text music score with natural voices. To play a demo music, run:<br>

```
$ GOPATH=$PWD/beep go get github.com/dbatbold/beep/cmd/beep
$ beep/bin/beep          # play beep sound
$ beep/bin/beep -vd      # download natural voice files
$ beep/bin/beep -m demo  # play demo music by Mozart
```

[Play demo with piano voice&nbsp; ▶](http://bmrust.com/dl/beep/demo-mozart-k33b-piano.mp3)

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
  -p: print demo music sheet (Mozart K33b)
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
 T#     - where # is 0-9, default is 4

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
Demo Music
==========
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
