package storage

import (
	"errors"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"web3_gui/chain/evm/common"

	"web3_gui/chain/config"
	db2 "web3_gui/chain/db"
	db "web3_gui/chain/db/leveldb"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/vm/environment"
	"web3_gui/chain/protos/go_protos"

	"github.com/golang/protobuf/proto"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

const (
	//key 为addressCoin 值为stateObject
	DBKEY_OBJECT = "stateObject_"
)

var (
	// 全局缓存合约对象
	// 地址:合约对象
	globalCacheObj = make(map[string]*go_protos.StateObject)
)

type Storage struct {
	db       *db.LevelDB
	dbTemp   *db.LevelDB
	cacheObj map[string]*go_protos.StateObject // 地址:合约对象
	upObj    map[string]struct{}               // 有变化更新至合约中
}

// 合约对象持久化
func NewStorage() *Storage {
	return &Storage{
		db:       db2.LevelDB,
		dbTemp:   db2.LevelTempDB,
		cacheObj: make(map[string]*go_protos.StateObject),
		upObj:    make(map[string]struct{}),
	}
}

// 更新缓存对象的余额
func (s *Storage) UpdateCacheObjBalance(address crypto.AddressCoin, value *evmutils.Int) {
	if obj, exist := s.cacheObj[address.B58String()]; exist {
		oldBalance := evmutils.New(int64(obj.Balance))
		newBalance := oldBalance.Add(value)
		obj.Balance = newBalance.Uint64()
		obj.BalanceChange = true
		s.upObj[address.B58String()] = struct{}{}
		s.cacheObj[address.B58String()] = obj

	}
}

// 合约对象持久化
// 缓存合约对象
// 通过地址获取合约缓存
func NewStorageWithCache(addresses ...crypto.AddressCoin) *Storage {
	s := &Storage{
		db:       db2.LevelDB,
		dbTemp:   db2.LevelTempDB,
		cacheObj: make(map[string]*go_protos.StateObject),
		upObj:    make(map[string]struct{}),
	}

	for i, _ := range addresses {
		address := addresses[i]
		if obj, ok := globalCacheObj[address.B58String()]; ok {
			s.cacheObj[address.B58String()] = obj
			return s
		}

		obj, err := s.getObj(address)
		if err != nil {
			engine.Log.Warn("通过地址(%s)获取合约对象,%v", address.B58String(), err)
		}
		s.cacheObj[address.B58String()] = obj
		// 同时也缓存到全局变量中
		globalCacheObj[address.B58String()] = obj
	}

	return s
}

// 获取合约对象
func (s *Storage) getObj(address crypto.AddressCoin) (*go_protos.StateObject, error) {
	//先从缓存中获取
	if obj, exist := s.cacheObj[address.B58String()]; exist {
		//engine.Log.Info("cache get %s", address.B58String())
		return obj, nil
	}

	//engine.Log.Info("db get %s", address.B58String())
	objectKey := append([]byte(config.DBKEY_CONTRACT_OBJECT), address...)
	objectValue, err := s.dbTemp.Find(objectKey)
	if err != nil || objectValue == nil || len(*objectValue) == 0 {
		obj := new(go_protos.StateObject)
		obj.Address = address
		obj.Balance, _ = s.getBalance(address)
		obj.BalanceChange = true
		obj.Nonce, _ = s.getNonce(address)
		obj.NonceChange = true
		obj.CacheStorage = make(map[string][]byte)
		s.upObj[address.B58String()] = struct{}{}
		s.cacheObj[address.B58String()] = obj
		return obj, nil
	}
	obj := new(go_protos.StateObject)
	proto.Unmarshal(*objectValue, obj)
	//这里重置余额
	if obj.CacheStorage == nil {
		obj.CacheStorage = make(map[string][]byte)
	}
	obj.Balance, _ = s.getBalance(address)
	obj.Nonce, _ = s.getNonce(address)
	s.cacheObj[address.B58String()] = obj
	return obj, err
}

// 获取合约对象
func (s *Storage) GetCacheObj(address crypto.AddressCoin) (*go_protos.StateObject, bool) {
	if obj, exist := s.cacheObj[address.B58String()]; exist {
		return obj, true
	}

	return nil, false
}

// 获取地址可用余额
func (s *Storage) getBalance(address crypto.AddressCoin) (uint64, error) {
	_, value := db2.GetNotSpendBalance(&address)
	//dbkey := config.BuildAddrValue(address)
	//valueBs, err := s.dbTemp.Find(dbkey)
	//if err != nil {
	//	return 0, err
	//}
	//value := utils.BytesToUint64(*valueBs)
	//engine.Log.Info("75行%s,%d", address.B58String(), value)
	return value, nil
}

// 获取地址可用余额
func (s *Storage) GetBalance(address crypto.AddressCoin) (*evmutils.Int, error) {
	obj, _ := s.getObj(address)
	if obj != nil {
		b := evmutils.New(int64(obj.Balance))
		return b, nil
	}
	dbkey := config.BuildAddrValue(address)
	valueBs, err := s.dbTemp.Find(dbkey)
	if err != nil {
		engine.Log.Info("87行")
		return evmutils.New(0), err
	}
	value := utils.BytesToUint64(*valueBs)
	obj.Balance = value
	obj.BalanceChange = true
	s.cacheObj[address.B58String()] = obj
	return evmutils.New(int64(value)), nil
}

// 获取合约代码
func (s *Storage) GetCode(address crypto.AddressCoin) ([]byte, error) {
	obj, err := s.getObj(address)
	if err != nil {
		return nil, err
	}
	if obj.Suicide {
		return nil, errors.New("contract  suicided")
	}
	return obj.Code, nil
}

// 获取合约代码大小
func (s *Storage) GetCodeSize(address crypto.AddressCoin) (*evmutils.Int, error) {
	obj, err := s.getObj(address)
	if err != nil {
		return evmutils.New(0), err
	}
	return evmutils.New(int64(len(obj.Code))), nil
}

// 获取合约代码hash
func (s *Storage) GetCodeHash(address crypto.AddressCoin) (*evmutils.Int, error) {
	obj, err := s.getObj(address)
	if err != nil {
		return evmutils.New(0), err
	}
	hash := evmutils.New(0)
	hash.SetBytes(obj.CodeHash)
	return hash, nil
}

// 根据区块高度获取区块hash
func (s *Storage) GetBlockHash(block *evmutils.Int) (*evmutils.Int, error) {
	hash := evmutils.New(0)
	//hashValue := mining.LoadBlockHashByHeight(block.Uint64())

	//bhash, err := s.db.Find([]byte(config.BlockHeight + strconv.Itoa(int(block.Int64()))))
	bhash, err := s.db.Find(append(config.BlockHeight, []byte(strconv.Itoa(int(block.Int64())))...))
	if err != nil {
		return hash, err
	}
	if bhash == nil || len(*bhash) == 0 {
		return hash, err
	}
	hash.SetBytes(*bhash)
	//hash.SetBytes(*hashValue)
	return hash, nil

}

// 获取区块链版本号
func (s *Storage) GetCurrentBlockVersion() uint32 {

	return uint32(0)
}

// 创建一个新的账户和相关的代码
func (s *Storage) CreateAddress(caller crypto.AddressCoin, tx environment.Transaction, addrType int32) crypto.AddressCoin {
	return s.CreateFixedAddress(caller, nil, tx, addrType)
	return nil
}

func (s *Storage) CreateFixedAddress(caller crypto.AddressCoin, salt *evmutils.Int, tx environment.Transaction, addrType int32) crypto.AddressCoin {
	addrPre := config.AddrPre
	if addrPre == "" {
		engine.Log.Info("设置地址前缀")
		addrPre = "TEST"
	}
	data := append(caller, tx.TxHash...)
	if salt != nil {
		data = append(data, salt.Bytes()...)
	}
	newAddr := crypto.BuildAddr(config.AddrPre, data)
	return newAddr
}

// 兼容eth，create action 时候根据调用合约的nonce生成地址
func (s *Storage) CreateAddressV1(caller, address crypto.AddressCoin, tx environment.Transaction, addrType int32) crypto.AddressCoin {
	obj, err := s.getObj(address)
	if err != nil {
		engine.Log.Warn("通过地址(%s)获取合约对象,%v", address.B58String(), err)
		return nil
	}
	nonceBs := new(big.Int).SetUint64(obj.Nonce + 1).Bytes()
	data := append(address, nonceBs...)
	newAddr := crypto.BuildAddr(config.AddrPre, data)
	s.setNonce(address, obj.Nonce+1)

	//engine.Log.Info("CreateAddressV1 caller:%s address:%s nonce:%d 创建合约地址:%s", caller.B58String(), address.B58String(), obj.Nonce, newAddr.B58String())
	return newAddr
}

// 兼容eth，create2 action 时候根据调用合约的salt生成地址
// 生成规则和eth保持一致
// keccak256(0xff ++ msg.sender ++ salt ++ keccak256(init_code))[12:]
func (s *Storage) CreateFixedAddressV1(caller, address crypto.AddressCoin, salt *evmutils.Int, tx environment.Transaction, addrType int32, hash []byte) crypto.AddressCoin {
	addr := common.Address(evmutils.AddressCoinToAddress(address))
	contractAddr := common.BytesToAddress(evmutils.Keccak256ByDatas([]byte{0xff}, addr.Bytes(), salt.Bytes(), hash)[12:])
	newAddr := evmutils.AddressToAddressCoin(contractAddr.Bytes())

	//engine.Log.Info("CreateFixedAddressV1 caller:%s address:%s new_address:%s 创建合约地址:%s", caller.B58String(), address.B58String(), hex.EncodeToString(contractAddr.Bytes()), newAddr.B58String())
	return newAddr
}

// 判断是否可以转账
func (s *Storage) CanTransfer(from crypto.AddressCoin, to crypto.AddressCoin, amount *evmutils.Int) bool {
	fromBalance, _ := s.GetBalance(from)
	return fromBalance.Cmp(amount.Int) >= 0
}

// 读取合约的变量
func (s *Storage) Load(n crypto.AddressCoin, k string) (*evmutils.Int, error) {
	obj, err := s.getObj(n)
	if err != nil {
		return evmutils.New(0), err
	}
	if value, ok := obj.CacheStorage[k]; ok {
		vi := evmutils.New(0)
		vi.SetBytes(value)
		return vi, nil
	}
	return evmutils.New(0), nil
}

// 存储合约的变量 ,key为name+::+key ,值为val
func (s *Storage) Store(address crypto.AddressCoin, key string, val *evmutils.Int) {
	obj, _ := s.getObj(address)
	obj.CacheStorage[key] = val.Bytes()
	s.upObj[address.B58String()] = struct{}{}
	s.cacheObj[address.B58String()] = obj

}

// 减余额
func (s *Storage) SubBalance(addr crypto.AddressCoin, value *evmutils.Int) {
	obj, _ := s.getObj(addr)
	oldBalance := evmutils.New(int64(obj.Balance))
	newBalance := oldBalance.Sub(value)
	obj.Balance = newBalance.Uint64()
	obj.BalanceChange = true
	s.upObj[addr.B58String()] = struct{}{}
	s.cacheObj[addr.B58String()] = obj
}

// 加余额
func (s *Storage) AddBalance(address crypto.AddressCoin, value *evmutils.Int) {
	obj, _ := s.getObj(address)
	oldBalance := evmutils.New(int64(obj.Balance))
	newBalance := oldBalance.Add(value)
	obj.Balance = newBalance.Uint64()
	obj.BalanceChange = true
	s.upObj[address.B58String()] = struct{}{}
	s.cacheObj[address.B58String()] = obj
}

// 设置合约代码
func (s *Storage) SetCode(address crypto.AddressCoin, code []byte) {
	obj, _ := s.getObj(address)
	obj.Code = code
	hash := evmutils.Keccak256(code)
	obj.CodeHash = hash
	s.upObj[address.B58String()] = struct{}{}
	s.cacheObj[address.B58String()] = obj
}

// 销毁合约
func (s *Storage) Suicide(address crypto.AddressCoin) {
	obj, _ := s.getObj(address)
	obj.Suicide = true
	s.upObj[address.B58String()] = struct{}{}
	s.cacheObj[address.B58String()] = obj
}

// 给某个账户退款
func (s *Storage) Refund(address crypto.AddressCoin, gasLeft uint64) {
	s.AddBalance(address, evmutils.New(int64(gasLeft)))
}

// 持久化提交
func (s *Storage) Commit() {
	//start := config.TimeNow()
	var count int64 = 0
	pool := make(chan struct{}, 10)
	wg := new(sync.WaitGroup)
	for k, v := range s.cacheObj {
		_, up := s.upObj[k]
		if !up {
			continue
		}
		wg.Add(1)
		pool <- struct{}{}
		go s.commitObject(v, wg, pool, &count)
	}
	wg.Wait()
	err := db2.DecContractCount(big.NewInt(count))
	if err != nil {
		engine.Log.Error("更新合约数量失败%s", err.Error())
	}

	//engine.Log.Info("保存合约数据耗时:%s", config.TimeNow().Sub(start))

}

// 持久化提交
func (s *Storage) commitObject(v *go_protos.StateObject, wg *sync.WaitGroup, pool chan struct{}, count *int64) {
	defer func() {
		wg.Done()
		<-pool
	}()
	if v.BalanceChange {
		//balanceKey := config.BuildAddrValue(v.Address)
		//balance := utils.Uint64ToBytes(v.Balance)
		addCoin := crypto.AddressCoin(v.Address)
		db2.SetNotSpendBalance(&addCoin, v.Balance)
		///s.dbTemp.Save(balanceKey, &balance)
	}

	if v.NonceChange {
		dbkey := config.BuildDBKeyAddrNonce(v.Address)
		nonceBs := new(big.Int).SetUint64(v.Nonce).Bytes()
		s.dbTemp.Save(*dbkey, &nonceBs)
	}

	//变更对象
	objKey := append([]byte(config.DBKEY_CONTRACT_OBJECT), v.Address...)
	utils.Log.Info().Hex("保存合约地址", v.Address).Hex("保存合约", objKey).Send()
	value, err := proto.Marshal(v)
	if err != nil {
		engine.Log.Error("编码失败")
	}
	err = s.dbTemp.Save(objKey, &value)
	if err != nil {
		coin := crypto.AddressCoin(v.Address)
		engine.Log.Error("保存账户对象失败", err.Error(), coin.B58String())
	}
	if v.Suicide {
		atomic.AddInt64(count, 1)
	}
}

// 获取地址Nonce
func (s *Storage) getNonce(address crypto.AddressCoin) (uint64, error) {
	dbkey := config.BuildDBKeyAddrNonce(address)
	valueBs, err := s.dbTemp.Find(*dbkey)
	if err != nil {
		return 0, err
	}
	nonce := new(big.Int).SetBytes(*valueBs)
	return nonce.Uint64(), nil
}

// 设置地址Nonce
func (s *Storage) setNonce(address crypto.AddressCoin, nonce uint64) {
	obj, _ := s.getObj(address)
	obj.Nonce = nonce
	obj.NonceChange = true
	s.upObj[address.B58String()] = struct{}{}
}
