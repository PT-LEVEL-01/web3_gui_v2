package transfer_manager

import (
	"github.com/oklog/ulid/v2"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/utils"
)

/*
文件hash缓存，异步计算hash
*/
type FileHashProcessManager struct {
	lock        *sync.RWMutex
	m           map[string]*FileHashProcess
	processChan chan []byte
	cacheLock   *sync.RWMutex      //
	cache       map[string]*[]byte //key:=文件绝对路径;value:=文件hash;
	//cacheDir    map[string][]*FileHashInfo //key:=文件夹名称;value:=文件hash信息;
}

func NewFileHashProcessManager() *FileHashProcessManager {
	fmp := FileHashProcessManager{
		lock:        new(sync.RWMutex),
		m:           make(map[string]*FileHashProcess),
		processChan: make(chan []byte, 100),
		cacheLock:   new(sync.RWMutex),
		cache:       make(map[string]*[]byte),
	}
	go fmp.loopCleanProcess()
	go fmp.runProcess()
	return &fmp
}

/*
添加一个任务
*/
func (this *FileHashProcessManager) AddProcess(filePaths []string) (*FileHashProcess, utils.ERROR) {
	dirFiles, filePrices, ERR := GetFileInfo(filePaths)
	if ERR.CheckFail() {
		return nil, ERR
	}
	fhp := FileHashProcess{
		Id:         ulid.Make().Bytes(),                       //
		Count:      uint64(len(filePrices)),                   //任务总量
		FilePrices: append(dirFiles, filePrices...),           //
		FinishTime: 0,                                         //完成时间，完成后，一定时间内不去取出则删除
		Cancel:     false,                                     //
		lock:       new(sync.RWMutex),                         //
		Signe:      make(chan *FileHashInfo, len(filePrices)), //
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
查询一个文件夹下的文件列表
*/
//func (this *FileHashProcessManager) FindDir(dir string) []*FileHashInfo {
//	this.cacheLock.RLock()
//	defer this.cacheLock.RUnlock()
//	fhp, ok := this.cacheDir[dir]
//	if !ok {
//		return nil
//	}
//	return fhp
//}

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
			//查缓存
			this.cacheLock.RLock()
			hashBs, ok := this.cache[one.AbsPath]
			this.cacheLock.RUnlock()
			if ok {
				one.Hash = *hashBs
			} else {
				//
				hashBs, err := utils.FileSHA3_256(filepath.Join(one.DirPath, one.Name))
				if err != nil {
					utils.Log.Error().Str("err", err.Error()).Send()
					continue
				}
				//utils.Log.Info().Hex("计算hash", hashBs).Send()
				one.Hash = hashBs
				//放缓存
				this.cacheLock.Lock()
				this.cache[one.AbsPath] = &hashBs
				this.cacheLock.Unlock()
			}
			select {
			case fhp.Signe <- one:
			default:
			}
		}
		fhp.SetFinishTime()
	}
}

type FileHashProcess struct {
	Id         []byte             //
	Count      uint64             //任务总量
	FilePrices []*FileHashInfo    //
	FinishTime int64              //完成时间，完成后，一定时间内不去取出则删除
	Cancel     bool               //是否取消任务
	lock       *sync.RWMutex      //
	Signe      chan *FileHashInfo //
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
func GetFileInfo(filePaths []string) ([]*FileHashInfo, []*FileHashInfo, utils.ERROR) {
	dirInfos := make([]*FileHashInfo, 0)
	fileInfos := make([]*FileHashInfo, 0)
	for _, dirPathOne := range filePaths {
		//utils.Log.Info().Str("文件夹路径", dirPathOne).Send()
		fi, err := os.Stat(dirPathOne)
		if err != nil {
			utils.Log.Error().Err(err).Send()
			return nil, nil, utils.NewErrorSysSelf(err)
		}
		if fi.IsDir() {
			dirEntry, err := os.ReadDir(dirPathOne)
			if err != nil {
				utils.Log.Error().Err(err).Send()
				return nil, nil, utils.NewErrorSysSelf(err)
			}
			for _, one := range dirEntry {
				absPath, err := filepath.Abs(filepath.Join(dirPathOne, one.Name()))
				if err != nil {
					return nil, nil, utils.NewErrorSysSelf(err)
				}
				dirOne := FileHashInfo{
					Name:    one.Name(),
					Hash:    nil,
					Size:    0,
					Time:    0,
					DirPath: dirPathOne,
					IsDir:   one.IsDir(),
					AbsPath: absPath,
				}
				if one.IsDir() {
					dirInfos = append(dirInfos, &dirOne)
				} else {
					fileInfos = append(fileInfos, &dirOne)
				}
			}
		} else {
			absPath, err := filepath.Abs(filepath.Join(dirPathOne, fi.Name()))
			if err != nil {
				return nil, nil, utils.NewErrorSysSelf(err)
			}
			fi := FileHashInfo{
				Name: fi.Name(), //真实文件名
				//Hash:    ,              //hash值
				Size:    fi.Size(),                //文件大小
				Time:    fi.ModTime().Unix(),      //创建时间
				DirPath: filepath.Dir(dirPathOne), //
				IsDir:   false,                    //
				AbsPath: absPath,                  //
			}
			fileInfos = append(fileInfos, &fi)
		}
	}
	//utils.Log.Info().Str("开始计算文件hash", "end").Send()
	return dirInfos, fileInfos, utils.NewErrorSuccess()
}

/*
获取一个文件夹下的所有文件夹
*/
func GetDirPaths(dirPaths string) ([]string, utils.ERROR) {
	if dirPaths == "" {
		return nil, utils.NewErrorSuccess()
	}
	paths := make([]string, 0)
	err := filepath.WalkDir(dirPaths, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return paths, utils.NewErrorSuccess()
}

/*
文件信息
*/
type FileHashInfo struct {
	Name    string //真实文件名
	Hash    []byte //hash值
	Size    int64  //文件大小
	Time    int64  //文件创建时间
	DirPath string //父文件夹路径
	IsDir   bool   //是否文件夹
	AbsPath string //文件绝对路径
}
