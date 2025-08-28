package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
	"web3_gui/keystore/v1/base58"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

const (
	//Count  int64 = 20 * 1024 * 1024 * 1024 //单个好友传输总量不超过20G
	Lenth  int64 = 200 * 1024 //每次传输大小（200kb）
	ErrNum int   = 5          //传输失败重试次数 5次
	Second int64 = 10         //传输速度统计时间间隔 10秒
)

type FileInfo struct {
	Name  string //原始文件名
	Hash  string
	Size  int64
	Path  string
	Index int64
	Data  []byte
	Speed map[string]int64 //传输速度统计
}

func NewFileInfo(path string) *FileInfo {
	return &FileInfo{Path: path}
}

func (f *FileInfo) FindFile(area *libp2parea.Area, toArea nodeStore.AddressNet) error {
	hasB, err := utils.FileSHA3_256(f.Path)
	if err != nil {
		utils.Log.Error().Msgf("文件hash失败:%s", err.Error())
		return err
	}
	f.Hash = string(base58.Encode(hasB))
	fd, err := json.Marshal(f)
	if err != nil {
		utils.Log.Error().Msgf("json.Marshal失败:%s", err.Error())
		return err
	}
	message, ok, _, err := area.SendP2pMsgWaitRequest(msg_id_searchSuper, &toArea, &fd, time.Second*100)
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

			f.Index = m.Index
			f.Speed = m.Speed
			f.Size = m.Size
		} else {
			utils.Log.Error().Msgf("P2p消息解析失败")
			return errors.New("P2p消息解析失败")
		}
	} else {
		utils.Log.Error().Msgf("发送P2p消息失败")
		return errors.New("发送P2p消息失败")
	}
	return nil
}

//分段传输，续传
/**
@param id 消息ID ，i 分段序号
@return fd 段数据 fileinfo 文件属性ok 是否传送完 err 错误
*/
func (f *FileInfo) ReadFileSlice() (fd []byte, fileinfo FileInfo, ok bool, errs error) {

	// TODO 判断是否是续传还要从远程去读取对方的文件信息
	//已经传完
	if f.Index >= f.Size && f.Size > 0 {
		return
	}
	index := f.Index //当前已传偏移量
	fi, err := os.Open(f.Path)
	if err != nil {
		fmt.Println(err)
		errs = err
		return
	}

	stat, err := fi.Stat()
	if err != nil {
		fmt.Println(err)
		errs = err
		return
	}
	size := stat.Size()
	start := index
	length := Lenth
	//如果偏移量小于文件大小，并且剩余大小小于长度，则长度为剩余大小(即最后一段)
	if start < size && size-index < Lenth {
		length = size - index
	}
	buf := make([]byte, length)
	_, err = fi.ReadAt(buf, start)
	if err != nil {
		fmt.Println(err)
		errs = err
		return
	}
	//下一次start位置
	nextstart := start + Lenth
	if nextstart >= size {
		ok = true
	}
	fmt.Println("文件发送中...", size, nextstart)
	f.Name = stat.Name()
	f.Size = size
	f.Index = nextstart
	f.Data = buf
	fmt.Printf("xxx%+v", f.Index)

	fd, err = json.Marshal(f)
	if err != nil {
		fmt.Println(err)
	}
	return
}

/*func (f *FileInfo) SendFile(area *libp2parea.Area, toArea nodeStore.AddressNet) (bl bool) {
	content := []byte("Nice to meet you!!!")
	_, _, _, err := area.SendP2pMsgWaitRequest(msg_id_p2p, &toArea, &content, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return false
	} else {
		utils.Log.Info().Msgf("发送成功")
	}
	return true
}*/

// 发送文件消息
func (f *FileInfo) SendFile(area *libp2parea.Area, toArea nodeStore.AddressNet) (bl bool) {
	utils.Log.Info().Msgf("SendFile---- cur:%s, to:%s", area.NodeManager.NodeSelf.IdInfo.Id.B58String(), toArea.B58String())
	var errnum int
	for i := int64(0); ; i++ {
	BEGIN:
		bs, _, okf, err := f.ReadFileSlice()
		if bs == nil { //已传输完，则退出
			break
		}
		if err != nil {
			//开始重传
			errnum++
			if errnum <= ErrNum {
				fmt.Println("resend slice...")
				goto BEGIN
			}
			break
		}

		message, ok, _, err := area.SendP2pMsgWaitRequest(msg_id_p2p, &toArea, &bs, time.Second*100)
		if err != nil {
			utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
			return false
		}
		if ok {
			if message != nil {
				//发送成功，对方已经接收到消息
				rate := int64(float64(f.Index) / float64(f.Size) * float64(100))
				utils.Log.Info().Msgf("发送的百分比：%d%%", rate)
				//fmt.Println("发送的百分比：", rate)
				f.SetSpeed(time.Now().Unix(), len(bs))
				speed := f.GetSpeed()
				fmt.Println("发送的速率：", speed)
				// TODO 更新日志
				if okf {
					bl = true
					break
				}
			} else {
				//发送失败，接收返回消息超时
				fmt.Println("fail")
				errnum++
				if errnum <= ErrNum {
					//开始重传
					fmt.Println("resend...")
					goto BEGIN
				}
				bl = false
				break
			}

		} else {
			bl = false
			break
		}
	}
	return
}

// 采集速度参数
func (f *FileInfo) SetSpeed(stime int64, size int) error {
	if f.Speed == nil {
		f.Speed = make(map[string]int64, 0)
	}

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
	return f.Speed["size"] / t * 100
}
