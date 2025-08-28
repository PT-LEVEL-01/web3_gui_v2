package transfer_manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

const (
	MinLenth                          int64 = 100 * 1024       //每次传最小（100kb）
	MaxLenth                          int64 = 20 * 1024 * 1024 //每次传输最大（20M）
	Transfer_slice_max_time_diff      int64 = 100              //每次传输最大时间差阀值（100ms）
	ErrNum                            int   = 5                //传输失败重试次数 5次
	Second                            int64 = 1                //传输速度统计时间间隔 1秒
	Recfilepath                             = "files"
	Transferlog_push_task_db_key            = "transfer_push_task_db_key"
	Transferlog_pull_task_db_key            = "transfer_pull_task_db_key"
	Transfer_push_task_id_max               = "transfer_push_task_id_max"
	Transfer_pull_task_id_max               = "transfer_pull_task_id_max"
	Transfer_task_expiration_interval       = 24 * 60 * 60 * time.Second        //任务有效期24小时
	Transfer_push_task_sharing_dirs         = "transfer_push_task_sharing_dirs" // 共享文件夹
	Transfer_pull_addr_white_list           = "transfer_pull_addr_white_list"   //有权限拉取文件的地址白名单
	Transfer_pull_task_if_atuo_db_key       = "transfer_pull_task_if_atuo_db_key"
	Transfer_p2p_mgs_timeout                = 5 * time.Second //p2p消息超时时间为5s
)

const (
	Transfer_pull_task_stautus_pending_confirmation = "pendingConfirmation" //任务待确认
	Transfer_pull_task_stautus_running              = "running"             // 运行中
	Transfer_pull_task_stautus_stop                 = "stop"                // 已停止
)

type TransferManger struct {
	PushTask                   *PushTask
	PullTask                   *PullTask
	exec                       chan *PullTask
	TaskMap                    *sync.Map
	PushTaskIDMax              uint64
	PullTaskIDMax              uint64
	PullIsAuto                 bool //是否自动
	area                       *libp2parea.Area
	levelDB                    *utilsleveldb.LevelDB
	lockPushTaskList           *sync.RWMutex
	lockPullTaskList           *sync.RWMutex
	pushFinishSignalSubscriber *sync.Map
	pushFinishSignalMonitor    chan *TransferMsg
	pullFinishSignalSubscriber *sync.Map
	pullFinishSignalMonitor    chan *PullTask
}

var TransferTaskManger *TransferManger

type PushTask struct {
	PushTaskID     uint64               `json:"push_task_id"`
	PullTaskID     uint64               `json:"pull_task_id"`
	To             nodeStore.AddressNet `json:"to"` //接收者
	Path           string               `json:"path"`
	Hash           []byte               `json:"hash"`
	ExpirationTime int64                `json:"expiration_time"`
	Size           int64                `json:"size"`
	Index          int64                `json:"index"` //下一次要从哪个偏移量
	Rate           int64                `json:"rate"`
}

type TransferMsg struct {
	PushTaskID uint64    `json:"push_task_id"`
	PullTaskID uint64    `json:"pull_task_id"`
	Lenth      int64     `json:"lenth"`
	FileInfo   *FileInfo `json:"file_info"`
}

type FileInfo struct {
	From  nodeStore.AddressNet `json:"from"` //传输者
	To    nodeStore.AddressNet `json:"to"`   //接收者
	Name  string               `json:"name"` //原始文件名
	Hash  []byte               `json:"hash"`
	Size  int64                `json:"size"`
	Path  string               `json:"path"`
	Index int64                `json:"index"` //下一次要从哪个偏移量
	Data  []byte               `json:"data"`
	Speed map[string]int64     `json:"speed"` //传输速度统计
	Rate  int64                `json:"rate"`
}

func NewTransferManger(area *libp2parea.Area, transfer_push, transfer_push_recv, transfer_pull, transfer_pull_recv, transfer_new_pull, transfer_new_pull_recv uint64) *TransferManger {
	msg_id_p2p_transfer_push = transfer_push
	msg_id_p2p_transfer_push_recv = transfer_push_recv
	msg_id_p2p_transfer_pull = transfer_pull
	msg_id_p2p_transfer_pull_recv = transfer_pull_recv
	msg_id_p2p_transfer_new_pull = transfer_new_pull
	msg_id_p2p_transfer_new_pull_recv = transfer_new_pull_recv

	tm := &TransferManger{
		area:                       area,
		levelDB:                    area.GetLevelDB(),
		lockPullTaskList:           new(sync.RWMutex),
		lockPushTaskList:           new(sync.RWMutex),
		exec:                       make(chan *PullTask, 20),
		TaskMap:                    new(sync.Map),
		pushFinishSignalSubscriber: new(sync.Map),
		pushFinishSignalMonitor:    make(chan *TransferMsg, 1),
		pullFinishSignalSubscriber: new(sync.Map),
		pullFinishSignalMonitor:    make(chan *PullTask, 1),
	}
	go tm.loopListenFinishSignal()

	area.Register_p2p(msg_id_p2p_transfer_push, tm.RecvNewPushTask)
	area.Register_p2p(msg_id_p2p_transfer_push_recv, tm.RecvNewPushTask_recv)

	area.Register_p2p(msg_id_p2p_transfer_pull, tm.FileSlicePush)
	area.Register_p2p(msg_id_p2p_transfer_pull_recv, tm.FileSlicePush_recv)

	area.Register_p2p(msg_id_p2p_transfer_new_pull, tm.RecvNewPullTask)
	area.Register_p2p(msg_id_p2p_transfer_new_pull_recv, tm.RecvNewPullTask_recv)

	TransferTaskManger = tm
	return tm
}

/*
获取文件发送完成信号
*/
func (this *TransferManger) GetListeningPushFinishSignal() (string, chan *TransferMsg) {
	c := make(chan *TransferMsg, 100000)
	id := uuid.NewV4().String()
	this.pushFinishSignalSubscriber.Store(id, c)
	return id, c
}

/*
获取文件接收完成信号
*/
func (this *TransferManger) GetListeningPullFinishSignal() (string, chan *PullTask) {
	c := make(chan *PullTask, 100000)
	id := uuid.NewV4().String()
	this.pullFinishSignalSubscriber.Store(id, c)
	return id, c
}

/*
循环监听文件传输完成信号
*/
func (this *TransferManger) loopListenFinishSignal() {
	for {
		select {
		case one := <-this.pushFinishSignalMonitor:
			this.pushFinishSignalSubscriber.Range(func(key, value interface{}) bool {
				c := value.(chan *TransferMsg)
				select {
				case c <- one:
				default:
				}
				return true
			})
		case one := <-this.pullFinishSignalMonitor:
			this.pullFinishSignalSubscriber.Range(func(key, value interface{}) bool {
				c := value.(chan *PullTask)
				select {
				case c <- one:
				default:
				}
				return true
			})
		}
	}
}

// 加载任务列表
func (this *TransferManger) Load() {
	go this.begin()

	newTransferPushKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transfer_push_task_id_max))
	if !ERR.CheckSuccess() {
		return
	}
	pushTaskIDMax, _ := this.levelDB.Find(*newTransferPushKey)
	if pushTaskIDMax != nil && pushTaskIDMax.Value != nil {
		this.PushTaskIDMax = utils.BytesToUint64(pushTaskIDMax.Value)
	}

	newTransferPullKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transfer_pull_task_id_max))
	if !ERR.CheckSuccess() {
		return
	}
	pullTaskIDMax, _ := this.levelDB.Find(*newTransferPullKey)
	if pullTaskIDMax != nil && pullTaskIDMax.Value != nil {
		this.PullTaskIDMax = utils.BytesToUint64(pullTaskIDMax.Value)
	}

	//加载拉取任务
	list, err := this.PullTask.PullTaskRecordList()
	if err != nil {
		return
	}
	for _, v := range list {
		if v.Status == Transfer_pull_task_stautus_running {
			v.CreatePullTask(this.pullFinishSignalMonitor)
		} else {
			this.TaskMap.Store(v.PullTaskID, v)
		}
	}
}

func (this *TransferManger) begin() {
	var timer = time.NewTicker(Transfer_task_expiration_interval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C: //定时清理过期推送任务
			go this.PushTask.clearExpirationPushTask()
		case s := <-this.exec: //执行一个拉取任务
			go s.Task()
		}
	}
}

/*
TransferPushTaskSharingDirs
@Description: 共享目录列表
@receiver this
@return sharingDirs
@return err
*/
func (this *TransferManger) TransferPushTaskSharingDirs() (sharingDirs map[string]string, err error) {
	return this.PushTask.TransferPushTaskSharingDirs()
}

/*
TransferPushTaskSharingDirsAdd
@Description: 共享目录添加
@receiver this
@param dir 目录绝对路径
@return error
*/
func (this *TransferManger) TransferPushTaskSharingDirsAdd(dir string) error {
	return this.PushTask.TransferPushTaskSharingDirsAdd(dir)
}

/*
TransferPushTaskSharingDirsDel
@Description: 共享目录删除
@receiver this
@param dir 目录绝对路径
@return error
*/
func (this *TransferManger) TransferPushTaskSharingDirsDel(dir string) error {
	return this.PushTask.TransferPushTaskSharingDirsDel(dir)
}

/*
TransferPullAddrWhiteList
@Description: 授权白名单地址列表
@receiver this
@return whiteList
@return err
*/
func (this *TransferManger) TransferPullAddrWhiteList() (list []string, err error) {
	whiteList, err := this.PushTask.TransferPullAddrWhiteList()
	if err != nil {
		return
	}
	for _, v := range whiteList {
		list = append(list, v.B58String())
	}
	return
}

/*
TransferPullAddrWhiteListAdd
@Description: 授权白名单地址添加
@receiver this
@param addr
@return error
*/
func (this *TransferManger) TransferPullAddrWhiteListAdd(addr nodeStore.AddressNet) error {
	return this.PushTask.TransferPullAddrWhiteListAdd(addr)
}

/*
TransferPullAddrWhiteListDel
@Description: 授权白名单地址删除
@receiver this
@param addr
@return error
*/
func (this *TransferManger) TransferPullAddrWhiteListDel(addr nodeStore.AddressNet) error {
	return this.PushTask.TransferPullAddrWhiteListDel(addr)
}

/*
NewPushTask
@Description: 传输文件申请(发起一个推送任务)
@receiver this
@param path
@param to
@return error
*/
func (this *TransferManger) NewPushTask(path string, to nodeStore.AddressNet) (*PushTask, error) {
	this.PushTask = &PushTask{
		To:             to,
		Path:           path,
		ExpirationTime: time.Now().Unix() + int64(Transfer_task_expiration_interval.Seconds()),
	}
	_, err := this.PushTask.CreatePushTask()
	if err != nil {
		return nil, err
	}
	return this.PushTask, nil
}

/*
PushTaskList
@Description: 推送任务列表
@receiver this
@return map[uint64]*PushTask
*/
func (this *TransferManger) PushTaskList() []*PushTask {
	m, _ := this.PushTask.pushTaskRecordList()
	if m == nil || len(m) == 0 {
		return nil
	}

	//按照任务id降序排列
	keys := make([]uint64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})
	list := make([]*PushTask, len(keys))
	for i, k := range keys {
		list[i] = m[k]
	}

	return list
}

/*
GetPushTaskByTaskId
@Description: 根据推送任务id获取推送任务
@receiver this
@param taskId 推送任务id
@return *PushTask
@return error
*/
func (this *TransferManger) GetPushTaskByTaskId(taskId uint64) (*PushTask, error) {
	list, _ := this.PushTask.pushTaskRecordList()
	task, ok := list[taskId]
	if !ok {
		return nil, errors.New("没有该任务！！")
	}
	return task, nil
}

/*
NewPullTask
@Description: 发起一个拉取文件任务
@receiver this
@param source 资源相对路径如:files/text.txt
@param path 相对路径如:files/text.txt
@param from 文件来源节点地址
@return error
*/
func (this *TransferManger) NewPullTask(source, path string, from nodeStore.AddressNet) error {
	//向from发起验证白名和共享文件夹，成功则加入拉取任务列表，
	sendMsg := TransferMsg{
		PullTaskID: TransferTaskManger.newTaskID(Transfer_pull_task_id_max),
		FileInfo: &FileInfo{
			To:   TransferTaskManger.area.GetNetId(),
			From: from,
			Path: source,
		},
	}

	fd, err := json.Marshal(sendMsg)
	if err != nil {
		utils.Log.Error().Msgf("json.Marshal失败:%s", err.Error())
		return err
	}

	message, ok, _, err := TransferTaskManger.area.SendP2pMsgWaitRequest(msg_id_p2p_transfer_new_pull, &from, &fd, Transfer_p2p_mgs_timeout)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return err
	}
	if ok {
		if message != nil {
			m, err := ParseMsg(*message)
			if err != nil {
				utils.Log.Error().Msgf("P2p消息解析失败:%s", err.Error())
				return err
			}

			//临时文件
			if path == "" {
				m.FileInfo.Path = filepath.Join(Recfilepath, m.FileInfo.Name)
			} else {
				m.FileInfo.Name = filepath.Base(path)
				m.FileInfo.Path = path
			}
			//添加传输任务
			this.PullTask = &PullTask{
				PushTaskID: m.PushTaskID,
				PullTaskID: m.PullTaskID,
				FileInfo:   m.FileInfo,
				exit:       make(chan struct{}),
				Status:     Transfer_pull_task_stautus_running,
			}
			err = this.PullTask.CreatePullTask(this.pullFinishSignalMonitor)
			if err != nil {
				return err
			}
		} else {
			utils.Log.Error().Msgf("对方节点创建推送任务失败")
			return errors.New("对方节点创建推送任务失败")
		}
	} else {
		utils.Log.Error().Msgf("发送P2p消息失败")
		return errors.New("发送P2p消息失败")
	}
	return nil
}

/*
PullTaskStart
@Description: 拉取任务开始
@receiver this
@param taskId 拉取任务id
@return ok
@return err
*/
func (this *TransferManger) PullTaskStart(taskId uint64, path string) error {
	task, err := this.PullTask.pullTaskRecordOne(taskId)
	if err != nil {
		return err
	}

	if task.Status == Transfer_pull_task_stautus_running {
		return errors.New("不能重复开启该任务")
	}

	if task.Status == Transfer_pull_task_stautus_stop && path != "" {
		return errors.New("该任务状态下不能更改存储目录")
	}

	//判断原来的零时目录
	if task.Status == Transfer_pull_task_stautus_pending_confirmation {
		//修改目录
		if path == "" {
			task.FileInfo.Path = filepath.Join(Recfilepath, task.FileInfo.Name)
		} else {
			task.FileInfo.Name = filepath.Base(path)
			task.FileInfo.Path = path
		}
	}

	task.Status = Transfer_pull_task_stautus_running
	err = task.CreatePullTask(this.pullFinishSignalMonitor)
	if err != nil {
		return err
	}
	return nil
}

/*
PullTaskStop
@Description: 拉取任务停止
@receiver this
@param taskId 拉取任务id
@return ok
@return err
*/
func (this *TransferManger) PullTaskStop(taskId uint64) error {
	//发送拉取消息
	v, ok := TransferTaskManger.TaskMap.Load(taskId)
	if !ok {
		return errors.New("未找到该任务")
	}
	task := v.(*PullTask)
	if task.Status == Transfer_pull_task_stautus_stop {
		return errors.New("不能重复停止该任务")
	}

	if task.exit == nil {
		return errors.New("不能停止该任务！")
	}

	task.exit <- struct{}{}
	return nil
}

/*
PullTaskDel
@Description: 拉取任务删除
@receiver this
@param taskId 拉取任务id
@return ok
@return err
*/
func (this *TransferManger) PullTaskDel(taskId uint64) error {
	task, err := this.PullTask.pullTaskRecordOne(taskId)
	if err != nil {
		return err
	}
	if task.Status == Transfer_pull_task_stautus_running {
		task.clear <- struct{}{}
	} else {
		err := task.pullTaskDel()
		return err
	}
	return nil
}

/*
PullTaskList
@Description: 拉取任务列表
@receiver this
@return []*PullTask
*/
func (this *TransferManger) PullTaskList() []*PullTask {
	list, _ := this.PullTask.PullTaskRecordList()
	return list
}

/*
GetPullTaskByTaskId
@Description: 根据拉取任务id获取拉取任务
@receiver this
@param taskId 拉取任务
@return *PullTask
@return error
*/
func (this *TransferManger) GetPullTaskByTaskId(taskId uint64) (*PullTask, error) {
	return this.PullTask.pullTaskRecordOne(taskId)
}

/*
PullTaskIsAutoSet
@Description: 设置是否自动拉取
@receiver this
@param auto  ture开启自动拉取
@return error
*/
func (this *TransferManger) PullTaskIsAutoSet(auto bool) error {
	b := []byte(strconv.FormatBool(auto))
	newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transfer_pull_task_if_atuo_db_key))
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	ERR = TransferTaskManger.levelDB.Save(*newKey, &b)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	return nil
}

/*
PullTaskIsAutoGet
@Description: 获取是否自动拉取状态
@receiver this
@return bool ture开启自动拉取
@return error
*/
func (this *TransferManger) PullTaskIsAutoGet() (bool, error) {
	return this.PullTask.PullTaskIsAutoGet()
}

/*
归还删除发送文件完成信号
*/
func (this *TransferManger) ReturnListeningPushFinishSignal(id string) {
	this.pushFinishSignalSubscriber.Delete(id)
}

/*
归还删除接收文件完成信号
*/
func (this *TransferManger) ReturnListeningPullFinishSignal(id string) {
	this.pullFinishSignalSubscriber.Delete(id)
}

// 创建一个push任务
func (t *PushTask) CreatePushTask() (*TransferMsg, error) {
	if !filepath.IsAbs(t.Path) {
		return nil, errors.New("文件路径必须是绝对路径！")
	}
	fi, err := os.Stat(t.Path)
	if err != nil {
		utils.Log.Error().Msgf("文件Stat失败:%s", err.Error())
		return nil, err
	}
	hasB, err := utils.FileSHA3_256(t.Path)
	if err != nil {
		utils.Log.Error().Msgf("文件hash失败:%s", err.Error())
		return nil, err
	}

	sendMsg := TransferMsg{
		PullTaskID: t.PullTaskID,
		PushTaskID: TransferTaskManger.newTaskID(Transfer_push_task_id_max),
		FileInfo: &FileInfo{
			To:   t.To,
			From: TransferTaskManger.area.GetNetId(),
			Name: fi.Name(),
			Size: fi.Size(),
			Hash: hasB,
		},
	}

	if sendMsg.FileInfo.Size < 1 {
		utils.Log.Error().Msgf("文件大小为空不能传输")
		return nil, err
	}

	if t.PullTaskID == 0 {
		fd, err := json.Marshal(sendMsg)
		if err != nil {
			utils.Log.Error().Msgf("json.Marshal失败:%s", err.Error())
			return nil, err
		}
		message, ok, _, err := TransferTaskManger.area.SendP2pMsgWaitRequest(msg_id_p2p_transfer_push, &sendMsg.FileInfo.To, &fd, Transfer_p2p_mgs_timeout)
		if err != nil {
			utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
			return nil, err
		}
		if ok {
			if message != nil {
				recv, err := ParseMsg(*message)
				if err != nil {
					utils.Log.Error().Msgf("P2p消息解析失败:%s", err.Error())
					return nil, err
				}
				if recv.PullTaskID < 1 {
					utils.Log.Error().Msgf("接收方任务id为空！！")
					return nil, err
				}
				t.PullTaskID = recv.PullTaskID
			} else {
				utils.Log.Error().Msgf("对方节点创建拉取任务失败")
				return nil, errors.New("对方节点创建拉取任务失败")
			}
		} else {
			utils.Log.Error().Msgf("发送P2p消息失败")
			return nil, errors.New("发送P2p消息失败")
		}
	}

	t.PushTaskID = sendMsg.PushTaskID
	t.Hash = hasB
	t.Size = sendMsg.FileInfo.Size
	err = t.pushTaskRecordSave()
	if err != nil {
		utils.Log.Error().Msgf(err.Error())
		return nil, err
	}
	return &sendMsg, nil

}

func (t *PushTask) pushTaskRecordSave() error {
	taskList, _ := t.pushTaskRecordList()
	if taskList == nil {
		taskList = make(map[uint64]*PushTask)
	}

	TransferTaskManger.lockPushTaskList.Lock()
	defer TransferTaskManger.lockPushTaskList.Unlock()

	taskList[t.PushTaskID] = t
	fd, err := json.Marshal(taskList)
	if err != nil {
		return err
	}
	newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transferlog_push_task_db_key))
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	ERR = TransferTaskManger.levelDB.Save(*newKey, &fd)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	return nil
}
func (t *PushTask) pushTaskRecordDel() error {
	taskList, _ := t.pushTaskRecordList()
	if taskList == nil {
		return nil
	}

	TransferTaskManger.lockPushTaskList.Lock()
	defer TransferTaskManger.lockPushTaskList.Unlock()

	delete(taskList, t.PushTaskID)
	fd, err := json.Marshal(taskList)
	if err != nil {
		return err
	}
	newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transferlog_push_task_db_key))
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	ERR = TransferTaskManger.levelDB.Save(*newKey, &fd)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	return nil
}
func (t *PushTask) pushTaskRecordList() (taskList map[uint64]*PushTask, err error) {
	TransferTaskManger.lockPushTaskList.RLock()
	defer TransferTaskManger.lockPushTaskList.RUnlock()

	newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transferlog_push_task_db_key))
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	item, err := TransferTaskManger.levelDB.Find(*newKey)
	if err != nil {
		//utils.Log.Info().Msgf("没有获取到db中的任务列表%s", err.Error())
		return nil, err
	}

	if item == nil || item.Value == nil || len(item.Value) == 0 {
		utils.Log.Error().Msgf("没有获取到db中的任务列表")
		return nil, errors.New("没有获取到任务列表")
	}

	decoder := json.NewDecoder(bytes.NewBuffer(item.Value))
	decoder.UseNumber()
	err = decoder.Decode(&taskList)
	if err != nil {
		utils.Log.Error().Msgf("db中的任务列表解析失败%s", err.Error())
		return
	}
	return
}

func (t *PushTask) checkPushTask(pushTaskId, pullTaskId uint64, sendId *nodeStore.AddressNet) (task *PushTask, err error) {
	taskList, err := t.pushTaskRecordList()
	if err != nil {
		return nil, err
	}
	task, ok := taskList[pushTaskId]
	if !ok {
		return nil, errors.New("没有该任务，无权限！！")
	}
	if !bytes.Equal(task.To, *sendId) {
		return nil, errors.New("节点无权限！！")
	}
	if task.PullTaskID != pullTaskId {
		return nil, errors.New("节点无任务权限！！")
	}
	if task.ExpirationTime > 0 && task.ExpirationTime <= time.Now().Unix() {
		return nil, errors.New("任务已过期！！")
	}

	return
}

func (t *PushTask) TransferPushTaskSharingDirs() (sharingDirs map[string]string, err error) {
	newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transfer_push_task_sharing_dirs))
	if !ERR.CheckSuccess() {
		return
	}
	b, err := TransferTaskManger.levelDB.Find(*newKey)
	if err != nil {
		utils.Log.Info().Msgf("没有获取到db中的共享目录%s", err.Error())
		return
	}

	if b == nil || b.Value == nil {
		utils.Log.Error().Msgf("没有获取到db中的共享目录")
		err = errors.New("没有获取到db中的共享目录")
		return
	}

	decoder := json.NewDecoder(bytes.NewBuffer(b.Value))
	decoder.UseNumber()
	err = decoder.Decode(&sharingDirs)
	if err != nil {
		utils.Log.Error().Msgf("db中的共享目录列表解析失败%s", err.Error())
		return
	}
	return
}

func (t *PushTask) TransferPushTaskSharingDirsAdd(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.New("不是一个有效目录")
	}

	key := filepath.Base(strings.TrimSuffix(filepath.ToSlash(dir), "/")) //获取最后一个目录名

	if key == "" {
		return errors.New("获取最后一个目录名失败")
	}

	list, _ := t.TransferPushTaskSharingDirs()
	if list == nil {
		list = make(map[string]string)
	} else {
		_, ok := list[key]
		if ok {
			return errors.New("目录key值重复，请更换最后一级目录名")
		}
	}

	list[key] = dir
	fd, err := json.Marshal(list)
	if err != nil {
		return err
	}
	newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transfer_push_task_sharing_dirs))
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	ERR = TransferTaskManger.levelDB.Save(*newKey, &fd)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	return nil
}

func (t *PushTask) TransferPushTaskSharingDirsDel(dir string) error {

	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.New("不是一个有效目录")
	}

	key := filepath.Base(strings.TrimSuffix(filepath.ToSlash(dir), "/")) //获取最后一个目录名

	if key == "" {
		return errors.New("获取最后一个目录名失败")
	}

	list, err := t.TransferPushTaskSharingDirs()
	if err != nil {
		return err
	}

	delete(list, key)

	fd, err := json.Marshal(list)
	if err != nil {
		return err
	}
	newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transfer_push_task_sharing_dirs))
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	ERR = TransferTaskManger.levelDB.Save(*newKey, &fd)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	return nil
}

func (t *PushTask) TransferPullAddrWhiteList() (whiteList []*nodeStore.AddressNet, err error) {
	newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transfer_pull_addr_white_list))
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	b, err := TransferTaskManger.levelDB.Find(*newKey)
	if err != nil {
		//utils.Log.Error().Msgf("没有获取到db中拉取地址的白名单%s", err.Error())
		return
	}

	if b == nil || b.Value == nil {
		//utils.Log.Error().Msgf("没有获取到db中拉取地址的白名单")
		err = errors.New("没有获取到db中拉取地址的白名单")
		return
	}

	decoder := json.NewDecoder(bytes.NewBuffer(b.Value))
	decoder.UseNumber()
	err = decoder.Decode(&whiteList)
	if err != nil {
		utils.Log.Error().Msgf("db中拉取地址的白名单列表解析失败%s", err.Error())
		return
	}
	return
}

func (t *PushTask) TransferPullAddrWhiteListAdd(addr nodeStore.AddressNet) error {
	list, _ := t.TransferPullAddrWhiteList()
	for k, _ := range list {
		if bytes.Equal(*list[k], addr) {
			return nil
		}
	}
	list = append(list, &addr)
	fd, err := json.Marshal(list)
	if err != nil {
		return err
	}
	newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transfer_pull_addr_white_list))
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	ERR = TransferTaskManger.levelDB.Save(*newKey, &fd)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	return nil
}

func (t *PushTask) TransferPullAddrWhiteListDel(addr nodeStore.AddressNet) error {
	list, err := t.TransferPullAddrWhiteList()
	if err != nil {
		return err
	}
	for k, _ := range list {
		if bytes.Equal(*list[k], addr) {
			list = append(list[:k], list[k+1:]...)
			break
		}
	}
	fd, err := json.Marshal(list)
	if err != nil {
		return err
	}
	newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transfer_pull_addr_white_list))
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	ERR = TransferTaskManger.levelDB.Save(*newKey, &fd)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	return nil
}

func (t *PushTask) clearExpirationPushTask() {
	taskList, _ := t.pushTaskRecordList()
	if len(taskList) == 0 {
		return
	}

	TransferTaskManger.lockPushTaskList.Lock()
	defer TransferTaskManger.lockPushTaskList.Unlock()
	newT := time.Now().Unix()
	var update bool
	for _, v := range taskList {
		if v.ExpirationTime > 0 && v.ExpirationTime <= newT {
			delete(taskList, v.PushTaskID)
			update = true
		}
	}

	if update {
		fd, err := json.Marshal(taskList)
		if err != nil {
			utils.Log.Info().Msgf("没有获取到db中的任务列表%s", err.Error())
			return
		}
		newKey, ERR := utilsleveldb.NewLeveldbKey([]byte(Transferlog_push_task_db_key))
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("没有生成出db中的任务key%s", ERR.String())
			return
		}
		TransferTaskManger.levelDB.Save(*newKey, &fd)
	}

	return
}

// 采集速度参数
func (f *FileInfo) SetSpeed(stime int64, size int) error {
	//if f.Speed == nil {
	//	f.Speed = make(map[string]int64, 0)
	//}

	if _, ok := f.Speed["time"]; !ok {
		f.Speed["time"] = stime
		f.Speed["size"] = int64(size)
	}
	if time.Now().Unix()-f.Speed["time"] > Second {
		f.Speed["time"] = stime
		f.Speed["size"] = 0
	} else {
		f.Speed["size"] += int64(size)
	}
	return nil
}

// 获取速度
func (f *FileInfo) GetSpeed() int64 {
	t := time.Now().Unix() - f.Speed["time"]
	if t < 1 {
		t = 1
	}
	return f.Speed["size"] / t / 1024
}

func (f *FileInfo) CheckFileHash() (err error) {
	hasB, err := utils.FileSHA3_256(f.Path)
	if err != nil {
		utils.Log.Error().Msgf("文件hash失败:%s", err.Error())
		return err
	}
	if !bytes.Equal(hasB, f.Hash) {
		utils.Log.Error().Msgf("文件上传错误！不完整或已损坏")
		return errors.New("文件上传错误！不完整或已损坏")
	}
	return err
}

// 解析消息
func ParseMsg(d []byte) (*TransferMsg, error) {
	msg := &TransferMsg{}
	// err := json.Unmarshal(d, msg)
	decoder := json.NewDecoder(bytes.NewBuffer(d))
	decoder.UseNumber()
	err := decoder.Decode(msg)
	if err != nil {
		fmt.Println(err)
	}
	return msg, err
}
