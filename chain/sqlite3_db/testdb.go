package sqlite3_db

import (
	"time"
	"web3_gui/libp2parea/adapter/engine"

	_ "github.com/go-xorm/xorm"
)

// var txitemLock = new(sync.Mutex)

type TestDB struct {
	Id         int64     `xorm:"pk autoincr unique 'id'"` //id
	Txid       []byte    `xorm:"Blob 'txid'"`             //交易id
	CreateTime time.Time `xorm:"created 'createtime'"`    //创建时间，这个Field将在Insert时自动赋值为当前时间
}

/*
添加一个未花费的交易记录
@return    int64    数据库id
*/
func (this *TestDB) AddTxItems(txitems *[]TestDB) error {
	if txitems == nil || len(*txitems) <= 0 {
		return nil
	}
	lenght := len(*txitems)
	pageOne := 100
	var err error

	// txitemLock.Lock()
	for i := 0; i < lenght/pageOne; i++ {
		items := (*txitems)[i*pageOne : (i+1)*pageOne]
		_, err = engineDB.Insert(&items)
		if err != nil {
			// txitemLock.Unlock()
			return err
		}
	}
	if lenght%pageOne > 0 {
		i := lenght / pageOne
		items := (*txitems)[i*pageOne : lenght]
		_, err = engineDB.Insert(&items)
		if err != nil {
			// txitemLock.Unlock()
			return err
		}
	}
	// txitemLock.Unlock()
	return nil
}

/*
删除多个已经使用了的余额记录
*/
func (this *TestDB) RemoveMoreKey(keys [][]byte) error {
	if keys == nil || len(keys) <= 0 {
		return nil
	}

	tis := make([]TestDB, 0)
	// txitemLock.Lock()
	err := engineDB.In("txid = ?", keys).Find(&tis)
	engine.Log.Info("查询结果:%d", len(tis))

	// params := make([]string, 0, len(keys))
	for i, _ := range keys {
		n, err := engineDB.Where("txid = ?", keys[i]).Unscoped().Delete(this)
		engine.Log.Info("删除记录数量:%d error:%v", n, err)
		// params = append(params, utils.Bytes2string(keys[i]))
	}

	// sql := "delete from tx_item where key in (?)"
	// res, err := engineDB.Exec(sql, params...)
	// engine.Log.Info("%+v\n%+v", res, err)

	// txitemLock.Lock()
	n, err := engineDB.In("txid = ?", keys).Unscoped().Delete(this)
	engine.Log.Info("删除记录数量:%d", n)
	// txitemLock.Unlock()
	return err
}
