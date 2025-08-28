@echo off

rd /S /Q .\node1\messagecache0\
rd /S /Q .\node2\messagecache0\
rd /S /Q .\node3\messagecache0\
rd /S /Q .\node4\messagecache0\
rd /S /Q .\node5\messagecache0\
rd /S /Q .\node6\messagecache0\

del /q .\node1\log.txt
del /q .\node2\log.txt
del /q .\node3\log.txt
del /q .\node4\log.txt
del /q .\node5\log.txt
del /q .\node6\log.txt

