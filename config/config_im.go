package config

import (
	"encoding/base64"
	"strings"
	"time"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

const (
	//聊天消息内容的类型
	MSG_type_text         = 1 //普通文字类型消息
	MSG_type_file_old     = 2 //文件传输 旧版
	MSG_type_image_base64 = 3 //图片base64编码
	MSG_type_file         = 4 //发送文件
	MSG_type_voice        = 5 //语音
	MSG_type_payment      = 6 //转账

	IMPROXY_datachain_status_notSend     = 1 //未发送状态
	IMPROXY_datachain_status_sendSuccess = 2 //发送成功状态

	//后端向前端及时通知类型
	SUBSCRIPTION_type_msg              = 1  //新聊天消息
	SUBSCRIPTION_type_addFriend        = 2  //申请添加好友
	SUBSCRIPTION_type_agreeFriend      = 3  //同意添加好友
	SUBSCRIPTION_type_update_userinfo  = 4  //好友个人信息更新
	SUBSCRIPTION_type_group_members    = 5  //群成员变动
	SUBSCRIPTION_type_msg_del          = 6  //双删单条消息
	SUBSCRIPTION_type_msg_clean        = 7  //双删所有消息
	SUBSCRIPTION_sharebox_fileHash     = 8  //计算文件hash异步任务推送
	SUBSCRIPTION_chain_payOrder_server = 11 //链端->收款端，支付订单已上链
	SUBSCRIPTION_chain_payOrder_client = 12 //链端->付款端，支付订单已上链

	MsgChanMaxLength = 10 //未发送成功消息最大数量

	//消息状态
	MSG_GUI_state_not_send = 1 //发送未送达
	MSG_GUI_state_success  = 2 //发送成功
	MSG_GUI_state_fail     = 3 //发送失败
	MSG_GUI_state_read     = 4 //已读

	ScreenShotMaxLength       = 100 * utils.KB                   //屏幕截图控制最大字节数
	ScreenShotMinLength       = 50 * utils.KB                    //屏幕截图控制最大字节数
	DataChainBlockSizeMax     = utils.MB / 10                    //单个数据链块内容最大容量
	DataChainBlockContentSize = DataChainBlockSizeMax - utils.KB //数据链保存文件时，除开头和尾，剩下的空间容量
	FILE_image_size_max       = 200 * utils.MB                   //发送的图片文件能直接显示的大小

	FILE_type_image_base64 = 1 //base64编码图片
	FILE_type_image_binary = 2 //图片文件
	FILE_type_video        = 3 //视频
	FILE_type_file         = 4 //文件

	IMPROXY_Command_server_init             = 1  //初始化
	IMPROXY_Command_server_forward          = 2  //转发消息
	IMPROXY_Command_server_msglog_add       = 3  //添加消息
	IMPROXY_Command_server_msglog_del       = 4  //删除消息
	IMPROXY_Command_server_setup            = 5  //设置消息
	IMPROXY_Command_server_group_create     = 6  //创建一个群
	IMPROXY_Command_server_group_update     = 7  //修改一个群
	IMPROXY_Command_server_group_members    = 8  //添加或删除群成员
	IMPROXY_Command_server_group_quit       = 9  //成员退出群聊
	IMPROXY_Command_server_group_dissolve   = 10 //解散群聊
	IMPROXY_Command_server_group_msglog_add = 11 //添加群消息
	IMPROXY_Command_server_group_msglog_del = 12 //删除群消息
	IMPROXY_Command_server_mark_receipt     = 13 //标记消息已接收

	IMPROXY_Command_client_addFriend        = 51 //申请添加好友
	IMPROXY_Command_client_agreeFriend      = 52 //同意添加好友
	IMPROXY_Command_client_del              = 53 //删除好友
	IMPROXY_Command_client_history_msg      = 54 //聊天历史记录
	IMPROXY_Command_client_sendText         = 55 //发送文本消息
	IMPROXY_Command_client_group_apply      = 56 //申请入群
	IMPROXY_Command_client_group_invitation = 57 //邀请入群
	IMPROXY_Command_client_group_accept     = 58 //接受邀请入群
	IMPROXY_Command_client_group_sendText   = 59 //发送群聊消息
	IMPROXY_Command_client_group_addMember  = 60 //管理员同意添加用户入群
	IMPROXY_Command_client_file             = 61 //发送文件
	IMPROXY_Command_client_remarksname      = 62 //备注昵称
	IMPROXY_Command_client_voice            = 63 //发送语音消息

	IMPROXY_client_download_interval = time.Minute / 2     //下载同步心跳间隔时间
	DissolveGroupOverTime            = time.Hour * 24 * 30 //群解散后，超过这个时间就可以清除数据了
	DHPUK_version_1                  = 1                   //
	IMPROXY_sync_total_once          = 100                 //一次同步记录总条数
	IM_nickname_length_max           = 100                 //昵称和备注昵称最大长度

)

var (
	Wallet_keystore_default_pwd = "123456789" //钱包默认密码 "123456789"

	FLOOD_key_searchUserInfo = BuildFloodKey(1000001) //搜索用户信息
	FLOOD_key_addFriend      = BuildFloodKey(1000002) //添加好友
	FLOOD_key_agreeFriend    = BuildFloodKey(1000003) //同意好友申请
	FLOOD_key_sendMsg        = BuildFloodKey(1000004) //发送消息

	DHPUK_info_version_1 = utilsleveldb.RegDbKeyExistPanicByUint64(DHPUK_version_1) //个人DH公钥信息版本号

	SyncNewBackoffTimer = []time.Duration{0, time.Second * 10, time.Second * 20, time.Second * 40, time.Second * 80,
		time.Minute * 2, time.Minute * 4, time.Minute * 8, time.Minute * 13}

	IM_RetrySend_interval = []int{1} //消息重试间隔轮次
	//IM_RetrySend_interval = []int{1, 2, 3, 5, 15, 30} //消息重试间隔轮次
)

/*
一个不会出错的方法
*/
func BuildFloodKey(keyNumber uint64) []byte {
	bs := utils.Uint64ToBytesByBigEndian(keyNumber)
	ok := engine.RegisterClassName(bs)
	if !ok {
		panic("key number exist")
	}
	return bs
}

func CreateFloodKey(major, minor []byte) (string, utils.ERROR) {
	dbkey, ERR := utilsleveldb.BuildLeveldbKey(minor)
	if !ERR.CheckSuccess() {
		return "", ERR
	}
	bs := append(major, dbkey.Byte()...)
	return utils.Bytes2string(bs), utils.NewErrorSuccess()
}

/*
构建公钥信息
*/
func BuildDhPukInfoV1(dhPuk []byte) ([]byte, utils.ERROR) {
	//版本号，方便以后升级
	//dhPuk := Area.Keystore.GetDHKeyPair().KeyPair.GetPublicKey()
	dhPukBs, ERR := utilsleveldb.LeveldbBuildKey(dhPuk[:])
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dhPukInfo := append(DHPUK_info_version_1.Byte(), dhPukBs...)
	return dhPukInfo, utils.NewErrorSuccess()
}

/*
解析公钥信息
*/
func ParseDhPukInfoV1(dhPukInfo []byte) (*keystore.Key, utils.ERROR) {
	//解析公钥信息
	keys, ERR := utilsleveldb.LeveldbParseKeyMore(dhPukInfo)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//解析公钥版本号
	versionBs, ERR := keys[0].BaseKey()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	version := utils.BytesToUint64ByBigEndian(versionBs)
	if version != DHPUK_version_1 {
		return nil, utils.NewErrorBus(ERROR_CODE_IM_dh_version_unknown, "")
	}
	//解析公钥内容
	dhPuk, ERR := keys[1].BaseKey()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	var memberPuk keystore.Key
	copy(memberPuk[:], dhPuk)
	return &memberPuk, utils.NewErrorSuccess()
}

/*
构建图片base64编码信息
@return    string    mimeType
@return    []byte    二进制图片内容字节码
*/
func BuildImgBase64(mimeType string, imgBinary []byte) string {
	imgBase64Str := base64.StdEncoding.EncodeToString(imgBinary)
	//"data:image/png;base64"
	imgStr := "data:" + mimeType + ";base64," + imgBase64Str
	utils.Log.Info().Str("从新编码后的图片", imgStr[:30]).Str("end", imgStr[len(imgStr)-10:]).Send()
	return imgStr
}

/*
构建图片base64编码信息
@return    string    mimeType
@return    []byte    二进制图片内容字节码
*/
func ParseImgBase64(imgBase64 string) (string, []byte, utils.ERROR) {
	utils.Log.Info().Str("解析图片源编码", imgBase64[:30]).Str("end", imgBase64[len(imgBase64)-10:]).Send()
	strs := strings.SplitN(imgBase64, ",", 2)
	if len(strs) != 2 {
		utils.Log.Info().Msgf("编码错误")
		return "", nil, utils.NewErrorBus(ERROR_CODE_IM_imgBase64_code_fail, "")
	}
	imgBasea64 := strs[1]
	strs = strings.SplitN(strs[0], ";", 2)
	if len(strs) != 2 {
		utils.Log.Info().Msgf("编码错误")
		return "", nil, utils.NewErrorBus(ERROR_CODE_IM_imgBase64_code_fail, "")
	}
	strs = strings.SplitN(strs[0], ":", 2)
	if len(strs) != 2 {
		utils.Log.Info().Msgf("编码错误")
		return "", nil, utils.NewErrorBus(ERROR_CODE_IM_imgBase64_code_fail, "")
	}
	//utils.Log.Info().Msgf("解析出来的图片编码:%d %s", len(strs[1]), strs[1][:500])
	mime := strs[1]
	imgBinary, err := base64.StdEncoding.DecodeString(imgBasea64)
	if err != nil {
		utils.Log.Error().Msgf("解析图片base64编码 错误:%s", err.Error())
		return "", nil, utils.NewErrorSysSelf(err)
	}
	return mime, imgBinary, utils.NewErrorSuccess()
}

/*
构建图片base64编码信息
*/
func BuildImgBase64InfoV1(imgBase64 string) ([]byte, utils.ERROR) {
	//if len(imgBase64) > 100 {
	//	utils.Log.Info().Msgf("base64图片编码源码:%s  %s", imgBase64[:100], imgBase64[len(imgBase64)-100:])
	//}
	strs := strings.SplitN(imgBase64, ",", 2)
	if len(strs) != 2 {
		utils.Log.Info().Msgf("编码错误")
		return nil, utils.NewErrorBus(ERROR_CODE_IM_imgBase64_code_fail, "")
	}
	//utils.Log.Info().Msgf("解析出来的图片编码:%d %s", len(strs[1]), strs[1][:500])
	mime := []byte(strs[0])
	imgBinary, err := base64.StdEncoding.DecodeString(strs[1])
	if err != nil {
		utils.Log.Error().Msgf("解析图片base64编码 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	imgInfo := make([]byte, len(mime)+len(imgBinary)+8+8)
	copy(imgInfo[:8], utils.Uint64ToBytesByBigEndian(uint64(len(mime))))
	copy(imgInfo[8:8+len(mime)], mime)
	copy(imgInfo[8+len(mime):8+len(mime)+8], utils.Uint64ToBytesByBigEndian(uint64(len(imgBinary))))
	copy(imgInfo[8+len(mime)+8:], imgBinary)
	return imgInfo, utils.NewErrorSuccess()
}

/*
解析图片base64编码信息
*/
func ParseImgBase64InfoV1(imgInfo []byte) (string, utils.ERROR) {
	imgMimeLen := utils.BytesToUint64ByBigEndian(imgInfo[:8])
	imgMine := imgInfo[8 : 8+imgMimeLen]
	imgBinary := imgInfo[8+imgMimeLen+8:]
	mimeStr := string(imgMine)
	imgBase64Str := base64.StdEncoding.EncodeToString(imgBinary)
	imgStr := mimeStr + "," + imgBase64Str
	//utils.Log.Info().Msgf("解析的图片base64编码信息:%s", imgStr)
	//if len(imgStr) > 100 {
	//	utils.Log.Info().Msgf("base64图片解码源码:%s  %s", imgStr[:100], imgStr[len(imgStr)-100:])
	//}
	return imgStr, utils.NewErrorSuccess()
}
