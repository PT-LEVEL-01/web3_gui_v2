@echo off

cd  ..\..\..\libp2parea\
git log > git_log
set /p texte=< git_log
set ff=%texte%
for /f "tokens=2 delims= " %%i in ("%ff%") do (
    set commit=%%i
)
echo commit:%commit%
git branch --show-current > git_log
set /p blog=< git_log
set branch=%blog%
set p2pinfo=%branch%:%commit%
echo %p2pinfo%
del git_log
cd ..\icom_chain\example\peer_root


go build -ldflags "-X libp2parea.CommitInfo=%p2pinfo%" .