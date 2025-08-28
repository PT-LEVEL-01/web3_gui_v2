package bind

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/utils"
)

type MetaData struct {
	mu   sync.Mutex
	Sigs map[string]string
	Bin  string
	ABI  string
	ab   *abi.ABI
}

func (m *MetaData) GetAbi() (*abi.ABI, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ab != nil {
		return m.ab, nil
	}
	if parsed, err := abi.JSON(strings.NewReader(m.ABI)); err != nil {
		return nil, err
	} else {
		m.ab = &parsed
	}
	return m.ab, nil
}

type BoundContract struct {
	api string  // network
	abi abi.ABI // Reflect based ABI to access the correct Ethereum methods
}
type RequestParams struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}
type TransactionOpts struct {
	Headers         map[string]string
	SrcAddress      string   `json:"srcaddress"`
	ContractAddress string   `json:"contractaddress"`
	Amount          *big.Int `json:"amount"`
	Gas             *big.Int `json:"gas"`
	FrozenHeight    *big.Int `json:"frozen_height"`
	Pwd             string   `json:"pwd"`
	Comment         string   `json:"comment"`
}

func (o *TransactionOpts) Copy() *TransactionOpts {
	return &TransactionOpts{
		SrcAddress:      o.SrcAddress,
		ContractAddress: o.ContractAddress,
		Amount:          o.Amount,
		Gas:             o.Gas,
		FrozenHeight:    o.FrozenHeight,
		Pwd:             o.Pwd,
		Comment:         o.Comment,
	}
}
func NewBoundContract(address string, abi abi.ABI) *BoundContract {
	return &BoundContract{
		api: address,
		abi: abi,
	}
}

// 返回交易hash
func (c *BoundContract) Transfer(opts *TransactionOpts, method string, params ...interface{}) (string, error) {
	//params是转换好的数据，比如address这种类型
	input, err := c.abi.Pack(method, params...)
	if err != nil {
		return "", err
	}
	res, err := c.sendRequest(opts, "callContract", common.Bytes2Hex(input))
	if err != nil {
		return "", err
	}
	if gjson.GetBytes(res, "code").Int() == 2000 {
		return gjson.GetBytes(res, "result.hash").String(), nil
	}
	return "", errors.New("message:" + gjson.GetBytes(res, "message").String())
}

// 返回input
func (c *BoundContract) GetInput(method string, params ...interface{}) (string, error) {
	//params是转换好的数据，比如address这种类型
	input, err := c.abi.PackV0(method, params...)
	if err != nil {
		return "", err
	}
	return common.Bytes2Hex(input), err
}

func (c *BoundContract) sendRequest(opts *TransactionOpts, method string, input string) ([]byte, error) {
	body := RequestParams{
		Method: method,
		Params: make(map[string]interface{}),
	}
	body.Params, _ = utils.ChangeMap(opts)
	if evmutils.Has0xPrefix(input) {
		input = input[2:]
	}
	body.Params["comment"] = input
	params, _ := json.Marshal(body)
	client := &http.Client{}
	req, err := http.NewRequest("POST", c.api, bytes.NewReader(params))
	if err != nil {
		return nil, err
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
func (c *BoundContract) Call(opts *TransactionOpts, results *[]interface{}, method string, params ...interface{}) error {
	if opts == nil {
		opts = new(TransactionOpts)
	}
	if results == nil {
		results = new([]interface{})
	}
	input, err := c.abi.Pack(method, params...)
	if err != nil {
		return err
	}
	response, err := c.sendRequest(opts, "staticCallContract", common.Bytes2Hex(input))
	if gjson.GetBytes(response, "code").Int() != 2000 {
		return errors.New("message:" + gjson.GetBytes(response, "message").String())
	}
	var output = common.Hex2Bytes(gjson.GetBytes(response, "result").String())
	if err != nil {
		return err
	}
	if len(*results) == 0 {
		res, err := c.abi.Unpack(method, output)
		*results = res
		return err
	}
	res := *results
	return c.abi.UnpackIntoInterface(res[0], method, output)

}

func (c *BoundContract) PreCall(opts *TransactionOpts, calldata []byte) (string, uint64, error) {
	if opts == nil {
		opts = new(TransactionOpts)
	}
	response, err := c.sendRequest(opts, "PreCallContract", common.Bytes2Hex(calldata))
	if err != nil {
		return "", 0, err
	}
	if gjson.GetBytes(response, "code").Int() != 2000 {
		return "", 0, errors.New("message:" + gjson.GetBytes(response, "message").String())
	}
	return gjson.GetBytes(response, "result.errMsg").String(), gjson.GetBytes(response, "result.gasUsed").Uint(), nil
}

// TODO
func (c *BoundContract) RawTransact(opts *TransactionOpts, calldata []byte) (string, error) {
	res, err := c.sendRequest(opts, "callContract", common.Bytes2Hex(calldata))
	if err != nil {
		return "", err
	}

	if gjson.GetBytes(res, "code").Int() == 2000 {
		return gjson.GetBytes(res, "result.hash").String(), nil
	}
	return "", errors.New("message:" + gjson.GetBytes(res, "message").String())
}

func (c *BoundContract) RawTransactStatic(opts *TransactionOpts, calldata []byte) (string, error) {
	res, err := c.sendRequest(opts, "staticCallContract", common.Bytes2Hex(calldata))
	if err != nil {
		return "", err
	}

	if gjson.GetBytes(res, "code").Int() == 2000 {
		return gjson.GetBytes(res, "result").String(), nil
	}
	return "", errors.New("message:" + gjson.GetBytes(res, "message").String())
}

// 合约地址，交易hash
func (c *BoundContract) DeployContract(opts *TransactionOpts, byteCode string, params ...interface{}) (string, string, error) {
	payload, err := c.abi.Pack("", params...)
	if err != nil {
		return "", "", err
	}
	input := byteCode + common.Bytes2Hex(payload)
	res, err := c.sendRequest(opts, "createContract", input)
	if err != nil {
		return "", "", err
	}
	if gjson.GetBytes(res, "code").Int() == 2000 {
		return gjson.GetBytes(res, "result.contract_address").String(), gjson.GetBytes(res, "result.hash").String(), nil
	}
	return "", "", errors.New(string(res))

}
