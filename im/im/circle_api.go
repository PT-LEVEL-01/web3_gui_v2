package im

import (
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/utils"
)

/*
添加分类
@name    string    分类名称
*/
func AddClass(name string) error {
	return db.AddClass(*config.DBKEY_circle_news_release, name)
}

/*
获取所有分类
*/
func GetClass() ([]string, utils.ERROR) {
	return db.GetClass(*config.DBKEY_circle_news_release)
}

/*
添加一条新闻到草稿箱列表
*/
func AddNewsToDraft(className, title, content string) ([]byte, error) {
	return db.AddNews(*config.DBKEY_circle_news_draft, className, title, content)
}

/*
添加一条新闻到发布列表
*/
func AddNewsToRelease(className, title, content string) ([]byte, error) {
	return db.AddNews(*config.DBKEY_circle_news_release, className, title, content)
}

/*
查询新闻在草稿箱列表
*/
func FindNewsDraft(class string) ([]*model.News, utils.ERROR) {
	return db.FindNews(*config.DBKEY_circle_news_draft, class)
}

/*
查询新闻在发布列表
*/
func FindNewsRelease(class string) ([]*model.News, utils.ERROR) {
	return db.FindNews(*config.DBKEY_circle_news_release, class)
}
