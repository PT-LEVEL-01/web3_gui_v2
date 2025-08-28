package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)


//接受数据结构体
type Respond struct {
	Code int `json:"code"`
	Jsonrpc string `json:"jsonrpc"`
	Result interface{} `json:"result"`
}

// Http 发送 http 请求
func Http(url string, method string, header map[string]string, byteData []byte) (err error, data string) {
	client := &http.Client {
	}
	/*byteData, err := json.Marshal(params)
	if err != nil {
		return err, ""
	}*/
	payload := bytes.NewReader(byteData)
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return err, ""
	}
	// 添加请求头
	for key,val := range header{
		req.Header.Add(key, val)
	}
	// 发送请求
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err, ""
	}
	defer res.Body.Close()
	// 读取响应数据
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err, ""
	}
	//fmt.Println(string(body))
	return nil, string(body)
}

// Get 发送 get 请求
func Get(url string, header map[string]string, byteData []byte) (err error, data string) {
	return Http(url, "GET", header, byteData)
}

// Post 发送 post 请求
func Post(url string, header map[string]string, byteData []byte) (err error, data string) {
	return Http(url, "POST", header, byteData)
}


// 把请求的结果转为struct
func JsonToStruct(result string) (Respond Respond, err error) {
	dec := json.NewDecoder(bytes.NewBuffer([]byte(result)))
	dec.UseNumber()
	err = dec.Decode(&Respond)
	return
}