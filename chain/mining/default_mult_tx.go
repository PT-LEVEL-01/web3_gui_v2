package mining

import (
	"bytes"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/keystore/adapter/crypto"
)

// 多签交易默认接口
// 实现该接口的交易只需调用 RequestMultTx 方法即可自处理内部流程
type DefaultMultTx interface {
	GetMultAddress() crypto.AddressCoin
	GetMultVins() []*Vin
	RequestMultTx(tx TxItr) error
	HandleMultTx(tx TxItr) error
}

type defaultMultTxImpl struct {
	MultAddress crypto.AddressCoin `json:"m_a"` //多签地址
	MultVins    []*Vin             `json:"m_v"` //多签集合,保证顺序,验证 2f/3 + 1
}

func NewDefaultMultTxImpl(multAddress crypto.AddressCoin, multVins []*Vin) DefaultMultTx {
	return &defaultMultTxImpl{
		MultAddress: multAddress,
		MultVins:    multVins,
	}
}

// 发起请求:多签交易
func (m *defaultMultTxImpl) RequestMultTx(txItr TxItr) error {
	//广播
	MulticastTx(txItr)

	//处理多签
	if err := m.HandleMultTx(txItr); err != nil {
		if err == config.ERROR_tx_multaddr_not_found || err == config.ERROR_tx_multaddr_pending {
			return err
		}
	} else {
		GetLongChain().TransactionManager.AddTx(txItr)
	}

	return nil
}

func (m *defaultMultTxImpl) GetMultAddress() crypto.AddressCoin {
	return m.MultAddress
}

func (m *defaultMultTxImpl) GetMultVins() []*Vin {
	return m.MultVins
}

// 处理多签交易
func (m *defaultMultTxImpl) HandleMultTx(txItr TxItr) error {
	class := txItr.Class()
	_, ok := GetMultsignAddrSet(m.MultAddress)
	if !ok {
		return config.ERROR_tx_multaddr_not_found
	}

	txid := *txItr.GetHash()
	key := config.BuildMultsignRequestTx(txid)
	txBs, err := db.LevelDB.Find(key)
	if err == nil && len(*txBs) != 0 {
		//聚合签名
		cacheTxItr, err := ParseTxBaseProto(class, txBs)
		if err != nil {
			return config.ERROR_tx_multaddr_pending
		}
		cachemtx := cacheTxItr.(DefaultMultTx)
		if len(cachemtx.GetMultVins()) == len(m.GetMultVins()) {
			for i, _ := range cachemtx.GetMultVins() {
				//索引位置相同,公钥相同,则合并签名
				if bytes.Equal(cachemtx.GetMultVins()[i].Puk, m.MultVins[i].Puk) && (len(m.MultVins[i].Sign) == 0) {
					m.GetMultVins()[i].Sign = cachemtx.GetMultVins()[i].Sign
				}
			}
		}
	}

	if txBs, err := txItr.Proto(); err == nil {
		db.LevelDB.Save(key, txBs)
		db.LevelDB.Commit(key)
	}

	//只有多签签名合法,才能添加到交易池
	if err := txItr.CheckSign(); err != nil {
		return err
	}

	return nil
}
