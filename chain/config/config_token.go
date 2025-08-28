/*
 * @Author: liaolei 136552740@qq.com
 * @Date: 2022-12-21 10:03:29
 * @LastEditors: liaolei 136552740@qq.com
 * @LastEditTime: 2022-12-21 14:44:55
 * @FilePath: \icomim-dev\config\config_rpc.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package config

/*
limit
*/

const (
	MethodDefault = "default"
	DriverToken   = "token"
	DriverToken2  = "token2"
)

var (
	DriverDefault = DriverToken2

	HandleTxTokens uint64 = 8000000
	DefaultTxToken uint64 = 1000
)

type LimitConf struct {
	Limit uint64
	Burst uint64
}

//限流器定义
var ReqLimiterMap = map[string]*LimitConf{
	//"default":       HandleTxTokens,
	//"handleTxLimit": HandleTxTokens, //交易限流Gas最大10000000

	"default":       {HandleTxTokens / 5, HandleTxTokens},
	"sendtoaddress": {HandleTxTokens / 5, HandleTxTokens},
}
