package message_center

import (
	"sync"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/message_center/doubleratchet"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

// var sessionManager = NewSessionManager()

var createHeMutex = KeyMutext{}

type SessionManager struct {
	nodeStore     map[string]map[string]*Session //key:string=节点地址;map[string=机器id];value:*Session=通道状态信息;
	nodeStoreLock sync.RWMutex
}

type KeyMutext struct {
	mutexMap sync.Map
}

func (k *KeyMutext) Lock(key interface{}) func() {
	mutexValue, _ := k.mutexMap.LoadOrStore(key, &sync.Mutex{})
	mutex := mutexValue.(*sync.Mutex)
	mutex.Lock()
	return func() {
		mutex.Unlock()
	}
}

/*
添加一个发送管道
*/
func (this *SessionManager) AddSendPipe(id *nodeStore.AddressNet, machineId []byte, sk, sharedHka, sharedNhkb [32]byte,
	keyPair *keystore.KeyPair) error {
	//fmt.Println("添加一个发送管道", hex.EncodeToString(sk[:]), hex.EncodeToString(sharedHka[:]), hex.EncodeToString(sharedNhkb[:]))
	session, err := doubleratchet.NewHE(sk, sharedHka, sharedNhkb, keyPair)
	if err != nil {
		return err
	}

	sessionKey := NewSessionKey(sk, sharedHka, sharedNhkb)
	sessionKey.sessionHE = session

	// 构建key值 id+machineId
	str := utils.Bytes2string(id.Data()) //id.B58String()

	// 加锁
	this.nodeStoreLock.Lock()
	defer this.nodeStoreLock.Unlock()

	// 查询和添加
	machineValue, ok := this.nodeStore[str]
	if ok && machineValue != nil {
		session, exist := machineValue[utils.Bytes2string(machineId)]
		if exist {
			session.sendPipe = &sessionKey
		} else {
			session := NewSession(id)
			session.sendPipe = &sessionKey
			machineValue[utils.Bytes2string(machineId)] = &session
		}
	} else {
		machineValue = make(map[string]*Session)
		session := NewSession(id)
		session.sendPipe = &sessionKey
		machineValue[utils.Bytes2string(machineId)] = &session

		this.nodeStore[str] = machineValue
	}

	return nil
}

/*
删除发送管道
*/
func (this *SessionManager) RemoveSendPipe(id *nodeStore.AddressNet, machineId []byte) {
	// 构建key值 id+machineId
	strKey := utils.Bytes2string(id.Data())

	this.nodeStoreLock.Lock()
	defer this.nodeStoreLock.Unlock()

	machineValue, ok := this.nodeStore[strKey]
	if ok && machineValue != nil {
		if len(machineId) > 0 {
			session, exist := machineValue[utils.Bytes2string(machineId)]
			if exist {
				session.sendPipe = nil
			}
		} else {
			for _, v := range machineValue {
				v.sendPipe = nil
			}
		}
	}
}

/*
添加一个接收管道
*/
func (this *SessionManager) AddRecvPipe(id *nodeStore.AddressNet, machineId []byte, sk, sharedHka, sharedNhkb, puk [32]byte) error {
	session, err := doubleratchet.NewHEWithRemoteKey(sk, sharedHka, sharedNhkb, puk)
	if err != nil {
		return err
	}

	sessionKey := NewSessionKey(sk, sharedHka, sharedNhkb)
	sessionKey.sessionHE = session

	// 构建key值 id+machineId
	str := utils.Bytes2string(id.Data())

	// 加锁
	this.nodeStoreLock.Lock()
	defer this.nodeStoreLock.Unlock()

	machineValue, ok := this.nodeStore[str]
	if ok && machineValue != nil {
		session, ok := machineValue[utils.Bytes2string(machineId)]
		if ok {
			session.recvPipe = &sessionKey
		} else {
			session := NewSession(id)
			session.recvPipe = &sessionKey
			machineValue[utils.Bytes2string(machineId)] = &session
		}
	} else {
		machineValue = make(map[string]*Session)
		session := NewSession(id)
		session.recvPipe = &sessionKey
		machineValue[utils.Bytes2string(machineId)] = &session

		this.nodeStore[str] = machineValue
	}

	return nil
}

/*
获取一个发送棘轮
*/
func (this *SessionManager) GetSendRatchet(id *nodeStore.AddressNet, machineId []byte) doubleratchet.SessionHE {
	// 构建key值 id+machineId
	str := utils.Bytes2string(id.Data())
	this.nodeStoreLock.RLock()
	defer this.nodeStoreLock.RUnlock()
	machineValue, ok := this.nodeStore[str]
	//if !ok || machineValue == nil {return nil}
	if ok && machineValue != nil {
		if len(machineId) > 0 {
			session, ok := machineValue[utils.Bytes2string(machineId)]
			if ok {
				if session.sendPipe != nil {
					return session.sendPipe.sessionHE
				}
			}
		} else {
			for _, v := range machineValue {
				if v != nil && v.sendPipe != nil {
					return v.sendPipe.sessionHE
				}
			}
		}
	}
	return nil
}

/*
获取一个接收棘轮
*/
func (this *SessionManager) GetRecvRatchet(id *nodeStore.AddressNet, machineId string) doubleratchet.SessionHE {
	// 构建key值 id+machineId
	str := utils.Bytes2string(id.Data()) //id.B58String()

	this.nodeStoreLock.RLock()
	defer this.nodeStoreLock.RUnlock()

	machineValue, ok := this.nodeStore[str]
	if ok && machineValue != nil {
		if machineId != "" {
			session, ok := machineValue[machineId]
			if ok {
				if session.recvPipe != nil {
					return session.recvPipe.sessionHE
				}
			}
		} else {
			for _, v := range machineValue {
				if v != nil && v.recvPipe != nil {
					return v.recvPipe.sessionHE
				}
			}
		}
	}

	return nil
}

/*
创建一个新的节点管理器
*/
func NewSessionManager() SessionManager {
	return SessionManager{nodeStore: make(map[string]map[string]*Session)}
}

type Session struct {
	Id       *nodeStore.AddressNet //节点地址
	recvPipe *SessionKey           //接收管道
	sendPipe *SessionKey           //发送管道
}

func NewSession(id *nodeStore.AddressNet) Session {
	return Session{
		Id: id, //节点地址
		// recvPipe: make([]SessionKey, 0), //接收管道
		// sendPipe: make([]SessionKey, 0), //发送管道
	}
}

/*
加密通道
*/
type SessionKey struct {
	sk         [32]byte                //协商密钥
	sharedHka  [32]byte                //随机密钥
	sharedNhkb [32]byte                //随机密钥
	sessionHE  doubleratchet.SessionHE //双棘轮算法状态
}

func NewSessionKey(sk, sharedHka, sharedNhkb [32]byte) SessionKey {
	return SessionKey{
		sk:         sk,         //协商密钥
		sharedHka:  sharedHka,  //随机密钥
		sharedNhkb: sharedNhkb, //随机密钥
	}
}
