package token

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestPublish(t *testing.T) {
	r, err := httppost(listaccountReq)
	if err != nil {
		t.Fatal(err)
	}
	addrs := []string{}
	for _, v := range r.Get("result").Array() {
		addr := v.Get("AddrCoin").String()
		addrs = append(addrs, addr)
	}

	if len(addrs) < 3 {
		t.Fatal("less address")
	}

	//发代币
	r, err = httppost(fmt.Sprintf(tokenpublishReq, addrs[0], addrs[0]))
	if err != nil {
		t.Fatal(err)
	}

	tokenId := r.Get("result").Get("hash").String()
	time.Sleep(time.Second * 2)
	t.Logf("Got TokenID: %s\n", tokenId)
	//转账
	for i := 0; i < 30; i++ {
		time.Sleep(time.Millisecond * 100)
		r, err = httppost(fmt.Sprintf(tokenpay, addrs[0], addrs[1], tokenId))
		if err != nil {
			t.Fatal(err)
		}
		r, err = httppost(fmt.Sprintf(querytoken, addrs[0], addrs[1], tokenId))
		if err != nil {
			t.Fatal(err)
		}
		t.Log()
		for _, v := range r.Get("result").Array() {
			addr := v.Get("AddrCoin").String()
			value := v.Get("Value").Int()
			lockvalue := v.Get("ValueLockup").Int()
			t.Logf("地址:%s 余额:%d 锁定:%d\n", addr, value, lockvalue)
		}
	}

	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)
		r, err = httppost(fmt.Sprintf(querytoken, addrs[0], addrs[1], tokenId))
		if err != nil {
			t.Fatal(err)
		}
		t.Log()
		for _, v := range r.Get("result").Array() {
			addr := v.Get("AddrCoin").String()
			value := v.Get("Value").Int()
			lockvalue := v.Get("ValueLockup").Int()
			t.Logf("地址:%s 余额:%d 锁定:%d\n", addr, value, lockvalue)
		}
	}
}

var (
	listaccountReq = `{
    "method": "listaccounts"
}`

	tokenpublishReq = `{
    "method": "tokenpublish",
    "params": {
        "srcaddress": "%s",
        "gas": 10000,
        "pwd": "123456789",
        "name": "TokenA",
        "symbol": "TA",
        "supply": "1000000",
        "owner": "%s"
    }
}`
	tokenpay = `{
    "method": "tokenpay",
    "params": {
        "srcaddress": "%s",
        "address": "%s",
        "gas": 1000,
        "pwd": "123456789",
        "amount": 10,
        "txid": "%s"
    }
}`
	querytoken = `{
    "method": "multiaccounts",
    "params": {
        "addresses": [
            "%s",
            "%s"
        ],
        "token_id": "%s"
    }
}`
)

func httppost(params string) (*gjson.Result, error) {
	url := "http://localhost:2080/rpc"
	method := "POST"
	header := http.Header{
		"user":     []string{"test"},
		"password": []string{"testp"},
	}
	client := &http.Client{
		Timeout: time.Second * 1,
	}

	req, err := http.NewRequest(method, url, strings.NewReader(params))
	if err != nil {
		return nil, err
	}
	req.Header = header
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var resultBs []byte

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	resultBs, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	r := gjson.ParseBytes(resultBs)
	if r.Get("code").Int() != 2000 {
		return nil, errors.New(r.String())
	}

	return &r, nil
}
