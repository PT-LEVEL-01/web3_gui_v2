package mining

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"web3_gui/chain/config"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/utils"
	"web3_gui/libp2parea/adapter/engine"
)

var offlineTxHandler = initOfflineTxHandler()

var (
	ErrStruct   = errors.New("must be struct")
	ErrReuqired = errors.New("must be required")
)

func initOfflineTxHandler() map[string]OfflineTxItr {
	return map[string]OfflineTxItr{
		"transfer":        new(TransferTxParams),        //转账
		"addCommunity":    new(AddCommunityTxParams),    //质押社区
		"cancelCommunity": new(CancelCommunityTxParams), //取消质押社区
		"addLight":        new(AddLightTxParams),        //质押轻节点
		"cancelLight":     new(CancelLightTxParams),     //取消质押轻节点
		"addVote":         new(AddVoteTxParams),         //添加投票
		"cancelVote":      new(CancelVoteTxParams),      //取消投票
		"transferErc20":   new(TransferErc20TxParams),   //代币转账
		"rewardWithdraw":  new(RewardWithdrawTxParams),  //奖励提现
	}
}

type OfflineTxItr interface {
	//解析并验证json
	ResolveAndCheck(data []byte) error
	//构建交易
	BuildTx() *OfflineTx
}

// 验证结构体字段
func validateStruct(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("param %w", ErrStruct)
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := v.Type().Field(i).Tag.Get("validate")
		if tag == "required" && field.Interface() == "" {
			return fmt.Errorf("%s %w", v.Type().Field(i).Name, ErrReuqired)
		}
		if strings.Contains(tag, ":") {
			var d uint64
			switch field.Kind() {
			case reflect.Uint16:
				d = uint64(field.Interface().(uint16))
			default:
				d = field.Interface().(uint64)
			}
			tags := strings.Split(tag, ":")
			s1, _ := strconv.ParseUint(tags[1], 10, 64)
			if !compare(tags[0], d, s1) {
				return fmt.Errorf("%s must be %s %d", v.Type().Field(i).Name, tags[0], s1)
			}
		}
	}
	return nil
}

func compare(t string, a, b uint64) bool {
	r := false
	switch t {
	case "eq":
		r = a == b
	case "gt":
		r = a > b
	case "le":
		r = a <= b
	}
	return r
}

//region 转账

// 转账交易参数
type TransferTxParams struct {
	SrcAddress string `json:"srcaddress" validate:"required"`
	Address    string `json:"address" validate:"required"`
	Amount     uint64 `json:"amount" validate:"gt:0"`
	Gas        uint64 `json:"gas"`
	GasPrice   uint64 `json:"gasPrice"`
	Comment    string `json:"comment"`
}

func (t *TransferTxParams) ResolveAndCheck(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	if err := validateStruct(*t); err != nil {
		return err
	}
	if t.GasPrice < config.DEFAULT_GAS_PRICE {
		t.GasPrice = config.DEFAULT_GAS_PRICE
	}
	return nil
}

func (t *TransferTxParams) BuildTx() *OfflineTx {
	return &OfflineTx{
		isContract: false,
		SrcAddress: t.SrcAddress,
		Address:    t.Address,
		Amount:     t.Amount,
		Gas:        t.Gas,
		Comment:    t.Comment,
		GasPrice:   t.GasPrice,
	}
}

//endregion

//region 质押社区

// 质押社区
type AddCommunityTxParams struct {
	SrcAddress     string `json:"srcAddress" validate:"required"`
	WitnessAddress string `json:"witnessAddress" validate:"required"`
	Amount         uint64 `json:"amount" validate:"gt:0"`
	Gas            uint64 `json:"gas" validate:"gt:0"`
	GasPrice       uint64 `json:"gasPrice" validate:"gt:0"`
	Name           string `json:"name"`
	Rate           uint16 `json:"rate" validate:"le:100"`
	Tag            string
}

func (t *AddCommunityTxParams) ResolveAndCheck(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	if err := validateStruct(*t); err != nil {
		return err
	}
	if t.Amount != config.Mining_vote {
		return errors.New("Community node deposit is " + strconv.Itoa(int(config.Mining_vote/1e8)))
	}
	return nil
}

func (t *AddCommunityTxParams) BuildTx() *OfflineTx {
	return &OfflineTx{
		SrcAddress: t.SrcAddress,
		Address:    t.WitnessAddress,
		Amount:     t.Amount,
		Gas:        t.Gas,
		GasPrice:   t.GasPrice,
		Rate:       t.Rate,
		Tag:        "votein",
		VoteType:   VOTE_TYPE_community,
		Comment:    t.Name,
	}
}

//endregion

//region 取消质押社区

// 取消质押社区
type CancelCommunityTxParams struct {
	SrcAddress string `json:"srcAddress" validate:"required"`
	Amount     uint64 `json:"amount" validate:"gt:0"`
	Gas        uint64 `json:"gas" validate:"gt:0"`
	GasPrice   uint64 `json:"gasPrice" validate:"gt:0"`
}

func (t *CancelCommunityTxParams) ResolveAndCheck(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	if err := validateStruct(*t); err != nil {
		return err
	}
	return nil
}

func (t *CancelCommunityTxParams) BuildTx() *OfflineTx {
	return &OfflineTx{
		SrcAddress: t.SrcAddress,
		Gas:        t.Gas,
		GasPrice:   t.GasPrice,
		Tag:        "voteout",
		VoteType:   VOTE_TYPE_community,
		Amount:     t.Amount,
	}
}

//endregion

//region 成为轻节点

// 质押轻节点
type AddLightTxParams struct {
	SrcAddress string `json:"srcAddress" validate:"required"`
	Amount     uint64 `json:"amount" validate:"gt:0"`
	Gas        uint64 `json:"gas" validate:"gt:0"`
	GasPrice   uint64 `json:"gasPrice" validate:"gt:0"`
	Name       string `json:"name"`
	Tag        string
}

func (t *AddLightTxParams) ResolveAndCheck(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	if err := validateStruct(*t); err != nil {
		return err
	}
	if t.Amount != config.Mining_light_min {
		return errors.New("Light node deposit is " + strconv.Itoa(int(config.Mining_light_min/1e8)))
	}
	return nil
}

func (t *AddLightTxParams) BuildTx() *OfflineTx {
	return &OfflineTx{
		SrcAddress: t.SrcAddress,
		Amount:     t.Amount,
		Gas:        t.Gas,
		GasPrice:   t.GasPrice,
		Tag:        "votein",
		VoteType:   VOTE_TYPE_light,
		Comment:    t.Name,
	}
}

//endregion

//region 取消质押轻节点

// 取消质押轻节点
type CancelLightTxParams struct {
	SrcAddress string `json:"srcAddress" validate:"required"`
	Amount     uint64 `json:"amount" validate:"gt:0"`
	Gas        uint64 `json:"gas" validate:"gt:0"`
	GasPrice   uint64 `json:"gasPrice" validate:"gt:0"`
}

func (t *CancelLightTxParams) ResolveAndCheck(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	if err := validateStruct(*t); err != nil {
		return err
	}
	return nil
}

func (t *CancelLightTxParams) BuildTx() *OfflineTx {
	return &OfflineTx{
		SrcAddress: t.SrcAddress,
		Gas:        t.Gas,
		GasPrice:   t.GasPrice,
		Tag:        "voteout",
		VoteType:   VOTE_TYPE_light,
		Amount:     t.Amount,
	}
}

//endregion

//region 投票

// 投票
type AddVoteTxParams struct {
	SrcAddress       string `json:"srcAddress" validate:"required"`
	CommunityAddress string `json:"communityAddress" validate:"required"`
	Amount           uint64 `json:"amount" validate:"gt:0"`
	Gas              uint64 `json:"gas" validate:"gt:0"`
	GasPrice         uint64 `json:"gasPrice" validate:"gt:0"`
	Tag              string
}

func (t *AddVoteTxParams) ResolveAndCheck(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	if err := validateStruct(*t); err != nil {
		return err
	}
	return nil
}

func (t *AddVoteTxParams) BuildTx() *OfflineTx {
	return &OfflineTx{
		SrcAddress: t.SrcAddress,
		Address:    t.CommunityAddress,
		Amount:     t.Amount,
		Gas:        t.Gas,
		GasPrice:   t.GasPrice,
		Tag:        "votein",
		VoteType:   VOTE_TYPE_vote,
	}
}

//endregion

//region 取消投票

// 取消投票
type CancelVoteTxParams struct {
	SrcAddress       string `json:"srcAddress" validate:"required"`
	CommunityAddress string `json:"communityAddress" validate:"required"`
	Amount           uint64 `json:"amount" validate:"gt:0"`
	Gas              uint64 `json:"gas" validate:"gt:0"`
	GasPrice         uint64 `json:"gasPrice" validate:"gt:0"`
}

func (t *CancelVoteTxParams) ResolveAndCheck(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	if err := validateStruct(*t); err != nil {
		return err
	}
	return nil
}

func (t *CancelVoteTxParams) BuildTx() *OfflineTx {
	return &OfflineTx{
		SrcAddress: t.SrcAddress,
		Address:    t.CommunityAddress,
		Amount:     t.Amount,
		Gas:        t.Gas,
		GasPrice:   t.GasPrice,
		Tag:        "voteout",
		VoteType:   VOTE_TYPE_vote,
	}
}

//endregion

//region 代币转账

// 代币转账
type TransferErc20TxParams struct {
	SrcAddress      string `json:"srcAddress" validate:"required"`
	ContractAddress string `json:"contractAddress" validate:"required"`
	ToAddress       string `json:"toAddress" validate:"required"`
	Amount          string `json:"amount" validate:"required"`
	Gas             uint64 `json:"gas" validate:"gt:0"`
	GasPrice        uint64 `json:"gasPrice" validate:"gt:0"`
	Decimal         uint8  `json:"decimal"`
}

func (t *TransferErc20TxParams) ResolveAndCheck(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	if err := validateStruct(*t); err != nil {
		return err
	}
	return nil
}

func (t *TransferErc20TxParams) BuildTx() *OfflineTx {
	amount := precompiled.StringToValue(t.Amount, t.Decimal)
	if amount == nil || amount.Cmp(big.NewInt(0)) == 0 {
		engine.Log.Error("amount cannot be zero")
		return nil
	}
	comment := precompiled.BuildErc20TransferBigInput(t.ToAddress, amount)
	return &OfflineTx{
		isContract: true,
		SrcAddress: t.SrcAddress,
		Address:    t.ContractAddress,
		Amount:     0,
		Gas:        t.Gas,
		Comment:    hex.EncodeToString(comment),
		GasPrice:   t.GasPrice,
	}
}

//endregion

//region 奖励提现

// 奖励提现
type RewardWithdrawTxParams struct {
	SrcAddress string `json:"srcAddress" validate:"required"`
	Gas        uint64 `json:"gas" validate:"gt:0"`
	GasPrice   uint64 `json:"gasPrice" validate:"gt:0"`
}

func (t *RewardWithdrawTxParams) ResolveAndCheck(data []byte) error {
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	if err := validateStruct(*t); err != nil {
		return err
	}
	return nil
}

func (t *RewardWithdrawTxParams) BuildTx() *OfflineTx {
	return &OfflineTx{
		SrcAddress: t.SrcAddress,
		Gas:        t.Gas,
		GasPrice:   t.GasPrice,
		Tag:        "reward",
	}
}

// 构建离线交易
func BuildOfflineTx(keyStorePath, pwd string, nonce, currentHeight, frozenHeight, domainType uint64, domain, tag, jsonData string) string {
	if pwd == "" {
		//engine.Log.Error("pwd can't be empty")
		return utils.Out(500, "pwd can't be empty")
	}
	target, ok := offlineTxHandler[tag]
	if !ok {
		//engine.Log.Error("not found tx type :%s", tag)
		return utils.Out(500, "not found tx type")
	}
	data := []byte(jsonData)
	if err := target.ResolveAndCheck(data); err != nil {
		//engine.Log.Error("parse tx fail :%s", err.Error())
		return utils.Out(500, err.Error())
	}

	tx := target.BuildTx()
	if tx == nil {
		return utils.Out(500, "tx build fail")
	}
	tx.keyStorePath = keyStorePath
	tx.nonce = nonce
	tx.currentHeight = currentHeight
	tx.frozenHeight = frozenHeight
	tx.Pwd = pwd
	tx.Domain = domain
	tx.DomainType = domainType
	return buildOfflineTx(tx)
}

// 离线交易
type OfflineTx struct {
	isContract    bool
	keyStorePath  string
	frozenHeight  uint64
	currentHeight uint64
	nonce         uint64
	SrcAddress    string
	Address       string
	Amount        uint64
	Gas           uint64
	GasPrice      uint64
	Comment       string
	Pwd           string
	Domain        string
	DomainType    uint64
	Abi           string
	Source        string
	VoteType      uint16
	Rate          uint16
	Tag           string
}

// 构建离线交易
func buildOfflineTx(tx *OfflineTx) string {
	var txData struct {
		Hash            string `json:"hash"`
		Tx              string `json:"tx"`
		ContractAddress string `json:"contractAddress"`
		IsContract      bool   `json:"isContract"`
	}

	var res, hash, addressContract string
	var err error
	if tx.isContract {
		res, hash, addressContract, err = CreateOfflineContractTx(tx.keyStorePath, tx.SrcAddress, tx.Address, tx.Pwd, tx.Comment, tx.Amount, tx.Gas, tx.frozenHeight, tx.GasPrice, tx.nonce, tx.currentHeight, tx.Domain, tx.DomainType, tx.Abi, tx.Source)
		if err != nil {
			return utils.Out(500, err.Error())
		}
	} else {
		if tx.Tag == "" {
			res, hash, err = CreateOfflineTx(tx.keyStorePath, tx.SrcAddress, tx.Address, tx.Pwd, tx.Comment, tx.Amount, tx.Gas, tx.frozenHeight, tx.nonce, tx.currentHeight, tx.Domain, tx.DomainType)
		} else {
			res, hash, err = CreateOfflineRewardTx(tx.keyStorePath, tx.SrcAddress, tx.Address, tx.Tag, tx.VoteType, tx.Rate, tx.Pwd, tx.Comment, tx.Amount, tx.Gas, tx.frozenHeight, tx.nonce, tx.currentHeight, tx.Domain, tx.DomainType)
		}
		if err != nil {
			return utils.Out(500, err.Error())
		}
	}
	txData.Hash = hash
	txData.Tx = res
	txData.ContractAddress = addressContract
	txData.IsContract = tx.isContract
	return utils.Out(200, txData)
}
