package addr_manager

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/utils"
)

/*
读取并解析本地的超级节点列表文件
*/
func (this *AddrManager) LoadSuperPeerEntry(areaName []byte) []string {

	//读取大于自己地址的最大连接数
	//eg : {"p2p_max_greater_conn":10}
	var tmpMaxConn int
	configFileBytes, err := ioutil.ReadFile(this.p2pMaxConnPath)
	if err != nil {
		utils.Log.Warn().Msgf("读取p2p配置文件失败")
	} else {
		var tmpconfig map[string]interface{}
		decoder := json.NewDecoder(bytes.NewBuffer(configFileBytes))
		err := decoder.Decode(&tmpconfig)
		if err != nil {
			utils.Log.Error().Msgf("解析p2p配置文件失败:%s", err.Error())
		}
		if len(tmpconfig) != 0 {
			if i, ok := tmpconfig["p2p_max_greater_conn"]; ok {
				iv, ok2 := i.(float64)
				if ok2 {
					tmpMaxConn = int(iv)
				} else {
					tmpMaxConn = 2
				}
			}
			if w, ok := tmpconfig["p2p_only_connect_list"]; ok {
				wTmp := make([]string, 0)
				wb, _ := json.Marshal(w)
				if err := json.Unmarshal(wb, &wTmp); err == nil {
					config.OnlyConnectList = wTmp
				} else {
					utils.Log.Info().Msgf(err.Error())
				}
			}
		}
	}
	if tmpMaxConn <= 0 {
		//GreaterThanSelfMaxConn 默认参数
		// config.GreaterThanSelfMaxConn = 20
		config.GreaterThanSelfMaxConn = 20
	} else {
		config.GreaterThanSelfMaxConn = tmpMaxConn
	}
	utils.Log.Info().Msgf("新规则中 节点单方向最大连接数 : %d 本节点大区哈希值 : %s", config.GreaterThanSelfMaxConn, hex.EncodeToString(areaName))
	if len(config.Entry) > 0 {
		for _, value := range config.Entry {
			this.AddSuperPeerAddr(areaName, value)
		}
		return config.Entry
		// AddSuperPeerAddr(Path_SuperPeerdomain)
	} else {

		fileBytes, err := ioutil.ReadFile(this.discoverEntryPath)
		if err != nil {
			// fmt.Println("读取超级节点地址列表失败", err)
			return nil
		}
		return this.parseSuperPeerEntry(areaName, fileBytes)
	}
}
