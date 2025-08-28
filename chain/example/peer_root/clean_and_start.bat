@echo off



rem 输出当前工作目录
echo remove log: %cd%

rmdir /s/q logs
rmdir /s/q messagecache
rmdir /s/q store
rmdir /s/q wallet
del /f/s/q *.prof
del /f/s/q conf\*.key
del /f/s/q conf\*.db
rem rmdir /s/q peer

rem pause