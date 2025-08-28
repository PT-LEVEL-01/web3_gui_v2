package update_version

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/message_center"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
新版本文件下载到本地成功后的回调方法
@newFilePath    string    新版本文件全路径
*/
type NewVersionFileCallBack func(newFilePath string)

type UpdateVersion struct {
	area            *libp2parea.Node
	RootNetAddr     nodeStore.AddressNet `json:"to"` //接收者
	Platform        string
	lockPull        *sync.Mutex
	versionFilePull *VersionFilePull
	Callback        NewVersionFileCallBack //新版本文件更新成功的回调方法
}

type VersionFilePull struct {
	Name  string `json:"name"` //原始文件名
	Size  int64  `json:"size"`
	Path  string `json:"path"`
	Index int64  `json:"index"`
	Data  []byte `json:"data"`
	Rate  int64  `json:"rate"`
}

var msg_id_p2p_find_last_version uint64 = 1001           //请求查询最新版本
var msg_id_p2p_find_last_version_recv uint64 = 1002      //请求查询最新版本回复
var msg_id_p2p_pull_version_file uint64 = 1003           //传输文件
var msg_id_p2p_pull_version_file_recv uint64 = 1004      //传输文件回复
var msg_id_p2p_origin_version_library uint64 = 1005      //远程版本文件列表
var msg_id_p2p_origin_version_library_recv uint64 = 1006 //远程版本文件列表回复

func NewUpdateVersion(area *libp2parea.Node, rootNetAddr nodeStore.AddressNet, platform string, find_last_version, find_last_version_recv, pull_version_file, pull_version_file_recv, origin_version_library, origin_version_library_recv uint64) *UpdateVersion {
	msg_id_p2p_find_last_version = find_last_version
	msg_id_p2p_find_last_version_recv = find_last_version_recv
	msg_id_p2p_pull_version_file = pull_version_file
	msg_id_p2p_pull_version_file_recv = pull_version_file_recv
	msg_id_p2p_origin_version_library = origin_version_library
	msg_id_p2p_origin_version_library_recv = origin_version_library_recv

	updateVersion := &UpdateVersion{
		area:        area,
		RootNetAddr: rootNetAddr,
		Platform:    platform,
		lockPull:    new(sync.Mutex),
	}

	area.Register_p2p(msg_id_p2p_find_last_version, updateVersion.RecvFindLastVersion)
	//area.Register_p2p(msg_id_p2p_find_last_version_recv, updateVersion.RecvFindLastVersion_recv)

	area.Register_p2p(msg_id_p2p_pull_version_file, updateVersion.FileSlicePush)
	//area.Register_p2p(msg_id_p2p_pull_version_file_recv, updateVersion.FileSlicePush_recv)

	area.Register_p2p(msg_id_p2p_origin_version_library, updateVersion.RecvOriginVersionLibrary)
	//area.Register_p2p(msg_id_p2p_origin_version_library_recv, updateVersion.RecvOriginVersionLibrary_recv)

	return updateVersion
}

/*
TickerUpdateVersion
@Description: 定时检查并更新
@receiver this
*/
func (this *UpdateVersion) TickerUpdateVersion() {
	utils.Go(func() {
		var timer = time.NewTicker(Update_version_expiration_interval)
		defer timer.Stop()
		firstTimer := true
		for {
			if !firstTimer {
				<-timer.C
			}
			firstTimer = false
			//utils.Log.Info().Msgf("定时任务执行！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！！")
			if bytes.Equal(this.area.GetNetId().GetAddr(), this.RootNetAddr.GetAddr()) {
				continue
			}
			latest, latestName, _, _, _, err := this.CheckLatestVersion()
			if err != nil {
				//utils.Log.Error().Msgf("定时更新错误:%s rootNetAddr:%s", err.Error(), this.RootNetAddr.B58String())
				continue
			}
			if !latest && len(latestName) > 0 {
				err = this.GetVersionFile(latestName)
				if err != nil {
					utils.Log.Error().Msgf("定时更新获取文件错误:%s rootNetAddr:%s", err.Error(), this.RootNetAddr.B58String())
					continue
				}
				//utils.Log.Info().Msgf("定时更新文件成功！")
			}
			//utils.Log.Info().Msgf("定时更新成功！")
		}
	}, nil)
}

/*
设置新版本文件下载到本地成功后的回调方法
@callback    NewVersionFileCallBack    回调函数
*/
func (this *UpdateVersion) SetCallback(callback NewVersionFileCallBack) {
	this.Callback = callback
}

func (this *UpdateVersion) SetPlatform(p string) {
	this.Platform = p
}

func (this *UpdateVersion) GetPlatform() string {
	return this.Platform
}

func (this *UpdateVersion) SetRootNetAddr(addr string) {
	this.RootNetAddr = nodeStore.AddressFromB58String(addr)
}

func (this *UpdateVersion) GetRootNetAddr() string {
	return this.RootNetAddr.B58String()
}

/*
CheckLatestVersion
@Description: 检查最新版本
@receiver this
@return isLatest        是否是最新
@return latestFileName  最新版本文件名
@return versionName     最新版本名称
@return localIndex  int 本地最新版本编号
@return remoteIndex int 远程最新版本编号
@return err
*/
func (this *UpdateVersion) CheckLatestVersion() (isLatest bool, latestFileName, versionName string, localIndex, remoteIndex int, err error) {
	//去root查找最新版本
	if this.Platform == "" {
		err = errors.New("Platform is empty")
		return
	}

	if len(this.RootNetAddr.GetAddr()) == 0 {
		err = errors.New("RootNetAddr is empty")
		return
	}

	platform := []byte(this.Platform)

	message, ERR := this.area.SendP2pMsgWaitRequest(msg_id_p2p_find_last_version, &this.RootNetAddr, &platform, P2p_mgs_timeout)
	if ERR.CheckFail() {
		return
	} else {
		if message != nil {
			//查找本地最新版本
			latestFileName, versionName, localIndex, err = this.LocalLatestVersion()
			if err != nil {
				return
			}
			origin := string(*message)
			//与root最新版本对比是否是最新
			if latestFileName == origin {
				return true, latestFileName, versionName, localIndex, localIndex, nil
			}
			versionName, remoteIndex = extractDataFromName(origin)
			return false, origin, versionName, localIndex, remoteIndex, nil
		}
		return false, "", versionName, localIndex, localIndex, errors.New("获取远端最新版本失败")
	}
	return false, "", versionName, localIndex, localIndex, errors.New("发送P2p消息失败")
}

/*
		LocalLatestVersion
		@Description: 获取本地最近版本文件名
		@receiver this
		@return fileName 最新版本文件名
	 	@return code 最新版本编号
		@return err
*/
func (this *UpdateVersion) LocalLatestVersion() (fileName, code string, index int, err error) {
	fileName, err = FindLatestVersionFileName(Version_dir)
	if err != nil || fileName == "" {
		return "", "", 0, err
	}
	code, index = extractDataFromName(fileName)
	return
}

/*
OriginVersionLibrary
@Description: 获取远端版本文件列表
@receiver this
@return []string
@return error
*/
func (this *UpdateVersion) OriginVersionLibrary() ([]string, error) {
	//去root查找最新版本
	if this.Platform == "" {
		return nil, errors.New("Platform is empty")
	}

	if this.RootNetAddr.GetAddr() == nil || len(this.RootNetAddr.GetAddr()) == 0 {
		return nil, errors.New("RootNetAddr is empty")
	}

	platform := []byte(this.Platform)

	message, ERR := this.area.SendP2pMsgWaitRequest(msg_id_p2p_origin_version_library, &this.RootNetAddr, &platform, P2p_mgs_timeout)
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", ERR.String())
		return nil, errors.New(ERR.String())
	} else {
		if message != nil {
			var list []string
			err := json.Unmarshal(*message, &list)
			return list, err
		}
		return nil, errors.New("获取远端版本文件列表失败")
	}
	return nil, errors.New("发送P2p消息失败")
}

/*
GetVersionFile
@Description: 获取版本文件
@receiver this
@param fileName
@return error
*/
func (this *UpdateVersion) GetVersionFile(fileName string) error {
	utils.Log.Info().Msgf("获取版本文件 %s", fileName)
	this.lockPull.Lock()
	defer this.lockPull.Unlock()
	defer func() {
		this.versionFilePull = nil
		//utils.Log.Info().Msgf("this.versionFilePull = nil")
	}()
	filePath := filepath.Join(this.Platform, fileName)
	//检查本地文件是否已经存在
	exist, err := utils.PathExists(filePath)
	if err != nil {
		return err
	}
	utils.Log.Info().Msgf("检查本地文件是否已经存在:%s %t", filePath, exist)
	if exist {
		return nil
	}

	//开始拉取文件
	this.versionFilePull = &VersionFilePull{
		Name: fileName,
		Path: filePath,
	}
	//utils.Log.Info().Msgf("开始拉取文件 %s", fileName)
	var errnum int
	for {
		okf, err := this.fileSlicePull()
		if okf == true { //已传输完，则退出
			break
		}
		if err != nil {
			//开始重传
			errnum++
			if errnum <= ErrNum {
				utils.Log.Info().Msgf("resend slice...")
				continue
			}
			return err
		}
	}
	if this.Callback == nil {
		return nil
	}
	this.Callback(filepath.Join(Version_dir, fileName))
	return nil
}

//分段传输，续传
/**
@return okf 是否传送完 err 错误
*/
func (this *UpdateVersion) fileSlicePull() (okf bool, err error) {
	//已经传完
	if this.versionFilePull.Index >= this.versionFilePull.Size && this.versionFilePull.Size > 0 {
		return true, nil
	}

	fd, err := json.Marshal(this.versionFilePull)
	if err != nil {
		return false, err
	}

	message, ERR := this.area.SendP2pMsgWaitRequest(msg_id_p2p_pull_version_file, &this.RootNetAddr, &fd, P2p_mgs_timeout)
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", ERR.String())
		return false, errors.New(ERR.String())
	} else {
		if message != nil {
			//发送成功，对方已经接收到消息
			m, err := parseVersionFilePullMsg(*message)
			if err != nil {
				utils.Log.Error().Msgf("P2p消息解析失败:%s", err.Error())
				return false, err
			}

			tmpPath := filepath.Join(Version_dir, this.versionFilePull.Name+TempFileSuffix)
			utils.CheckCreateDir(Version_dir)
			if this.versionFilePull.Index == 0 {
				//删除旧临时文件
				ok, err := utils.PathExists(tmpPath)
				if err != nil {
					utils.Log.Error().Msgf("删除旧临时文件失败", err)
					return false, err
				}
				if ok {
					err = os.Remove(tmpPath)
					if err != nil {
						return false, err
					}
				}
			}

			fi, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				utils.Log.Error().Msgf("临时文件打开失败:%s", err.Error())
				return false, err
			}

			if this.versionFilePull.Index == 0 && m.Size > 0 {
				err = fi.Truncate(m.Size)
				if err != nil {
					utils.Log.Error().Msgf("空文件新建失败:%s", err.Error())

					return false, errors.New(err.Error())
				}
			}

			fi.Seek(this.versionFilePull.Index, 0)
			fi.Write(m.Data)
			defer fi.Close()

			this.versionFilePull.Index = m.Index

			//utils.Log.Info().Msgf("接收的start：%d", this.versionFilePull.FileInfo.Index)
			this.versionFilePull.Rate = int64(float64(m.Index) / float64(m.Size) * float64(100))
			//utils.Log.Info().Msgf("接收的百分比：%d%%", this.versionFilePull.Rate)

			//传输完成，则更新状态
			if this.versionFilePull.Rate >= 100 {
				fi.Close()
				os.Rename(tmpPath, filepath.Join(Version_dir, this.versionFilePull.Name))
				//utils.Log.Info().Msgf("文件完成：%d %d%%", this.versionFilePull.Index, this.versionFilePull.Rate)
				// 发送完成
				return true, nil
			}
			return false, nil
		} else {
			utils.Log.Error().Msgf("文件拉取失败")
			return false, errors.New("文件拉取失败")
		}
	}
}

// 分片推送
func (this *UpdateVersion) FileSlicePush(message *message_center.MessageBase) {
	content := message.Content
	m, err := parseVersionFilePullMsg(content)
	if err == nil {
		index := m.Index //当前已传偏移量
		fi, err := os.Open(path.Join(Root_version_library, m.Path))
		if err != nil {
			utils.Log.Error().Msgf("文件打开失败：%s", err.Error())
			this.area.SendP2pReplyMsg(message, nil)
			return
		}
		defer fi.Close()
		stat, err := fi.Stat()
		if err != nil {
			utils.Log.Error().Msgf("文件打开失败：%s", err.Error())
			this.area.SendP2pReplyMsg(message, nil)
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
			utils.Log.Error().Msgf("读取文件流失败：%s", err.Error())
			this.area.SendP2pReplyMsg(message, nil)
			return
		}
		//下一次start位置
		nextstart := start + Lenth
		if nextstart > size {
			nextstart = size
		}
		m.Size = size
		m.Index = nextstart
		m.Data = buf

		fd, err := json.Marshal(m)
		if err != nil {
			utils.Log.Error().Msgf("Marshal err:%s", err.Error())
			return
		}
		this.area.SendP2pReplyMsg(message, &fd)

		//更新推送日志
		//rate := int64(float64(nextstart) / float64(size) * float64(100))
		//utils.Log.Info().Msgf("推送的百分比：%d%%", rate)
		//if nextstart >= size {
		//	utils.Log.Info().Msgf("文件推送完成 %s", m.Name)
		//}
		return
	} else {
		utils.Log.Error().Msgf("P2p消息解析失败：%s", err.Error())
		this.area.SendP2pReplyMsg(message, nil)
	}
	return
}

func (this *UpdateVersion) RecvFindLastVersion(message *message_center.MessageBase) {
	content := message.Content
	if content != nil || len(content) > 0 {
		fileName, err := FindLatestVersionFileName(path.Join(Root_version_library, string(content)))

		if err != nil {
			utils.Log.Error().Msgf("获取最新版本失败:%s  sendId:%s", err.Error(), message.SenderAddr.B58String())
		}
		recv := []byte(fileName)
		this.area.SendP2pReplyMsg(message, &recv)
		return
	}
	utils.Log.Error().Msgf("获取platform失败 sendId:%s", message.SenderAddr.B58String())
	this.area.SendP2pReplyMsg(message, nil)
	return
}

func (this *UpdateVersion) RecvOriginVersionLibrary(message *message_center.MessageBase) {
	content := message.Content
	if content != nil || len(content) > 0 {
		fileNames, err := scanning(path.Join(Root_version_library, string(content)))

		if err != nil {
			utils.Log.Error().Msgf("获取文件列表失败:%s  sendId:%s", err.Error(), message.SenderAddr.B58String())
		}

		b, err := json.Marshal(fileNames)
		if err != nil {
			utils.Log.Error().Msgf("Marshal err:%s", err.Error())
			return
		}
		this.area.SendP2pReplyMsg(message, &b)
		return
	}
	utils.Log.Error().Msgf("获取platform失败 sendId:%s", message.SenderAddr.B58String())
	this.area.SendP2pReplyMsg(message, nil)
	return
}

//func (this *UpdateVersion) RecvFindLastVersion_recv(message *message_center.MessageBase) {
//	// utils.Log.Info().Msgf("收到P2P消息返回 from:%s", message.Head.Sender.B58String())
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}
//func (this *UpdateVersion) FileSlicePush_recv(message *message_center.MessageBase) {
//	// utils.Log.Info().Msgf("收到SearchSuper消息返回 from:%s", message.Head.Sender.B58String())
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}
//func (this *UpdateVersion) RecvOriginVersionLibrary_recv(message *message_center.MessageBase) {
//	// utils.Log.Info().Msgf("收到SearchSuper消息返回 from:%s", message.Head.Sender.B58String())
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
scanning
@Description: 扫描文件
@param src
@return fileNames
@return err
*/
func scanning(src string) (fileNames []string, err error) {
	utils.CheckCreateDir(src)

	err = filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		//跳过子目录
		if info.IsDir() && path != src {
			return filepath.SkipDir
		}

		//排除文件夹 和 tmp临时文件
		if !info.IsDir() && !strings.HasSuffix(info.Name(), TempFileSuffix) {
			fileNames = append(fileNames, info.Name())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(fileNames) > 0 {
		sort.Slice(fileNames, func(i, j int) bool {
			_, index := extractDataFromName(fileNames[i])
			_, index1 := extractDataFromName(fileNames[j])
			return index < index1
		})
		return fileNames, nil
	}
	return
}

/*
findLatestVersionFileName
@Description: 查找最新版本文件名
@param src
@return string
@return error
*/
func FindLatestVersionFileName(src string) (string, error) {
	fileNames, err := scanning(src)
	if err != nil {
		return "", err
	}
	if len(fileNames) > 0 {
		return fileNames[len(fileNames)-1], nil
	}
	return "", nil
}

/*
extractDataFromName
@Description: 截取文件名中版本号用于排序，如文件名为 XXXXX_v3.0.1_1.exe 或 XXXX_v3.0.1_1 或 XXXX.XX_v3.0.1_1 都取得 code:v3.0.1 index:1
@param fileName
@return code 版本编号 用于显示用的
@return index 版本号序列号 用于排序
*/
func extractDataFromName(fileName string) (code string, index int) {
	lastUnderscoreIndex := strings.LastIndex(fileName, "_")
	if lastUnderscoreIndex == -1 {
		return
	}
	lastDotIndex := strings.LastIndex(fileName, ".")
	if lastDotIndex == -1 || lastDotIndex < lastUnderscoreIndex {
		lastDotIndex = len(fileName)
	}

	if lastDotIndex == 0 {
		return
	}

	dataString := fileName[lastUnderscoreIndex+1 : lastDotIndex]
	lastUnderscoreIndex1 := strings.LastIndex(fileName[:lastUnderscoreIndex], "_")
	if lastUnderscoreIndex1 == -1 {
		return
	}
	code = fileName[lastUnderscoreIndex1+1 : lastUnderscoreIndex]
	index, _ = strconv.Atoi(dataString)
	return
}

// 解析消息
func parseVersionFilePullMsg(d []byte) (*VersionFilePull, error) {
	msg := &VersionFilePull{}
	decoder := json.NewDecoder(bytes.NewBuffer(d))
	decoder.UseNumber()
	err := decoder.Decode(msg)
	if err != nil {
		fmt.Println(err)
	}
	return msg, err
}
