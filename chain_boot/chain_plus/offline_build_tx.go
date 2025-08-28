package chain_plus

import (
	"encoding/base64"
	"fmt"
	"math/big"
	chainconfig "web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2"
	"web3_gui/keystore/v2/coin_address"
	keystconfig "web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
创建离线交易，转账
*/
func OfflineTxCreateSendAddress(keyStorePath, srcaddress, address, pwd, comment string, amount, gas, frozenHeight uint64, nonce uint64,
	currentHeight uint64, domain string, domainType uint64) (tx, hash string, ERR utils.ERROR) {
	srcaddr := srcaddress
	src := crypto.AddressFromB58String(srcaddr)
	addr := address
	dst := crypto.AddressFromB58String(addr)
	key, ERR := LoadWalletCheckPwd(keyStorePath, src.GetPre(), pwd)
	if ERR.CheckFail() {
		return "", "", ERR
	}
	txpay, ERR := CreateTxPayByKey(key, &src, &dst, amount, gas, frozenHeight, pwd, comment, nonce, currentHeight, domain, domainType)
	if ERR.CheckFail() {
		return "", "", ERR
	}
	paybs, _ := txpay.Proto()
	hash = fmt.Sprintf("%x", txpay.Hash)
	tx = base64.StdEncoding.EncodeToString(*paybs)
	return
}

/*
*
初始化keystore 临时测试
*/
func LoadWalletCheckPwd(filePath, addrPre, pwd string) (*keystore.Keystore, utils.ERROR) {
	w := keystore.NewWallet(filePath, addrPre)
	//加载钱包文件
	ERR := w.Load()
	if ERR.CheckFail() {
		return nil, ERR
	}
	keyst, ERR := w.GetKeystoreUse()
	if ERR.CheckFail() {
		return nil, ERR
	}
	ok, err := keyst.CheckSeedPassword(pwd)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if !ok {
		return nil, utils.NewErrorBus(keystconfig.ERROR_code_seed_password_fail, "")
	}
	return keyst, ERR
}

/*
创建离线合约交易
*/
func OfflineTxCreateContract(keyStorePath, srcaddress, address, pwd, comment string, amount, gas, frozenHeight, gasPrice uint64,
	nonce uint64, currentHeight uint64, domain string, domainType uint64, abi, source string) (string, string, string, utils.ERROR) {
	tx, hash, addressContract, err := mining.CreateOfflineContractTx(keyStorePath, srcaddress, address, pwd, comment,
		amount, gas, frozenHeight, gasPrice, nonce, currentHeight, "", 0, abi, source)
	if err != nil {
		return "", "", "", utils.NewErrorSysSelf(err)
	}
	return tx, hash, addressContract, utils.NewErrorSuccess()
}

/*
创建离线交易，社区质押
*/
func OfflineTxCommunityDepositIn(walletPath string, witnessAddress, address coin_address.AddressCoin, rate uint16, amount, gas uint64,
	nonce uint64, currentHeight uint64, domain string, domainType uint64, pwd, comment string) (*mining.Tx_vote_in, utils.ERROR) {

	key, ERR := LoadWalletCheckPwd(walletPath, address.GetPre(), pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	txpay, ERR := CreateTxVoteinByKey(key, address, witnessAddress, mining.VOTE_TYPE_community, rate, amount, gas, pwd, comment,
		nonce, currentHeight, domain, domainType)
	return txpay, ERR
}

/*
创建离线交易，社区质押
*/
func OfflineTxCommunityDepositOut(walletPath string, address coin_address.AddressCoin, amount, gas uint64,
	nonce uint64, currentHeight uint64, domain string, domainType uint64, pwd, comment string) (*mining.Tx_vote_out, utils.ERROR) {

	key, ERR := LoadWalletCheckPwd(walletPath, address.GetPre(), pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	out, ERR := CreateTxVoteOutByKey(key, address, mining.VOTE_TYPE_community, amount, gas, pwd, comment,
		nonce, currentHeight, domain, domainType)
	return out, ERR
}

/*
创建离线交易，轻节点质押
*/
func OfflineTxLightDepositIn(walletPath string, address coin_address.AddressCoin, rate uint16, amount, gas uint64,
	nonce uint64, currentHeight uint64, domain string, domainType uint64, pwd, comment string) (*mining.Tx_vote_in, utils.ERROR) {

	key, ERR := LoadWalletCheckPwd(walletPath, address.GetPre(), pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	txpay, ERR := CreateTxVoteinByKey(key, address, nil, mining.VOTE_TYPE_light, rate, amount, gas, pwd, comment,
		nonce, currentHeight, domain, domainType)
	return txpay, ERR
}

/*
创建离线交易，轻节点取消质押
*/
func OfflineTxLightDepositOut(walletPath string, address coin_address.AddressCoin, amount, gas uint64,
	nonce uint64, currentHeight uint64, domain string, domainType uint64, pwd, comment string) (*mining.Tx_vote_out, utils.ERROR) {

	key, ERR := LoadWalletCheckPwd(walletPath, address.GetPre(), pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	out, ERR := CreateTxVoteOutByKey(key, address, mining.VOTE_TYPE_light, amount, gas, pwd, comment,
		nonce, currentHeight, domain, domainType)
	return out, ERR
}

/*
创建离线交易，轻节点投票
*/
func OfflineTxLightVoteIn(walletPath string, communityAddress, address coin_address.AddressCoin, rate uint16, amount, gas uint64,
	nonce uint64, currentHeight uint64, domain string, domainType uint64, pwd, comment string) (*mining.Tx_vote_in, utils.ERROR) {

	key, ERR := LoadWalletCheckPwd(walletPath, address.GetPre(), pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	txpay, ERR := CreateTxVoteinByKey(key, address, communityAddress, mining.VOTE_TYPE_vote, rate, amount, gas, pwd, comment,
		nonce, currentHeight, domain, domainType)
	return txpay, ERR
}

/*
创建离线交易，轻节点取消投票
*/
func OfflineTxLightVoteOut(walletPath string, address coin_address.AddressCoin, amount, gas uint64,
	nonce uint64, currentHeight uint64, domain string, domainType uint64, pwd, comment string) (*mining.Tx_vote_out, utils.ERROR) {

	key, ERR := LoadWalletCheckPwd(walletPath, address.GetPre(), pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	out, ERR := CreateTxVoteOutByKey(key, address, mining.VOTE_TYPE_vote, amount, gas, pwd, comment,
		nonce, currentHeight, domain, domainType)
	return out, ERR
}

/*
创建一个votein交易
*/
func CreateTxVoteinByKey(keyst *keystore.Keystore, vote, voteTo coin_address.AddressCoin, voteType, rate uint16,
	amount, gas uint64, pwd, comment string, nonceInt uint64, currentHeight uint64, domain string,
	domainType uint64) (*mining.Tx_vote_in, utils.ERROR) {
	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}
	addrInfo := keyst.FindCoinAddr(&vote)
	puk, _, ERR := addrInfo.DecryptPrkPuk(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	//查找余额
	nonce := big.NewInt(int64(nonceInt))
	vins := make([]*mining.Vin, 0)
	vin := mining.Vin{
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	vout := mining.Vout{
		Value: amount, //输出金额 = 实际金额 * 100000000
		//Address:      *address,     //钱包地址
		Address:      crypto.AddressCoin(vote), //钱包地址
		FrozenHeight: 0,                        //
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)

	base := NewTxBase(chainconfig.Wallet_tx_type_vote_in, &vins, &vouts, gas, currentHeight, 0)
	base.Comment = commentbs

	voteIn := &mining.Tx_vote_in{
		TxBase: *base,
		//Vote:     crypto.AddressCoin(voteTo),
		VoteType: voteType,
		Rate:     rate,
	}
	if voteTo != nil && len(voteTo) > 0 {
		voteIn.Vote = crypto.AddressCoin(voteTo)
	}
	//给输出签名，防篡改
	ERR = SignTxVin(voteIn, keyst, pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}
	voteIn.BuildHash()
	return voteIn, utils.NewErrorSuccess()
}

/*
创建一个voteOut交易
*/
func CreateTxVoteOutByKey(keyst *keystore.Keystore, address coin_address.AddressCoin, voteType uint16, amount, gas uint64,
	pwd, comment string, nonceInt uint64, currentHeight uint64, domain string, domainType uint64) (*mining.Tx_vote_out, utils.ERROR) {

	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}
	addrInfo := keyst.FindCoinAddr(&address)
	puk, _, ERR := addrInfo.DecryptPrkPuk(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}

	//查找余额
	nonce := big.NewInt(int64(nonceInt))
	vins := make([]*mining.Vin, 0)
	vin := mining.Vin{
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	vout := mining.Vout{
		Value:        amount,                      //输出金额 = 实际金额 * 100000000
		Address:      crypto.AddressCoin(address), //钱包地址
		FrozenHeight: 0,                           //
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)

	base := NewTxBase(chainconfig.Wallet_tx_type_vote_out, &vins, &vouts, gas, currentHeight, 0)
	base.Comment = commentbs

	voteOut := &mining.Tx_vote_out{
		TxBase:   *base,
		Vote:     crypto.AddressCoin(address), //见证人地址
		VoteType: voteType,
	}

	//给输出签名，防篡改
	ERR = SignTxVin(voteOut, keyst, pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}
	voteOut.BuildHash()
	return voteOut, utils.NewErrorSuccess()
}
