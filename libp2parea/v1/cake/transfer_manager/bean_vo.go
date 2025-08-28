package transfer_manager

import (
	"encoding/hex"
)

type FileinfoVO struct {
	PushTaskID uint64           //
	PullTaskID uint64           //
	From       string           //传输者
	To         string           //接收者
	Name       string           //原始文件名
	Hash       string           //
	Size       int64            //
	Path       string           //
	Index      int64            //
	Data       string           //
	Speed      map[string]int64 //传输速度统计
	Rate       int64            //
	Status     string           //ture 为传输中 false为暂停中
}

/*
转换文件下载信息
*/
func ConverTransferMsgVO(tm *TransferMsg) FileinfoVO {
	return FileinfoVO{
		PushTaskID: tm.PushTaskID,
		PullTaskID: tm.PullTaskID,
		From:       tm.FileInfo.From.B58String(), //传输者
		To:         tm.FileInfo.To.B58String(),   //接收者
		Name:       tm.FileInfo.Name,             //原始文件名
		Hash:       hex.EncodeToString(tm.FileInfo.Hash),
		Size:       tm.FileInfo.Size,
		Path:       tm.FileInfo.Path,
		Index:      tm.FileInfo.Index,
		// Data       :hex
		Speed: tm.FileInfo.Speed, //传输速度统计
		Rate:  tm.FileInfo.Rate,
	}
}

/*
转换文件下载信息
*/
func ConverPullTaskVO(tm *PullTask) FileinfoVO {
	return FileinfoVO{
		PushTaskID: tm.PushTaskID,
		PullTaskID: tm.PullTaskID,
		From:       tm.FileInfo.From.B58String(), //传输者
		To:         tm.FileInfo.To.B58String(),   //接收者
		Name:       tm.FileInfo.Name,             //原始文件名
		Hash:       hex.EncodeToString(tm.FileInfo.Hash),
		Size:       tm.FileInfo.Size,
		Path:       tm.FileInfo.Path,
		Index:      tm.FileInfo.Index,
		// Data       :hex
		Speed:  tm.FileInfo.Speed, //传输速度统计
		Rate:   tm.FileInfo.Rate,
		Status: tm.Status,
	}
}
