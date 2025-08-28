@echo off



rem 输出当前工作目录
echo remove wallet: %cd%



rmdir /s/q peer_root\wallet
rmdir /s/q peer_super\wallet
rmdir /s/q peer_super2\wallet


rem pause