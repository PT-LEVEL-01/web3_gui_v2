@echo off

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64

go build -o peer_root -ldflags "-w -s" -buildmode=exe
