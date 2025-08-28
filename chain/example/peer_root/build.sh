#!/bin/bash
# 上面中的 #! 是一种约定标记, 它可以告诉系统这个脚本需要什么样的解释器来执行;
go build -x -v -ldflags "-s -w" -o chainim firstPeer.go
