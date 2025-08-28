package utils

import "regexp"

// dns判定是否正确的正则表达式
const dnsRegexp = `^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`

/*
正则判断地址是否是域名
*/
func IsDNS(dnsName string) (bool, error) {
	isOk, err := regexp.MatchString(dnsRegexp, dnsName)
	if err != nil {
		return false, err
	}
	return isOk, nil
}
