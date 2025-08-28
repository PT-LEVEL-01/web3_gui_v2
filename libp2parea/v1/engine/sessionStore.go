package engine

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"web3_gui/utils"
)

type sessionBase struct {
	name       string
	attrbutes  *sync.Map //
	machineID  string    // 机器Id
	areaName   string    // 区域名
	setGodTime int64     // 被设置为超级代理的时间
	connType   uint8     // 连接类型
}

func (this *sessionBase) Set(name string, value interface{}) {
	this.attrbutes.Store(name, value)
}
func (this *sessionBase) Get(name string) interface{} {
	v, _ := this.attrbutes.Load(name)
	return v
}
func (this *sessionBase) GetName() string {
	return this.name
}

func (this *sessionBase) SetName(name string) {
	this.name = name
}

func (this sessionBase) GetConnType() string {
	if this.connType == CONN_TYPE_TCP {
		return "TCP"
	} else if this.connType == CONN_TYPE_QUIC {
		return "QUIC"
	}

	return "UNKNOWN"
}

func (this *sessionBase) Close() {}

type Session interface {
	GetIndex() uint64
	Send(msgID uint64, data, datapuls *[]byte, timeout time.Duration) error
	Close()
	Set(name string, value interface{})
	Get(name string) interface{}
	GetName() string
	SetName(name string)
	GetRemoteHost() string
	GetMachineID() string
	GetAreaName() string
	GetSetGodTime() int64
	GetConnType() string
}

type sessionStore struct {
	lock               *sync.RWMutex
	indexMax           uint64                     //
	sessionByIndex     map[uint64]Session         //
	customNameStore    map[string]*utils.SyncList //
	upSessionByIndex   map[string]map[uint64]Session
	downSessionByIndex map[string]map[uint64]Session
}

/*
 * 只添加session到customNameStore和upSession
 */
func (this *sessionStore) addOnlyUpSession(customName string, session Session, areaName []byte) {
	this.addUpSession(areaName, session)

	this.lock.Lock()
	slist, ok := this.customNameStore[customName]
	if ok && slist != nil {
		slist.Add(session)
	} else {
		slist := utils.NewSyncList()
		slist.Add(session)
		this.customNameStore[customName] = slist
	}
	if _, ok := this.sessionByIndex[session.GetIndex()]; !ok {
		this.sessionByIndex[session.GetIndex()] = session
	}
	this.lock.Unlock()
}

/*
 * 只添加session到customNameStore和downSession
 */
func (this *sessionStore) addOnlyDownSession(customName string, session Session, areaName []byte) {
	this.addDownSession(areaName, session)

	this.lock.Lock()
	slist, ok := this.customNameStore[customName]
	if ok && slist != nil {
		slist.Add(session)
	} else {
		slist := utils.NewSyncList()
		slist.Add(session)
		this.customNameStore[customName] = slist
	}
	if _, ok := this.sessionByIndex[session.GetIndex()]; !ok {
		this.sessionByIndex[session.GetIndex()] = session
	}
	this.lock.Unlock()
}

/*
 * 添加session
 */
func (this *sessionStore) addSession(customName string, session Session, ss string, areaName []byte) {
	// addrNet := AddressNet([]byte(customName))
	// utils.Log.Info().Msgf("添加session:%d %s", index, addrNet.B58String())
	if ss == AddressUp {
		this.addUpSession(areaName, session)
	} else if ss == AddressDown {
		this.addDownSession(areaName, session)
	}

	index := session.GetIndex()
	this.lock.Lock()
	_, ok := this.sessionByIndex[index]
	if ok {
		this.lock.Unlock()
		return
	}
	this.sessionByIndex[index] = session
	slist, ok := this.customNameStore[customName]
	if ok && slist != nil {
		slist.Add(session)
	} else {
		slist := utils.NewSyncList()
		slist.Add(session)
		this.customNameStore[customName] = slist
	}
	this.lock.Unlock()
}

// 管理根据新规则连接的 地址比自己大的session
func (this *sessionStore) addUpSession(areaName []byte, session Session) {
	this.lock.Lock()
	defer this.lock.Unlock()
	index := session.GetIndex()
	if _, ok := this.upSessionByIndex[utils.Bytes2string(areaName)][index]; ok {
		return
	}

	if _, ok := this.upSessionByIndex[utils.Bytes2string(areaName)]; ok {
		this.upSessionByIndex[utils.Bytes2string(areaName)][index] = session
	} else {
		tmp := make(map[uint64]Session)
		tmp[index] = session
		this.upSessionByIndex[utils.Bytes2string(areaName)] = tmp
	}
}

// 管理根据新规则连接的 地址比自己小的session
func (this *sessionStore) addDownSession(areaName []byte, session Session) {
	this.lock.Lock()
	defer this.lock.Unlock()
	index := session.GetIndex()
	if _, ok := this.downSessionByIndex[utils.Bytes2string(areaName)][index]; ok {
		return
	}
	if _, ok := this.downSessionByIndex[utils.Bytes2string(areaName)]; ok {
		this.downSessionByIndex[utils.Bytes2string(areaName)][index] = session
	} else {
		tmp := make(map[uint64]Session)
		tmp[index] = session
		this.downSessionByIndex[utils.Bytes2string(areaName)] = tmp
	}
}

/*
 * 根据customName获取特定session
 */
func (this *sessionStore) getSession(areaName []byte, customName string) (Session, bool) {
	// addrNet := AddressNet([]byte(customName))
	// utils.Log.Info().Msgf("获取session:%s", AddressNet([]byte(customName)).B58String())
	this.lock.RLock()
	slist, ok := this.customNameStore[customName]
	if slist == nil || !ok {
		// utils.Log.Info().Msgf("获取session fail:%s", addrNet.B58String())
		this.lock.RUnlock()
		return nil, false
	}
	// utils.Log.Info().Msgf("获取session:%s", addrNet.B58String())
	ssItr := slist.GetAll()
	for i, _ := range ssItr {
		ssOne, ok := ssItr[i].(Session)
		if !ok {
			continue
		}
		if ssOne.GetAreaName() != utils.Bytes2string(areaName) {
			continue
		}
		this.lock.RUnlock()
		return ssOne, true
	}
	this.lock.RUnlock()
	return nil, false
}

/*
 * getSessionAll 根据目标获取所有的连接信息
 *
 * @param 	customName	string		目标地址
 * @return	sessions	[]Session	目标对应的所有连接
 * @return	success		bool		是否存在与目标的连接信息
 */
func (this *sessionStore) getSessionAll(areaName []byte, customName string) ([]Session, bool) {
	this.lock.RLock()
	slist, ok := this.customNameStore[customName]
	if slist == nil || !ok {
		this.lock.RUnlock()
		return nil, false
	}
	ssItr := slist.GetAll()
	ss := make([]Session, 0, len(ssItr))
	for i, _ := range ssItr {
		ssOne, ok := ssItr[i].(Session)
		if !ok {
			continue
		}
		if ssOne.GetAreaName() != utils.Bytes2string(areaName) {
			continue
		}
		ss = append(ss, ssOne)
	}
	this.lock.RUnlock()
	return ss, true
}

// func (this *sessionStore) getSessionByHost(hostName string) Session {
// 	this.lock.RLock()
// 	ss, ok := this.hostNameStore[hostName]
// 	if !ok {
// 		this.lock.RUnlock()
// 		return nil
// 	}
// 	this.lock.RUnlock()
// 	return ss
// }

/*
 * 特定session是否存在
 */
func (this *sessionStore) checkInSessionByIndex(ss Session) bool {
	if _, ok := this.sessionByIndex[ss.GetIndex()]; ok {
		return true
	} else {
		return false
	}
}

/*
 * 在真实节点downSession中删除特定session
 */
func (this *sessionStore) delNodeDownSession(areaName []byte, ss Session) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.downSessionByIndex[utils.Bytes2string(areaName)]; ok {
		delete(this.downSessionByIndex[utils.Bytes2string(areaName)], ss.GetIndex())
	}
}

/*
 * 在真实节点upSession中删除特定session
 */
func (this *sessionStore) delNodeUpSession(areaName []byte, ss Session) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.upSessionByIndex[utils.Bytes2string(areaName)]; ok {
		delete(this.upSessionByIndex[utils.Bytes2string(areaName)], ss.GetIndex())
	}
}

/*
 * 在真实节点customNameStore中删除特定session
 */
func (this *sessionStore) removeCustomSession(ss Session) {
	this.lock.Lock()
	defer this.lock.Unlock()

	slist, ok := this.customNameStore[ss.GetName()]
	if ok && slist != nil {
		ssItr := slist.GetAll()
		for i, _ := range ssItr {
			session := ssItr[i].(Session)
			if session.GetIndex() == ss.GetIndex() {
				slist.Remove(i)
				if len(ssItr) == 1 {
					delete(this.customNameStore, ss.GetName())
				}
				break
			}
		}
	}
}

/*
 * sessionStore中删除特定session
 */
func (this *sessionStore) removeSession(areaName string, ss Session) {
	// addrNet := AddressNet([]byte(ss.GetName()))
	// utils.Log.Info().Msgf("删除一个连接session:%d %s", ss.GetIndex(), addrNet.B58String())
	var hasIndex bool
	var session Session

	this.lock.Lock()

	if ses, ok := this.sessionByIndex[ss.GetIndex()]; ok {
		session = ses
		hasIndex = ok
	}
	if ses, ok := this.upSessionByIndex[areaName][ss.GetIndex()]; ok {
		session = ses
		hasIndex = ok
	}
	if ses, ok := this.downSessionByIndex[areaName][ss.GetIndex()]; ok {
		session = ses
		hasIndex = ok
	}

	if !hasIndex {
		this.lock.Unlock()
		return
	}
	slist, ok := this.customNameStore[session.GetName()]
	if ok && slist != nil {
		ssItr := slist.GetAll()
		for i, _ := range ssItr {
			session := ssItr[i].(Session)
			if session.GetIndex() == ss.GetIndex() {
				slist.Remove(i)
				break
			}
		}
	}
	//如果list长度为0,则删除
	if ok && len(slist.GetAll()) == 0 {
		delete(this.customNameStore, session.GetName())
	}

	if _, ok := this.sessionByIndex[ss.GetIndex()]; ok {
		delete(this.sessionByIndex, ss.GetIndex())
	}
	if _, ok := this.upSessionByIndex[areaName][ss.GetIndex()]; ok {
		delete(this.upSessionByIndex[areaName], ss.GetIndex())
	}
	if _, ok := this.downSessionByIndex[areaName][ss.GetIndex()]; ok {
		delete(this.downSessionByIndex[areaName], ss.GetIndex())
	}
	this.lock.Unlock()
}

/*
 * sessionStore中重命名session
 */
func (this *sessionStore) renameSession(index uint64, customName string) {
	this.lock.Lock()
	session, ok := this.sessionByIndex[index]
	// session, ok := this.hostNameStore[hostName]
	if !ok {
		this.lock.Unlock()
		return
	}
	//从原来的集合中删除
	oldCustomName := session.GetName()
	slist, ok := this.customNameStore[oldCustomName]
	if !ok || slist == nil {
		this.lock.Unlock()
		return
	}
	//
	ssItr := slist.GetAll()
	for i, _ := range ssItr {
		ss := ssItr[i].(Session)
		// hostNameOne := ss.GetRemoteHost()
		if ss.GetIndex() == index {
			//找到了
			slist.Remove(i)
			break
		}
	}
	//如果list长度为0,则删除
	if len(slist.GetAll()) == 0 {
		// utils.Log.Info().Msgf("removeSessionByHost:%s", hex.EncodeToString([]byte(oldCustomName)))
		delete(this.customNameStore, oldCustomName)
	}
	//保存到新的名称集合中
	slist, ok = this.customNameStore[customName]
	if ok && slist != nil {
		slist.Add(session)
	} else {
		slist := utils.NewSyncList()
		slist.Add(session)
		// utils.Log.Info().Msgf("removeSessionByHost rename:%s", hex.EncodeToString([]byte(customName)))
		this.customNameStore[customName] = slist
	}
	this.lock.Unlock()
}

/*
 * 获取特定区域名的所有session
 */
func (this *sessionStore) getAllSession(areaName []byte) []Session {
	ss := make([]Session, 0)
	strAreaName := utils.Bytes2string(areaName)
	this.lock.RLock()
	for _, session := range this.sessionByIndex {
		if session.GetAreaName() != strAreaName {
			continue
		}
		ss = append(ss, session)
	}
	this.lock.RUnlock()
	return ss
}

/*
 * 获取真实节点的所有upSession
 */
func (this *sessionStore) getAllUpSession(areaName []byte) []Session {
	ss := make([]Session, 0)
	if _, ok := this.upSessionByIndex[utils.Bytes2string(areaName)]; !ok {
		return ss
	}
	this.lock.RLock()
	for i := range this.upSessionByIndex[utils.Bytes2string(areaName)] {
		ss = append(ss, this.upSessionByIndex[utils.Bytes2string(areaName)][i])
	}
	this.lock.RUnlock()

	return ss
}

/*
 * 获取真实节点的所有downSession
 */
func (this *sessionStore) getAllDownSession(areaName []byte) []Session {
	ss := make([]Session, 0)
	if _, ok := this.downSessionByIndex[utils.Bytes2string(areaName)]; !ok {
		return ss
	}
	this.lock.RLock()
	for i := range this.downSessionByIndex[utils.Bytes2string(areaName)] {
		ss = append(ss, this.downSessionByIndex[utils.Bytes2string(areaName)][i])
	}
	this.lock.RUnlock()

	return ss
}

/*
 * 获取特定区域名的所有session
 */
func (this *sessionStore) getAllSessionName(areaName []byte) []string {
	names := make([]string, 0)
	nameM := make(map[string]int)
	strAreaName := utils.Bytes2string(areaName)
	this.lock.RLock()
	for _, one := range this.sessionByIndex {
		if _, ok := nameM[one.GetName()]; ok {
			continue
		}
		if one.GetAreaName() != strAreaName {
			continue
		}
		nameM[one.GetName()] = 0
		names = append(names, one.GetName())
	}
	this.lock.RUnlock()
	return names
}

/*
获得一个未使用的服务器连接
*/
func (this *sessionStore) getClientConn(areaName []byte, engine *Engine) *Client {
	this.lock.Lock()
	for {
		_, ok := this.sessionByIndex[this.indexMax]
		if !ok {
			break
		}
		this.indexMax++
	}
	index := this.indexMax
	this.indexMax++
	this.lock.Unlock()

	contextRoot, canceRoot := context.WithCancel(context.Background())
	sessionBase := sessionBase{
		attrbutes: new(sync.Map),
		areaName:  utils.Bytes2string(areaName),
		connType:  CONN_TYPE_TCP,
	}
	clientConn := &Client{
		sessionBase: sessionBase,
		index:       index,
		engine:      engine,
		sendQueue:   NewSendQueue(SendQueueCacheNum, contextRoot, canceRoot, engine.name),
		allowClose:  make(chan bool, 1),
		heartbeat:   make(chan bool, 1),
		getDataSign: make(chan bool, 1),
	}
	return clientConn

}

/*
获得一个未使用的服务器连接
*/
func (this *sessionStore) getServerConn(engine *Engine, areaName string) *ServerConn {
	this.lock.Lock()
	for {
		_, ok := this.sessionByIndex[this.indexMax]
		if !ok {
			break
		}
		this.indexMax++
	}
	index := this.indexMax
	this.indexMax++
	this.lock.Unlock()

	contextRoot, canceRoot := context.WithCancel(context.Background())
	//创建一个新的session
	sessionBase := sessionBase{
		attrbutes: new(sync.Map),
		areaName:  areaName,
		connType:  CONN_TYPE_TCP,
	}
	serverConn := &ServerConn{
		sessionBase: sessionBase,
		index:       index,
		engine:      engine,
		sendQueue:   NewSendQueue(SendQueueCacheNum, contextRoot, canceRoot, engine.name),
		allowClose:  make(chan bool, 1),
	}
	serverConn.controller = &ControllerImpl{
		lock:       new(sync.RWMutex),
		engine:     engine,
		attributes: make(map[string]interface{}),
	}
	return serverConn
}

/*
获得一个未使用的服务器连接
*/
func (this *sessionStore) getClientQuicConn(areaName []byte, engine *Engine) *ClientQuic {
	this.lock.Lock()
	for {
		_, ok := this.sessionByIndex[this.indexMax]
		if !ok {
			break
		}
		this.indexMax++
	}
	index := this.indexMax
	this.indexMax++
	this.lock.Unlock()

	contextRoot, canceRoot := context.WithCancel(context.Background())
	sessionBase := sessionBase{
		attrbutes: new(sync.Map),
		areaName:  utils.Bytes2string(areaName),
		connType:  CONN_TYPE_QUIC,
	}
	clientConn := &ClientQuic{
		sessionBase: sessionBase,
		index:       index,
		engine:      engine,
		sendQueue:   NewSendQueue(SendQueueCacheNum, contextRoot, canceRoot, engine.name),
		allowClose:  make(chan bool, 1),
		heartbeat:   make(chan bool, 1),
		getDataSign: make(chan bool, 1),
		quicConf: &quic.Config{
			MaxIdleTimeout:  time.Second,
			KeepAlivePeriod: 500 * time.Millisecond,
		},
		tlsConf: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-p2p-project"},
		},
	}

	return clientConn
}

/*
获得一个未使用的服务器连接
*/
func (this *sessionStore) getServerQuicConn(engine *Engine, areaName string) *ServerQuicConn {
	this.lock.Lock()
	for {
		_, ok := this.sessionByIndex[this.indexMax]
		if !ok {
			break
		}
		this.indexMax++
	}
	index := this.indexMax
	this.indexMax++
	this.lock.Unlock()

	contextRoot, canceRoot := context.WithCancel(context.Background())
	//创建一个新的session
	sessionBase := sessionBase{
		attrbutes: new(sync.Map),
		areaName:  areaName,
		connType:  CONN_TYPE_QUIC,
	}
	serverConn := &ServerQuicConn{
		sessionBase: sessionBase,
		index:       index,
		engine:      engine,
		sendQueue:   NewSendQueue(SendQueueCacheNum, contextRoot, canceRoot, engine.name),
		allowClose:  make(chan bool, 1),
	}
	serverConn.controller = &ControllerImpl{
		lock:       new(sync.RWMutex),
		engine:     engine,
		attributes: make(map[string]interface{}),
	}
	return serverConn
}

func NewSessionStore() *sessionStore {
	sessionStore := new(sessionStore)
	sessionStore.lock = new(sync.RWMutex)
	sessionStore.sessionByIndex = make(map[uint64]Session)
	sessionStore.downSessionByIndex = make(map[string]map[uint64]Session)
	sessionStore.upSessionByIndex = make(map[string]map[uint64]Session)
	sessionStore.customNameStore = make(map[string]*utils.SyncList)
	// sessionStore.hostNameStore = make(map[string]Session)
	return sessionStore
}

/*
 * 获取特定区域名的session总数
 */
func (this *sessionStore) getSessionCnt(areaName []byte) (res int64) {
	strAreaName := utils.Bytes2string(areaName)
	this.lock.RLock()
	for _, one := range this.sessionByIndex {
		if one.GetAreaName() != strAreaName {
			continue
		}
		res++
	}
	this.lock.RUnlock()
	return res
}
