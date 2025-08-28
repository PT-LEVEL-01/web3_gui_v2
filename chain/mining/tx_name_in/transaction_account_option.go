package tx_name_in

import (
	"web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/nodeStore"
)

/*
注册域名，缴押金
*/
func NameIn(srcAddr, addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string,
	nametype int, name string, netIds []nodeStore.AddressNet, addrCoins []crypto.AddressCoin) (mining.TxItr, error) {

	//缴纳押金注册一个名称
	txItr, err := mining.GetLongChain().GetBalance().BuildOtherTx(config.Wallet_tx_type_account,
		srcAddr, addr, amount, gas, frozenHeight, pwd, comment, nametype, name, netIds, addrCoins)
	if err != nil {
		// fmt.Println("缴纳域名押金失败", err)
	} else {
		// fmt.Println("缴纳域名押金完成")
	}
	return txItr, err
}
