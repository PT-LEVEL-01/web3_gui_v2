package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"web3_gui/keystore/v1/base58"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center"
	"web3_gui/utils"
)

const (
	recfilepath   = "files"
	filerecording = "filerecording.log"
)

func (this *TestPeer) FindFileMsg(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//content := message.Body.Content
	//utils.Log.Info().Msgf("接收到的消息:%s", content)

	content := *message.Body.Content
	m, err := ParseMsg(content)
	if err == nil {
		filerecordingMap := ReadFilerecording()
		value, ok := filerecordingMap[m.Hash]

		if !ok {
			fd, _ := json.Marshal(&FileInfo{})
			utils.Log.Info().Msgf("没有找到文件:%s", m.Path)
			this.area.SendP2pReplyMsg(message, msg_id_searchSuper_recv, &fd)
			return
		}

		value.Path = m.Path
		value.Name = m.Name
		fd, err := json.Marshal(value)
		if err != nil {
			utils.Log.Error().Msgf("json.Marshal失败:%s", err.Error())
			this.area.SendP2pReplyMsg(message, msg_id_searchSuper_recv, nil)
			return
		}
		this.area.SendP2pReplyMsg(message, msg_id_searchSuper_recv, &fd)
		return

	}
	this.area.SendP2pReplyMsg(message, msg_id_searchSuper_recv, nil)
}

/*
接收发送的文件消息
*/
func (this *TestPeer) FileMsg(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	var bs []byte
	//	var rev *FileInfo
	//	fmt.Println("这个文本消息是自己的")
	//发送给自己的，自己处理
	content := *message.Body.Content
	m, err := ParseMsg(content)
	if err == nil {
		//发送者
		sendId := message.Head.Sender.B58String()
		m.Path = filepath.Join(recfilepath, m.Name)
		num := 0
	Rename:
		//如果文件存在，则重命名为新的文件
		if ok, _ := utils.PathExists(m.Path); ok {
			num++
			filenamebase := filepath.Base(m.Name)
			fileext := filepath.Ext(m.Name)
			filename := strings.TrimSuffix(filenamebase, fileext)
			newname := filename + "_" + strconv.Itoa(num) + fileext
			m.Path = filepath.Join(recfilepath, newname)
			if ok1, _ := utils.PathExists(m.Path); ok1 {
				goto Rename
			}
			m.Name = newname
		}

		utils.CheckCreateDir(recfilepath)

		//临时文件，传输完成后改为原来文件名
		tmpPath := filepath.Join(recfilepath, m.Name+"_"+sendId+"_tmp")
		fi, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			utils.Log.Error().Msgf("临时文件打开失败:%s", err.Error())
			this.area.SendP2pReplyMsg(message, msg_id_p2p_recv, &bs)
			return
		}
		start := m.Index - Lenth
		if start >= m.Size {
			start = m.Size
		}

		//刚开始传时新建一个和原文件大小相同的空文件来占位所需空间
		if start == 0 {
			err = fi.Truncate(m.Size)
			if err != nil {
				utils.Log.Error().Msgf("空文件新建失败:%s", err.Error())
				this.area.SendP2pReplyMsg(message, msg_id_p2p_recv, &bs)
				return
			}
		}

		fi.Seek(start, 0)
		fi.Write(m.Data)

		defer fi.Close()
		utils.Log.Info().Msgf("发送者：%v 接收的start：%d", sendId, start)

		rate := int64(float64(m.Index) / float64(m.Size) * float64(100))
		utils.Log.Info().Msgf("接收的百分比：%d%%", rate)
		m.SetSpeed(time.Now().Unix(), len(content))
		speed := m.GetSpeed()
		utils.Log.Info().Msgf("接收的速率：%d", speed)

		//传输完成，则更新状态
		if rate >= 100 {
			//传输完成，则重命名文件名
			fi.Close()
			os.Rename(tmpPath, m.Path)

			err = checkHash(m)
			if err != nil {
				DelFilerecording(m)
				os.Remove(m.Path)
				utils.Log.Error().Msgf(err.Error())
				this.area.SendP2pReplyMsg(message, msg_id_p2p_recv, &bs)
				return
			}
		}
		err = SaveFilerecording(m)
		if err != nil {
			utils.Log.Error().Msgf(err.Error())
			this.area.SendP2pReplyMsg(message, msg_id_p2p_recv, &bs)
			return
		}
		bs = utils.Int64ToBytes(start)
		utils.Log.Info().Msgf("文件完成：%d %d%%", start, rate)
	}
	//回复发送者，自己已经收到消息ID

	this.area.SendP2pReplyMsg(message, msg_id_p2p_recv, &bs)
}

func checkHash(fileInfo *FileInfo) (err error) {
	hasB, err := utils.FileSHA3_256(fileInfo.Path)
	if err != nil {
		utils.Log.Error().Msgf("文件hash失败:%s", err.Error())
		return err
	}
	fHash := base58.Decode(fileInfo.Hash)
	if !bytes.Equal(hasB, fHash) {
		return errors.New("文件上传错误！不完整或已被破坏")
	}
	return err
}

func ReadFilerecording() (filerecordingMap map[string]*FileInfo) {
	bs, err := ioutil.ReadFile(filerecording)
	if err != nil {
		utils.Log.Info().Msgf("读取filerecording失败:%s", err.Error())
		return
	}
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&filerecordingMap)
	if err != nil {
		utils.Log.Info().Msgf("解析filerecording失败:%s", err.Error())
		return
	}
	return
}

func SaveFilerecording(fileInfo *FileInfo) (err error) {
	filerecordingMap := ReadFilerecording()

	if filerecordingMap == nil {
		filerecordingMap = make(map[string]*FileInfo)
	}
	fileInfo.Data = nil
	filerecordingMap[fileInfo.Hash] = fileInfo
	bs, err := json.Marshal(filerecordingMap)
	if err != nil {
		utils.Log.Info().Msgf("解析filerecording失败:%s", err.Error())
		return err
	}
	err = utils.SaveFile(filerecording, &bs)
	return
}

func DelFilerecording(fileInfo *FileInfo) (err error) {
	filerecordingMap := ReadFilerecording()
	if filerecordingMap != nil {
		delete(filerecordingMap, fileInfo.Hash)
		bs, err := json.Marshal(filerecordingMap)
		if err != nil {
			utils.Log.Info().Msgf("解析filerecording失败:%s", err.Error())
			return err
		}
		err = utils.SaveFile(filerecording, &bs)
		return err
	}
	return
}

// 解析消息
func ParseMsg(d []byte) (*FileInfo, error) {
	msg := &FileInfo{}
	// err := json.Unmarshal(d, msg)
	decoder := json.NewDecoder(bytes.NewBuffer(d))
	decoder.UseNumber()
	err := decoder.Decode(msg)
	if err != nil {
		fmt.Println(err)
	}
	return msg, err
}
