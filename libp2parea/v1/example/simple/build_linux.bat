@echo off

rem 输出当前工作目录
echo build: %cd%

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64

go build -ldflags "-w -s" -buildmode=exe

rem pause