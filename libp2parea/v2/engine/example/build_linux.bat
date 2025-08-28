@echo off

rem 输出当前工作目录
echo build: %cd%

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64

echo %cd%
go build simple_tcp_server.go

