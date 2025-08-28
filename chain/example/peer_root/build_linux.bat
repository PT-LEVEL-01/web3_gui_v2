@echo off



rem 输出当前工作目录
echo build: %cd%

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64

cd peer_root
echo %cd%
go build -ldflags "-w -s" -buildmode=exe



rem cd ../peer_super
rem echo %cd%
rem go build -o peer_root.exe -ldflags "-w -s" -buildmode=exe



cd ../peer_dll
echo %cd%
go build -ldflags "-w -s" -buildmode=c-shared -o peer.dll peer_dll.go
go build --buildmode=plugin peer_dll.go

cd ..

cd peer_root
del peer.dll
cd ..
robocopy "peer_dll" "peer_root" peer.dll

echo %cd%

cd ..\example\


rem pause