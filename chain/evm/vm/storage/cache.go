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
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/vm/environment"
	"web3_gui/keystore/adapter/crypto"
)

type Cache map[string]*evmutils.Int
type CacheUnderAddress map[string]Cache

func (c CacheUnderAddress) Get(address string, key string) *evmutils.Int {
	if c[address] == nil {
		return nil
	} else {
		return c[address][key]
	}
}

func (c CacheUnderAddress) Set(address string, key string, v *evmutils.Int) {
	if c[address] == nil {
		c[address] = Cache{}
	}

	c[address][key] = v
}

type balance struct {
	Address crypto.AddressCoin
	Balance *evmutils.Int
}

type BalanceCache map[string]*balance

type Log struct {
	Topics  [][]byte
	Data    []byte
	Context environment.Context
}

type LogCache map[string][]Log

type InternalTx struct {
	TxHash   string `json:"txHash"`
	Contract string `json:"contract"`
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
}

type InternalTxCache map[string][]InternalTx

type ResultCache struct {
	//OriginalData CacheUnderAddress
	CachedData CacheUnderAddress

	Balance   BalanceCache
	Logs      LogCache
	Destructs Cache

	InternalTxs InternalTxCache
}

type CodeCache map[string][]byte

type readOnlyCache struct {
	Code      CodeCache
	CodeSize  Cache
	CodeHash  Cache
	BlockHash Cache
}

func MergeResultCache2211(src *ResultCache, to *ResultCache) {
	// fix bug for multi cross call of contract
	for k, v := range src.CachedData {
		vTo, exist := to.CachedData[k]
		if !exist {
			to.CachedData[k] = v
		} else {
			vSrc := (map[string]*evmutils.Int)(v)
			for vKSrc, vVSrc := range vSrc {
				vTo[vKSrc] = vVSrc
			}
		}
	}

	for k, v := range src.Balance {
		if to.Balance[k] != nil {
			to.Balance[k].Balance.Add(v.Balance)
		} else {
			to.Balance[k] = v
		}
	}

	for k, v := range src.Logs {
		to.Logs[k] = append(to.Logs[k], v...)
	}

	for k, v := range src.Destructs {
		to.Destructs[k] = v
	}

	for k, v := range src.InternalTxs {
		to.InternalTxs[k] = append(to.InternalTxs[k], v...)
	}
}

// 合并缓存
func MergeResultCache(src *ResultCache, to *ResultCache) {
	//for k, v := range src.OriginalData {
	//	to.OriginalData[k] = v
	//}

	for k, v := range src.CachedData {
		to.CachedData[k] = v
	}

	for k, v := range src.Balance {
		if to.Balance[k] != nil {
			to.Balance[k].Balance.Add(v.Balance)
		} else {
			to.Balance[k] = v
		}
	}

	for k, v := range src.Logs {
		to.Logs[k] = append(to.Logs[k], v...)
	}

	for k, v := range src.Destructs {
		to.Destructs[k] = v
	}

	for k, v := range src.InternalTxs {
		to.InternalTxs[k] = append(to.InternalTxs[k], v...)
	}
}
