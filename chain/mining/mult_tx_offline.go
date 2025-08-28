package mining

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"web3_gui/chain/config"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/base58"
	"web3_gui/keystore/adapter/crypto"
)

// 离线交易:创建多签地址
func CreateOfflineTxBuild_CreateMultsignAddrTx(keyStorePath, srcaddress string, gas, nonceInt, currentHeight uint64, pwd string, comment string, pukArrays []string) (string, string, error) {
	if len(pukArrays) < config.MultsignSet_addrs_puk_min || len(pukArrays) > config.MultsignSet_addrs_puk_max {
		return "", "", errors.New(fmt.Sprintf("limit number %d~%d", config.MultsignSet_addrs_puk_min, config.MultsignSet_addrs_puk_max))
	}

	if runeLength := len([]rune(comment)); runeLength > 1024 {
		return "", "", errors.New("comment over length max")
	}

	key := InitKeyStore(keyStorePath, pwd)
	src := crypto.AddressFromB58String(srcaddress)
	puk, ok := key.GetPukByAddr(src)
	if !ok {
		return "", "", config.ERROR_public_key_not_exist
	}

	vins := make([]*Vin, 0)

	nonce := big.NewInt(int64(nonceInt))
	nonce.Add(nonce, big.NewInt(1))
	vin := Vin{
		Nonce: *nonce,
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//生成多签地址
	puks := [][]byte{}
	for _, pukStr := range pukArrays {
		multPuk := base58.Decode(pukStr)
		if len(multPuk) != ed25519.PublicKeySize {
			return "", "", errors.New("public key size")
		}
		puks = append(puks, multPuk)
	}
	sortPuks, multAddress, randNum := GenerateMultAddressWithoutCheck(puks...)
	multVins := make([]*Vin, len(puks))
	for i, _ := range sortPuks {
		multVins[i] = &Vin{
			Puk: sortPuks[i],
		}
	}

	//构建交易输出
	vouts := make([]*Vout, 0)

	var pay *Tx_Multsign_Addr
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_multsign_addr, //交易类型
		Vin_total:  uint64(len(vins)),                   //输入交易数量
		Vin:        vins,                                //交易输入
		Vout_total: uint64(len(vouts)),                  //输出交易数量
		Vout:       vouts,                               //交易输出
		Gas:        gas,
		LockHeight: currentHeight + config.Wallet_tx_lockHeight, //锁定高度
		Payload:    []byte{},                                    //
		Comment:    []byte(comment),
	}
	pay = &Tx_Multsign_Addr{
		TxBase:      base,
		MultAddress: multAddress,
		RandNum:     randNum,
		MultVins:    multVins,
	}

	//给输出签名，防篡改
	for i, one := range pay.Vin {
		_, prk, err := key.GetKeyByPuk(one.Puk, pwd)
		if err != nil {
			return "", "", err
		}
		sign := pay.GetSign(&prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}

	pay.BuildHash()

	paybs, _ := pay.Proto()
	return base64.StdEncoding.EncodeToString(*paybs), fmt.Sprintf("%x", pay.Hash), nil
}

// 离线交易:多签支付
func CreateOfflineTxBuild_MultsignPay(keyStorePath, multAddressStr, addressStr string, amount, gas, nonceInt, frozenHeight, currentHeight uint64, pwd string, comment string, multPukStr string, pukArrays []string) (string, string, []string, error) {
	key := InitKeyStore(keyStorePath, pwd)
	multAddress := crypto.AddressFromB58String(multAddressStr)
	address := crypto.AddressFromB58String(addressStr)

	if err := CheckTxPayFreeGasWithParams(config.Wallet_tx_type_multsign_pay, address, amount, gas, comment); err != nil {
		return "", "", nil, errors.New("gas too little")
	}

	multpuk := base58.Decode(multPukStr)

	vins := make([]*Vin, 0)

	nonce := big.NewInt(int64(nonceInt))
	nonce.Add(nonce, big.NewInt(1))
	vin := Vin{
		Nonce: *nonce,
		Puk:   multpuk, //公钥
	}
	vins = append(vins, &vin)

	//多签vin
	multVins := make([]*Vin, len(pukArrays))
	for i, pukStr := range pukArrays {
		puk := base58.Decode(pukStr)
		multVins[i] = &Vin{
			Puk: puk,
		}
	}

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      address,      //钱包地址
		FrozenHeight: frozenHeight, //
	}
	vouts = append(vouts, &vout)

	var mtx *Tx_Multsign_Pay
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_multsign_pay,                   //交易类型
		Vin_total:  uint64(len(vins)),                                    //输入交易数量
		Vin:        vins,                                                 //交易输入
		Vout_total: uint64(len(vouts)),                                   //输出交易数量
		Vout:       vouts,                                                //交易输出
		Gas:        gas,                                                  //交易手续费
		LockHeight: currentHeight + config.Wallet_multsign_tx_lockHeight, //锁定高度
		Payload:    []byte{},                                             //
		Comment:    []byte(comment),
	}
	mtx = &Tx_Multsign_Pay{
		TxBase:        base,
		DefaultMultTx: NewDefaultMultTxImpl(multAddress, multVins),
	}

	//若自己是其中之一,就先签名
	signAddrs := []string{}
	for j, one := range mtx.GetMultVins() {
		if _, prk, err := key.GetKeyByPuk(one.Puk, pwd); err == nil {
			sign := mtx.GetSign(&prk, uint64(j))
			mtx.GetMultVins()[j].Sign = *sign
			signAddrs = append(signAddrs, one.GetPukToAddr().B58String())
		}
	}

	mtx.BuildHash()

	paybs, _ := mtx.Proto()
	return base64.StdEncoding.EncodeToString(*paybs), fmt.Sprintf("%x", mtx.Hash), signAddrs, nil
}

func CreateOfflineTxBuild_MultTxSetSign(keyStorePath, pwd, tx string) (string, string, []string, error) {
	txjsonBs, e := base64.StdEncoding.DecodeString(tx)
	if e != nil {
		return "", "", nil, fmt.Errorf("MultTxSetSign DecodeString fail:%s", e.Error())
	}

	txItr, e := ParseTxBaseProto(0, &txjsonBs)
	if e != nil {
		return "", "", nil, fmt.Errorf("MultTxSetSign ParseTxBaseProto fail:%s", e.Error())
	}

	if txItr.Class() == config.Wallet_tx_type_multsign_pay {
		mtx, ok := txItr.(*Tx_Multsign_Pay)
		if !ok {
			return "", "", nil, fmt.Errorf("MultTxSetSign  txItr to Tx_Multsign_Pay fail")
		}

		key := InitKeyStore(keyStorePath, pwd)
		//若自己是其中之一,就先签名
		signAddrs := []string{}
		for j, one := range mtx.GetMultVins() {
			if _, prk, err := key.GetKeyByPuk(one.Puk, pwd); err == nil {
				buf := mtx.GetWaitSign(uint64(j))
				sign := keystore.Sign(prk, *buf)
				mtx.GetMultVins()[j].Sign = sign
				signAddrs = append(signAddrs, one.GetPukToAddr().B58String())
			}
		}
		if len(signAddrs) == 0 {
			return "", "", nil, fmt.Errorf("MultTxSetSign not found puk")
		}

		paybs, _ := mtx.Proto()
		return base64.StdEncoding.EncodeToString(*paybs), fmt.Sprintf("%x", mtx.Hash), signAddrs, nil
	}
	return "", "", nil, fmt.Errorf("MultTxSetSign fail")
}

func ParseOfflineTxMultsignPay(tx string) (interface{}, error) {
	txjsonBs, e := base64.StdEncoding.DecodeString(tx)
	if e != nil {
		return nil, fmt.Errorf("DecodeString fail:%s", e.Error())
	}

	txItr, e := ParseTxBaseProto(0, &txjsonBs)
	if e != nil {
		return nil, fmt.Errorf("ParseTxBaseProto fail:%s", e.Error())
	}

	if txItr.Class() == config.Wallet_tx_type_multsign_pay {
		t, ok := txItr.(*Tx_Multsign_Pay)
		if !ok {
			return nil, fmt.Errorf("ParseOfflineTx txItr to Tx_Multsign_Pay fail")
		}

		return t.GetVOJSON(), nil
	}
	return nil, fmt.Errorf("ParseOfflineTx txType fail")
}
