package db

import (
	"errors"
	"time"
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/im/protos/go_protos"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
添加新闻分类
@className    string    分类名称
*/
func AddClass(dbListType utilsleveldb.LeveldbKey, className string) error {
	classNameKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(className))
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	_, ERR = config.LevelDB.SaveMapInList(dbListType, *classNameKey, nil, nil)
	if !ERR.CheckSuccess() {
		return errors.New(ERR.Msg)
	}
	return nil
}

/*
获取所有新闻分类
@className    string    分类名称
*/
func GetClass(dbListType utilsleveldb.LeveldbKey) ([]string, utils.ERROR) {
	items, ERR := config.LevelDB.FindMapInListAll(dbListType)
	if ERR.CheckFail() {
		return nil, ERR
	}
	names := make([]string, 0)
	for _, one := range items {
		baseKey, ERR := one.Key.BaseKey()
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		names = append(names, string(baseKey))
	}
	return names, utils.NewErrorSuccess()
}

/*
添加一个新闻
*/
func AddNews(dbListType utilsleveldb.LeveldbKey, className, title, content string) ([]byte, error) {
	classNameKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(className))
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	createTime := time.Now().Unix()
	news := go_protos.News{
		Class:      []byte(className),
		Title:      []byte(title),
		Content:    []byte(content),
		CreateTime: createTime,
		UpdateTime: createTime,
	}
	bs, err := news.Marshal()
	if err != nil {
		return nil, err
	}
	index, ERR := config.LevelDB.SaveMapInList(dbListType, *classNameKey, bs, nil)
	//再通过修改的方式，把index保存到记录中
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	news = go_protos.News{
		Class:      []byte(className),
		Index:      index,
		Title:      []byte(title),
		Content:    []byte(content),
		CreateTime: createTime,
		UpdateTime: createTime,
	}
	bs, err = news.Marshal()
	if err != nil {
		return nil, err
	}
	ERR = config.LevelDB.SaveOrUpdateMapInListByIndex(dbListType, *classNameKey, index, bs, nil)
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	return index, nil
}

/*
修改一个新闻
*/
func UpdateNews(dbListType utilsleveldb.LeveldbKey, index []byte, className, title, content string) ([]byte, error) {
	classNameKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(className))
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	item, ERR := config.LevelDB.FindMapInListByIndex(dbListType, *classNameKey, index)
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	oldNews, err := model.ParseNews(&item.Value)
	updateTime := time.Now().Unix()
	news := go_protos.News{
		Class:      []byte(className),
		Index:      index,
		Title:      []byte(title),
		Content:    []byte(content),
		CreateTime: oldNews.CreateTime,
		UpdateTime: updateTime,
	}
	bs, err := news.Marshal()
	if err != nil {
		return nil, err
	}
	ERR = config.LevelDB.SaveOrUpdateMapInListByIndex(dbListType, *classNameKey, index, bs, nil)
	if !ERR.CheckSuccess() {
		return nil, errors.New(ERR.Msg)
	}
	return index, nil
}

/*
查询一个类别下的所有新闻条目总数
*/
func FindNewsCount(dbListType utilsleveldb.LeveldbKey, className string) (uint64, error) {
	classNameKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(className))
	if !ERR.CheckSuccess() {
		return 0, errors.New(ERR.Msg)
	}
	count, _, _, ERR := config.LevelDB.FindMapInListTotal(dbListType, *classNameKey)
	if !ERR.CheckSuccess() {
		return 0, errors.New(ERR.Msg)
	}
	return count, nil
}

/*
查询一个类别下的所有新闻
*/
func FindNews(dbListType utilsleveldb.LeveldbKey, className string) ([]*model.News, utils.ERROR) {
	classNameKey, ERR := utilsleveldb.BuildLeveldbKey([]byte(className))
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	items, err := config.LevelDB.FindMapInListAllByKey(dbListType, *classNameKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	news := make([]*model.News, 0)
	for _, one := range items {
		newsOne, err := model.ParseNews(&one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		news = append(news, newsOne)
	}
	return news, utils.NewErrorSuccess()
}
