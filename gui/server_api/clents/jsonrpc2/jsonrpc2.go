package jsonrpc2

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"
	"web3_gui/gui/server_api/clents"
)

const AmountCv = uint64(1e8) //金额需要乘以1e8

type peerHttpPolling struct {
	endpoint string
	rpcuser  string
	rpcpwd   string
}

func NewClient(endpoint, rpcuser, rpcpwd string) clents.PeerClient {
	return &peerHttpPolling{
		endpoint: endpoint,
		rpcuser:  rpcuser,
		rpcpwd:   rpcpwd,
	}
}

func (n *peerHttpPolling) GetNewAddress(pwd string) error {
	params := map[string]interface{}{
		"method": "getnewaddress",
		"params": map[string]interface{}{
			"password": pwd,
		},
	}
	_, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return err
	}
	return nil
}

func (n *peerHttpPolling) SendToAddress(srcAddr string, toAddr string, pwd string, amount float64, gas float64) (*clents.RespSendToAddress, error) {
	amount = AmountMuulCV(amount)
	gas = AmountMuulCV(gas)
	params := map[string]interface{}{
		"method": "sendtoaddress",
		"params": map[string]interface{}{
			"srcaddress":    srcAddr,
			"address":       toAddr,
			"changeaddress": srcAddr,
			"amount":        amount,
			"gas":           gas,
			"frozen_height": 7,
			"pwd":           pwd,
		},
	}
	res, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(res.Result)
	if err != nil {
		return nil, err
	}

	result := &clents.RespSendToAddress{}
	if err := json.Unmarshal(b, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (n *peerHttpPolling) Depositin(amount float64, gas float64, rate float64, pwd string, payload string) error {
	amount = AmountMuulCV(amount)
	gas = AmountMuulCV(gas)
	params := map[string]interface{}{
		"method": "depositin",
		"params": map[string]interface{}{
			"amount":  amount,
			"gas":     gas,
			"pwd":     pwd,
			"rate":    rate,
			"payload": payload,
		},
	}
	_, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return err
	}
	return nil
}

func (n *peerHttpPolling) Depositout(witnessAddr string, amount float64, gas float64, pwd string) error {
	amount = AmountMuulCV(amount)
	gas = AmountMuulCV(gas)
	params := map[string]interface{}{
		"method": "depositout",
		"params": map[string]interface{}{
			"witness": witnessAddr,
			"amount":  amount,
			"gas":     gas,
			"pwd":     pwd,
		},
	}
	_, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return err
	}
	return nil
}

func (n *peerHttpPolling) VoteIn(votetype uint16, address string, witness string, amount float64, gas float64, rate float64, pwd string, payload string) error {
	amount = AmountMuulCV(amount)
	gas = AmountMuulCV(gas)
	params := map[string]interface{}{
		"method": "votein",
		"params": map[string]interface{}{
			"votetype": votetype,
			"address":  address,
			"witness":  witness,
			"amount":   amount,
			"gas":      gas,
			"pwd":      pwd,
			"rate":     rate,
			"payload":  payload,
		},
	}
	_, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return err
	}
	return nil
}
func (n *peerHttpPolling) VoteOut(votetype uint16, address string, amount float64, gas float64, rate float64, pwd string, payload string) error {
	amount = AmountMuulCV(amount)
	gas = AmountMuulCV(gas)
	params := map[string]interface{}{
		"method": "voteout",
		"params": map[string]interface{}{
			"votetype": votetype,
			"address":  address,
			"amount":   amount,
			"gas":      gas,
			"pwd":      pwd,
			"rate":     rate,
			"payload":  payload,
		},
	}
	_, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return err
	}
	return nil
}

func (n *peerHttpPolling) CommunityDistribute(srcaddress string, gas float64, pwd string) error {
	gas = AmountMuulCV(gas)
	params := map[string]interface{}{
		"method": "communitydistribute",
		"params": map[string]interface{}{
			"srcaddress": srcaddress,
			"gas":        gas,
			"pwd":        pwd,
		},
	}
	_, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return err
	}
	return nil
}

func (n *peerHttpPolling) NameIn(address string, amount float64, gas float64, pwd string, name string, netids string, addrcoins string, comment string) error {
	amount = AmountMuulCV(amount)
	gas = AmountMuulCV(gas)

	netidsArr := []string{}
	if netids != "" {
		netidsArr = strings.Split(netids, ",")
	}

	addrcoinsArr := []string{}
	if addrcoins != "" {
		addrcoinsArr = strings.Split(addrcoins, ",")
	}

	params := map[string]interface{}{
		"method": "namesin",
		"params": map[string]interface{}{
			"address":   address,
			"amount":    amount,
			"gas":       gas,
			"pwd":       pwd,
			"name":      name,
			"netids":    netidsArr,
			"addrcoins": addrcoinsArr,
			"comment":   comment,
		},
	}
	_, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return err
	}
	return nil
}

func (n *peerHttpPolling) NameOut(address string, amount float64, gas float64, pwd string, name string, comment string) error {
	amount = AmountMuulCV(amount)
	gas = AmountMuulCV(gas)

	params := map[string]interface{}{
		"method": "namesout",
		"params": map[string]interface{}{
			"address": address,
			"amount":  amount,
			"gas":     gas,
			"pwd":     pwd,
			"name":    name,
			"comment": comment,
		},
	}
	_, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return err
	}
	return nil
}

func (n *peerHttpPolling) GetNames() ([]*clents.NameinfoVO, error) {
	params := map[string]interface{}{
		"method": "getnames",
	}
	res, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(res.Result)
	if err != nil {
		return nil, err
	}

	result := []*clents.Nameinfo{}
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}

	resultVO := []*clents.NameinfoVO{}
	for _, v := range result {
		resultVO = append(resultVO, &clents.NameinfoVO{
			Name:           v.Name,
			Owner:          v.Owner,
			NetIds:         v.NetIds,
			AddrCoins:      v.AddrCoins,
			Height:         v.Height,
			NameOfValidity: v.NameOfValidity,
			Deposit:        AmountDivCV(v.Deposit),
		})
	}

	return resultVO, nil
}

func (n *peerHttpPolling) ListAccounts() ([]*clents.RespListAccounts, error) {
	params := map[string]interface{}{
		"method": "listaccounts",
		"params": map[string]interface{}{
			"page":      1,
			"page_size": 20,
		},
	}
	res, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(res.Result)
	if err != nil {
		return nil, err
	}

	result := []*clents.RespListAccounts{}
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (n *peerHttpPolling) GetInfo() (*clents.RespGetInfo, error) {
	params := map[string]interface{}{
		"method": "getinfo",
		"params": map[string]interface{}{},
	}
	res, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(res.Result)
	if err != nil {
		return nil, err
	}

	result := &clents.RespGetInfo{}
	if err := json.Unmarshal(b, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (n *peerHttpPolling) GetAccount(addr string) (*clents.RespGetAccount, error) {
	params := map[string]interface{}{
		"method": "listaccounts",
		"params": map[string]interface{}{
			"address": addr,
		},
	}
	res, err := Post(n.endpoint, n.rpcuser, n.rpcpwd, params)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(res.Result)
	if err != nil {
		return nil, err
	}

	result := &clents.RespGetAccount{}
	if err := json.Unmarshal(b, result); err != nil {
		return nil, err
	}

	return result, nil
}

func Post(addr, rpcUser, rpcPwd string, params map[string]interface{}) (*clents.PostResult, error) {
	url := "/rpc"
	method := "POST"
	header := http.Header{
		"user":     []string{rpcUser},
		"password": []string{rpcPwd},
	}
	client := &http.Client{Timeout: time.Second * 6}
	bs, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, "http://"+addr+url, strings.NewReader(string(bs)))
	if err != nil {
		return nil, err
	}
	req.Header = header
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New(err.Error() + "Offline")
	}
	defer resp.Body.Close()

	var resultBs []byte

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	resultBs, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := new(clents.PostResult)
	buf := bytes.NewBuffer(resultBs)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err = decoder.Decode(result)
	if err != nil {
		return nil, err
	}

	if result.Code != 2000 {
		return nil, errors.New(result.Message)
	}
	return result, nil
}

func AmountDivCV(amount uint64) float64 {
	return float64(amount) / float64(AmountCv)
}

func AmountMuulCV(amount float64) float64 {
	ft := big.NewFloat(amount)
	ftConst := big.NewFloat(float64(AmountCv))
	res, _ := ft.Mul(ft, ftConst).Float64()
	return res
}
