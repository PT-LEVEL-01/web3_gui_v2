package storage

import (
	"github.com/oklog/ulid/v2"
	"os"
	"path/filepath"
	"sync"
	"time"
	"web3_gui/chain_boot/object_beans"
	"web3_gui/chain_orders"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/utils"
)

var FileHashM = NewFileHashProcessManager()

type FileHashProcessManager struct {
	lock        *sync.RWMutex
	m           map[string]*FileHashProcess
	processChan chan []byte
}

func NewFileHashProcessManager() *FileHashProcessManager {
	fmp := FileHashProcessManager{
		lock:        new(sync.RWMutex),
		m:           make(map[string]*FileHashProcess),
		processChan: make(chan []byte, 100),
	}
	go fmp.loopCleanProcess()
	go fmp.runProcess()
	return &fmp
}

/*
添加一个任务
*/
func (this *FileHashProcessManager) AddProcess(filePaths []string) (*FileHashProcess, utils.ERROR) {
	filePrices, ERR := GetFileInfo(filePaths)
	if ERR.CheckFail() {
		return nil, ERR
	}
	fhp := FileHashProcess{
		Id: ulid.Make().Bytes(), //
		//Count      uint64            //任务总量
		FilePrices: filePrices, //
		FinishTime: 0,          //完成时间，完成后，一定时间内不去取出则删除
		Cancel:     false,
		lock:       new(sync.RWMutex),
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	this.m[utils.Bytes2string(fhp.Id)] = &fhp
	select {
	case this.processChan <- fhp.Id:
	default:
		return nil, utils.NewErrorBus(config.ERROR_CODE_sharebox_process_full, "")
	}
	return &fhp, utils.NewErrorSuccess()
}

/*
查询一个任务
*/
func (this *FileHashProcessManager) FindProcess(id []byte) *FileHashProcess {
	this.lock.RLock()
	defer this.lock.RUnlock()
	fhp, ok := this.m[utils.Bytes2string(id)]
	if !ok {
		return nil
	}
	return fhp
}

/*
取消一个任务
*/
func (this *FileHashProcessManager) CancelProcess(id []byte) {
	this.lock.Lock()
	fhp, ok := this.m[utils.Bytes2string(id)]
	if ok {
		fhp.SetCancel(true)
		delete(this.m, utils.Bytes2string(id))
	}
	this.lock.Unlock()
}

/*
定时删除过期任务
*/
func (this *FileHashProcessManager) loopCleanProcess() {
	const timeout = time.Minute
	for range time.NewTicker(time.Minute).C {
		now := time.Now()
		this.lock.Lock()
		for k, one := range this.m {
			if one.GetCancel() {
				delete(this.m, k)
			}
			finishTime := one.GetFinishTime()
			if finishTime == 0 {
				continue
			}
			if time.Unix(finishTime, 0).Add(timeout).Before(now) {
				delete(this.m, k)
			}
		}
		this.lock.Unlock()
	}
}

/*
执行计算hash任务
*/
func (this *FileHashProcessManager) runProcess() {
	for oneProcess := range this.processChan {
		fhp := this.FindProcess(oneProcess)
		if fhp == nil {
			continue
		}
		for i, _ := range fhp.FilePrices {
			if fhp.GetCancel() {
				break
			}
			one := fhp.FilePrices[i]
			//
			hashBs, err := utils.FileSHA3_256(filepath.Join(one.DirPath, one.Name))
			if err != nil {
				utils.Log.Error().Str("err", err.Error()).Send()
				continue
			}
			utils.Log.Info().Hex("计算hash", hashBs).Send()
			one.Hash = hashBs

			utils.Log.Info().Hex("打印hash", fhp.FilePrices[i].Hash).Send()
			//查询定价
			filePrice, ERR := FindPrice(hashBs)
			if ERR.CheckFail() {
				utils.Log.Error().Str("ERR", ERR.String()).Send()
				continue
			}
			if filePrice == nil {
				continue
			}
			if filePrice.Name != "" {
				one.Name = filePrice.Name
			}
			one.Price = filePrice.Price

			//messageOne := &model.MessageContent{
			//	//Type:       config.MSG_type_text, //消息类型
			//	FromIsSelf: true,                          //是否自己发出的
			//	Time:       time.Now().Unix(),             //时间
			//	State:      config.MSG_GUI_state_not_send, //
			//	Index:      fhp.Id,                        //
			//
			//}
			//msgVO := messageOne.ConverVO()
			//msgVO.Subscription = config.SUBSCRIPTION_sharebox_fileHash
			//msgVO.State = messageOne.State
			//subscription.AddSubscriptionMsg(msgVO)
		}
		fhp.SetFinishTime()
	}
}

type FileHashProcess struct {
	Id         []byte             //
	Count      uint64             //任务总量
	FilePrices []*model.FilePrice //
	FinishTime int64              //完成时间，完成后，一定时间内不去取出则删除
	Cancel     bool               //是否取消任务
	lock       *sync.RWMutex      //
}

func (this *FileHashProcess) SetFinishTime() {
	this.lock.Lock()
	this.FinishTime = time.Now().Unix()
	this.lock.Unlock()
}

func (this *FileHashProcess) GetFinishTime() int64 {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.FinishTime
}

func (this *FileHashProcess) SetCancel(cancel bool) {
	this.lock.Lock()
	this.Cancel = cancel
	this.lock.Unlock()
}

func (this *FileHashProcess) GetCancel() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.Cancel
}

/*
获取多个文件、文件夹中的所有文件的详细信息
*/
func GetFileInfo(filePaths []string) ([]*model.FilePrice, utils.ERROR) {
	dirInfos := make([]*model.FilePrice, 0)
	fileInfos := make([]*model.FilePrice, 0)
	for _, dirPathOne := range filePaths {
		//utils.Log.Info().Str("文件夹路径", dirPathOne).Send()
		fi, err := os.Stat(dirPathOne)
		if err != nil {
			utils.Log.Error().Err(err).Send()
			return nil, utils.NewErrorSysSelf(err)
		}
		if fi.IsDir() {
			dirEntry, err := os.ReadDir(dirPathOne)
			if err != nil {
				utils.Log.Error().Err(err).Send()
				return nil, utils.NewErrorSysSelf(err)
			}
			for _, one := range dirEntry {
				dirOne := model.FilePrice{
					Name:    one.Name(),
					Hash:    nil,
					Price:   0,
					Size:    0,
					Time:    0,
					DirPath: dirPathOne,
					IsDir:   one.IsDir(),
				}
				if one.IsDir() {
					dirInfos = append(dirInfos, &dirOne)
				} else {
					fileInfos = append(fileInfos, &dirOne)
				}
			}
		} else {
			fi := model.FilePrice{
				Name: fi.Name(), //真实文件名
				//Hash:    ,              //hash值
				Price:   0,                        //价格
				Size:    fi.Size(),                //文件大小
				Time:    fi.ModTime().Unix(),      //创建时间
				DirPath: filepath.Dir(dirPathOne), //
				IsDir:   false,                    //
			}
			fileInfos = append(fileInfos, &fi)
		}
	}
	//utils.Log.Info().Str("开始计算文件hash", "end").Send()
	return append(dirInfos, fileInfos...), utils.NewErrorSuccess()
}

/*
循环获取目录及子目录中的文件路径
@dirPath    string    文件夹路径
@return     []string    文件夹名称
@return     []string    文件路径
@return     error
*/
func GetFilePathForDir(dirPath string) ([]string, []string, error) {
	dirNames := make([]string, 0)
	filePaths := make([]string, 0)
	dirEntry, err := os.ReadDir(dirPath)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return nil, nil, err
	}
	//utils.Log.Info().Int("文件数量", len(dirEntry)).Send()
	for _, one := range dirEntry {
		if one.IsDir() {
			dirNames = append(dirNames, one.Name())
		} else {
			filePaths = append(filePaths, filepath.Join(dirPath, one.Name()))
		}
	}
	return dirNames, filePaths, nil
}

/*
查询一个文件的价格
*/
func FindPrice(fileHash []byte) (*object_beans.OrderShareboxGoods, utils.ERROR) {
	orderGoods, ERR := db.Sharebox_server_FindGoods(fileHash)
	if ERR.CheckFail() {
		return nil, ERR
	}
	return orderGoods, ERR
}

/*
查询多个文件的价格
*/
func FindFilePriceMore(fileHash [][]byte) ([]*model.FilePrice, utils.ERROR) {
	//filePrice, ERR := db.Sharebox_FindFilePriceMore(fileHash)
	//if ERR.CheckFail() {
	//	return nil, ERR
	//}
	//return filePrice, ERR
	return nil, utils.NewErrorSuccess()
}

/*
保存价格
*/
func SavePrice(fileHash []byte, price uint64) utils.ERROR {
	goods := object_beans.OrderShareboxGoods{
		ObjectBase:     object_beans.ObjectBase{Class: object_beans.CLASS_SHAREBOX_goods_file},
		GoodsId:        fileHash,
		Price:          price,
		SalesTotalMany: true,
		SalesTotal:     0,
		SoldTotal:      0,
		LockTotal:      0,
	}
	ERR := db.Sharebox_server_SaveOrUpdateGoodsPrice(&goods)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = chain_orders.OrderServer.UpGoods(chain_orders.GOODS_sharebox_file, fileHash, true, 0, price)
	return ERR
}

/*
查询文件价格列表
*/
func GetPriceListRange(startFileHash []byte, limit uint64) ([]model.FilePrice, utils.ERROR) {
	//filePrices, ERR := db.Sharebox_GetFilePriceRange(startFileHash, limit)
	//if ERR.CheckFail() {
	//	return nil, ERR
	//}
	//return filePrices, ERR
	return nil, utils.NewErrorSuccess()
}
