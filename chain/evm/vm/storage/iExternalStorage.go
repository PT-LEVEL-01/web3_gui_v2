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

// Teh External Storage,provding a Storage for touching out of current evm
type IExternalStorage interface {
	SubBalance(coin crypto.AddressCoin, value *evmutils.Int)
	AddBalance(coin crypto.AddressCoin, value *evmutils.Int)
	GetBalance(address crypto.AddressCoin) (*evmutils.Int, error)
	SetCode(address crypto.AddressCoin, code []byte)
	GetCode(address crypto.AddressCoin) ([]byte, error)
	GetCodeSize(address crypto.AddressCoin) (*evmutils.Int, error)
	GetCodeHash(address crypto.AddressCoin) (*evmutils.Int, error)
	GetBlockHash(block *evmutils.Int) (*evmutils.Int, error)
	GetCurrentBlockVersion() uint32

	CreateAddress(caller crypto.AddressCoin, tx environment.Transaction, addrType int32) crypto.AddressCoin
	CreateFixedAddress(caller crypto.AddressCoin, salt *evmutils.Int, tx environment.Transaction, addrType int32) crypto.AddressCoin
	CreateAddressV1(caller, address crypto.AddressCoin, tx environment.Transaction, addrType int32) crypto.AddressCoin
	CreateFixedAddressV1(caller, address crypto.AddressCoin, salt *evmutils.Int, tx environment.Transaction, addrType int32, hash []byte) crypto.AddressCoin

	CanTransfer(from crypto.AddressCoin, to crypto.AddressCoin, amount *evmutils.Int) bool

	Load(n crypto.AddressCoin, k string) (*evmutils.Int, error)
	Store(address crypto.AddressCoin, key string, val *evmutils.Int)
	Suicide(address crypto.AddressCoin)
}
