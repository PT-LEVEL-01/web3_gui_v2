@echo off



rem 输出当前工作目录
echo remove log: %cd%

rmdir /s/q peer_root\logs
rmdir /s/q peer_super\logs
rmdir /s/q peer_super2\logs
rem rmdir /s/q peer_super3\logs

rem pause