package storage

import (
	"sync"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
权限管理
*/
type AuthManager struct {
	lock          *sync.RWMutex         //
	addr          string                //
	userSpaces    map[string]utils.Byte //用户已购买空间容量 单位：byte
	userSpacesUse map[string]utils.Byte //用户已使用空间容量 单位：byte
}

/*
创建
*/
func CreateAuthManager(addr string) *AuthManager {
	auth := AuthManager{
		lock:          new(sync.RWMutex),
		addr:          addr,
		userSpaces:    make(map[string]utils.Byte),
		userSpacesUse: make(map[string]utils.Byte),
	}
	//ids, err := db.StorageServer_GetStorageServerInUseOrdersIDs()
	//if err != nil {
	//	return nil, err
	//}

	return &auth
}

/*
查询一个用户的存储空间
@return    uint64    购买的总空间
@return    uint64    已经使用的空间
@return    uint64    剩余空间
*/
func (this *AuthManager) QueryUserSpaces(addr nodeStore.AddressNet) (utils.Byte, utils.Byte, utils.Byte) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	spacesTotal := this.userSpaces[utils.Bytes2string(addr.GetAddr())]
	spacesUse := this.userSpacesUse[utils.Bytes2string(addr.GetAddr())]
	remain := utils.Byte(0)
	if spacesTotal > spacesUse {
		remain = spacesTotal - spacesUse
	}
	//utils.Log.Info().Msgf("查询用户存储空间:%d %d %d", spacesTotal, spacesUse, remain)
	return spacesTotal, spacesUse, remain
}

/*
查询用户列表
@return    uint64    购买的总空间
@return    uint64    已经使用的空间
@return    uint64    剩余空间
*/
func (this *AuthManager) QueryUserList() []nodeStore.AddressNet {
	this.lock.RLock()
	defer this.lock.RUnlock()
	list := make([]nodeStore.AddressNet, 0, len(this.userSpaces))
	for k, _ := range this.userSpaces {
		addr := nodeStore.NewAddressNet([]byte(k)) //nodeStore.AddressNet([]byte(k))
		list = append(list, *addr)
	}
	return list
}

/*
删除用户
*/
func (this *AuthManager) DelUser(addr nodeStore.AddressNet) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	delete(this.userSpaces, utils.Bytes2string(addr.GetAddr()))
	delete(this.userSpacesUse, utils.Bytes2string(addr.GetAddr()))
}

/*
添加购买空间
*/
func (this *AuthManager) AddPurchaseSpace(addr nodeStore.AddressNet, count utils.Byte) {
	this.lock.Lock()
	defer this.lock.Unlock()
	//utils.Log.Info().Msgf("添加用户购买空间:%d", count)
	total, ok := this.userSpaces[utils.Bytes2string(addr.GetAddr())]
	if !ok {
		//utils.Log.Info().Msgf("保存用户购买空间:%d", count)
		this.userSpaces[utils.Bytes2string(addr.GetAddr())] = count
		this.userSpacesUse[utils.Bytes2string(addr.GetAddr())] = 0
		return
	}
	total += count
	//utils.Log.Info().Msgf("添加用户购买空间:%d", total)
	this.userSpaces[utils.Bytes2string(addr.GetAddr())] = total
}

/*
减少购买空间
*/
func (this *AuthManager) SubPurchaseSpace(addr nodeStore.AddressNet, count utils.Byte) {
	this.lock.Lock()
	defer this.lock.Unlock()
	total, ok := this.userSpaces[utils.Bytes2string(addr.GetAddr())]
	if !ok {
		utils.Log.Error().Msgf("减少空间错误:%s %d", addr.B58String(), count)
		return
	}
	utils.Log.Info().Msgf("减少购买的空间数量:%d %d", total, count)
	if count > total {
		utils.Log.Error().Msgf("减少空间错误:%s %d %d", addr.B58String(), count, total)
		return
	}
	total -= count
	//if total == 0 {
	//	delete(this.userSpaces, utils.Bytes2string(addr))
	//	delete(this.userSpacesUse, utils.Bytes2string(addr))
	//	return
	//}
	//当空间为0时，这里不能直接删除
	this.userSpaces[utils.Bytes2string(addr.GetAddr())] = total
}

/*
添加使用空间
空间不足则返回false
*/
func (this *AuthManager) AddUseSpace(addr nodeStore.AddressNet, count utils.Byte) bool {
	//utils.Log.Info().Msgf("判断空间是否足够")
	this.lock.Lock()
	defer this.lock.Unlock()
	total, ok := this.userSpaces[utils.Bytes2string(addr.GetAddr())]
	if !ok {
		utils.Log.Info().Msgf("空间不足")
		return false
	}
	usetTotal, ok := this.userSpacesUse[utils.Bytes2string(addr.GetAddr())]
	if total < usetTotal+count {
		utils.Log.Info().Msgf("空间不足:%d %d %d", total, usetTotal, count)
		return false
	}
	usetTotal += count
	this.userSpacesUse[utils.Bytes2string(addr.GetAddr())] = usetTotal
	return true
}

/*
减少使用空间
*/
func (this *AuthManager) SubUseSpace(addr nodeStore.AddressNet, count utils.Byte) bool {
	this.lock.Lock()
	defer this.lock.Unlock()

	//total, ok := this.userSpaces[utils.Bytes2string(addr)]
	//if !ok {
	//	utils.Log.Info().Msgf("空间不足")
	//	return false
	//}
	//usetTotal, ok := this.userSpacesUse[utils.Bytes2string(addr)]
	//if total < usetTotal+count {
	//	utils.Log.Info().Msgf("空间不足:%d %d %d", total, usetTotal, count)
	//	return false
	//}

	total, ok := this.userSpacesUse[utils.Bytes2string(addr.GetAddr())]
	if !ok {
		utils.Log.Error().Msgf("减少使用空间错误:%s %d", addr.B58String(), count)
		return false
	}
	if count > total {
		utils.Log.Error().Msgf("减少使用空间错误:%s %d %d", addr.B58String(), count, total)
		return false
	}
	total -= count
	this.userSpacesUse[utils.Bytes2string(addr.GetAddr())] = total
	return true
}
