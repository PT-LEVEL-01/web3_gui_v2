package file_transfer

import (
	"github.com/oklog/ulid/v2"
	"sync"
	"web3_gui/config"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

// var manager = NewManager()
var ManagerStatic *Manager

type Manager struct {
	class                      *sync.Map              //key:uint64=类型ID;value:*Transfer=传输管理器;
	area                       *libp2parea.Node       //
	pushFinishSignalSubscriber *sync.Map              //保存订阅上传完成信号管道
	pushFinishSignalMonitor    chan *FileTransferTask //
	pullFinishSignalSubscriber *sync.Map              //保存订阅下载完成信号管道
	pullFinishSignalMonitor    chan *DownloadStep     //
}

func NewManager(area *libp2parea.Node) *Manager {
	m := Manager{
		class:                      new(sync.Map),
		area:                       area,
		pushFinishSignalSubscriber: new(sync.Map),
		pushFinishSignalMonitor:    make(chan *FileTransferTask, 10),
		pullFinishSignalSubscriber: new(sync.Map),
		pullFinishSignalMonitor:    make(chan *DownloadStep, 10),
	}
	go m.loopListenFinishSignal()
	m.RegisterHandlers()
	return &m
}

/*
创建一个类型的文件传输管理器
*/
func (this *Manager) CreateClass(classID uint64) *Transfer {
	t := NewTransfer(classID, this.area, this.pullFinishSignalMonitor)
	this.class.Store(classID, t)
	return t
}

/*
获取一个类型的文件传输管理器
*/
func (this *Manager) GetClass(classID uint64) *Transfer {
	value, ok := this.class.Load(classID)
	if !ok {
		return nil
	}
	t := value.(*Transfer)
	return t
}

/*
删除一个类型的文件传输管理器
*/
func (this *Manager) DelClass(name string) {
	this.class.Delete(name)
}

/*
设置一个接口
*/
func (this *Manager) SetFile(classID uint64, uploadFile UploadFile, downloadFile DownloadFile) utils.ERROR {
	tf := this.GetClass(classID)
	if tf == nil {
		return utils.NewErrorBus(config.ERROR_CODE_file_transfer_classID_not_find, "")
	}
	tf.uploadFile = uploadFile
	tf.downloadFile = downloadFile
	return utils.NewErrorSuccess()
}

/*
获取文件发送完成信号
*/
func (this *Manager) GetListeningPushFinishSignal() (string, chan *FileTransferTask) {
	c := make(chan *FileTransferTask, 10000)
	id := ulid.Make().String() // uuid.NewV4().String()
	this.pushFinishSignalSubscriber.Store(id, c)
	return id, c
}

/*
获取文件接收完成信号
*/
func (this *Manager) GetListeningPullFinishSignal() (string, chan *DownloadStep) {
	c := make(chan *DownloadStep, 10000)
	id := ulid.Make().String() // uuid.NewV4().String()
	this.pullFinishSignalSubscriber.Store(id, c)
	return id, c
}

/*
循环监听文件传输完成信号
*/
func (this *Manager) loopListenFinishSignal() {
	for {
		select {
		case one := <-this.pushFinishSignalMonitor:
			this.delTransferTask(one)
			this.pushFinishSignalSubscriber.Range(func(key, value interface{}) bool {
				c := value.(chan *FileTransferTask)
				select {
				case c <- one:
				default:
				}
				return true
			})
		case one := <-this.pullFinishSignalMonitor:
			this.delDownloadStep(one)
			this.pullFinishSignalSubscriber.Range(func(key, value interface{}) bool {
				c := value.(chan *DownloadStep)
				select {
				case c <- one:
				default:
				}
				return true
			})
		}
	}
}

/*
归还删除发送文件完成信号
*/
func (this *Manager) ReturnListeningPushFinishSignal(id string) {
	this.pushFinishSignalSubscriber.Delete(id)
}

/*
归还删除接收文件完成信号
*/
func (this *Manager) ReturnListeningPullFinishSignal(id string) {
	this.pullFinishSignalSubscriber.Delete(id)
}

/*
发送文件
*/
func (this *Manager) SendFile(classID uint64, addr nodeStore.AddressNet, localPath string, pullID []byte) ([]byte, utils.ERROR) {
	tf := this.GetClass(classID)
	if tf == nil {
		return nil, utils.NewErrorBus(config.ERROR_CODE_file_transfer_classID_not_find, "")
	}
	supplierID, ERR := tf.SendFile(addr, localPath, pullID)
	return supplierID, ERR
}

/*
获取下载列表
*/
func (this *Manager) GetDownloadList() []*DownloadStepVO {
	dss := make([]*DownloadStepVO, 0)
	this.class.Range(func(k, v interface{}) bool {
		t := v.(*Transfer)
		dss = append(dss, t.GetDownloadList()...)
		return true
	})
	return dss
}

/*
获取上传列表
*/
func (this *Manager) GetUploadList() []*FileTransferTaskVO {
	dss := make([]*FileTransferTaskVO, 0)
	this.class.Range(func(k, v interface{}) bool {
		t := v.(*Transfer)
		dss = append(dss, t.GetUploadList()...)
		return true
	})
	return dss
}

/*
删除一个下载任务
*/
func (this *Manager) delDownloadStep(ds *DownloadStep) {
	v, ok := this.class.Load(ds.FileIndex.ClassID)
	if !ok {
		return
	}
	t := v.(*Transfer)
	t.delDownloadStep(ds.PullID)
}

/*
删除一个下载任务
*/
func (this *Manager) delTransferTask(ft *FileTransferTask) {
	v, ok := this.class.Load(ft.ClassID)
	if !ok {
		return
	}
	t := v.(*Transfer)
	t.delTransferTask(ft.SupplierID)
}
