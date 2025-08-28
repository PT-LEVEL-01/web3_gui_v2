/*
 * Copyright 2020 The SealEVM Authors
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package storage

import (
	"encoding/hex"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/vm/environment"
	"web3_gui/chain/evm/vm/utils"
	"web3_gui/keystore/adapter/crypto"
)

type Storage struct {
	ResultCache     ResultCache
	ExternalStorage IExternalStorage
	readOnlyCache   readOnlyCache
}

func New(extStorage IExternalStorage) *Storage {
	s := &Storage{
		ResultCache: ResultCache{
			//OriginalData: CacheUnderAddress{},
			CachedData:  CacheUnderAddress{},
			Balance:     BalanceCache{},
			Logs:        LogCache{},
			Destructs:   Cache{},
			InternalTxs: InternalTxCache{},
		},
		ExternalStorage: extStorage,
		readOnlyCache: readOnlyCache{
			Code:      CodeCache{},
			CodeSize:  Cache{},
			CodeHash:  Cache{},
			BlockHash: Cache{},
		},
	}

	return s
}

func (s *Storage) SLoad(n crypto.AddressCoin, k *evmutils.Int) (*evmutils.Int, error) {
	//fmt.Println("SLoad", n.String(), "k", k.String())
	//if s.ResultCache.OriginalData == nil || s.ResultCache.CachedData == nil || s.ExternalStorage == nil {
	if s.ResultCache.CachedData == nil || s.ExternalStorage == nil {
		return nil, utils.ErrStorageNotInitialized
	}

	//nsStr := hex.EncodeToString(n.Bytes())
	nsStr := n.B58String()
	keyStr := hex.EncodeToString(k.Bytes())
	//nsStr := n.String()
	//keyStr := k.String()

	var err error = nil
	i := s.ResultCache.CachedData.Get(nsStr, keyStr)
	if i == nil {
		i, err = s.ExternalStorage.Load(n, keyStr)
		if err != nil {
			return nil, utils.NoSuchDataInTheStorage(err)
		}

		//s.ResultCache.OriginalData.Set(nsStr, keyStr, i)
		if s.GetCurrentBlockVersion() != 220 {
			s.ResultCache.CachedData.Set(nsStr, keyStr, i)
		}

	}

	return i, nil
}

func (s *Storage) SStore(n crypto.AddressCoin, k, v *evmutils.Int) {
	//nsStr := hex.EncodeToString(n.Bytes())
	nsStr := n.B58String()
	keyStr := hex.EncodeToString(k.Bytes())
	s.ResultCache.CachedData.Set(nsStr, keyStr, v)
	s.ExternalStorage.Store(n, keyStr, v)
	//fmt.Println("SStore", n.String(), "k", k.String(), "v", v.String())
}

func (s *Storage) BalanceModify(address crypto.AddressCoin, value *evmutils.Int, neg bool) {
	//kString := address.String()
	//kString := hex.EncodeToString(address.Bytes())
	kString := address.B58String()
	b, exist := s.ResultCache.Balance[kString]
	if !exist {
		//b = &balance{
		//	Address: evmutils.FromBigInt(address.Int),
		//	Balance: evmutils.New(0),
		//}
		b = &balance{
			Address: address,
			Balance: evmutils.New(0),
		}

		s.ResultCache.Balance[kString] = b
	}

	if neg {
		b.Balance.Int.Sub(b.Balance.Int, value.Int)
		s.ExternalStorage.SubBalance(address, value)
	} else {
		b.Balance.Int.Add(b.Balance.Int, value.Int)
		s.ExternalStorage.AddBalance(address, value)
	}
}

// Log指令,以太坊中在区块插入的时候写入并发送消息给订阅者
func (s *Storage) Log(address crypto.AddressCoin, topics [][]byte, data []byte, context environment.Context) {
	//kString := address.String()
	//kString := hex.EncodeToString(address.Bytes())
	kString := address.B58String()
	var theLog = Log{
		Topics:  topics,
		Data:    data,
		Context: context,
	}
	l := s.ResultCache.Logs[kString]
	s.ResultCache.Logs[kString] = append(l, theLog)

	return
}

// address 合约地址
func (s *Storage) Destruct(address crypto.AddressCoin) {
	//s.ResultCache.Destructs[address.String()] = address
	key := address.B58String()
	//s.ResultCache.Destructs[hex.EncodeToString(address.Bytes())] = address
	addressInt := evmutils.New(0)
	addressN := evmutils.AddressCoinToAddress(address)
	addressInt.SetBytes(addressN.GetBytes())
	s.ResultCache.Destructs[key] = addressInt
	s.ExternalStorage.Suicide(address)
}

type commonGetterFunc func(*evmutils.Int) (*evmutils.Int, error)

func (s *Storage) commonGetter(key *evmutils.Int, cache Cache, getterFunc commonGetterFunc) (*evmutils.Int, error) {
	//keyStr := key.String()
	keyStr := hex.EncodeToString(key.Bytes())
	if b, exists := cache[keyStr]; exists {
		return evmutils.FromBigInt(b.Int), nil
	}

	b, err := getterFunc(key)
	if err == nil {
		cache[keyStr] = b
	}

	return b, err
}

func (s *Storage) Balance(address crypto.AddressCoin) (*evmutils.Int, error) {
	return s.ExternalStorage.GetBalance(address)
}
func (s *Storage) SetCode(address crypto.AddressCoin, code []byte) {
	//keyStr := address.String()
	//keyStr := hex.EncodeToString(address.Bytes())
	keyStr := address.B58String()
	s.readOnlyCache.Code[keyStr] = code
	s.ExternalStorage.SetCode(address, code)
}
func (s *Storage) GetCode(address crypto.AddressCoin) ([]byte, error) {
	//keyStr := address.String()
	//keyStr := hex.EncodeToString(address.Bytes())
	keyStr := address.B58String()

	if b, exists := s.readOnlyCache.Code[keyStr]; exists {
		return b, nil
	}

	b, err := s.ExternalStorage.GetCode(address)
	if err == nil {
		s.readOnlyCache.Code[keyStr] = b
	}

	return b, err
}
func (s *Storage) SetCodeSize(address crypto.AddressCoin, size *evmutils.Int) {
	//keyStr := address.String()
	//keyStr := hex.EncodeToString(address.Bytes())
	keyStr := address.B58String()
	s.readOnlyCache.CodeSize[keyStr] = size
}
func (s *Storage) GetCodeSize(address crypto.AddressCoin) (*evmutils.Int, error) {
	//keyStr := address.String()
	//keyStr := hex.EncodeToString(address.Bytes())
	keyStr := address.B58String()
	if size, exists := s.readOnlyCache.CodeSize[keyStr]; exists {
		return size, nil
	}

	size, err := s.ExternalStorage.GetCodeSize(address)
	if err == nil {
		s.readOnlyCache.CodeSize[keyStr] = size
	}

	return size, err
}
func (s *Storage) SetCodeHash(address crypto.AddressCoin, codeHash *evmutils.Int) {
	//keyStr := address.String()
	//keyStr := hex.EncodeToString(address.Bytes())
	keyStr := address.B58String()
	s.readOnlyCache.CodeHash[keyStr] = codeHash
}
func (s *Storage) GetCodeHash(address crypto.AddressCoin) (*evmutils.Int, error) {
	//keyStr := address.String()
	//keyStr := hex.EncodeToString(address.Bytes())
	keyStr := address.B58String()
	if hash, exists := s.readOnlyCache.CodeHash[keyStr]; exists {
		return hash, nil
	}

	hash, err := s.ExternalStorage.GetCodeHash(address)
	if err == nil {
		s.readOnlyCache.CodeHash[keyStr] = hash
	}

	return hash, err
}

func (s *Storage) GetBlockHash(block *evmutils.Int) (*evmutils.Int, error) {
	//keyStr := block.String()
	keyStr := hex.EncodeToString(block.Bytes())
	if hash, exists := s.readOnlyCache.BlockHash[keyStr]; exists {
		return hash, nil
	}

	hash, err := s.ExternalStorage.GetBlockHash(block)
	if err == nil {
		s.readOnlyCache.BlockHash[keyStr] = hash
	}

	return hash, err
}

func (s *Storage) GetCurrentBlockVersion() uint32 {
	return s.ExternalStorage.GetCurrentBlockVersion()
}

func (s *Storage) CreateAddress(caller crypto.AddressCoin, tx environment.Transaction, addrType int32) crypto.AddressCoin {
	return s.ExternalStorage.CreateAddress(caller, tx, addrType)
}

func (s *Storage) CreateFixedAddress(caller crypto.AddressCoin, salt *evmutils.Int, tx environment.Transaction, addrType int32) crypto.AddressCoin {
	return s.ExternalStorage.CreateFixedAddress(caller, salt, tx, addrType)
}

func (s *Storage) CreateAddressV1(caller, address crypto.AddressCoin, tx environment.Transaction, addrType int32) crypto.AddressCoin {
	return s.ExternalStorage.CreateAddressV1(caller, address, tx, addrType)
}

func (s *Storage) CreateFixedAddressV1(caller, address crypto.AddressCoin, salt *evmutils.Int, tx environment.Transaction, addrType int32, hash []byte) crypto.AddressCoin {
	return s.ExternalStorage.CreateFixedAddressV1(caller, address, salt, tx, addrType, hash)
}

func (s *Storage) CanTransfer(from crypto.AddressCoin, to crypto.AddressCoin, amount *evmutils.Int) bool {
	return s.ExternalStorage.CanTransfer(from, to, amount)
}
func (s *Storage) Transfer(from crypto.AddressCoin, to crypto.AddressCoin, amount *evmutils.Int) {
	s.SubBalance(from, amount)
	s.AddBalance(to, amount)
}
func (s *Storage) SubBalance(address crypto.AddressCoin, value *evmutils.Int) {
	oldBalance, _ := s.Balance(address)
	newBalance := oldBalance.Sub(value)
	s.ResultCache.Balance[address.B58String()] = &balance{Address: address, Balance: newBalance}
	s.ExternalStorage.SubBalance(address, value)
}
func (s *Storage) AddBalance(address crypto.AddressCoin, value *evmutils.Int) {
	oldBalance, _ := s.Balance(address)
	newBalance := oldBalance.Add(value)
	s.ResultCache.Balance[address.B58String()] = &balance{Address: address, Balance: newBalance}
	s.ExternalStorage.AddBalance(address, value)
}

func (s *Storage) InternalTx(contract crypto.AddressCoin, txHash []byte, from crypto.AddressCoin, to crypto.AddressCoin, amount *evmutils.Int) {
	kString := contract.B58String()
	var tx = InternalTx{
		TxHash:   hex.EncodeToString(txHash),
		Contract: contract.B58String(),
		From:     from.B58String(),
		To:       to.B58String(),
		Value:    amount.String(),
	}
	l := s.ResultCache.InternalTxs[kString]
	s.ResultCache.InternalTxs[kString] = append(l, tx)
	return
}
