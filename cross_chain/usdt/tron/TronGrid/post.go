package TronGrid

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func Post(url string, heads map[string]string, body []byte) ([]byte, error) {
	client := &http.Client{}
	//url := fmt.Sprintf("%s/wallet/getaccount", c.Config.NetworkURL)
	//requestBody := fmt.Sprintf(`{"address": "%s", "visible": true}`, address)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	for k, v := range heads {
		req.Header.Add(k, v)
	}
	req.Header.Add("Content-Type", "application/json")
	if APIKey != "" {
		req.Header.Add("TRON-PRO-API-KEY", APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	} else {
		return nil, errors.New("http status code:" + strconv.Itoa(resp.StatusCode))
	}
}

/*
获取一个收款地址的所有余额
*/
func GetAccountBalance(address string) (int64, error) {
	url := fmt.Sprintf("%s/wallet/getaccount", Default_tron_net)
	requestBody := fmt.Sprintf(`{"address": "%s", "visible": true}`, address)

	respBody, err := Post(url, nil, []byte(requestBody))
	if err != nil {
		return 0, err
	}
	var result struct {
		Balance int64 `json:"balance"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, err
	}
	return result.Balance, nil
}

/*
创建新账号
*/
func CreateAccount() (*Account, error) {
	url := fmt.Sprintf("%s/wallet/createaccount", Default_tron_net)
	respBody, err := Post(url, nil, nil)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(respBody)
	var account Account
	if err := json.NewDecoder(bodyReader).Decode(&account); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &account, nil
}

// Account represents a TRON network account, including its basic properties.
type Account struct {
	// Address is the account's address in hexadecimal.
	Address string `json:"address"`
	// Balance is the account's balance in Sun (1 TRX = 1,000,000 Sun).
	Balance int64 `json:"balance"`
	// PublicKey is the account's public key/
	PublicKey string `json:"public_key"`
}

/*
转账
*/
func TransferTRX(fromAddress, toAddress string, amount int64) (string, error) {
	url := fmt.Sprintf("%s/wallet/createtransaction", Default_tron_net)
	// Generate a transaction, sign it with the private key, and then broadcast it using the broadcast method.
	payload := map[string]interface{}{
		"to_address":    toAddress,
		"owner_address": fromAddress,
		"amount":        amount,
		"visible":       true,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshalling payload: %w", err)
	}
	head := map[string]string{
		"accept": "application/json",
	}
	respBody, err := Post(url, head, payloadBytes)
	if err != nil {
		return "", err
	}
	var response struct {
		TxID string `json:"txID"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("unmarshalling response: %w", err)
	}
	if response.TxID == "" {
		return "", fmt.Errorf("transaction ID is empty, response: %s", string(respBody))
	}
	return response.TxID, nil
}

/*
广播离线签名交易
*/
func BroadcastTransaction(signedTx string) (string, error) {
	url := fmt.Sprintf("%s/wallet/broadcasttransaction", Default_tron_net)
	requestBody := fmt.Sprintf(`{"raw_data_hex": "%s"}`, signedTx)

	respBody, err := Post(url, nil, []byte(requestBody))
	if err != nil {
		return "", err
	}
	var result struct {
		// Adjust according to the actual API response
		TxID string `json:"txID"`
	}
	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	return result.TxID, nil
}
