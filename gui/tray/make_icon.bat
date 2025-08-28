@ECHO OFF

IF "%GOPATH%"=="" GOTO NOGO
IF NOT EXIST %GOPATH%\bin\2goarray.exe GOTO INSTALL
:POSTINSTALL
2goarray systemTrayPngData tray < system_tray.png > system_tray_data.go
2goarray systemTrayNewMsgPngData tray < system_tray_new_msg.png > system_tray_new_msg_data.go

GOTO DONE

:CREATEFAIL
ECHO Unable to create output file
GOTO DONE

:INSTALL
ECHO Installing 2goarray...
go get github.com/cratonica/2goarray
IF ERRORLEVEL 1 GOTO GETFAIL
GOTO POSTINSTALL

:GETFAIL
ECHO Failure running go get github.com/cratonica/2goarray.  Ensure that go and git are in PATH
GOTO DONE

:NOGO
ECHO GOPATH environment variable not set
GOTO DONE

:NOICO
ECHO Please specify a .ico file
GOTO DONE

:BADFILE
ECHO %1 is not a valid file
GOTO DONE

:DONE