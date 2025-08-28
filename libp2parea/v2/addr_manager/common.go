package addr_manager

import (
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"net"
	"time"
	"web3_gui/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*
检查一个地址的计算机是否在线
@return idOnline    是否在线
*/
func CheckOnline(addr string) (isOnline bool) {
	//尝试连接节点，看是否在线
	utils.Log.Info().Msgf("Try connecting nodes to see if they are online %s", addr)
	conn, err := net.DialTimeout("tcp", addr, time.Second*5)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

/*
其他域地址
*/
type OtherAreaMultiaddrs struct {
	Addrs []string //
}

func NewOtherAreaMultiaddrs(addrs []string) *OtherAreaMultiaddrs {
	return &OtherAreaMultiaddrs{Addrs: addrs}
}

func (this *OtherAreaMultiaddrs) Json() ([]byte, error) {
	return json.Marshal(this)
}

func ParseOtherAreaMultiaddrs(bs []byte) (*OtherAreaMultiaddrs, error) {
	oam := new(OtherAreaMultiaddrs)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(oam)
	if err != nil {
		return nil, err
	}
	return oam, nil
}

/*
检查地址是否可用
*/
//func (this *AddrManager) CheckAddr() {
//	//先获得一个拷贝
//	oldSuperPeerEntry := make(map[string][]byte)
//	this.Sys_superNodeEntry.Range(func(k, v interface{}) bool {
//		key, ok := k.(string)
//		if !ok {
//			return false
//		}
//		value, ok := v.([]byte)
//		if !ok {
//			return false
//		}
//		oldSuperPeerEntry[key] = value
//		return true
//	})
//	// for key, value := range Sys_superNodeEntry {
//	// 	oldSuperPeerEntry[key] = value
//	// }
//	/*
//		一个地址一个地址的判断是否可用
//	*/
//	for key, value := range oldSuperPeerEntry {
//		// utils.Log.Info().Msgf("检查节点 1111111111111")
//		//如果地址是本节点，则退出
//		// if config.Init_LocalIP+":"+strconv.Itoa(int(config.Init_LocalPort)) == key {
//		// 	continue
//		// }
//		if CheckOnline(key) {
//			this.AddSuperPeerAddr(value, key)
//		} else {
//			// delete(Sys_superNodeEntry, key)
//			this.Sys_superNodeEntry.Delete(key)
//		}
//	}
//}

/*
删除一个节点地址
*/
//func (this *AddrManager) RemoveIP(ip string, port uint16) {
//	key := net.JoinHostPort(ip, strconv.Itoa(int(port)))
//	// delete(Sys_superNodeEntry, key)
//	this.Sys_superNodeEntry.Delete(key)
//}

/*
解析超级节点地址列表
*/
//func (this *AddrManager) parseSuperPeerEntry(areaName []byte, fileBytes []byte) []string {
//	var tempSuperPeerEntry map[string]string
//	decoder := json.NewDecoder(bytes.NewBuffer(fileBytes))
//	decoder.UseNumber()
//	err := decoder.Decode(&tempSuperPeerEntry)
//	if err != nil {
//		utils.Log.Error().Msgf("解析超级节点地址列表失败:%s", err.Error())
//		return nil
//	}
//
//	addrs := make([]string, 0, len(tempSuperPeerEntry))
//	for key, _ := range tempSuperPeerEntry {
//		this.AddSuperPeerAddr(areaName, key)
//		addrs = append(addrs, key)
//	}
//	// AddSuperPeerAddr(Path_SuperPeerdomain)
//	return addrs
//}

// //匹配文档中的IP字段
// func RegexpIp(str string) []string {
// 	reg, _ := regexp.Compile(`[[:digit:]]{1,3}\.[[:digit:]]{1,3}\.[[:digit:]]{1,3}\.[[:digit:]]{1,3}`)
// 	s := reg.FindAllString(str, -1)
// 	sort.Strings(s)
// 	return removeDuplicates(s)
// }

// //去除重复字符串和空格Ip限定
// func removeDuplicates(a []string) (ret []string) {
// 	a_len := len(a)
// 	for i := 0; i < a_len; i++ {
// 		if (i > 0 && a[i-1] == a[i]) || len(a[i]) == 0 {
// 			continue
// 		}
// 		if validIPAddress(net.ParseIP(string(a[i]))) {
// 			continue
// 		}
// 		ret = append(ret, a[i])
// 	}
// 	return
// }

// //去除重复字符串和空格
// func removeDuplicatesDns(a []string) (ret []string) {
// 	a_len := len(a)
// 	for i := 0; i < a_len; i++ {
// 		if (i > 0 && a[i-1] == a[i]) || len(a[i]) == 0 {
// 			continue
// 		}
// 		ret = append(ret, a[i])
// 	}
// 	return
// }

// //判断是否为IP且排除内网IP
// func validIPAddress(ip net.IP) bool {
// 	if ip.IsLoopback() {
// 		return true
// 	}
// 	ip4 := ip.To4()
// 	if ip4 == nil {
// 		return true
// 	}
// 	return ip4[0] == 10 || // 10.0.0.0/8
// 		(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
// 		(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16
// 		(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
// }
