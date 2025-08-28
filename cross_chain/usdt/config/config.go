package config

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"web3_gui/keystore/v2"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

var (
	PATH_config_json = "config.json"
	PATH_Wallet      = "wallet.bin"
	PATH_db          = "leveldbData"

	WALLET_address_pre = "TEST"
	WALLET_password    = "123456789"

	HTTP_PORT = uint16(27330)

	Wallet_keystore *keystore.Keystore //
	Leveldb         *utilsleveldb.LevelDB

	SCAN_BLOCK_HEIGHT_trx = uint64(5000000) //波场统计起始高度
	Trc20PreBytes         []byte            //
	TrcAddressPreBytes    []byte

	NODE_rpc_addr_trx = []string{"8.210.166.55:50051", "118.69.78.91:50051"} //波场节点地址列表

	Block_confirmation_number_trx = 3 //交易确认区块数量->波场

	GAS_trx = int64(1e6) //手续费->波场

	Balance_sweep_interval_height_trx = int64(20)
)

// 交易类型
const (
	ContractTransfer = "ContractTransfer"
	Transfer         = "Transfer"
	ContractCall     = "ContractCall"
	CreateContract   = "CreateContract"
)

func init() {
	var err error
	Trc20PreBytes, err = hex.DecodeString("a9059cbb")
	if err != nil {
		panic(err)
	}
	TrcAddressPreBytes, err = hex.DecodeString("41")
	if err != nil {
		panic(err)
	}
}

/*
检查区块确认数量
@cNumber       int       区块确认数量
@scanHeight    uint64    统计高度
@pullHeight    uint64    节点同步高度
@return    bool    是否确认
*/
func CheckBlockConfirmationNumber(cNumber int, scanHeight, pullHeight uint64) bool {
	if pullHeight <= scanHeight+uint64(cNumber) {
		return false
	}
	return true
}

// // 正式合约地址
const (
	EthUsdtAddr   = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	EthUsdcAddr   = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	EthDaiAddr    = "0x6b175474e89094c44da98b954eedeac495271d0f"
	EthUniAddr    = "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984"
	EthShibAddr   = "0x95ad61b0a150d79219dcf64e1e6cc01f0b64c4ce"
	MaticDaiAddr  = "0x8f3cf7ad23cd3cadbd9735aff958023239c6a063"
	MaticUsdtAddr = "0xc2132d05d31c914a87c6611c10748aeb04b58e8f"
	//MaticUsdtAddr   = "0xc2132D05D31c914a87C6611C10748AEb04B58e8F"
	MaticUsdcAddr   = "0x2791bca1f2de4661ed88a30c99a7a9449aa84174"
	TrxUsdcAddr     = "TEkxiTehnzSmSe2XqrBj4w32RUN966rdz8"
	TrxTestUsdtAddr = "TXLAQ63Xg1NAzckPwKHvzw7CSEmLMEqcdj"
	TrxUsdtAddr     = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	EthChainID      = int64(1)
	EtcChainID      = int64(61)
	BnbChainID      = int64(56)
	AVAXChainID     = int64(43114)
	MaticChainID    = int64(137)
	TrxChainID      = int64(66)
)

type ConfigInfo struct {
	WalletPath string //钱包文件路径
	DbPath     string //数据库文件路径
	HttpPort   uint16 //
}

func LoadConfig(filePath string) (*ConfigInfo, utils.ERROR) {
	ok, err := utils.PathExists(filePath)
	if err != nil {
		// panic("检查配置文件错误：" + err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !ok {
		return nil, utils.NewErrorBus(ERROR_CODE_File_not_exist, filePath)
	}
	bs, err := os.ReadFile(filePath)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	config := new(ConfigInfo)
	err = json.Unmarshal(bs, config)
	return config, utils.NewErrorSysSelf(err)
}

func InitConfig(config *ConfigInfo) {
	PATH_Wallet = config.WalletPath
	PATH_db = config.DbPath
	HTTP_PORT = config.HttpPort
}

func SaveConfig() utils.ERROR {
	config := new(ConfigInfo)
	config.DbPath = PATH_db
	config.HttpPort = HTTP_PORT
	config.WalletPath = PATH_Wallet
	err := utils.SaveJsonFile(PATH_config_json, config)
	return utils.NewErrorSysSelf(err)
}
