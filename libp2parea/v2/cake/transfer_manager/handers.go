package transfer_manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"web3_gui/libp2parea/v2/message_center"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

//func (this *TransferManger) RecvNewPushTask_recv(message *message_center.MessageBase) {
//	// utils.Log.Info().Msgf("收到P2P消息返回 from:%s", message.Head.Sender.B58String())
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}
//
//func (this *TransferManger) FileSlicePush_recv(message *message_center.MessageBase) {
//	// utils.Log.Info().Msgf("收到SearchSuper消息返回 from:%s", message.Head.Sender.B58String())
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}
//
//func (this *TransferManger) RecvNewPullTask_recv(message *message_center.MessageBase) {
//	// utils.Log.Info().Msgf("收到SearchSuper消息返回 from:%s", message.Head.Sender.B58String())
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

// 接收传输文件申请并创建pull任务
func (this *TransferManger) RecvNewPushTask(message *message_center.MessageBase) {
	content := message.Content
	m, err := ParseMsg(content)
	if err == nil {

		pullTaskID := TransferTaskManger.newTaskID(Transfer_pull_task_id_max)
		//回复任务id
		m.PullTaskID = pullTaskID
		recv, err := json.Marshal(m)
		if err != nil {
			fmt.Println(err)
		}
		TransferTaskManger.area.SendP2pReplyMsg(message, &recv)

		//添加传输任务
		task := &PullTask{
			PushTaskID: m.PushTaskID,
			PullTaskID: pullTaskID,
			FileInfo:   m.FileInfo,
		}
		//路径默认为
		task.FileInfo.Path = filepath.Join(Recfilepath, m.FileInfo.Name)
		task.Status = Transfer_pull_task_stautus_running
		//判断是否自动拉取
		autoPull, _ := task.PullTaskIsAutoGet()
		if !autoPull {
			task.Status = Transfer_pull_task_stautus_pending_confirmation
		}

		task.CreatePullTask(this.pullFinishSignalMonitor)
		return
	}
	utils.Log.Error().Msgf("p2p消息解析败:%s", err.Error())
	TransferTaskManger.area.SendP2pReplyMsg(message, nil)
	return
}

// 接收一个Pull任务申请，并创建对应的push任务
func (this *TransferManger) RecvNewPullTask(message *message_center.MessageBase) {
	content := message.Content
	m, err := ParseMsg(content)
	if err == nil {
		//验证白名单
		var task *PushTask
		whiteList, _ := task.TransferPullAddrWhiteList()
		if len(whiteList) > 0 {
			var have bool
			for k, _ := range whiteList {
				if bytes.Equal(whiteList[k].GetAddr(), m.FileInfo.To.GetAddr()) {
					have = true
					break
				}
			}
			if !have {
				utils.Log.Info().Msgf("验证白名单失败,节点无权限")
				TransferTaskManger.area.SendP2pReplyMsg(message, nil)
				return
			}
		}

		//验证地址
		sharingDirs, _ := task.TransferPushTaskSharingDirs()
		strs := strings.Split(strings.TrimLeft(strings.TrimLeft(filepath.ToSlash(m.FileInfo.Path), "./"), "/"), "/")
		if len(strs) <= 1 {
			utils.Log.Info().Msgf("解析共享目录地址key失败，不是有效路径！")
			TransferTaskManger.area.SendP2pReplyMsg(message, nil)
			return
		}
		dir, ok := sharingDirs[strs[0]]
		if !ok {
			utils.Log.Info().Msgf("解析共享目录地址key失败，无权限")
			TransferTaskManger.area.SendP2pReplyMsg(message, nil)
			return
		}
		//创建推送任务
		task = &PushTask{
			PullTaskID: m.PullTaskID,
			To:         m.FileInfo.To,
			Path:       filepath.Join(dir, filepath.Join(strs[1:]...)),
		}
		sendMsg, ERR := task.CreatePushTask()
		if ERR.CheckFail() {
			TransferTaskManger.area.SendP2pReplyMsg(message, nil)
			utils.Log.Info().Msgf("创建推送任务失败:%s", ERR.String())
			return
		}

		fd, err := json.Marshal(sendMsg)
		if err != nil {
			utils.Log.Error().Msgf("json.Marshal失败:%s", err.Error())
			return
		}
		TransferTaskManger.area.SendP2pReplyMsg(message, &fd)
		return
	}
	utils.Log.Info().Msgf("解析msg失败:%s", err.Error())
	TransferTaskManger.area.SendP2pReplyMsg(message, nil)
	return
}

func (t *PullTask) PullTaskIsAutoGet() (bool, utils.ERROR) {
	newKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(Transfer_pull_task_if_atuo_db_key))
	if !ERR.CheckSuccess() {
		return false, ERR
	}
	b, err := TransferTaskManger.levelDB.Find(*newKey)
	if err != nil || b == nil || b.Value == nil {
		return false, utils.NewErrorSysSelf(err)
	}
	auto, err := strconv.ParseBool(string(b.Value))
	if err != nil {
		return false, utils.NewErrorSysSelf(err)
	}
	return auto, utils.NewErrorSuccess()
}

func (this *TransferManger) newTaskID(idtype string) uint64 {
	if idtype == Transfer_push_task_id_max {
		b := utils.Uint64ToBytes(atomic.AddUint64(&this.PushTaskIDMax, 1))
		newKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(Transfer_push_task_id_max))
		if !ERR.CheckSuccess() {
			return 0
		}
		this.levelDB.Save(*newKey, &b)
		return this.PushTaskIDMax
	}
	if idtype == Transfer_pull_task_id_max {
		b := utils.Uint64ToBytes(atomic.AddUint64(&this.PullTaskIDMax, 1))
		newKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(Transfer_pull_task_id_max))
		if !ERR.CheckSuccess() {
			return 0
		}
		this.levelDB.Save(*newKey, &b)
		return this.PullTaskIDMax
	}
	return 0
}

func (t *PullTask) PullTaskRecordList() ([]*PullTask, utils.ERROR) {
	taskList := make([]*PullTask, 0)
	TransferTaskManger.lockPullTaskList.RLock()
	defer TransferTaskManger.lockPullTaskList.RUnlock()

	newKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(Transferlog_pull_task_db_key))
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	b, err := TransferTaskManger.levelDB.Find(*newKey)
	if err != nil {
		//utils.Log.Info().Msgf("没有获取到db中的任务列表%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if b == nil || b.Value == nil {
		//utils.Log.Error().Msgf("没有获取到db中的任务列表")
		return taskList, utils.NewErrorBus(ERROR_CODE_task_list_zero, "") //errors.New("没有获取到任务列表")
	}

	decoder := json.NewDecoder(bytes.NewBuffer(b.Value))
	decoder.UseNumber()
	err = decoder.Decode(&taskList)
	if err != nil {
		utils.Log.Error().Msgf("db中的任务列表解析失败%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	return taskList, utils.NewErrorSuccess()
}

// 拉取文件任务
type PullTask struct {
	PushTaskID    uint64    `json:"push_task_id"`
	PullTaskID    uint64    `json:"pull_task_id"`
	FileInfo      *FileInfo `json:"file_info"`
	Status        string    `json:"status"` //ture 为传输中 false为暂停中
	Fault         string    `json:"fault"`  //传输文件过程中的异常
	exit          chan struct{}
	clear         chan struct{}
	finishSiglnal chan *PullTask
}

func (t *PullTask) CreatePullTask(tag chan *PullTask) utils.ERROR {
	utils.Log.Info().Msgf("CreatePullTask任务id:%d", t.PullTaskID)
	//初始化
	t.exit = make(chan struct{})
	t.clear = make(chan struct{})
	t.FileInfo.Speed = make(map[string]int64, 0)
	t.Fault = ""
	t.finishSiglnal = tag
	//保存任务
	ERR := t.pullTaskRecordSave()
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("保存错误:%s", ERR.String())
		return ERR
	}

	if t.Status == Transfer_pull_task_stautus_running {
		dir := filepath.Dir(t.FileInfo.Path)
		utils.CheckCreateDir(dir)
		tmpPath := filepath.Join(dir, t.FileInfo.Name+"_"+t.FileInfo.From.B58String()+"_"+strconv.FormatUint(t.PullTaskID, 10)+"_tmp")
		fi, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			utils.Log.Error().Msgf("临时文件新建失败:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
		defer fi.Close()

		err = fi.Truncate(t.FileInfo.Size)
		if err != nil {
			utils.Log.Error().Msgf("空文件新建失败:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}

		TransferTaskManger.exec <- t
	}
	return utils.NewErrorSuccess()
}

func (t *PullTask) Task() {
	var errnum int
	var lenth = MinLenth
	for {
		select {
		case <-t.exit:
			t.Status = Transfer_pull_task_stautus_stop
			utils.Log.Info().Msgf("退出任务id:%d", t.PullTaskID)
			ERR := t.pullTaskRecordSave()
			if ERR.CheckFail() {
				utils.Log.Error().Msgf("保存错误:%s", ERR.String())
			}
			return
		case <-t.clear:
			utils.Log.Info().Msgf("删除任务id:%d", t.PullTaskID)
			ERR := t.pullTaskDel()
			if ERR.CheckFail() {
				utils.Log.Error().Msgf("删除任务错误:%s", ERR.String())
			}
			return
		default:
			okf, p2pUseTime, ERR := t.fileSlicePull(lenth)
			if okf == true { //已传输完，则退出
				t.Status = Transfer_pull_task_stautus_stop
				ERR = t.pullTaskRecordDel()
				if ERR.CheckFail() {
					utils.Log.Error().Msgf("PullTaskID %d 已传输完，清理错误:%s", t.PullTaskID, ERR.String())
				}
				select {
				case t.finishSiglnal <- t:
				default:
				}
				return
			}
			if ERR.CheckFail() {
				//开始重传
				errnum++
				if errnum <= ErrNum {
					utils.Log.Info().Msgf("resend slice...PullTaskID:%d", t.PullTaskID)
					continue
				}
				t.Status = Transfer_pull_task_stautus_stop
				t.Fault = ERR.String()
				ERR := t.pullTaskRecordSave()
				if ERR.CheckFail() {
					utils.Log.Error().Msgf("PullTaskID %d 保存错误:%s", t.PullTaskID, ERR.String())
				}
				return
			}

			lenth = getNextLenth(p2pUseTime, lenth)
			//utils.Log.Info().Msgf("lenth---------------------:%d", lenth)
		}
	}

}

// 根据处理速度获取一下次拉取数据量
func getNextLenth(p2pUseTime, lenth int64) int64 {
	if Transfer_slice_max_time_diff-p2pUseTime < 0 {
		if (lenth / MinLenth) > 1 {
			//utils.Log.Info().Msgf("开始限速----------------:%d", p2pUseTime)
			return lenth / 2
		}
	} else if lenth < MaxLenth {
		//utils.Log.Info().Msgf("开始加速+++++++++++++++:%d", p2pUseTime)
		if lenth*2 > MaxLenth {
			return (MaxLenth - lenth) + lenth
		}
		return lenth * 2
	}
	//utils.Log.Info().Msgf("不限速=====================================:%d", p2pUseTime)
	return lenth
}

//分段传输，续传
/**
@return okf 是否传送完 err 错误
*/
func (t *PullTask) fileSlicePull(lenth int64) (okf bool, p2pUseTime int64, ERR utils.ERROR) {
	//已经传完
	if t.FileInfo.Index >= t.FileInfo.Size && t.FileInfo.Size > 0 {
		return true, p2pUseTime, utils.NewErrorSuccess()
	}

	sendMsg := &TransferMsg{
		PushTaskID: t.PushTaskID,
		PullTaskID: t.PullTaskID,
		Lenth:      lenth,
		FileInfo:   t.FileInfo,
	}

	fd, err := json.Marshal(sendMsg)
	if err != nil {
		return false, p2pUseTime, utils.NewErrorSysSelf(err)
	}
	milli := time.Now().UnixMilli()
	message, ERR := TransferTaskManger.area.SendP2pMsgWaitRequest(msg_id_p2p_transfer_pull, &t.FileInfo.From, &fd, Transfer_p2p_mgs_timeout)
	p2pUseTime = time.Now().UnixMilli() - milli
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", ERR.String())
		return false, p2pUseTime, ERR
	} else {
		if message != nil {
			//发送成功，对方已经接收到消息
			m, err := ParseMsg(*message)
			if err != nil {
				utils.Log.Error().Msgf("P2p消息解析失败:%s", err.Error())
				return false, p2pUseTime, utils.NewErrorSysSelf(err)
			}

			tmpPath := filepath.Join(filepath.Dir(t.FileInfo.Path), t.FileInfo.Name+"_"+t.FileInfo.From.B58String()+"_"+
				strconv.FormatUint(t.PullTaskID, 10)+"_tmp")
			fi, err := os.OpenFile(tmpPath, os.O_RDWR, os.ModePerm)
			if err != nil {
				utils.Log.Error().Msgf("临时文件打开失败:%s", err.Error())
				return false, p2pUseTime, utils.NewErrorSysSelf(err)
			}
			fi.Seek(t.FileInfo.Index, 0)
			fi.Write(m.FileInfo.Data)
			defer fi.Close()

			t.FileInfo.Index = m.FileInfo.Index

			//utils.Log.Info().Msgf("接收的start：%d", t.FileInfo.Index)
			t.FileInfo.Rate = int64(float64(m.FileInfo.Index) / float64(m.FileInfo.Size) * float64(100))
			//utils.Log.Info().Msgf("接收的百分比：%d%%", t.FileInfo.Rate)
			t.FileInfo.SetSpeed(time.Now().Unix(), len(*message))
			//speed := t.FileInfo.GetSpeed()
			//utils.Log.Info().Msgf("接收的速率：%d KB/s", speed)
			//传输完成，则更新状态
			if t.FileInfo.Rate >= 100 {
				//传输完成，则重命名文件名
				num := 0
			Rename:
				//如果文件存在，则重命名为新的文件
				if ok, _ := utils.PathExists(t.FileInfo.Path); ok {
					num++
					filenamebase := filepath.Base(t.FileInfo.Name)
					fileext := filepath.Ext(t.FileInfo.Name)
					filename := strings.TrimSuffix(filenamebase, fileext)
					newname := filename + "_" + strconv.Itoa(num) + fileext
					t.FileInfo.Path = filepath.Join(filepath.Dir(t.FileInfo.Path), newname)
					if ok1, _ := utils.PathExists(t.FileInfo.Path); ok1 {
						goto Rename
					}
					t.FileInfo.Name = newname
				}

				fi.Close()
				os.Rename(tmpPath, t.FileInfo.Path)
				utils.Log.Info().Msgf("PullTaskID %d 文件完成：%d %d", t.PullTaskID, t.FileInfo.Index, t.FileInfo.Rate)
				//检查文件完整性
				ERR = t.FileInfo.CheckFileHash()
				okf = true // 发送完成
				t.finishSiglnal <- t
				return
			} else {
				t.pullTaskRecordSave()
			}
		} else {
			utils.Log.Error().Msgf("文件拉取失败")
			return false, p2pUseTime, utils.NewErrorBus(ERROR_CODE_file_pull_fail, "")
		}
	}
	return
}

// 分片推送
func (this *TransferManger) FileSlicePush(message *message_center.MessageBase) {
	content := message.Content
	m, err := ParseMsg(content)
	if err == nil {

		//验证taskID
		sendId := message.SenderAddr
		var task *PushTask
		task, ERR := task.checkPushTask(m.PushTaskID, m.PullTaskID, sendId)
		if ERR.CheckFail() {
			utils.Log.Error().Msgf("验证task失败：%s", ERR.String())
			TransferTaskManger.area.SendP2pReplyMsg(message, nil)
			return
		}

		index := m.FileInfo.Index //当前已传偏移量
		fi, err := os.Open(task.Path)
		if err != nil {
			utils.Log.Error().Msgf("文件打开失败：%s", err.Error())
			TransferTaskManger.area.SendP2pReplyMsg(message, nil)
			return
		}
		defer fi.Close()
		stat, err := fi.Stat()
		if err != nil {
			utils.Log.Error().Msgf("文件打开失败：%s", err.Error())
			TransferTaskManger.area.SendP2pReplyMsg(message, nil)
			return
		}
		size := stat.Size()
		start := index

		if m.Lenth > MaxLenth {
			m.Lenth = MaxLenth
		}

		length := m.Lenth
		//如果偏移量小于文件大小，并且剩余大小小于长度，则长度为剩余大小(即最后一段)
		if start < size && size-index < m.Lenth {
			length = size - index
		}
		buf := make([]byte, length)
		_, err = fi.ReadAt(buf, start)
		if err != nil {
			utils.Log.Error().Msgf("读取文件流失败：%s", err.Error())
			TransferTaskManger.area.SendP2pReplyMsg(message, nil)
			return
		}
		//下一次start位置
		nextstart := start + m.Lenth
		if nextstart > size {
			nextstart = size
		}
		m.FileInfo.Size = size
		m.FileInfo.Index = nextstart
		m.FileInfo.Data = buf

		fd, err := json.Marshal(m)
		if err != nil {
			utils.Log.Error().Msgf("Marshal err:%s", err.Error())
			return
		}
		TransferTaskManger.area.SendP2pReplyMsg(message, &fd)

		//更新推送日志
		task.Rate = int64(float64(nextstart) / float64(size) * float64(100))
		//utils.Log.Info().Msgf("推送的百分比：%d%%", task.Rate)
		task.Index = nextstart

		if nextstart >= size {
			utils.Log.Info().Msgf("PushTaskID: %d 文件推送完成", m.PushTaskID)
			task.pushTaskRecordDel()
			this.pushFinishSignalMonitor <- m
			return
		}
		task.pushTaskRecordSave()
		return
	} else {
		utils.Log.Error().Msgf("P2p消息解析失败：%s", err.Error())
		TransferTaskManger.area.SendP2pReplyMsg(message, nil)
	}
	return
}

func (t *PullTask) pullTaskDel() utils.ERROR {
	ERR := t.pullTaskRecordDel()
	if ERR.CheckFail() {
		return ERR
	}
	//删除临时文件
	tmpPath := filepath.Join(filepath.Dir(t.FileInfo.Path), t.FileInfo.Name+"_"+t.FileInfo.From.B58String()+"_"+
		strconv.FormatUint(t.PullTaskID, 10)+"_tmp")
	ok, err := utils.PathExists(tmpPath)
	if err != nil {
		utils.Log.Error().Msgf("删除临时文件失败：%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if ok {
		err = os.Remove(tmpPath)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
	}
	return utils.NewErrorSuccess()
}

func (t *PullTask) pullTaskRecordOne(pullTaskID uint64) (*PullTask, utils.ERROR) {
	//先在内存中拿
	if v, ok := TransferTaskManger.TaskMap.Load(pullTaskID); ok {
		task := v.(*PullTask)
		return task, utils.NewErrorSuccess()
	}
	//拿不到就去db里拿
	taskList, ERR := t.PullTaskRecordList()
	if ERR.CheckFail() {
		return nil, ERR
	}
	for _, v := range taskList {
		if v.PullTaskID == pullTaskID {
			TransferTaskManger.TaskMap.Store(pullTaskID, v)
			return v, utils.NewErrorSuccess()
		}
	}
	return nil, utils.NewErrorBus(ERROR_CODE_task_not_exist, "") //errors.New("任务不存在")
}

func (t *PullTask) pullTaskRecordDel() utils.ERROR {
	taskList, ERR := t.PullTaskRecordList()
	if ERR.CheckFail() {
		return ERR
	}

	TransferTaskManger.lockPullTaskList.Lock()
	defer TransferTaskManger.lockPullTaskList.Unlock()
	TransferTaskManger.TaskMap.Delete(t.PullTaskID)

	for k, v := range taskList {
		if v.PullTaskID == t.PullTaskID {
			taskList = append(taskList[:k], taskList[k+1:]...)
			break
		}
	}

	fd, err := json.Marshal(taskList)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	newKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(Transferlog_pull_task_db_key))
	if !ERR.CheckSuccess() {
		return ERR
	}
	ERR = TransferTaskManger.levelDB.Save(*newKey, &fd)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

func (t *PullTask) pullTaskRecordSave() utils.ERROR {
	taskList, _ := t.PullTaskRecordList()

	TransferTaskManger.lockPullTaskList.Lock()
	defer TransferTaskManger.lockPullTaskList.Unlock()
	TransferTaskManger.TaskMap.Store(t.PullTaskID, t)

	have := false
	for k, v := range taskList {
		if v.PullTaskID == t.PullTaskID {
			taskList[k] = t
			have = true
			break
		}
	}
	if !have {
		taskList = append(taskList, t)
	}

	fd, err := json.Marshal(taskList)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	newKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(Transferlog_pull_task_db_key))
	if !ERR.CheckSuccess() {
		return ERR
	}
	ERR = TransferTaskManger.levelDB.Save(*newKey, &fd)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}
