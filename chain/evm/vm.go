package evm

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/storage"
	"web3_gui/chain/evm/vm"
	"web3_gui/chain/evm/vm/environment"
	"web3_gui/chain/evm/vm/opcodes"
	storage2 "web3_gui/chain/evm/vm/storage"
	"web3_gui/chain/protos/go_protos"

	"github.com/golang/protobuf/proto"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

type VmRun struct {
	from           crypto.AddressCoin
	to             crypto.AddressCoin
	txHash         []byte
	timestamp      int64
	height         uint64
	block          *environment.Block
	contractEvents []*go_protos.ContractEvent
	s              *storage.Storage
	gasLimit       uint64
	isReward       bool
	payload        []byte
}

func NewRewardVmRun(from, to crypto.AddressCoin, txhash []byte, block *environment.Block) *VmRun {
	if !bytes.Equal(txhash[:8], utils.Uint64ToBytes(uint64(config.Wallet_tx_type_mining))) {
		return nil
	}

	vmRun := NewVmRun(from, to, txhash, block)
	vmRun.isReward = true
	return vmRun
}

func NewVmRun(from, to crypto.AddressCoin, txhash []byte, block *environment.Block) *VmRun {
	return &VmRun{
		from:           from,
		to:             to,
		txHash:         txhash,
		block:          block,
		contractEvents: []*go_protos.ContractEvent{},
	}
}

func NewCountVmRun(block *environment.Block) *VmRun {
	return &VmRun{
		block:          block,
		contractEvents: []*go_protos.ContractEvent{},
	}
}

var (
	//出块奖励
	//DISTRIBUTE_ID = "d0558c34"
	DISTRIBUTE_ID = []byte{208, 85, 140, 52}

	// 全局缓存合约对象
	// 地址:合约对象
	globalCacheContractObj = make(map[string]*go_protos.StateObject)
)

// 奖励合约地址
var RewardContract crypto.AddressCoin

func InitRewardContract() {
	RewardContract = evmutils.AddressToAddressCoin(common.HexToAddress(config.RewardContractAddr).Bytes())
}

// 验证是否出块奖励
func (v *VmRun) checkMiningReward(payload []byte) bool {
	if !v.isReward && bytes.Equal(v.to, RewardContract) {
		if len(payload) < 4 {
			return true
		}
		return !bytes.Equal(payload[:4], DISTRIBUTE_ID)
		//return hex.EncodeToString(payload[:4]) != DISTRIBUTE_ID
	}
	return true
}

// 仅当address不为nil时,缓存合约对象
func (v *VmRun) SetStorage(addressess ...crypto.AddressCoin) {
	if addressess == nil || len(addressess) == 0 {
		v.s = storage.NewStorage()
	} else {
		v.s = storage.NewStorageWithCache(addressess...)
	}
}

func (v *VmRun) SetBlock(b *environment.Block) {
	v.block = b
}

// 缓存全局合约对象
func (v *VmRun) CacheStorage(address crypto.AddressCoin) {
	if obj, ok := v.s.GetCacheObj(address); ok {
		globalCacheContractObj[address.B58String()] = obj
	}
}

// 更新缓存对象的余额
func (v *VmRun) UpdateCacheObjBalance(address crypto.AddressCoin, value uint64) {
	if !v.isReward {
		return
	}

	if _, ok := v.s.GetCacheObj(address); ok {
		v.s.UpdateCacheObjBalance(address, evmutils.New(int64(value)))
	}
}

func (v *VmRun) SetReward(t bool) {
	v.isReward = t
}

func (v *VmRun) SetTxContext(from, to crypto.AddressCoin, txhash []byte, timestamp int64, height uint64, difficulty, gasLimit *evmutils.Int) {
	v.from = from
	v.to = to
	v.txHash = txhash
	v.timestamp = timestamp
	v.height = height
	if v.block != nil {
		v.block.Difficulty = difficulty
		v.block.GasLimit = gasLimit
	}
}

func Run(payload []byte, gas, gasPrice, value uint64, isCreate bool, v *VmRun) (vm.ExecuteResult, []*go_protos.ContractEvent, error) {
	if !v.checkMiningReward(payload) {
		return vm.ExecuteResult{}, nil, errors.New("mining reward tx is warn")
	}
	//engine.Log.Info("evm.storage  %p", v.s)
	var contract environment.Contract
	//var block environment.Block
	//构建消息对象
	message := environment.Message{
		Caller: v.from,
		Value:  evmutils.New(int64(value)),
	}

	//构建交易对象
	evmTransaction := environment.Transaction{
		TxHash:   v.txHash,
		Origin:   v.from, // creator address
		GasPrice: evmutils.New(int64(gasPrice)),
		GasLimit: evmutils.New(int64(gas)),
	}
	v.gasLimit = gas
	//构建合约对象
	if isCreate {
		contractHash, _ := v.s.GetCodeHash(v.to)
		if contractHash.Cmp(evmutils.New(0).Int) != 0 {
			return vm.ExecuteResult{}, nil, errors.New("contract address collision")
		}

		hash := evmutils.Keccak256(payload)
		i := evmutils.New(0)
		i.SetBytes(hash)
		contract = environment.Contract{
			Address: v.to,
			Code:    payload,
			Hash:    i,
		}
	} else {
		code, err := v.s.GetCode(v.to)
		if err != nil {
			return vm.ExecuteResult{}, nil, err
		}
		if len(code) == 0 {
			return vm.ExecuteResult{}, nil, errors.New("contract code is null")
		}
		codeHash, _ := v.s.GetCodeHash(v.to)
		contract = environment.Contract{
			Address: v.to,
			Code:    code,
			Hash:    codeHash,
		}
		message.Data = payload
		v.payload = payload
	}
	var call vm.EVMResultCallback
	call = v.callback
	//构建区块对象
	if v.block == nil {
		v.block = defaultBlock(v.from, v.timestamp, v.height)
		call = v.callbackReadOnly
	}

	evmObj := vm.New(vm.EVMParam{
		MaxStackDepth:  1024,
		ExternalStore:  v.s,
		ResultCallback: call,
		Context: &environment.Context{
			Block:       *v.block,
			Contract:    contract,
			Transaction: evmTransaction, //影响evm解释器的gasRemaining 来源为transaction的GasLimit
			Message:     message,
			Parameters:  nil,
			Cfg:         environment.Config{},
		},
	})
	result, err := evmObj.ExecuteContract(isCreate, 0)
	if err != nil {
		return vm.ExecuteResult{}, nil, err
	}
	return result, v.contractEvents, nil
}

func (v *VmRun) Run(payload []byte, gas, gasPrice, value uint64, isCreate bool) (vm.ExecuteResult, []*go_protos.ContractEvent, error) {
	if !v.checkMiningReward(payload) {
		return vm.ExecuteResult{}, nil, errors.New("mining reward tx is warn")
	}

	var contract environment.Contract
	//var block environment.Block
	//构建消息对象
	message := environment.Message{
		Caller: v.from,
		Value:  evmutils.New(int64(value)),
	}
	s := storage.NewStorage()
	v.s = s
	//构建交易对象
	evmTransaction := environment.Transaction{
		TxHash:   v.txHash,
		Origin:   v.from, // creator address
		GasPrice: evmutils.New(int64(gasPrice)),
		GasLimit: evmutils.New(int64(gas)),
	}
	v.gasLimit = gas
	//构建合约对象
	if isCreate {
		contractHash, _ := v.s.GetCodeHash(v.to)
		if contractHash.Cmp(evmutils.New(0).Int) != 0 {
			return vm.ExecuteResult{}, nil, errors.New("contract address collision")
		}

		hash := evmutils.Keccak256(payload)
		i := evmutils.New(0)
		i.SetBytes(hash)
		contract = environment.Contract{
			Address: v.to,
			Code:    payload,
			Hash:    i,
		}
	} else {
		code, err := s.GetCode(v.to)
		if err != nil {
			return vm.ExecuteResult{}, nil, err
		}
		if len(code) == 0 {
			return vm.ExecuteResult{}, nil, errors.New("contract code is null")
		}
		codeHash, _ := s.GetCodeHash(v.to)
		contract = environment.Contract{
			Address: v.to,
			Code:    code,
			Hash:    codeHash,
		}
		message.Data = payload
		v.payload = payload
	}
	var call vm.EVMResultCallback
	call = v.callback
	//构建区块对象
	if v.block == nil {
		v.block = defaultBlock(v.from, v.timestamp, v.height)
		call = v.callbackReadOnly
	}

	evmObj := vm.New(vm.EVMParam{
		MaxStackDepth:  1024,
		ExternalStore:  s,
		ResultCallback: call,
		Context: &environment.Context{
			Block:       *v.block,
			Contract:    contract,
			Transaction: evmTransaction, //影响evm解释器的gasRemaining 来源为transaction的GasLimit
			Message:     message,
			Parameters:  nil,
			Cfg:         environment.Config{},
		},
	})
	result, err := evmObj.ExecuteContract(isCreate, 0)
	if err != nil {
		return vm.ExecuteResult{}, nil, err
	}
	return result, v.contractEvents, nil
}
func (v *VmRun) callback(result vm.ExecuteResult, err error) {
	if result.ExitOpCode == opcodes.REVERT || err != nil {
		return
	}

	//保存合约部署代码
	if len(result.ByteCodeHead) > 0 && len(result.ByteCodeBody) > 0 {
		v.s.SetCode(v.to, result.ResultData)
	}
	if result.GasLeft > 0 {
		//v.s.Refund(v.from, result.GasLeft)
	}
	//持久化合约状态
	v.s.Commit()

	//save tx gas_used
	gasUsedKey := append([]byte(config.DBKEY_TxCONTRACT_GASUSED), v.txHash...)
	gasUsedBt := utils.Uint64ToBytes(v.gasLimit - result.GasLeft)
	err = db.LevelDB.Save(gasUsedKey, &gasUsedBt)
	if err != nil {
		engine.Log.Error("save contract_tx gas used error: %s", err.Error())
	}
	contractEvents, err := v.emitContractEvent(result)
	if err != nil {
		engine.Log.Error(err.Error())
	}
	if len(contractEvents) > 0 {
		v.contractEvents = contractEvents
		//保存事件日志,key为区块高度,字段为交易hash,值为事件
		key := append([]byte(config.DBKEY_BLOCK_EVENT), v.block.Number.Bytes()...)
		var eventsInfo []*go_protos.ContractEventInfo
		for _, event := range contractEvents {
			eventInfo := &go_protos.ContractEventInfo{
				BlockHeight:     v.block.Number.Int64(),
				BlockHash:       hex.EncodeToString(v.block.Hash.Bytes()),
				Topic:           event.Topic,
				TxId:            event.TxId,
				ContractAddress: event.ContractAddress,
				EventData:       event.EventData,
			}
			eventsInfo = append(eventsInfo, eventInfo)
		}
		list := &go_protos.ContractEventInfoList{ContractEvents: eventsInfo}
		listBytes, err := proto.Marshal(list)
		if err != nil {
			engine.Log.Error(err.Error())
		}
		_, err = db.LevelDB.GetDB().HSet(key, v.txHash, listBytes)
		if err != nil {
			engine.Log.Error("保存合约事件日志失败%s", err.Error())
		}
	}

	if _, err := v.emitContractInternalTx(result); err != nil {
		engine.Log.Error("保存合约内部交易失败%s", err.Error())
	}
}
func (v *VmRun) checkTopicStr(topicHex string) error {
	topicLen := len(topicHex)
	if topicLen == 0 {
		return fmt.Errorf("topic can not empty")
	}
	if topicLen > 255 {
		return fmt.Errorf("topic too long")
	}
	return nil
}

func (v *VmRun) emitContractInternalTx(result vm.ExecuteResult) ([]storage2.InternalTx, error) {
	itxMap := result.StorageCache.InternalTxs

	var itxs []storage2.InternalTx
	for _, txs := range itxMap {
		for _, tx := range txs {
			itxs = append(itxs, tx)
		}
	}

	key := append([]byte(config.DBKEY_INTERNAL_TX), v.block.Number.Bytes()...)
	itxsBs, err := json.Marshal(&itxs)
	if err != nil {
		return nil, err
	}
	_, err = db.LevelDB.GetDB().HSet(key, v.txHash, itxsBs)
	if err != nil {
		return nil, err
	}

	return itxs, nil
}

func (v *VmRun) emitContractEvent(result vm.ExecuteResult) ([]*go_protos.ContractEvent, error) {
	logsMap := result.StorageCache.Logs
	//txHash := string(v.txHash)
	txHash := hex.EncodeToString(v.txHash)
	//contractAddr := v.to.B58String()
	var contractEvents []*go_protos.ContractEvent
	for _, logs := range logsMap {
		for _, log := range logs {
			if len(log.Topics) > 15 {
				return nil, fmt.Errorf("事件数据太多")
			}
			contractEvent := &go_protos.ContractEvent{
				TxId: txHash,
				//ContractAddress: contractAddr,
				ContractAddress: log.Context.Contract.Address.B58String(),
			}
			topics := log.Topics
			for index, topic := range topics {
				//将第一个topic赋值给contractEvent的topic,第一个主题通常为事件名称及其参数类型签名
				if index == 0 && topic != nil {
					topicHexStr := hex.EncodeToString(topic)
					if err := v.checkTopicStr(topicHexStr); err != nil {
						return nil, err
					}
					contractEvent.Topic = topicHexStr
					continue
				}
				//参数声明为 'indexed' 被视为主题
				topicIndexHexStr := hex.EncodeToString(topic)
				contractEvent.EventData = append(contractEvent.EventData, topicIndexHexStr)
			}
			data := log.Data
			dataHexStr := hex.EncodeToString(data)
			if len(dataHexStr) > 65535 {
				return nil, fmt.Errorf("事件数据太长")
			}
			//EventData最后一个元素是log，前面的都是主题
			contractEvent.EventData = append(contractEvent.EventData, dataHexStr)

			contractEvents = append(contractEvents, contractEvent)
		}
	}
	return contractEvents, nil
}
func (v *VmRun) callbackReadOnly(result vm.ExecuteResult, err error) {
	if err != nil {
		fmt.Println("callback err info:", err)
	}

	//fmt.Println("callback result info ResultData:", result.ResultData)
	//fmt.Println("callback result info ResultData:", hex.EncodeToString(result.ResultData))
	//fmt.Println("callback result info GasLeft:", result.GasLeft)
	//fmt.Println("callback result info StorageCache:", result.StorageCache)
	//fmt.Println("callback result info ExitOpCode:", result.ExitOpCode)
	//fmt.Println("callback result info ByteCodeHead:", result.ByteCodeHead)
	//fmt.Println("callback result info ByteCodeBody:", result.ByteCodeBody)

}
func defaultBlock(from crypto.AddressCoin, timestamp int64, height uint64) *environment.Block {
	return &environment.Block{
		Coinbase:   from,
		Timestamp:  evmutils.New(timestamp),
		Number:     evmutils.New(int64(height)),
		Difficulty: evmutils.New(0),
		GasLimit:   evmutils.New(1e10),
	}
}
