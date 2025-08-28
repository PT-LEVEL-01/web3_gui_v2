package keystore

type DHKeyPair struct {
	Index     uint64  `json:"index"`     //棘轮数量
	KeyPair   KeyPair `json:"keypair"`   //
	CheckHash []byte  `json:"checkhash"` //主私钥和链编码加密验证hash值
	SubKey    []byte  `json:"subKey"`    //子密钥
}
