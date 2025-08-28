@echo off

go build -o peer_root.exe -ldflags "-w -s" -buildmode=exe
