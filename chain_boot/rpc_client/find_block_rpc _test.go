package rpc_client

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"testing"
	chainconfig "web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/chain/mining/tx_name_in"
	"web3_gui/chain/mining/tx_name_out"
	"web3_gui/chain/rpc"
	enginea "web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

func TestFindBlockRpc(t *testing.T) {
	start()
}

func init() {
	mining.RegisterTransaction(chainconfig.Wallet_tx_type_account, new(tx_name_in.AccountController))
	mining.RegisterTransaction(chainconfig.Wallet_tx_type_account_cancel, new(tx_name_out.AccountController))
	// mining.RegisterTransaction(config.Wallet_tx_type_spaces_mining_in, new(spaces_mining_in.AccountController))
	// mining.RegisterTransaction(config.Wallet_tx_type_spaces_mining_out, new(spaces_mining_out.AccountController))
}

func start() {
	peer1 := Peer{
		// Addr:       "127.0.0.1:5081",        //节点地址及端口，本地
		Addr:       "127.0.0.1:27331",       //节点地址及端口，本地
		AddressMax: 50000,                   //收款地址总数
		RPCUser:    "test",                  //rpc用户名
		RPCPwd:     "testp",                 //rpc密码
		WalletPwd:  "witness2",              //
		PayChan:    make(chan *Address, 10), //
	}

	info, ERR := peer1.GetInfo()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	startHeight := info.StartingBlock
	endHeight := info.CurrentBlock

	for i := startHeight; i <= endHeight; i++ {
		bhvo, ERR := peer1.GetBlockAndTxProto(i)
		if ERR.CheckFail() {
			panic(ERR.String())
		}
		PrintBlock(bhvo)
	}
	return
}

type Peer struct {
	Addr       string        //节点地址及端口
	AddressMax uint64        //收款地址总数
	RPCUser    string        //rpc用户名
	RPCPwd     string        //rpc密码
	WalletPwd  string        //钱包支付密码
	PayChan    chan *Address //
}

/*
查询区块高度
*/
func (this *Peer) GetInfo() (*Info, utils.ERROR) {
	return GetInfo(this.Addr, this.RPCUser, this.RPCPwd)
}

/*
给多人转账
*/
func (this *Peer) TxPayMore(payNumber []PayNumber, gas uint64) bool {
	//{"method":"sendtoaddress","params":{"address":"1Hy62rv8BDypQgLpGeUPwGQzVkkkBBU8v22",
	//"amount":1000000000000000,"gas":100000000,"pwd":"123456789","comment":"test"}}

	fmt.Println("多人转账", payNumber, gas)

	paramsChild := map[string]interface{}{
		"addresses": payNumber,
		"gas":       gas,
		"pwd":       this.WalletPwd,
	}
	params := map[string]interface{}{
		"method": "sendtoaddressmore",
		"params": paramsChild,
	}
	result := Post(this.Addr, this.RPCUser, this.RPCPwd, params)
	bs, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误", err.Error())
		return false
	}
	fmt.Println(result.Code, result.Message, string(bs))
	return true
}

/*
查询指定账户余额
*/
func (this *Peer) GetAccount(address string) *Account {
	//{"method":"sendtoaddress","params":{"address":"1Hy62rv8BDypQgLpGeUPwGQzVkkkBBU8v22",
	//"amount":1000000000000000,"gas":100000000,"pwd":"123456789","comment":"test"}}
	paramsChild := map[string]interface{}{
		"address": address,
	}
	params := map[string]interface{}{
		"method": "getaccount",
		"params": paramsChild,
	}
	result := Post(this.Addr, this.RPCUser, this.RPCPwd, params)
	bs, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误", err.Error())
		return nil
	}

	account := new(Account)
	buf := bytes.NewBuffer(bs)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err = decoder.Decode(&account)

	return account
}

/*
获取指定高度的区块及交易
*/
func (this *Peer) GetBlockAndTxProto(height uint64) (*mining.BlockHeadVO, utils.ERROR) {
	return BlockProto64ByHeight(this.Addr, this.RPCUser, this.RPCPwd, height)
}

// 帐号余额
type Account struct {
	Balance       uint64 `json:"Balance"`
	BalanceFrozen uint64 `json:"BalanceFrozen"`
}

/*
自定义请求head，body，method，参数用body传递
获取添加金币记录
*/
func Post(addr, rpcUser, rpcPwd string, params map[string]interface{}) *PostResult {
	url := "/rpc"
	method := "POST"
	fmt.Println("22222222222")
	header := http.Header{"user": []string{rpcUser}, "password": []string{rpcPwd}}
	client := &http.Client{}
	//req, err := http.NewRequest("GET", "http://www.baidu.com/", nil)
	bs, err := json.Marshal(params)
	req, err := http.NewRequest(method, "http://"+addr+url, strings.NewReader(string(bs)))
	if err != nil {
		fmt.Println("创建request错误")
		return nil
	}
	fmt.Println("22222222222")
	req.Header = header
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求服务器错误", err.Error())
		return nil
	}
	fmt.Println("22222222222")
	fmt.Println("response:", resp.StatusCode)

	var resultBs []byte

	if resp.StatusCode == 200 {
		resultBs, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("读取body内容错误")
			return nil
		}
		// fmt.Println(string(resultBs))

		result := new(PostResult)

		buf := bytes.NewBuffer(resultBs)
		decoder := json.NewDecoder(buf)
		decoder.UseNumber()
		err = decoder.Decode(result)
		return result
	}
	fmt.Println("StatusCode:", resp.StatusCode)
	return nil
}

type PostResult struct {
	Jsonrpc string      `json:"jsonrpc"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

/*
获得一个随机数(0 - n]，包含0，不包含n
*/
func GetRandNum(n int64) int64 {
	if n == 0 {
		return 0
	}
	result, _ := crand.Int(crand.Reader, big.NewInt(int64(n)))
	return result.Int64()
}

type PayPlan struct {
	SrcIndex int    //转出账户索引地址
	DstIndex int    //转入账户索引地址
	Value    uint64 //转账额度
}

/*
打印区块内容
*/
func PrintBlock(bhVO *mining.BlockHeadVO) {
	bh := bhVO.BH
	enginea.Log.Info("===============================================")
	enginea.Log.Info("第%d个块 %s 交易数量 %d", bh.Height, hex.EncodeToString(bh.Hash), len(bh.Tx))
	txs := make([]string, 0)
	for _, one := range bh.Tx {
		txs = append(txs, hex.EncodeToString(one))
	}
	bhvo := rpc.BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		GroupHeightGrowth: bh.GroupHeightGrowth,                     //
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
	}
	bs, _ := json.Marshal(bhvo)
	enginea.Log.Info("%s", string(bs))

	for _, one := range bhVO.Txs {
		// engine.Log.Info("%v", one)
		enginea.Log.Info("-----------------------------------------------")
		enginea.Log.Info("交易id:%s", string(hex.EncodeToString(*one.GetHash())))
		// one.GetVOJSON()
		txItr := one.GetVOJSON()
		bs, _ := json.Marshal(txItr)
		enginea.Log.Info("%s", string(bs))
	}

}
