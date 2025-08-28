package db

import (
	"math/big"

	"github.com/golang/protobuf/proto"
	"web3_gui/chain/config"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/libp2parea/adapter/engine"
)

type Erc20Info struct {
	From        string   `json:"from"` //发行人
	Address     string   `json:"address"`
	Name        string   `json:"name"`
	Symbol      string   `json:"symbol"`
	Decimals    uint8    `json:"decimals"`
	TotalSupply *big.Int `json:"-"`
	Sort        int      `json:"-"`
}

type Erc20Sort []Erc20Info

func (e Erc20Sort) Len() int {
	return len(e)
}

func (e Erc20Sort) Less(i, j int) bool {
	return e[i].Sort < e[j].Sort
}

func (e Erc20Sort) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type Erc721Info struct {
	From        string   `json:"from"` //发行人
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	TotalSupply *big.Int `json:"-"`
}

type Erc1155Info struct {
	From        string   `json:"from"` //发行人
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	TotalSupply *big.Int `json:"-"`
}

// 保存erc20合约
func SaveErc20Info(info Erc20Info) error {
	key := []byte(config.DBKEY_ERC20_CONTRACT)
	erc := go_protos.Erc20{
		From:        info.From,
		Name:        info.Name,
		Symbol:      info.Symbol,
		Decimals:    uint32(info.Decimals),
		TotalSupply: info.TotalSupply.String(),
	}
	ercBs, err := erc.Marshal()
	if err != nil {
		engine.Log.Error("保存erc20合约失败：%s", err.Error())
		return err
	}

	_, err = LevelDB.HSet(key, []byte(info.Address), ercBs)
	if err != nil {
		engine.Log.Error("保存erc20合约失败：%s", err.Error())
		return err
	}

	return nil
}

func GetAllErc20Info() []Erc20Info {
	key := []byte(config.DBKEY_ERC20_CONTRACT)
	values, err := LevelDB.HGetAll(key)
	if err != nil {
		return nil
	}

	list := make([]Erc20Info, 0)

	for _, v := range values {
		e20 := new(go_protos.Erc20)
		err := proto.Unmarshal(v.Value, e20)
		if err != nil {
			continue
		}
		totalSupply, ok := new(big.Int).SetString(e20.TotalSupply, 10)
		if !ok {
			totalSupply = big.NewInt(0)
		}
		list = append(list, Erc20Info{
			From:        e20.From,
			Address:     string(v.Field),
			Name:        e20.Name,
			Symbol:      e20.Symbol,
			Decimals:    uint8(e20.Decimals),
			TotalSupply: totalSupply,
		})
	}

	return list
}

func GetErc20Info(addr string) Erc20Info {
	key := []byte(config.DBKEY_ERC20_CONTRACT)
	value, err := LevelDB.HGet(key, []byte(addr))
	if err != nil {
		return Erc20Info{}
	}

	e20 := new(go_protos.Erc20)
	err = proto.Unmarshal(value, e20)
	if err != nil {
		return Erc20Info{}
	}

	totalSupply, ok := new(big.Int).SetString(e20.TotalSupply, 10)
	if !ok {
		totalSupply = big.NewInt(0)
	}
	return Erc20Info{
		From:        e20.From,
		Address:     addr,
		Name:        e20.Name,
		Symbol:      e20.Symbol,
		Decimals:    uint8(e20.Decimals),
		TotalSupply: totalSupply,
	}
}

// 保存erc721合约
func SaveErc721Info(info Erc721Info) error {
	key := []byte(config.DBKEY_ERC721_CONTRACT)
	erc := go_protos.Erc721{
		From:        info.From,
		Name:        info.Name,
		Address:     info.Address,
		TotalSupply: info.TotalSupply.String(),
	}
	ercBs, err := erc.Marshal()
	if err != nil {
		engine.Log.Error("保存erc721合约失败：%s", err.Error())
		return err
	}

	_, err = LevelDB.HSet(key, []byte(info.Address), ercBs)
	if err != nil {
		engine.Log.Error("保存erc721合约失败：%s", err.Error())
		return err
	}

	return nil
}

func GetErc721Info(addr string) Erc721Info {
	key := []byte(config.DBKEY_ERC721_CONTRACT)
	value, err := LevelDB.HGet(key, []byte(addr))
	if err != nil {
		return Erc721Info{}
	}

	e721 := new(go_protos.Erc721)
	err = proto.Unmarshal(value, e721)
	if err != nil {
		return Erc721Info{}
	}

	totalSupply, ok := new(big.Int).SetString(e721.TotalSupply, 10)
	if !ok {
		totalSupply = big.NewInt(0)
	}
	return Erc721Info{
		From:        e721.From,
		Name:        e721.Name,
		Address:     addr,
		TotalSupply: totalSupply,
	}
}

// 保存erc1155合约
func SaveErc1155Info(info Erc1155Info) error {
	key := []byte(config.DBKEY_ERC1155_CONTRACT)
	erc := go_protos.Erc1155{
		From:        info.From,
		Name:        info.Name,
		Address:     info.Address,
		TotalSupply: info.TotalSupply.String(),
	}
	ercBs, err := erc.Marshal()
	if err != nil {
		engine.Log.Error("保存erc1155合约失败：%s", err.Error())
		return err
	}

	_, err = LevelDB.HSet(key, []byte(info.Address), ercBs)
	if err != nil {
		engine.Log.Error("保存erc1155合约失败：%s", err.Error())
		return err
	}

	return nil
}

func GetErc1155Info(addr string) Erc1155Info {
	key := []byte(config.DBKEY_ERC1155_CONTRACT)
	value, err := LevelDB.HGet(key, []byte(addr))
	if err != nil {
		return Erc1155Info{}
	}

	e1155 := new(go_protos.Erc1155)
	err = proto.Unmarshal(value, e1155)
	if err != nil {
		return Erc1155Info{}
	}

	totalSupply, ok := new(big.Int).SetString(e1155.TotalSupply, 10)
	if !ok {
		totalSupply = big.NewInt(0)
	}
	return Erc1155Info{
		From:        e1155.From,
		Name:        e1155.Name,
		Address:     addr,
		TotalSupply: totalSupply,
	}
}
