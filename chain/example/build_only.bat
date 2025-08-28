@echo off



rem 输出当前工作目录
echo build: %cd%

cd peer_root
echo %cd%
go build -ldflags "-w -s" -buildmode=exe



rem cd ../peer_super
rem echo %cd%
rem go build -o peer_root.exe -ldflags "-w -s" -buildmode=exe

rem cd ../peer_super2
rem echo %cd%
rem go build -o peer_root.exe -ldflags "-w -s" -buildmode=exe

rem cd ../peer_super3
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

rem cd peer_super
rem del peer.dll
rem cd ..
rem robocopy "peer_dll" "peer_super" peer.dll

rem cd peer_super2
rem del peer.dll
rem cd ..
rem robocopy "peer_dll" "peer_super2" peer.dll










rem pause