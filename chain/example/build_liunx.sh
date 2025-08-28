#!/bin/bash
cleardata(){
  rm -rf  peer_root_*/imdb
  rm -rf  peer_root_*/logs
  rm -rf  peer_root_*/conf/*.key
  rm -rf  peer_root_*/*.db
  rm -rf  peer_root_*/pid
  rm -rf  peer_root_*/*.log

}

stop(){
   files=$(ls | grep "peer_root_")
   for file in $files;do
     if [ -f $file/pid ];then
        kill -9 `cat $file/pid`
     fi
   done
   echo "SUCCESS"
}

start(){
   files=$(ls | grep "peer_root_")
   i=0
   for file in $files;do
     let  i+=1
     cd $file
      if [[ $i -eq 1 ]];then
        nohup ./peer$i init > peer.log  2>&1 & echo $! > pid
      else
        nohup ./peer$i > peer.log  2>&1 & echo $! > pid
      fi
     cd ..
   done
  echo "SUCCESS"

}
build(){
  go build -o bin/peer -ldflags '-s -w' peer_root/firstPeer.go
  echo "SUCCESS"
}

sync(){
  read -p "sync node num:" nodeNum
  if [[ $nodeNum -gt 1 ]];then
    for((i=1;i<=$nodeNum;i++))
    do
        mkdir -p peer_root_$i
        mkdir -p peer_root_$i/conf
        cp -r  conf peer_root_$i
        cp -f bin/peer peer_root_$i/peer$i
         # 替换端口
        sed -i "s/<<i>>/$i/g"  peer_root_$i/conf/config.json
    done
    echo "SUCCESS"

  fi
}


menu(){
  select cmd in "build" "sync" "clear" "start" "stop" "restart" "quit"
do
  case $cmd in
    "build")
    build
    menu
    ;;
    "start")
    start
    menu
    ;;
    "sync")
    sync
    menu
    ;;
    "restart")
    stop
    start
    menu
    ;;
   "clear")
    stop
    cleardata
    menu
    ;;
    "stop")
    stop
    menu
    ;;
    "quit")
    echo "88"
    exit 0
    ;;
  esac

done
}
menu
