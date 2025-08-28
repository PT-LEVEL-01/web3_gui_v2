package keystore

import (
	"os"
	"path/filepath"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

/*
持久化接口
*/
type Store interface {
	Load() ([]byte, utils.ERROR) //读取
	Save(bs []byte) utils.ERROR  //保存
}

/*
保存到文件
*/
type StoreFile struct {
	filePath string //文件路径
}

func NewStoreFile(filePath string) *StoreFile {
	return &StoreFile{filePath}
}

/*
从文件中读取
*/
func (this *StoreFile) Load() ([]byte, utils.ERROR) {
	if err := utils.RenameTempFile(this.filePath); err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//判断文件是否存在
	exist, err := utils.PathExists(this.filePath)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if !exist {
		return nil, utils.NewErrorBus(config.ERROR_code_wallet_file_not_exist, "")
	}
	bs, err := os.ReadFile(this.filePath)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return bs, utils.NewErrorSysSelf(err)
}

/*
保存到文件中
*/
func (this *StoreFile) Save(bs []byte) utils.ERROR {
	//utils.Log.Info().Str("保存文件路径", this.filePath).Send()
	err := utils.CheckCreateDir(filepath.Dir(this.filePath))
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	err = utils.SaveFile(this.filePath, &bs)
	return utils.NewErrorSysSelf(err)
}
