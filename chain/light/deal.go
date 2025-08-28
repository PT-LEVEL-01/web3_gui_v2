package light

import (
	"encoding/json"
	"web3_gui/keystore/adapter"
)

func MapChangeAddressInfo(m interface{}) keystore.AddressInfo {
	m1 := m.(map[string]interface{})
	marshal, _ := json.Marshal(m1)
	temp := keystore.AddressInfo{}

	json.Unmarshal(marshal, &temp)
	return temp
}
