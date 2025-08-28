package file_transfer

import (
	"github.com/oklog/ulid/v2"
	"path/filepath"
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type Transfer struct {
	lock                    *sync.RWMutex                //
	classID                 uint64                       //
	area                    *libp2parea.Node             //
	whiteList               *WhiteList                   //白名单
	shareDir                map[string]string            //
	transferTaskByID        map[string]*FileTransferTask //作为 任务 key:string=SupplierID（提供者ID）;value:=;
	downloadSteps           map[string]*DownloadStep     //作为 任务 key:string=pullID（下载者ID）;value:=;
	autoReceive             uint32                       //是否自动接收文件。0=否;1=是;
	defaultDownloadPath     string                       //默认下载文件到本地地址
	pullFinishSignalMonitor chan *DownloadStep           //
	uploadFile              UploadFile                   //
	downloadFile            DownloadFile                 //
}

func NewTransfer(classID uint64, area *libp2parea.Node, pullFinishSignalMonitor chan *DownloadStep) *Transfer {
	t := Transfer{
		lock:                    new(sync.RWMutex),
		classID:                 classID,
		area:                    area,
		whiteList:               NewWhiteList(classID, area.GetLevelDB()),
		shareDir:                make(map[string]string),
		transferTaskByID:        make(map[string]*FileTransferTask),
		downloadSteps:           make(map[string]*DownloadStep),
		defaultDownloadPath:     FILEPATH_Recfilepath,
		pullFinishSignalMonitor: pullFinishSignalMonitor,
	}
	return &t
}

/*
删除一个上传任务
*/
func (this *Transfer) delTransferTask(supplierID []byte) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.transferTaskByID, utils.Bytes2string(supplierID))
}

/*
删除一个下载任务
*/
func (this *Transfer) delDownloadStep(pullID []byte) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.downloadSteps, utils.Bytes2string(pullID))
}

/*
查询传输任务
*/
func (this *Transfer) FindTransferTask(supplierID []byte) *FileTransferTask {
	this.lock.RLock()
	defer this.lock.RUnlock()
	ft, ok := this.transferTaskByID[utils.Bytes2string(supplierID)]
	if ok {
		return ft
	}
	return nil
}

/*
查询传输任务
*/
func (this *Transfer) FindDownload(pullID []byte) *DownloadStep {
	this.lock.RLock()
	defer this.lock.RUnlock()
	ft, ok := this.downloadSteps[utils.Bytes2string(pullID)]
	if ok {
		return ft
	}
	return nil
}

/*
设置是否自动接收文件
*/
func (this *Transfer) SetAutoReceive(auto bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if auto {
		this.autoReceive = 1
	} else {
		this.autoReceive = 0
	}
	saveAutoReceive(this.area.GetLevelDB(), this.classID, this.autoReceive)
}

/*
获取是否自动接收文件
*/
func (this *Transfer) GetAutoReceive() bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if this.autoReceive == 0 {
		return false
	} else {
		return true
	}
}

/*
设置接收文件地址
*/
func (this *Transfer) SetReceiveFilePath(filePath string) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	this.defaultDownloadPath = filePath
}

/*
获取接收文件地址
*/
func (this *Transfer) GetReceiveFilePath() string {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.defaultDownloadPath
}

/*
获取共享文件夹列表
*/
func (this *Transfer) GetShareDir(key string) string {
	this.lock.RLock()
	defer this.lock.RUnlock()
	v, ok := this.shareDir[key]
	if ok {
		return v
	}
	return ""
}

/*
添加共享文件夹
*/
func (this *Transfer) AddShareDir(dirPath string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	dirPath = filepath.Clean(dirPath)
	_, dir := filepath.Split(dirPath)
	this.shareDir[dir] = dirPath
	//保存到数据库
	list := make([]string, 0, len(this.shareDir))
	for _, one := range this.shareDir {
		list = append(list, one)
	}
	saveShareDirs(this.area.GetLevelDB(), this.classID, list)
}

/*
删除共享文件夹
*/
func (this *Transfer) DelShareDir(dir string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.shareDir, dir)
	list := make([]string, 0, len(this.shareDir))
	for _, one := range this.shareDir {
		list = append(list, one)
	}
	saveShareDirs(this.area.GetLevelDB(), this.classID, list)
}

/*
添加一个白名单地址
*/
func (this *Transfer) AddWhiteList(addr nodeStore.AddressNet) {
	this.whiteList.AddAddr(addr)
}

/*
删除一个白名单地址
*/
func (this *Transfer) DelWhiteList(addr nodeStore.AddressNet) {
	this.whiteList.DelAddr(addr)
}

/*
验证一个白名单地址
*/
func (this *Transfer) CheckWhiteList(addr nodeStore.AddressNet) bool {
	return this.whiteList.CheckAddr(addr)
}

func transferJoinKey(addr nodeStore.AddressNet, src string) string {
	return utils.Bytes2string(append(addr.GetAddr(), []byte(src)...))
}

/*
主动给对方发送文件
@sAddr         nodeStore.AddressNet    对方节点地址
@localPath     string                  保存到本地的文件路径
*/
func (this *Transfer) SendFile(sAddr nodeStore.AddressNet, localPath string, pullID []byte) ([]byte, utils.ERROR) {
	//创建任务
	ft, ERR := NewFileTransferTask(this.area.GetLevelDB(), this.classID, localPath, this.uploadFile, pullID)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//先把任务创建好
	this.lock.Lock()
	this.transferTaskByID[utils.Bytes2string(ft.SupplierID)] = ft
	this.lock.Unlock()

	fileIndex, err := NewFileIndex(ft)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}

	bs, err := fileIndex.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := this.area.SendP2pMsgHEWaitRequest(NET_CODE_apply_send_file, &sAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("上传文件索引错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	return ft.SupplierID, utils.NewErrorSuccess()
}

/*
从共享文件夹下载文件
@sAddr         nodeStore.AddressNet    对方节点地址
@remotePath    string                  对方文件路径
@localPath     string                  保存到本地的文件路径
*/
func (this *Transfer) DownloadFromShare(sAddr nodeStore.AddressNet, remotePath, localPath string) ([]byte, utils.ERROR) {
	remotePath = filepath.Clean(remotePath)
	localPath = filepath.Clean(localPath)
	//idBs, err := GetGenID(this.area.GetLevelDB())
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	fileIndex := &FileIndex{
		ClassID:        this.classID, //
		RemoteFilePath: remotePath,   //文件的路径
		PullID:         ulid.Make().Bytes(),
	}
	bs, err := fileIndex.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := this.area.SendP2pMsgHEWaitRequest(NET_CODE_apply_download_shareBox, &sAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("上传文件索引错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	newFileIndex, err := ParseFileIndex(nr.Data)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	newFileIndex.ClassID = fileIndex.ClassID
	newFileIndex.PullID = fileIndex.PullID
	ds := NewDownloadStep(this.area, sAddr, newFileIndex, localPath, this.pullFinishSignalMonitor, this.downloadFile, this)
	utils.Log.Info().Msgf("申请发送文件:%v", ds.FileIndex.PullID)
	this.lock.Lock()
	this.downloadSteps[utils.Bytes2string(ds.FileIndex.PullID)] = ds
	this.lock.Unlock()
	ds.Start()

	return ds.FileIndex.PullID, utils.NewErrorSuccess()
}

/*
获取下载任务详情
*/
//func (this *Transfer) GetDownloadByPullID(pullID []byte) *FileIndex {
//	this.lock.RLock()
//	defer this.lock.RUnlock()
//	ds, _ := this.downloadSteps[utils.Bytes2string(pullID)]
//	return ds.FileIndex
//}

/*
获取下载列表
*/
func (this *Transfer) GetDownloadList() []*DownloadStepVO {
	this.lock.RLock()
	defer this.lock.RUnlock()
	dss := make([]*DownloadStepVO, 0)
	for _, one := range this.downloadSteps {
		dss = append(dss, one.ConverVO())
	}
	return dss
}

/*
获取上传列表
*/
func (this *Transfer) GetUploadList() []*FileTransferTaskVO {
	this.lock.RLock()
	defer this.lock.RUnlock()
	dss := make([]*FileTransferTaskVO, 0)
	for _, one := range this.transferTaskByID {
		dss = append(dss, one.ConverVO())
	}
	return dss
}
