#!/bin/sh

cd  ../../../libp2parea/
COMMIT_ID=`git log |head -n 1| awk '{print $2;}'`
BRANCH_NAME=`git branch --show-current '{ print $2; }'`
SERVICE_INFO="$BRANCH_NAME:$COMMIT_ID"
echo $SERVICE_INFO
cd ../icom_chain/example/peer_root


go build -ldflags "-X libp2parea.CommitInfo=$SERVICE_INFO" .