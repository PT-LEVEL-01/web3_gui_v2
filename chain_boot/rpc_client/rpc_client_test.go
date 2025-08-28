package rpc_client

import (
	"fmt"
	"testing"
	chainconfig "web3_gui/chain/config"
)

const (
	Addr       = "127.0.0.1:27331" //节点地址及端口，本地
	AddressMax = 50000             //收款地址总数
	RPCUser    = "test"            //rpc用户名
	RPCPwd     = "testp"           //rpc密码
	WalletPwd  = "123456789"       //
)

func TestGetInfo(t *testing.T) {
	info, ERR := GetInfo(Addr, RPCUser, RPCPwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	fmt.Printf("result:%+v\n", info)
}

func TestBlockProto64ByHeight(t *testing.T) {
	bhvo, ERR := BlockProto64ByHeight(Addr, RPCUser, RPCPwd, 1)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	fmt.Printf("result:%+v\n", bhvo)
	fmt.Printf("head:%+v\n", bhvo.BH)
	for _, one := range bhvo.Txs {
		fmt.Printf("tx:%+v\n", one)
	}
}

func TestAddressList(t *testing.T) {
	accList, ERR := AddressList(Addr, RPCUser, RPCPwd, "", 100)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	fmt.Printf("result:%+v\n", accList)
}

func TestAddressCreate(t *testing.T) {
	addr, ERR := AddressCreate(Addr, RPCUser, RPCPwd, "", WalletPwd, WalletPwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	fmt.Printf("result:%+v\n", addr)
}

func TestSendToAddress(t *testing.T) {
	accList, ERR := AddressList(Addr, RPCUser, RPCPwd, "", 100)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	txPay, ERR := SendToAddress(Addr, RPCUser, RPCPwd, "", accList[1].AddrCoin, 1e8, 1e8, WalletPwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	fmt.Printf("result:%+v\n", txPay)
}

func TestSendToAddressMore(t *testing.T) {
	accList, ERR := AddressList(Addr, RPCUser, RPCPwd, "", 100)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	payMore := make([]PayNumber, 0)
	payMore = append(payMore, PayNumber{
		Address: accList[1].AddrCoin,
		Amount:  1e8,
	})
	payMore = append(payMore, PayNumber{
		Address: accList[2].AddrCoin,
		Amount:  1e8,
	})

	txPay, ERR := SendToAddressMore(Addr, RPCUser, RPCPwd, "", payMore, 1e8, WalletPwd, "")
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	fmt.Printf("result:%+v\n", txPay)
}

func TestAddressBalanceMore(t *testing.T) {
	addrs := make([]string, 0)
	addrs = append(addrs, "HIVEBimzz5h5ru3ufCKoYL4NcvfrQahSqVviU5")
	addrs = append(addrs, "TESTFMKBXxUTKm5pZvTxQvJ2zxfG6cSQFxkib5")
	//addrs = append(addrs, "TESTfocgFBoMY66krrAyR6aP5zWKue1szfB7QroKBq5BDoPXywgww5")
	//addrs = append(addrs, "TESTkJWmf1WqnfwPjBxBncqEmdxB8VqDyrRfhVdJnCnNbucSc7T8i5")

	balance, balanceFrozen, ERR := AddressBalanceMore(Addr, RPCUser, RPCPwd, addrs)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	fmt.Printf("result:%+v %+v\n", balance, balanceFrozen)
	//for i, addrOne := range addrs {
	//	fmt.Printf("address:%s balance:%d balanceFrozen:%d\n", addrOne, balance[i], balanceFrozen[i])
	//}
}

func TestOfflineTxCreate(t *testing.T) {
	txHash, txStrB64, ERR := OfflineTxCreate(Addr, RPCUser, RPCPwd, "D:\\test\\test_local_cmd\\peer2\\conf/wallet.bin",
		"TESTFzZbqVFQDGMFnXoSNrsvEPCdrxtzRzFTQKu6zH3btSRF2fZkx5", "TEST2PUgNJ3Npzp3xSeEs92FXAYWHaTamBxAPe8VcPJWGr29YHqHQH5",
		100, 1, 10, 0, chainconfig.Wallet_tx_gas_min, WalletPwd, "")
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	fmt.Printf("txHash:%s\n", txHash)

	_, ERR = PushTxProto64(Addr, RPCUser, RPCPwd, txStrB64, false)
	if ERR.CheckFail() {
		panic(ERR.String())
	}

}
