package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/vm"
	"web3_gui/chain/evm/vm/opcodes"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
合约交易
*/
type Tx_Contract struct {
	TxBase
	Action        string `json:"action"` //create or call 表示创建交易还是调用交易
	GzipSource    []byte //压缩的源代码
	GzipAbi       []byte //压缩abi
	ContractClass uint64 `json:"contract_class"` //合约类型
	GasPrice      uint64 `json:"gas_price"`      //燃料价格

	Bloom []byte `json:"bloom"` //bloom过滤器
}
type Tx_Contract_VO struct {
	TxBaseVO
	Action   string `json:"action"`
	GasUsed  uint64 `json:"gas_used"`
	GasPrice uint64 `json:"gas_price"`
}

/*
用于地址和txid格式化显示
*/
func (this *Tx_Contract) GetVOJSON() interface{} {
	contract := Tx_Contract_VO{
		TxBaseVO: this.TxBase.ConversionVO(),
		Action:   this.Action,
		GasUsed:  GetContractTxGasUsed(this.Hash),
		GasPrice: this.GasPrice,
	}
	contract.Payload = hex.EncodeToString([]byte(contract.Payload))
	return contract
}

func (this *Tx_Contract) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	buf.Write([]byte(this.Action))
	buf.Write(this.GzipSource)
	buf.Write(utils.Uint64ToBytes(this.ContractClass))
	buf.Write(utils.Uint64ToBytes(this.GasPrice))
	*bs = buf.Bytes()
	return bs
}

/*
构建hash值得到交易id
*/
func (this *Tx_Contract) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}

	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_contract)
	// jsonBs, _ := this.Json()

	// fmt.Println("序列化输出 111", string(*jsonBs))
	// fmt.Println("序列化输出 222", len(*bs), hex.EncodeToString(*bs))
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
格式化成[]byte
*/
func (this *Tx_Contract) Proto() (*[]byte, error) {
	vins := make([]*go_protos.Vin, 0)
	for _, one := range this.Vin {
		vinOne := &go_protos.Vin{
			// Txid: one.Txid,
			// Vout: one.Vout,
			Puk:  one.Puk,
			Sign: one.Sign,
			// Nonce: one.Nonce.Bytes(),
		}
		// if len(one.Nonce.Bytes()) == 0 {
		vinOne.Nonce = one.Nonce.Bytes()
		// }
		vins = append(vins, vinOne)
	}
	vouts := make([]*go_protos.Vout, 0)
	for _, one := range this.Vout {
		vouts = append(vouts, &go_protos.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
			Domain:       one.Domain,
			DomainType:   one.DomainType,
		})
	}
	txBase := go_protos.TxBase{
		Hash:       this.Hash,
		Type:       this.Type,
		VinTotal:   this.Vin_total,
		Vin:        vins,
		VoutTotal:  this.Vout_total,
		Vout:       vouts,
		Gas:        this.Gas,
		LockHeight: this.LockHeight,
		Payload:    this.Payload,
		BlockHash:  this.BlockHash,
		GasUsed:    this.GasUsed,
		Comment:    this.Comment,
	}

	txPay := go_protos.TxContract{
		TxBase:        &txBase,
		Action:        this.Action,
		GzipSource:    this.GzipSource,
		ContractClass: this.ContractClass,
		GasPrice:      this.GasPrice,
		GzipAbi:       this.GzipAbi,
	}
	// txPay.Marshal()
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
获取签名
*/
func (this *Tx_Contract) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)
	*signDst = append(*signDst, []byte(this.Action)...)
	*signDst = append(*signDst, this.GzipSource...)
	*signDst = append(*signDst, this.GzipAbi...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.ContractClass)...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.GasPrice)...)
	//engine.Log.Info("验证签名前的字节:%d %s", len(*signDst), hex.EncodeToString(*signDst))
	sign := keystore.Sign(*key, *signDst)
	return &sign
}

/*
验证是否合法
*/
func (this *Tx_Contract) CheckSign() error {
	// fmt.Println("开始验证交易合法性 Tx_Contract")
	if len(this.Vin) != 1 {
		return config.ERROR_pay_vin_too_much
	}
	// 特殊处理云存储合约类型
	if len(this.Vin[0].Nonce.Bytes()) == 0 && this.ContractClass != uint64(config.CLOUD_STORAGE_PROXY_CONTRACT) {
		// engine.Log.Info("txid:%s nonce is nil", txItr.GetHash())
		return config.ERROR_pay_nonce_is_nil
	}
	if this.Vout_total > config.Mining_pay_vout_max {
		return config.ERROR_pay_vout_too_much
	}

	if len(this.Vin) > 1 {
		return config.ERROR_pay_vin_too_much
	}
	//不能出现余额为0的转账
	//for i, one := range this.Vout {
	//	if i != 0 && one.Value <= 0 {
	//		return config.ERROR_amount_zero
	//	}
	//}

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费;3.输入不能重复。
	// vinMap := make(map[string]int)
	// inTotal := uint64(0)
	for i, one := range this.Vin {

		sign := this.GetSignSerialize(nil, uint64(i))
		*sign = append(*sign, []byte(this.Action)...)
		*sign = append(*sign, this.GzipSource...)
		*sign = append(*sign, this.GzipAbi...)
		*sign = append(*sign, utils.Uint64ToBytes(this.ContractClass)...)
		*sign = append(*sign, utils.Uint64ToBytes(this.GasPrice)...)
		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
		}
		// fmt.Printf("txid:%x puk:%x sign:%x", md5.Sum(one.Txid), md5.Sum(one.Puk), md5.Sum(one.Sign))
		if !ed25519.Verify(puk, *sign, one.Sign) {

			// engine.Log.Debug("ERROR_sign_fail 222 %s %d", hex.EncodeToString(one.Txid), one.Vout)
			// engine.Log.Debug("ed25519.Verify: puk: %x; waitSignData: %x; sign: %x\n", puk, *sign, one.Sign)
			// engine.Log.Debug("ERROR_sign_fail 222 %s %d", hex.EncodeToString(one.Txid), one.Vout)
			return config.ERROR_sign_fail
		}
		// engine.Log.Debug("CheckBase 5555555555555555 %s", config.TimeNow().Sub(start))

	}
	//
	//if err := this.TxBase.CheckBase(); err != nil {
	//	return err
	//}
	//for _, vout := range this.Vout {
	//	if vout.Value <= 0 {
	//		return config.ERROR_amount_zero
	//	}
	//}

	return nil
}
func (this *Tx_Contract) GetGas() uint64 {
	return this.Gas * this.GasPrice
}
func (this *Tx_Contract) GetGasLimit() uint64 {
	return this.Gas
}
func (this *Tx_Contract) GetGasUsed() uint64 {
	return this.GasUsed
}

/*
获取本交易总共花费的余额
*/
func (this *Tx_Contract) GetSpend() uint64 {
	spend := this.Gas * this.GasPrice
	for _, vout := range this.Vout {
		spend += vout.Value
	}
	return spend
}

/*
是否验证通过
*/
func (this *Tx_Contract) CheckRepeatedTx(txs ...TxItr) bool {
	//判断是否出现双花
	// return this.MultipleExpenditures(txs...)
	//需要判断交易发起方及其nonce
	input, _ := this.GetContractInfo()
	delVoteTotal := big.NewInt(0) //取消的总票数

	//验证域名重复注册
	//var regrec *ens.RegisterRecord
	//thisName := ""
	//if hex.EncodeToString(this.Payload[:4]) == ens.REGISTER_ID {
	//	regrec = ens.GetEnsRegisterRecord()
	//	if regrec == nil {
	//		return false
	//	}
	//	thisName = regrec.GetRegisterName(this.GetPayload())
	//	if thisName == "" {
	//		return false
	//	}
	//}

	for _, tx := range txs {
		//engine.Log.Info("248行%s", hex.EncodeToString(*tx.GetHash()))
		if tx.Class() != config.Wallet_tx_type_contract {
			continue
		}
		txContract, ok := tx.(*Tx_Contract)
		if !ok {
			continue
		}
		nonce := txContract.Vin[0].Nonce
		if input.Nonce.Cmp(&nonce) == 0 &&
			bytes.Equal(*input.GetPukToAddr(), *txContract.Vin[0].GetPukToAddr()) {
			return false
		}
		addrSelf := this.Vin[0].GetPukToAddr()
		vout := *tx.GetVout()
		//如果是奖励合约，判断是否重复质押,重复取消质押，以及临界值取消投票
		if bytes.Equal(vout[0].Address, precompiled.RewardContract) {
			vin := *tx.GetVin()
			//如果是同一个地址调用，则进行验证
			if bytes.Equal(*addrSelf, *vin[0].GetPukToAddr()) {
				//
				//engine.Log.Info("268行")
				myVoteType, _, myIsAdd := precompiled.UnpackPayloadV1(this.Payload)
				//加票操作，判断重复质押
				if myVoteType > 0 && myIsAdd {
					voteType, _, _ := precompiled.UnpackPayloadV1(tx.GetPayload())
					if voteType > 0 {
						//engine.Log.Info("273行%d", voteType)
						return false
					}
				}
				//减票操作，判断重复取消质押和临界值取消投票
				if myVoteType > 0 && !myIsAdd {
					voteType, _, vote, _ := precompiled.UnpackPayloadForCheck(tx.GetPayload())
					switch myVoteType {
					case VOTE_TYPE_community, VOTE_TYPE_light:
						if myVoteType == voteType {
							//engine.Log.Info("283行%d", myVoteType)
							return false
						}
					case VOTE_TYPE_vote:
						//验证总票数是否超过现有票数
						delVoteTotal = big.NewInt(0).Add(delVoteTotal, vote)
						lightDetail := precompiled.GetLightDetail(*addrSelf)
						if delVoteTotal.Cmp(lightDetail.Vote) > 0 {
							//engine.Log.Info("291行")
							return false
						}
					}
				}

			}

		}

		////验证domain 是否重复注册
		//if thisName != "" && hex.EncodeToString(txContract.Payload[:4]) == ens.REGISTER_ID {
		//	if thisName == regrec.GetRegisterName(txContract.GetPayload()) {
		//		return false
		//	}
		//}
	}

	//预执行校验,验证执行是否成功,成功的可以上链
	//errPre := this.PreExec()
	//if errPre != nil {
	//	return false
	//}
	return true
}
func (this *Tx_Contract) GetContractInfo() (*Vin, *Vout) {
	return this.Vin[0], this.Vout[0]
}

/*
统计交易余额，只扣除gas费用
*/
func (this *Tx_Contract) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vout)+len(this.Vin)),
		Nonce:    make(map[string]big.Int),
	}

	// 特殊处理云存储合约, 不减余额
	if this.ContractClass == uint64(config.CLOUD_STORAGE_PROXY_CONTRACT) {
		return &itemCount
	}

	totalValue := this.Gas * this.GasPrice
	//余额中减去。
	from := this.Vin[0].GetPukToAddr()
	itemCount.Nonce[utils.Bytes2string(*from)] = this.Vin[0].Nonce
	frozenMap, ok := itemCount.AddItems[utils.Bytes2string(*from)]
	if ok {
		oldValue, ok := (*frozenMap)[0]
		if ok {
			oldValue -= int64(totalValue)
			(*frozenMap)[0] = oldValue
		} else {
			(*frozenMap)[0] = (0 - int64(totalValue))
		}
	} else {
		frozenMap := make(map[uint64]int64, 0)
		frozenMap[0] = (0 - int64(totalValue))
		itemCount.AddItems[utils.Bytes2string(*from)] = &frozenMap
	}
	return &itemCount
}

func CreateTxContractNew(srcAddress, address *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment,
	source string, class uint64, gasPrice uint64, abi string) (*Tx_Contract, error) {
	//engine.Log.Info("start CreateTxContract")
	commentbs := []byte{}
	if comment != "" {
		commentbs = common.Hex2Bytes(comment)
		//commentbs = []byte(comment)
	}

	//压缩源代码
	sourceBytes := []byte{}
	if source != "" {
		//sourceBytes, _ = common.ZipBytes(common.Hex2Bytes(source))
		sourceBytes, _ = common.ZipBytes([]byte(source))
	}

	//压缩abi
	abiBytes := []byte{}
	if abi != "" {
		//sourceBytes, _ = common.ZipBytes(common.Hex2Bytes(source))
		abiBytes, _ = common.ZipBytes([]byte(abi))
	}

	chain := forks.GetLongChain()

	currentHeight := chain.GetCurrentBlock()

	//engine.Log.Info("currentHeight:%d", currentHeight)

	//查找余额
	vins := make([]*Vin, 0)

	total, item := chain.Balance.BuildPayVinNew(srcAddress, amount+gas)

	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough
	}

	puk, ok := Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := chain.GetBalance().FindNonce(item.Addr)

	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	//engine.Log.Info("新交易nonce:%d", vin.Nonce.Uint64())
	vins = append(vins, &vin)
	// }
	//默认为合约调用
	action := "call"
	//构建交易输出
	vouts := make([]*Vout, 0)
	var vout Vout
	if address == nil {
		//合约地址生成方式待选择
		data := append(*srcAddress, vin.Nonce.Bytes()...)
		addCoin := crypto.BuildAddr(config.AddrPre, data)
		action = "create"
		//addCoin, _ := config.Area.Keystore.GetNewAddr(*config.WalletPwd)
		address = &addCoin
	}

	vout = Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *address,     //钱包地址
		FrozenHeight: frozenHeight, //
	}

	vouts = append(vouts, &vout)

	var pay *Tx_Contract
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_contract,                  //交易类型
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: uint64(len(vouts)),                              //输出交易数量
			Vout:       vouts,                                           //交易输出
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                       //
			Comment:    []byte{},
		}
		pay = &Tx_Contract{
			TxBase:        base,
			Action:        action,
			GzipSource:    sourceBytes,
			ContractClass: class,
			GasPrice:      gasPrice,
			GzipAbi:       abiBytes,
		}
		//给输出签名，防篡改
		for i, one := range pay.Vin {
			_, prk, err := Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}
			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))
			sign := pay.GetSign(&prk, uint64(i))
			pay.Vin[i].Sign = *sign
		}

		pay.BuildHash()

		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}

	//验证合约
	gasUsed, err := pay.PreExecNew()
	if err != nil {
		engine.Log.Error("预执行失败%s %s", hex.EncodeToString(pay.Hash), err.Error())
		return nil, err
	}
	pay.GasUsed = gasUsed

	// engine.Log.Info("交易签名 耗时 %s", config.TimeNow().Sub(start))
	//engine.Log.Info("start CreateTxContract %s", hex.EncodeToString(*pay.GetHash()))
	// chain.Balance.Frozen(items, pay)
	chain.Balance.AddLockTx(pay)
	return pay, nil
}

func (this *Tx_Contract) PreExec() error {
	//预执行校验,验证执行是否成功,成功的可以上链
	payTx := this
	contractVin, contractVout := this.GetContractInfo()
	from := *contractVin.GetPukToAddr()
	to := contractVout.Address
	var (
		evmErr error
		result vm.ExecuteResult
	)
	vmRun := evm.NewVmRun(from, to, *payTx.GetHash(), nil)
	if payTx.Action == "create" {
		//创建合约，提前构建相应结构
		result, _, evmErr = vmRun.Run(payTx.GetPayload(), payTx.GetGasLimit(), payTx.GasPrice, contractVout.Value, true)
		if evmErr != nil || result.ExitOpCode == opcodes.REVERT {
			if result.ExitOpCode == opcodes.REVERT {
				msg, err := abi.UnpackRevert(result.ResultData)
				if err != nil {
					return errors.New("预执行合约创建失败,退出码:" + result.ExitOpCode.String() + err.Error())
				}
				return errors.New("预执行合约创建失败,退出码:" + result.ExitOpCode.String() + ",失败原因:" + msg)
			}
			if evmErr != nil {
				return errors.New("预执行合约创建失败，退出码" + result.ExitOpCode.String() + ",失败原因:" + evmErr.Error())
			}
			return errors.New("预执行合约创建失败")
		}

	} else {
		//调用合约
		result, _, evmErr = vmRun.Run(payTx.GetPayload(), payTx.GetGasLimit(), payTx.GasPrice, contractVout.Value, false)
		if evmErr != nil || result.ExitOpCode == opcodes.REVERT {
			if result.ExitOpCode == opcodes.REVERT {
				msg, err := abi.UnpackRevert(result.ResultData)
				if err != nil {
					return errors.New("预执行合约失败,退出码:" + result.ExitOpCode.String() + err.Error())
				}
				return errors.New("预执行合约失败,退出码:" + result.ExitOpCode.String() + ",失败原因:" + msg)
			}
			if evmErr != nil {
				return errors.New("预执行合约调用失败，退出码" + result.ExitOpCode.String() + ",失败原因:" + evmErr.Error())
			}
			return errors.New("预执行合约调用失败，退出码" + result.ExitOpCode.String())
		}
	}
	//这里判断gas_used
	gasUsed := payTx.GetGasLimit() - result.GasLeft
	if gasUsed > config.EVM_GAS_MAX && !bytes.Equal(to, precompiled.RewardContract) {
		return errors.New("gas spend too many")
	}

	//engine.Log.Error("*********************合约：%s，使用gas：%d", hex.EncodeToString(*payTx.GetHash()), gasUsed)
	return nil
}

func (this *Tx_Contract) PreExecNew() (uint64, error) {
	//预执行校验,验证执行是否成功,成功的可以上链
	payTx := this
	contractVin, contractVout := this.GetContractInfo()
	from := *contractVin.GetPukToAddr()
	to := contractVout.Address
	var (
		evmErr error
		result vm.ExecuteResult
	)
	start := config.TimeNow()
	vmRun := evm.NewVmRun(from, to, *payTx.GetHash(), nil)
	if payTx.Action == "create" {
		//创建合约，提前构建相应结构
		result, _, evmErr = vmRun.Run(payTx.GetPayload(), payTx.GetGasLimit(), payTx.GasPrice, contractVout.Value, true)
		if evmErr != nil || result.ExitOpCode == opcodes.REVERT {
			if result.ExitOpCode == opcodes.REVERT {
				msg, err := abi.UnpackRevert(result.ResultData)

				engine.Log.Error("666666666666666666666666666666666666666")
				if err != nil {
					return 0, errors.New("预执行合约创建失败,退出码:" + result.ExitOpCode.String() + err.Error())
				}
				return 0, errors.New("预执行合约创建失败,退出码:" + result.ExitOpCode.String() + ",失败原因:" + msg)
			}
			if evmErr != nil {
				return 0, errors.New("预执行合约创建失败，退出码" + result.ExitOpCode.String() + ",失败原因:" + evmErr.Error())
			}
			return 0, errors.New("预执行合约创建失败")
		}

	} else {
		//调用合约
		result, _, evmErr = vmRun.Run(payTx.GetPayload(), payTx.GetGasLimit(), payTx.GasPrice, contractVout.Value, false)
		if evmErr != nil || result.ExitOpCode == opcodes.REVERT {
			if result.ExitOpCode == opcodes.REVERT {
				engine.Log.Error("55555555555555555555555555555")
				fmt.Println(result.ResultData)
				fmt.Println(fmt.Sprintf("%x", payTx.GetPayload()))
				engine.Log.Error("555555555555555555555555555551")
				msg, err := abi.UnpackRevert(result.ResultData)
				if err != nil {
					return 0, errors.New("预执行合约失败,退出码:" + result.ExitOpCode.String() + err.Error())
				}
				return 0, errors.New("预执行合约失败,退出码:" + result.ExitOpCode.String() + ",失败原因:" + msg)
			}
			if evmErr != nil {
				return 0, errors.New("预执行合约调用失败，退出码" + result.ExitOpCode.String() + ",失败原因:" + evmErr.Error())
			}
			return 0, errors.New("预执行合约调用失败，退出码" + result.ExitOpCode.String())
		}
	}

	//engine.Log.Error(payTx.GetHash())
	// engine.Log.Info("交易hash: %s", fmt.Sprintf("%x", payTx.Hash))
	engine.Log.Info("count block 合约 time:%s", config.TimeNow().Sub(start))

	//这里判断gas_used
	gasUsed := payTx.GetGasLimit() - result.GasLeft
	if gasUsed > config.EVM_GAS_MAX && !bytes.Equal(to, precompiled.RewardContract) {
		return 0, errors.New("gas spend too many")
	}

	//engine.Log.Error("*********************合约：%s，使用gas：%d", hex.EncodeToString(*payTx.GetHash()), gasUsed)
	return gasUsed, nil
}

func (this *Tx_Contract) PreExecV1(vmRun *evm.VmRun) error {
	vmRun.SetBlock(nil)

	//预执行校验,验证执行是否成功,成功的可以上链
	_, contractVout := this.GetContractInfo()
	var (
		evmErr error
		result vm.ExecuteResult
	)
	if this.Action == "create" {
		//创建合约，提前构建相应结构
		result, _, evmErr = evm.Run(this.GetPayload(), this.GetGasLimit(), this.GasPrice, contractVout.Value, true, vmRun)
		if evmErr != nil || result.ExitOpCode == opcodes.REVERT {
			if result.ExitOpCode == opcodes.REVERT {
				msg, err := abi.UnpackRevert(result.ResultData)
				if err != nil {
					return errors.New("预执行合约创建失败,退出码:" + result.ExitOpCode.String() + err.Error())
				}
				return errors.New("预执行合约创建失败,退出码:" + result.ExitOpCode.String() + ",失败原因:" + msg)
			}
			if evmErr != nil {
				return errors.New("预执行合约创建失败，退出码" + result.ExitOpCode.String() + ",失败原因:" + evmErr.Error())
			}
			return errors.New("预执行合约创建失败")
		}

	} else {
		//engine.Log.Error("打包预执行 hash:%s", hex.EncodeToString(this.Hash))
		//调用合约
		//特殊处理云存储合约
		if this.ContractClass == uint64(config.CLOUD_STORAGE_PROXY_CONTRACT) {
			result, _, evmErr = evm.Run(this.GetPayload(), config.Mining_coin_total, this.GasPrice, 1000000*1e8, false, vmRun)
		} else {
			result, _, evmErr = evm.Run(this.GetPayload(), this.GetGasLimit(), this.GasPrice, contractVout.Value, false, vmRun)
		}
		if evmErr != nil || result.ExitOpCode == opcodes.REVERT {
			if result.ExitOpCode == opcodes.REVERT {
				msg, err := abi.UnpackRevert(result.ResultData)
				if err != nil {
					return errors.New("预执行合约失败,退出码:" + result.ExitOpCode.String() + err.Error())
				}
				return errors.New("预执行合约失败,退出码:" + result.ExitOpCode.String() + ",失败原因:" + msg)
			}
			if evmErr != nil {
				return errors.New("预执行合约调用失败，退出码" + result.ExitOpCode.String() + ",失败原因:" + evmErr.Error())
			}
			return errors.New("预执行合约调用失败，退出码" + result.ExitOpCode.String())
		}

	}
	//这里判断gas_used
	gasUsed := this.GetGasLimit() - result.GasLeft
	if gasUsed > config.EVM_GAS_MAX && !bytes.Equal(contractVout.Address, precompiled.RewardContract) {
		return errors.New("gas spend too many")
	}

	//engine.Log.Error("*********************合约：%s，使用gas：%d", hex.EncodeToString(*payTx.GetHash()), gasUsed)
	return nil
}

// 获取bloom
func (this *Tx_Contract) GetBloom() []byte {
	key := append(config.DBKEY_tx_bloom, this.Hash...)
	bs, err := db.LevelDB.Find(key)
	if err != nil {
		return nil
	}
	return *bs
}

// 设置bloom
func (this *Tx_Contract) SetBloom(bs []byte) {
	key := append(config.DBKEY_tx_bloom, this.Hash...)
	db.LevelDB.Save(key, &bs)
}
