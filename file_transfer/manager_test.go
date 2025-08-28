package file_transfer

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"web3_gui/keystore/v2"
	keystoreconfig "web3_gui/keystore/v2/config"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/utils"
)

const (
	CLASS_share_dir = "CLASS_share_dir"
)

func TestManager(*testing.T) {
	exampleFileTransfer()
}

func exampleFileTransfer() {
	fileTransferProto()
	//return
	area1 := StartOnePeer(0)
	area2 := StartOnePeer(1)

	area2.WaitAutonomyFinish()

	m1 := NewManager(area1)
	m2 := NewManager(area2)

	m1.CreateClass(1)
	tf := m1.GetClass(1)
	tf.AddShareDir("D:/迅雷下载")
	tf.SetAutoReceive(true)
	tf.SetReceiveFilePath("D:\\test\\temp/download")

	m2.CreateClass(1)
	tf2 := m2.GetClass(1)

	//测试下载文件
	cid, c := m2.GetListeningPullFinishSignal()
	pullID, ERR := tf2.DownloadFromShare(*area1.GetNetId(), "迅雷下载/7z2107-x64.exe", "D:/test/7z2107-x64.exe")
	if !ERR.CheckSuccess() {
		fmt.Println("下载错误:", ERR.String())
		return
	}
	finishID := <-c
	if bytes.Equal(finishID.PullID, pullID) {
		fmt.Println("下载完成")
	}
	m2.ReturnListeningPullFinishSignal(cid)

	//测试主动发送文件
	utils.Log.Info().Msgf("测试主动发送文件")
	cid, c = m1.GetListeningPullFinishSignal()
	tf2.SendFile(*area1.GetNetId(), "D:\\迅雷下载/amd-software-adrenalin-edition-24.1.1-combined-minimalsetup-240122_web.exe", nil)
	finishID = <-c
	if bytes.Equal(finishID.PullID, pullID) {
		utils.Log.Info().Msgf("传送文件完成")
	}
	m1.ReturnListeningPullFinishSignal(cid)

}

var (
	addrPre    = "SELF"
	areaName   = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd     = "123456789"
	serverHost = "124.221.170.43"
	host       = "127.0.0.1"
	basePort   = 19960
)

func StartOnePeer(i int) *libp2parea.Node {
	keyPath1 := filepath.Join("D:\\test", "keystore"+strconv.Itoa(i)+".key")

	wallet, ERR := InitWallet(keyPath1, addrPre, keyPwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("报错", ERR.String()).Send()
		return nil
	}
	keyst, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		utils.Log.Error().Str("报错", ERR.String()).Send()
		return nil
	}

	area, ERR := libp2parea.NewNode(areaName, addrPre, keyst, keyPwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("报错", ERR.String()).Send()
		return nil
	}
	area.SetLeveldbPath(config.Path_leveldb + strconv.Itoa(i))
	//area.SetNetTypeToTest()
	area.SetLeveldbPath(filepath.Join("D:\\test", "messagecache"+strconv.Itoa(i)))

	//area.OpenVnode()

	//serverHost
	area.SetDiscoverPeer([]string{"/ip4/" + host + "/tcp/" + strconv.Itoa(basePort) + "/ws"})
	area.StartUP(uint16(basePort + i))

	return area
}

func fileTransferProto() {
	filePath := "D:/迅雷下载/test/"
	filePath = filepath.Clean(filePath)

	filepath.Split(filePath)

	s1, s2 := filepath.Split(filePath)

	paths := filepath.SplitList(filePath)
	fmt.Println("拆分路径:", len(paths), paths, filePath, filepath.Dir(filePath), s1, s2)

	fileIndex := FileIndex{
		SupplierID: []byte{3}, //提供者ID
		PullID:     []byte{3}, //文件下载者ID
		Name:       "nihao",
	}
	utils.Log.Info().Msgf("参数:%+v", fileIndex)
	bs, err := fileIndex.Proto()
	if err != nil {
		utils.Log.Info().Msgf("错误:%s", err.Error())
		return
	}
	fi, err := ParseFileIndex(*bs)
	if err != nil {
		utils.Log.Info().Msgf("错误:%s", err.Error())
		return
	}
	utils.Log.Info().Msgf("解析的文件索引:%+v", fi)

}

func InitWallet(filePath, addrPre, pwd string) (*keystore.Wallet, utils.ERROR) {
	w := keystore.NewWallet(filePath, addrPre)
	//加载钱包文件
	ERR := w.Load()
	if ERR.CheckFail() {
		//文件不存在
		if ERR.Code == keystoreconfig.ERROR_code_wallet_incomplete {
			utils.Log.Error().Str("钱包文件损坏", filePath).Send()
			return nil, ERR
		} else if ERR.Code == keystoreconfig.ERROR_code_wallet_file_not_exist {
			utils.Log.Error().Str("钱包文件不存在", filePath).Send()
		} else {
			utils.Log.Error().Str("加载钱包文件报错", ERR.String()).Send()
			return nil, ERR
		}
	}
	list := w.List()
	if len(list) == 0 {
		ERR = w.AddKeystoreRand(pwd)
		if ERR.CheckFail() {
			utils.Log.Error().Str("添加一个随机数钱包报错", ERR.String()).Send()
			return nil, ERR
		}
	}
	keyst, ERR := w.GetKeystoreUse()
	if ERR.CheckFail() {
		utils.Log.Error().Str("使用密钥库错误", ERR.String()).Send()
		return nil, ERR
	}
	if len(keyst.GetCoinAddrAll()) == 0 {
		_, ERR = keyst.CreateCoinAddr("", pwd, pwd)
		if ERR.CheckFail() {
			utils.Log.Error().Str("创建收款地址错误", ERR.String()).Send()
			return nil, ERR
		}
	}
	if len(keyst.GetNetAddrAll()) == 0 {
		_, ERR = keyst.CreateNetAddr(pwd, pwd)
		if ERR.CheckFail() {
			utils.Log.Error().Str("创建收款地址错误", ERR.String()).Send()
			return nil, ERR
		}
	}
	if len(keyst.GetDHAddrInfoAll()) == 0 {
		_, ERR = keyst.CreateDHKey(pwd, pwd)
		if ERR.CheckFail() {
			utils.Log.Error().Str("创建收款地址错误", ERR.String()).Send()
			return nil, ERR
		}
	}
	//验证一下密码
	ok, err := keyst.CheckSeedPassword(pwd)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if !ok {
		return w, utils.NewErrorBus(keystoreconfig.ERROR_code_password_fail, "")
	}
	//addrList := keyst.GetCoinAddrAll()
	//for _, one := range addrList {
	//	utils.Log.Error().Str("地址列表", one.GetAddrStr()).Send()
	//}
	return w, utils.NewErrorSuccess()
}
