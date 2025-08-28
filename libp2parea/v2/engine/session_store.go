package engine

import (
	"sync"
	"web3_gui/utils"
)

type sessionStore struct {
	lock     *sync.RWMutex      //
	sessions map[string]Session //
}

func NewSessionStore() *sessionStore {
	sessionStore := new(sessionStore)
	sessionStore.lock = new(sync.RWMutex)
	sessionStore.sessions = make(map[string]Session)
	return sessionStore
}

/*
 * 添加session
 */
func (this *sessionStore) addSession(session Session) {
	this.lock.Lock()
	this.sessions[utils.Bytes2string(session.GetId())] = session
	this.lock.Unlock()
}

/*
 * 根据id查询session
 */
func (this *sessionStore) getSessionById(id []byte) Session {
	this.lock.RLock()
	session := this.sessions[utils.Bytes2string(id)]
	this.lock.RUnlock()
	return session
}

/*
 * getSessionAll 根据目标获取所有的连接信息
 *
 * @param 	customName	string		目标地址
 * @return	sessions	[]Session	目标对应的所有连接
 * @return	success		bool		是否存在与目标的连接信息
 */
func (this *sessionStore) getSessionAll() []Session {
	this.lock.RLock()
	defer this.lock.RUnlock()
	ss := make([]Session, 0, len(this.sessions))
	for _, one := range this.sessions {
		ss = append(ss, one)
	}
	return ss
}

/*
 * sessionStore中删除特定session
 */
func (this *sessionStore) removeSession(ss Session) {
	this.lock.Lock()
	delete(this.sessions, utils.Bytes2string(ss.GetId()))
	this.lock.Unlock()
}

/*
 * sessionStore中重命名session
 */
func (this *sessionStore) renameSession(oldId, newId []byte) {
	this.lock.Lock()
	defer this.lock.Unlock()
	session, ok := this.sessions[utils.Bytes2string(oldId)]
	if !ok {
		return
	}
	delete(this.sessions, utils.Bytes2string(oldId))
	this.sessions[utils.Bytes2string(newId)] = session
}
