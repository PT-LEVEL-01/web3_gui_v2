/*
此文件负责存储功能中文件上传和下载的状态记录
*/
package storage

import (
	"time"
	"web3_gui/config"
	"web3_gui/utils"
)

/*
保存下载任务
@plainHash    []byte    文件加密后的hash
@pwd          []byte    密码
*/
func StorageClient_SaveDownloadTask(downloadStep []DownloadStep) ([][]byte, utils.ERROR) {
	bss := make([][]byte, 0)
	for _, one := range downloadStep {
		bs, err := one.Proto()
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		bss = append(bss, *bs)
	}

	ids, ERR := config.LevelDB.SaveListMore(*config.DBKEY_storage_client_downloading_list, bss, nil)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	return ids, ERR
}

/*
移动下载任务到完成列表
@id    []byte    数据库id
*/
func StorageClient_MoveDownloadTaskFinish(downloadStep *DownloadStep) utils.ERROR {
	downloadStep.finishTime = time.Now()
	bs, err := downloadStep.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//batch := new(leveldb.Batch)
	//删除旧列表
	ERR := config.LevelDB.RemoveListByIndex(*config.DBKEY_storage_client_downloading_list, downloadStep.dbId, nil)
	if ERR.CheckFail() {
		return ERR
	}
	//保存到新列表
	id, ERR := config.LevelDB.SaveList(*config.DBKEY_storage_client_download_finish_list, *bs, nil)
	if ERR.CheckFail() {
		return ERR
	}
	downloadStep.dbId = id
	return utils.NewErrorSuccess()
}

/*
查询下载任务列表
*/
func StorageClient_LoadDownloadTask() ([]DownloadStep, utils.ERROR) {
	items, err := config.LevelDB.FindListAll(*config.DBKEY_storage_client_downloading_list)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	dss := make([]DownloadStep, 0, len(items))
	for _, item := range items {
		ds, err := ParseDownloadStep(item.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		dss = append(dss, *ds)
	}
	return dss, utils.NewErrorSuccess()
}

/*
查询下载完成列表
@id    []byte    数据库id
*/
func StorageClient_LoadDownloadTaskFinish(startIndex []byte, limit uint64) ([]DownloadStep, utils.ERROR) {
	//config.LevelDB.FindListTotal(config.DBKEY_storage_client_download_finish_list)
	items, ERR := config.LevelDB.FindListRange(*config.DBKEY_storage_client_download_finish_list, startIndex, limit, false)
	if ERR.CheckFail() {
		return nil, ERR
	}
	dss := make([]DownloadStep, 0, len(items))
	for _, item := range items {
		ds, err := ParseDownloadStep(item.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		ds.dbId = item.Index
		dss = append(dss, *ds)
	}
	return dss, utils.NewErrorSuccess()
}
