* ### compile

1. The main function file of the project is located in 'chain\_boot\\boot\\hive\_v2\\example\\chain\_node.go'
2. Use the command in the 'example' directory：go build -o peer\_root.exe -ldflags "-w -s" -buildmode=exe
3. Generate the executable file 'peer\_root. exe' after successful compilation in the previous step



* ### Genesis Node Deployment

1. Create a new folder 'peer1'.Copy the 'peer\_root. exe' file to the 'peer1' folder.
2. Create a new folder named 'conf' in the 'peer1' directory to store configuration files.
3. Go to the 'conf' folder and create a text file named 'config. json'
4. Copy the contents of the 'config. json' file as follows:
   {"AreaName":"test","AddrPre":"HIVE","Port":25331,"RpcServer":true,"AdminPassword":"J.7(f+vB","RegistAddress":\["/ip4/127.0.0.1/tcp/25331/ws"]}
5. Genesis Node Start Command: ./peer\_root.exe init
6. The 'init' parameter indicates that the founding node will start and overwrite the previous data.
7. A 'wallet. bin' file will be generated in the 'conf' directory, which is the node key file that stores the node address identity.
8. The 'BlockChainData' folder will be generated in the 'peer1' folder,Storing blockchain data,This node serves as a witness node and is responsible for generating blocks.
9. The genesis node has been launched and completed



* ### Explanation of file parameters for 'config. json'

1. AreaName: Domain network name.When multiple nodes are connected to each other, the names should be set to be the same. If they are not the same, they cannot be connected to the network.
2. AddrPre: Blockchain payment address prefix.
3. Port: HTTP RPC and WebSocket ports
4. RpcServer: Do you want to open RPC request
5. AdminPassword: Administrator Password
6. RegistAddress: Network discovers ports of other nodes



* ### Synchronize node startup

1. Copy a folder of the founding node 'peer1' and rename it as' peer2 '.Delete the "BlockChainData" folder in "peer2" and the "wallet. bin" file in "conf".
2. Modify the "RegistAddress" parameter in the "config. json" file to "\["/ip4/127.0.0.1/tcp/25331/ws "]".The address is the genesis node address.Enable this node to connect to the genesis node.
3. The 'RegistAddress' parameter can add multiple IPs.Just connect one of the addresses.
4. Start synchronization node command: ./peer\_root.exe
5. Synchronize node startup without any parameters.The 'peer2' folder will generate the 'BlockChainData' folder and the 'conf/wallet. bin' file.
