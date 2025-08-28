package boot

import (
	"github.com/pkg/errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"web3_gui/chain/boot/hardware"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/light"
	"web3_gui/chain/mining"
	"web3_gui/chain/mining/snapshot"
	_ "web3_gui/chain/mining/tx_name_in"
	_ "web3_gui/chain/mining/tx_name_out"
	"web3_gui/chain/startblock"
	"web3_gui/libp2parea/adapter"
	"web3_gui/utils"

	_ "github.com/hyahm/golog"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/ntp"
)

// var prk *ecdsa.PrivateKey

// var coinbase *keystore.Address

var Area *libp2parea.Area

/*
通过命令行启动，要解析命令行传入的参数
*/
func Register(area *libp2parea.Area) error {
	// 非创世节点等待网络自治完成
	config.ParseInitFlag()
	Area = area
	return register(area)
}

/*
通过API启动，解析传入的参数
*/
func RegisterWithParamsJSON(area *libp2parea.Area, configJSONStr string) error {
	err := config.ParseConfigWithConfigJSON(configJSONStr, "")
	if err != nil {
		return err
	}
	Area = area
	return register(area)
}

func register(area *libp2parea.Area) error {
	engine.Log.Info("start")
	if !ntp.WaitSyncTimeFinish() {
		return errors.New("sync ntp time failed")
	}
	engine.Log.Info("sync ntp time success %s", config.TimeNow().Format("2006-01-02 15:04:05"))
	// 非创世节点等待网络自治完成
	if !config.ParseInitFlag() {
		area.WaitAutonomyFinish()
	}

	Area = area
	// golog.InitLogger("logs/randeHash.txt", 0, true)
	//golog.Infof("start %s", "log")

	engine.Log.Info("CPUNUM :%d", config.CPUNUM)
	// if config.CPUNUM < 8 {
	// 	config.CPUNUM = 8
	// }
	go func() {
		for {
			engine.Log.Info("NumGoroutine:%d", runtime.NumGoroutine())
			time.Sleep(time.Minute)
			// log.Error(http.ListenAndServe(":6060", nil))
		}
	}()

	err := utils.StartSystemTime() //StartOtherTime()
	if err != nil {
		return err
	}

	config.ParseInitFlag()

	// mining.ModuleEnable = true
	if config.InitNode {
		os.RemoveAll(filepath.Join(config.Wallet_path))
	}
	//检查目录是否存在，不存在则创建
	utils.CheckCreateDir(filepath.Join(config.Wallet_path))

	//检查是否满足最低硬件要求
	peerIps := []string{} //邻居超级节点IP集合
	for _, addr := range Area.GetNodeManager().GetSuperNodeIps() {
		peerIp, _, err := net.SplitHostPort(addr)
		if err != nil {
			continue
		}
		peerIps = append(peerIps, peerIp)
	}
	if err := hardware.CheckHardwareRequirement(config.MinCPUCores, config.MinFreeMemory, config.MinFreeDisk, config.MinNetworkBandwidth, peerIps); err != nil {
		engine.Log.Error("最低配置要求检测:%v", err)
		os.Exit(0)
	}

	//启动和链接leveldb数据库
	err = db.InitDB(config.DB_path, config.DB_path_temp)
	if err != nil {
		panic(err)
	}

	//初始化虚拟机快照
	//snapshot.InitContractSnap()
	//mining.InitAccountSnap()

	//删除数据库中的域名

	mining.SetArea(Area)

	// 订单薄引擎启动
	mining.OrderBookEngineSetup()

	//检查leveldb数据库中是否有创始区块
	bhvo := mining.LoadStartBlock()
	if bhvo == nil {
		config.DB_is_null = true
	}

	// _, err = db.Find(config.Key_block_start)
	// if err != nil {
	// 	//认为这是一个空数据库
	// 	// engine.Log.Info("这是一个空数据库")

	// 	config.DB_is_null = true
	// }

	// fmt.Println("数据库是否为空", config.DB_is_null)

	// go client.Start()
	// go server.Start()
	mining.RegisterMSG()
	//注册轻节点模块 area注册一次即可
	light.RegisterBaseMsg(area)
	light.RegisterWitnessMsg()
	light.RegisterDomainMsg()
	light.RegisterMinerMsg()
	light.RegisterContractMsg()
	light.RegisterEnsMsg()
	light.RegisterMulticastMsg()

	//启动交易缓存任务
	mining.InitTxCacheTask()

	//创始节点方式启动
	if config.InitNode {
		bhvo, err = startblock.BuildFirstBlock(area)
		if err != nil {
			return err
		}
		engine.Log.Info("create initiation block build chain")
		// engine.Log.Info("%+v", bhvo)
		config.StartBlockHash = bhvo.BH.Hash
		//构建创始区块成功
		mining.BuildFirstChain(bhvo)
		// 初始化快照
		mining.InitChainSnap()
		mining.SetHighestBlock(config.Mining_block_start_height)
		mining.GetLongChain().SyncBlockFinish = true
		// mining.GetLongChain().NoticeLoadBlockForDB()
		mining.GetLongChain().WitnessChain.BuildMiningTime()

		return nil
	}

	//拉起节点方式启动
	if config.LoadNode {
		// 快照启动方式一
		// 使用load命令，检测是否存在快照和是否禁用快照
		if snapshot.Height() > 0 && !config.DisableSnapshot {
			return mining.StartChainSnapOfLoad()
		}

		bhvo := mining.LoadStartBlock()
		if bhvo == nil {
			return errors.New("加载本地区块数据失败")
		}
		engine.Log.Info("load db initiation block build chain")
		config.StartBlockHash = bhvo.BH.Hash
		//从本地数据库创始区块构建链
		mining.BuildFirstChain(bhvo)

		// 初始化快照
		mining.InitChainSnap()
		// config.InitNode = true
		mining.SetHighestBlock(db.GetHighstBlock())

		mining.FindBlockHeight()

		if err := mining.GetLongChain().LoadBlockChain(); err != nil {
			return err
		}
		mining.FinishFirstLoadBlockChain()
		engine.Log.Info("build chain success")
		// if err := mining.GetLongChain().FirstDownloadBlock(); err != nil {
		// 	return err
		// }

		return nil
	}

	//快照方式拉起节点启动
	//if config.TestSnapshotNode {
	//	return mining.StartChainSnap()
	//}

	if config.Model == config.Model_light {
		light.StartModelLight()
		return nil
	}

	// 快照启动方式二
	// 普通启动方式，检测是否存在快照和是否禁用快照
	if snapshot.Height() > 0 && !config.DisableSnapshot {
		return mining.StartChainSnapOfNormal()
	}

	//普通启动方式
	bhvo = mining.LoadStartBlock()
	if bhvo == nil {
		engine.Log.Info("neighbor initiation block build chain")
		//从邻居节点同步区块
		err = mining.GetFirstBlock()
		if err != nil {
			engine.Log.Error("get first block error: %s", err.Error())
			panic(err.Error())
		}
		// engine.Log.Info("用邻居节点区块构建链2")
		mining.FindBlockHeight()
	} else {
		engine.Log.Info("load db initiation block build chain")
		config.StartBlockHash = bhvo.BH.Hash
		//从本地数据库创始区块构建链
		mining.BuildFirstChain(bhvo)
		mining.FindBlockHeight()
	}

	// 初始化快照
	mining.InitChainSnap()
	if err := mining.GetLongChain().FirstDownloadBlock(); err != nil {
		engine.Log.Info("build chain success:%s", err.Error())
		return err
	}

	engine.Log.Info("build chain success")

	mining.GetLongChain().NoticeLoadBlockForDB()
	return nil
}
