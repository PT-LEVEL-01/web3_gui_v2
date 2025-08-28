@echo off



rem 输出当前工作目录
echo %cd%

rem call clean_log.bat
rem call clean_wallet.bat
call build_only.bat

robocopy "peer_root" "D:\test\test_local_cmd" peer.dll
robocopy "peer_root" "D:\test\test_local_cmd" peer_root.exe

rem pause