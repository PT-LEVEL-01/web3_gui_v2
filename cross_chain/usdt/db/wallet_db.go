package db

import (
	usdtconfig "web3_gui/cross_chain/usdt/config"
	"web3_gui/cross_chain/usdt/models"
	"web3_gui/utils"
)

/*
保存钱包地址
*/
func SaveWalletAddress(addrInfo *models.WalletAddressInfo) utils.ERROR {
	bs, err := addrInfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	_, ERR := usdtconfig.Leveldb.SaveList(*DBKEY_wallet_address, *bs, nil)
	return ERR
}

/*
保存钱包地址
*/
func LoadWalletAddress() ([]models.WalletAddressInfo, utils.ERROR) {
	items, err := usdtconfig.Leveldb.FindListAll(*DBKEY_wallet_address)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	addrInfos := make([]models.WalletAddressInfo, 0, len(items))
	for _, one := range items {
		addrInfo, err := models.ParseWalletAddressInfo(&one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		addrInfos = append(addrInfos, *addrInfo)
	}
	return addrInfos, utils.NewErrorSuccess()
}
