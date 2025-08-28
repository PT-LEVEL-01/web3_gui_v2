package im

import (
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/utils"
)

const (
	CircleMultcastIntervalMin = time.Second * 10
	CircleMultcastIntervalMax = CircleMultcastIntervalMin * 11
)

var circleClass = new(sync.Map)
var circleClassHaveNew = make(chan bool, 1)

/*
启动定时广播，以及收集广播
*/
func InitCircleClassMultcast() {
	go LoopMultcastCircleClass()
}

/*
循环广播贴子
当最小间隔时间内未收到广播，则广播自己的贴子。
当最小间隔内收到广播，则最大间隔时间广播自己的贴子。
*/
func LoopMultcastCircleClass() {
	ticker := time.NewTicker(CircleMultcastIntervalMin)
	count := int64(0)
	for range ticker.C {
		count++
		have := false
		select {
		case <-circleClassHaveNew:
			have = true
		default:
			have = false
		}
		if time.Duration(count)*CircleMultcastIntervalMin >= CircleMultcastIntervalMax {
			count = 0
			//执行一次广播
			CircleClassMultcastOnce()
			continue
		}
		if have {
			continue
		}
		//执行一次广播
		CircleClassMultcastOnce()
	}
}

/*
广播一次博客
*/
func CircleClassMultcastOnce() {
	//utils.Log.Info().Msgf("广播一次博客")
	class, ERR := db.GetClass(*config.DBKEY_circle_news_release)
	if ERR.CheckFail() {
		return
	}
	for _, one := range class {
		news, ERR := db.FindNews(*config.DBKEY_circle_news_release, one)
		if !ERR.CheckSuccess() {
			continue
		}
		if len(news) <= 0 {
			continue
		}
		newsOne := news[len(news)-1]
		bs, err := newsOne.Proto()
		if err != nil {
			return
		}
		np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
		bs, err = np.Proto()
		if err != nil {
			//return nil, utils.NewErrorSysSelf(err)
			utils.Log.Error().Msgf("格式化广播消息错误:%s", err.Error())
			return
		}
		Node.SendMulticastMsg(config.MSGID_circle_multicast_news_recv, bs)
	}
}

/*
添加广播的博客
@className    string    类别
@addr         string    用户网络地址
@title        string    标题
@content      string    内容
*/
func AddMultcastNews(addr string, news *model.News) {
	utils.Log.Info().Msgf("接收博客广播 %s %s", addr, news.Title)
	select {
	case circleClassHaveNew <- false:
	default:
	}
	key := addr + news.Title
	valueItr, ok := circleClass.Load(news.Class)
	if ok {
		value := valueItr.(*sync.Map)
		value.Store(key, news)
		return
	}
	newsMap := new(sync.Map)
	newsMap.Store(key, news)
	circleClass.Store(news.Class, newsMap)
}

/*
获取广播的博客圈子及博客数量
*/
func GetMultcastClassCount() []model.ClassCount {
	classCounts := make([]model.ClassCount, 0)
	circleClass.Range(func(k, v interface{}) bool {
		key := k.(string)
		value := v.(*sync.Map)
		count := uint64(0)
		value.Range(func(k, v interface{}) bool { count++; return true })
		one := model.ClassCount{
			Name:  key,
			Count: count,
		}
		classCounts = append(classCounts, one)
		return true
	})
	return classCounts
}

/*
查询指定的圈子类别中，广播的博客圈子及博客列表
*/
func GetMultcastClassNewsList(className string) []model.News {
	news := make([]model.News, 0)
	newsMapItr, ok := circleClass.Load(className)
	if !ok {
		return news
	}
	newMap := newsMapItr.(*sync.Map)
	newMap.Range(func(k, v interface{}) bool {
		value := v.(*model.News)
		news = append(news, *value)
		return true
	})
	return news
}
