@echo off
REM Windows build script (requires MinGW)
set LIB=c:\mingw\lib
set INCLUDE=c:\mingw\include
set PATH=%PATH%;c:\mingw\bin
go build
