package im

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"web3_gui/chain_boot/object_beans"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/im/protos/go_protos"
	"web3_gui/im/subscription"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/cake/transfer_manager"
	"web3_gui/libp2parea/v2/message_center"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func RegisterHandlers() {
	RegisterRPC(Node)
	Node.Register_multicast(config.MSGID_multicast_message_recv, RecvMulticastMessage) //获取广播消息
	//Node.Register_p2pHE(config.MSGID_get_friend_info, GetFriendInfoP2P)                //获取好友基本信息
	//Node.Register_p2pHE(config.MSGID_get_friend_info_recv, GetFriendInfoP2P_recv)      //获取好友基本信息 返回
	//Node.Register_p2pHE(config.MSGID_add_friend, AddFriendP2P)                            //申请添加好友
	//Node.Register_p2pHE(config.MSGID_add_friend_recv, AddFriendP2P_recv)                  //申请添加好友 返回
	//Node.Register_p2pHE(config.MSGID_agree_add_friend, AgreeAddFriendP2P)                 //同意添加好友
	//Node.Register_p2pHE(config.MSGID_agree_add_friend_recv, AgreeAddFriendP2P_recv)       //同意添加好友 返回
	//Node.Register_p2pHE(config.MSGID_send_friend_message, SendFriendMessageP2P)           //发送私聊消息
	//Node.Register_p2pHE(config.MSGID_send_friend_message_recv, SendFriendMessageP2P_recv) //发送私聊消息 返回
	Node.Register_multicast(config.MSGID_multicast_online_recv, RecvMulticastOnline) //用户上线广播消息

	Node.Register_multicast(config.MSGID_IM_PROXY_multicast_nodeinfo_recv, MulticastImProxy_recv) //提供离线消息存储节点广播消息 返回
	Node.Register_multicast(config.MSGID_IM_PROXY_multicast_search, MulticastSearchProxy)         //广播搜索某节点的代理节点
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_multicast_search_recv, MulticastSearchProxy_recv)   //广播搜索某节点的代理节点 返回
	Node.Register_p2pHE(config.MSGID_IM_PROXY_getorders, GetOrders) //客户端向代理获取一个租用时间的订单
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_getorders_recv, GetOrders_recv)                     //客户端向代理获取一个租用时间的订单 返回
	Node.Register_p2pHE(config.MSGID_IM_PROXY_getRenewalOrders, GetRenewalOrders) //客户端向代理获取续费订单
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_getRenewalOrders_recv, GetRenewalOrders_recv)       //客户端向代理获取续费订单 返回

	Node.Register_p2pHE(config.MSGID_IM_PROXY_getDataChainIndex, GetDataChainIndex) //查询代理节点已经同步的数据链索引
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_getDataChainIndex_recv, GetDataChainIndex_recv) //查询代理节点已经同步的数据链索引 返回
	Node.Register_p2pHE(config.MSGID_IM_PROXY_upload_datachain, UploadDatachain) //上传数据链
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_upload_datachain_recv, UploadDatachain_recv)                           //上传数据链 返回
	Node.Register_p2pHE(config.MSGID_IM_PROXY_sync_datachain, SyncDatachainForClient) //从客户端同步数据链
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_sync_datachain_recv, SyncDatachainForClient_recv)                      //从客户端同步数据链 返回
	Node.Register_p2pHE(config.MSGID_IM_PROXY_send_datachain, SendDatachain) //发送数据链
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_send_datachain_recv, SendDatachain_recv)                               //发送数据链 返回
	Node.Register_p2pHE(config.MSGID_IM_PROXY_download_datachain, DownloadDatachain) //下载数据链
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_download_datachain_recv, DownloadDatachain_recv)                       //下载数据链 返回
	Node.Register_p2pHE(config.MSGID_IM_PROXY_group_send_datachain, SendGroupDatachain) //发送群聊消息
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_group_send_datachain_recv, SendGroupDatachain_recv)                    //发送群聊消息 返回
	Node.Register_p2pHE(config.MSGID_IM_PROXY_group_multicast_datachain, multicaseGroupDatachain)          //群内广播消息
	Node.Register_p2pHE(config.MSGID_IM_PROXY_group_download_datachain_start, DownloadGroupDatachainStart) //成员查询群数据链起始记录
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_group_download_datachain_start_recv, DownloadGroupDatachainStart_recv) //成员查询群数据链起始记录 返回
	Node.Register_p2pHE(config.MSGID_IM_PROXY_group_download_datachain, DownloadGroupDatachain) //下载群聊消息
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_group_download_datachain_recv, DownloadGroupDatachain_recv)            //下载群聊消息 返回
	Node.Register_p2pHE(config.MSGID_IM_PROXY_get_sendindex, GetSendIndex) //获取自己的sendindex
	//Node.Register_p2pHE(config.MSGID_IM_PROXY_get_sendindex_recv, GetSendIndex_recv)                                 //获取自己的sendindex 返回

	Node.Register_p2pHE(config.MSGID_file_getShareboxList, GetShareboxListP2P) //查询共享文件列表
	//Node.Register_p2pHE(config.MSGID_file_getShareboxList_recv, GetShareboxListP2P_recv) //查询共享文件列表 返回

	Node.Register_p2pHE(config.MSGID_circle_getClassNames, GetClassNames) //查询好友博客圈子列表
	//Node.Register_p2pHE(config.MSGID_circle_getClassNames_recv, GetClassNames_recv)         //查询好友博客圈子列表 返回
	Node.Register_p2pHE(config.MSGID_circle_multicast_news_recv, MulticastCircleClass_recv) //查询好友博客圈子列表 返回

	Node.Register_p2pHE(config.MSGID_SHAREBOX_Order_getOrder, Sharebox_GetOrder) //获取订单

}

/*
给对方返回一个系统错误
*/
func ReplyError(area *libp2parea.Node, version uint64, message *message_center.MessageBase, code uint64, msg string) {
	nr := utils.NewNetResult(version, code, msg, nil)
	bs, err := nr.Proto()
	if err != nil {
		area.SendP2pReplyMsgHE(message, nil)
		return
	}
	area.SendP2pReplyMsgHE(message, bs)
}

/*
给对方返回成功
*/
func ReplySuccess(area *libp2parea.Node, version uint64, message *message_center.MessageBase, bs *[]byte) {
	resultBs := []byte{}
	if bs != nil {
		resultBs = *bs
	}
	nr := utils.NewNetResult(version, config.ERROR_CODE_success, "", resultBs)
	bs, err := nr.Proto()
	if err != nil {
		area.SendP2pReplyMsgHE(message, nil)
		return
	}
	area.SendP2pReplyMsgHE(message, bs)
}

/*
接收广播消息
*/
func RecvMulticastMessage(message *message_center.MessageBase) {
	//utils.Log.Info().Msgf("接收到广播消息 :%d", len(message.Content))
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		utils.Log.Info().Msgf("接收云存储空间提供者广播，解析参数出错:%s", err.Error())
		return
	}
	//utils.Log.Info().Msgf("广播消息内容:%+v", np.Data)
	msgInfo := model.NewMessageInfo(config.SUBSCRIPTION_type_msg, config.MSG_type_text, false, message.SenderAddr.B58String(),
		"", string(np.Data), time.Now().Unix(), 0, nil, "", "", 0, nil)
	subscription.AddMuticastMsg(msgInfo)
}

/*
接收用户上线广播消息
*/
func RecvMulticastOnline(message *message_center.MessageBase) {
	//utils.Log.Info().Msgf("接收到广播消息")
	_, err := utils.ParseNetParams(message.Content)
	if err != nil {
		utils.Log.Info().Msgf("接收云存储空间提供者广播，解析参数出错:%s", err.Error())
		return
	}
	StaticProxyServerManager.syncServer.ForwardDataChain(*message.SenderAddr)
}

/*
获取好友基本信息
*/
//func GetFriendInfoP2P(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	//utils.Log.Info().Msgf("获取好友基本信息 1111111")
//	replyMsgID := uint64(config.MSGID_get_friend_info_recv)
//	_, err := utils.ParseNetParams(*message.Body.Content)
//	if err != nil {
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_params_format, err.Error())
//		return
//	}
//	userinfo, ERR := db.GetSelfInfo(Node.GetNetId())
//	if !ERR.CheckSuccess() {
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
//		//Node.SendP2pReplyMsgHE(message, config.MSGID_get_friend_info_recv, &config.ERROR_byte_nomarl)
//		return
//	}
//	// utils.Log.Info().Msgf("获取好友基本信息 22222")
//	if userinfo == nil {
//		userinfo = model.NewUserInfo(Node.GetNetId())
//	}
//
//	classNames, ERR := GetClass()
//	if ERR.CheckFail() {
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, ERR.Code, ERR.Msg)
//		//Node.SendP2pReplyMsgHE(message, config.MSGID_get_friend_info_recv, &config.ERROR_byte_nomarl)
//		return
//	}
//	userinfo.CircleClass = make([]model.ClassCount, 0)
//	for i, one := range classNames {
//		count, err := db.FindNewsCount(config.DBKEY_circle_news_release, one)
//		if err != nil {
//			ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
//			//Node.SendP2pReplyMsgHE(message, config.MSGID_get_friend_info_recv, &config.ERROR_byte_nomarl)
//			return
//		}
//		userinfo.CircleClass = append(userinfo.CircleClass, model.ClassCount{
//			Name:  classNames[i],
//			Count: count,
//		})
//	}
//
//	//版本号，方便以后升级
//	dhPuk := Node.Keystore.GetDHKeyPair().KeyPair.GetPublicKey()
//	dhPukInfo, ERR := config.BuildDhPukInfoV1(dhPuk[:])
//	if !ERR.CheckSuccess() {
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, ERR.Code, ERR.Msg)
//		return
//	}
//
//	//dhPukBs, ERR := utilsleveldb.LeveldbBuildKey(dhPuk[:])
//	//if !ERR.CheckSuccess() {
//	//	ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
//	//	return
//	//}
//	//dhPukInfo := append(config.DHPUK_info_version_1.Byte(), dhPukBs...)
//	userinfo.GroupDHPuk = dhPukInfo
//	//utils.Log.Info().Msgf("DH公钥:%+v", userinfo)
//	bs, err := userinfo.Proto()
//	if err != nil {
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
//		//Node.SendP2pReplyMsgHE(message, config.MSGID_get_friend_info_recv, nil)
//		return
//	}
//
//	//utils.Log.Info().Msgf("DH公钥:%d %+v", len(*bs), userinfo)
//	// utils.Log.Info().Msgf("获取好友基本信息 33333")
//	ReplySuccess(Node, config.NET_protocol_version_v1, message, replyMsgID, bs)
//	//Node.SendP2pReplyMsgHE(message, config.MSGID_get_friend_info_recv, bs)
//}

/*
获取好友基本信息 返回
*/
//func GetFriendInfoP2P_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
申请添加好友
*/
//func AddFriendP2P(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	utils.Log.Info().Msgf("申请添加好友p2p 11111111111")
//	replyMsgID := uint64(config.MSGID_add_friend_recv)
//	np, err := utils.ParseNetParams(*message.Body.Content)
//	if err != nil {
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_params_format, err.Error())
//		return
//	}
//	userInfo, err := model.ParseUserInfo(&np.Data)
//	if err != nil {
//		utils.Log.Info().Msgf("申请添加好友p2p 22222222")
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_params_format, err.Error())
//		//Node.SendP2pReplyMsgHE(message, config.MSGID_add_friend_recv, &config.ERROR_byte_nomarl)
//		return
//	}
//
//	//判断好友列表里面有没有
//	userList, err := db.GetUserList(config.DBKEY_friend_userlist)
//	if err != nil {
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
//		//Node.SendP2pReplyMsgHE(message, config.MSGID_add_friend_recv, &config.ERROR_byte_nomarl)
//		return
//	}
//	if userList != nil {
//		for _, one := range userList.UserList {
//			if bytes.Equal(*message.Head.Sender, one.Addr) {
//				ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_IM_In_the_friend_list, "")
//				//Node.SendP2pReplyMsgHE(message, config.MSGID_add_friend_recv, &config.ERROR_byte_exist)
//				return
//			}
//		}
//	}
//
//	utils.Log.Info().Msgf("申请添加好友p2p 22222222")
//	ERR := db.AddUserList(userInfo, config.DBKEY_apply_remote_userlist)
//	if !ERR.CheckSuccess() {
//		utils.Log.Info().Msgf("申请添加好友p2p 22222222")
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, ERR.Code, ERR.Msg)
//		//Node.SendP2pReplyMsgHE(message, config.MSGID_add_friend_recv, &config.ERROR_byte_nomarl)
//		return
//	}
//	utils.Log.Info().Msgf("申请添加好友p2p 22222222")
//	//给前端发送一个通知
//	msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_addFriend}
//	AddSubscriptionMsg(&msgInfo)
//	ReplySuccess(Node, config.NET_protocol_version_v1, message, replyMsgID, nil)
//	//Node.SendP2pReplyMsgHE(message, config.MSGID_add_friend_recv, nil)
//	return
//}

/*
申请添加好友 返回
*/
//func AddFriendP2P_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
同意添加好友
*/
//func AgreeAddFriendP2P(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	replyMsgID := uint64(config.MSGID_agree_add_friend_recv)
//	_, err := utils.ParseNetParams(*message.Body.Content)
//	if err != nil {
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_params_format, err.Error())
//		return
//	}
//	//判断申请列表里面有没有这个好友
//	userList, err := db.GetUserList(config.DBKEY_apply_local_userlist)
//	if err != nil {
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
//		return
//	}
//	var userInfo *model.UserInfo
//	for i, one := range userList.UserList {
//		if bytes.Equal(*message.Head.Sender, one.Addr) {
//			userInfo = userList.UserList[i]
//			break
//		}
//	}
//	if userInfo == nil {
//		//申请列表里面没有，则判断好友列表里面有没有
//		userList, err = db.GetUserList(config.DBKEY_friend_userlist)
//		if err != nil {
//			ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
//			//Node.SendP2pReplyMsgHE(message, config.MSGID_agree_add_friend_recv, &config.ERROR_byte_nomarl)
//			return
//		}
//		for i, one := range userList.UserList {
//			if bytes.Equal(*message.Head.Sender, one.Addr) {
//				userInfo = userList.UserList[i]
//				break
//			}
//		}
//		//好友列表中也没有
//		if userInfo == nil {
//			ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_IM_invalid_Agree_Add_Friend, "")
//			//Node.SendP2pReplyMsgHE(message, config.MSGID_agree_add_friend_recv, &config.ERROR_byte_nomarl)
//			return
//		}
//	}
//	ERR := db.AddUserList(userInfo, config.DBKEY_friend_userlist)
//	if !ERR.CheckSuccess() {
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, ERR.Code, ERR.Msg)
//		//Node.SendP2pReplyMsgHE(message, config.MSGID_add_friend_recv, &config.ERROR_byte_nomarl)
//		return
//	}
//	//给前端发送一个通知
//	msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
//	AddSubscriptionMsg(&msgInfo)
//	ReplySuccess(Node, config.NET_protocol_version_v1, message, replyMsgID, nil)
//	//Node.SendP2pReplyMsgHE(message, config.MSGID_agree_add_friend_recv, nil)
//	return
//}

/*
同意添加好友 返回
*/
//func AgreeAddFriendP2P_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
发送私聊消息
*/
//func SendFriendMessageP2P(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	//utils.Log.Info().Msgf("收到消息")
//	replyMsgID := uint64(config.MSGID_send_friend_message_recv)
//	np, err := utils.ParseNetParams(*message.Body.Content)
//	if err != nil {
//		//utils.Log.Info().Msgf("收到消息")
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_params_format, err.Error())
//		return
//	}
//	//查看是否在好友列表中
//	userinfo, err := FindUserInFriendList(*message.Head.Sender)
//	if err != nil {
//		//utils.Log.Info().Msgf("收到消息")
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
//		return
//	}
//	//不在好友列表中，则添加到好友申请列表中
//	if userinfo == nil {
//		//在好友申请列表中查询
//		userinfo, ERR := FindUserInApplyRemoteList(*message.Head.Sender)
//		if !ERR.CheckSuccess() {
//			//utils.Log.Info().Msgf("收到消息")
//			ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, ERR.Code, ERR.Msg)
//			//Node.SendP2pReplyMsgHE(message, config.MSGID_send_friend_message_recv, &config.ERROR_byte_nomarl)
//			return
//		}
//		//好友申请列表中没有
//		if userinfo == nil {
//			//添加到好友申请列表
//			ERR := db.AddUserList(userinfo, config.DBKEY_apply_remote_userlist)
//			if !ERR.CheckSuccess() {
//				ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, ERR.Code, ERR.Msg)
//				return
//			}
//		} else {
//			ERR := db.UpdateUserList(*userinfo, config.DBKEY_apply_remote_userlist)
//			if !ERR.CheckSuccess() {
//				ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, ERR.Code, ERR.Msg)
//				return
//			}
//		}
//	}
//	//保存聊天记录
//	mc, err := model.ParseMessageContent(&np.Data)
//	if err != nil {
//		//utils.Log.Info().Msgf("收到消息")
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
//		return
//	}
//	//判断消息重复发送
//	if mc.SendID != nil && len(mc.SendID) > 0 {
//		oldMc, ERR := db.FindMessageContentSend(mc.From, mc.To, mc.SendID)
//		if !ERR.CheckSuccess() {
//			ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, ERR.Code, ERR.Msg)
//			return
//		}
//		if oldMc != nil {
//			//utils.Log.Info().Msgf("收到消息")
//			//重复发送消息
//			ReplySuccess(Node, config.NET_protocol_version_v1, message, replyMsgID, nil)
//			return
//		}
//	}
//	//utils.Log.Info().Msgf("收到消息")
//	recvID := ulid.Make()
//	mc.RecvID = recvID[:]
//	mc.From = *message.Head.Sender
//	mc.To = Node.GetNetId()
//	mc.FromIsSelf = false
//	mc.Time = time.Now().Unix()
//	mc.State = config.MSG_GUI_state_success
//	//utils.Log.Info().Msgf("收到消息:%+v", mc)
//	mc, ERR := db.AddMessageHistory(mc)
//	if err != nil {
//		//utils.Log.Info().Msgf("收到消息")
//		ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, ERR.Code, ERR.Msg)
//		//Node.SendP2pReplyMsgHE(message, config.MSGID_send_friend_message_recv, &config.ERROR_byte_nomarl)
//		return
//	}
//	//utils.Log.Info().Msgf("收到消息:%+v", mc)
//	//给前端发送一个通知
//	msgInfo := mc.ConverVO()
//	msgInfo.Subscription = config.SUBSCRIPTION_type_msg
//	//msgInfo := model.NewMessageInfo(config.SUBSCRIPTION_type_msg, mc.Type, false, message.Head.Sender.B58String(),
//	//	config.NetAddr, string(mc.Content), utils.FormatTimeToSecond(time.Now()), mc.PullAndPushID, nil,
//	//	hex.EncodeToString(mc.SendID), hex.EncodeToString(mc.RecvID), config.MSG_GUI_state_success, mc.Index)
//	AddSubscriptionMsg(msgInfo)
//	ReplySuccess(Node, config.NET_protocol_version_v1, message, replyMsgID, nil)
//	//Node.SendP2pReplyMsgHE(message, config.MSGID_send_friend_message_recv, nil)
//}

/*
发送私聊消息 返回
*/
//func SendFriendMessageP2P_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
获取共享文件列表
*/
func GetShareboxListP2P(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_file_getShareboxList_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}

	fileListPro := go_protos.Filelist{
		//ErrorCode: ERROR_CODE_protobuf_marshal_error,
	}
	fileMap, ERR := transfer_manager.TransferMangerStatic.TransferPushTaskSharingDirs()
	if ERR.CheckFail() {
		if ERR.Msg == leveldb.ErrNotFound.Error() {
			bs, err := fileListPro.Marshal()
			if err != nil {
				ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
				return
			}
			ReplySuccess(Node, config.NET_protocol_version_v1, message, &bs)
			return
		}
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, ERR.Msg)
		return
	}
	dirPath := string(np.Data)
	//utils.Log.Info().Msgf("要查询的路径:%s", dirPath)

	// filelist := make([][]byte, 0)
	files := make([]*go_protos.File, 0)
	if dirPath == "" {
		list := make([]string, 0, len(fileMap))
		for k, _ := range fileMap {
			list = append(list, k)
		}
		sort.Strings(list)
		for _, key := range list {
			value := fileMap[key]
			fileInfo, err := os.Stat(value)
			if err != nil {
				continue
			}
			fileOne := go_protos.File{Name: []byte(key), IsDir: true, UpdateTime: fileInfo.ModTime().Unix()}
			files = append(files, &fileOne)
		}
	} else {
		dirPath = strings.Trim(dirPath, " ") //去掉前后空格
		if dirPath != "" {
			dirPath = filepath.Clean(dirPath)     //统一格式化为"\file1\file2"
			dirPath = strings.Trim(dirPath, "\\") //去掉前后的"\"符号
		}
		dirNames := strings.Split(dirPath, "\\")
		if len(dirNames) <= 0 {
			ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_sharebox_Request_path_format_error, dirPath)
			return
		}
		//判断参数中文件名是否在本地共享列表中
		dirAbs, ok := fileMap[dirNames[0]]
		if !ok {
			//此路径不在共享目录中
			ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_sharebox_Request_path_not_found, "")
			return
		}
		//utils.Log.Info().Msgf("绝对路径:%s", dirAbs)

		fileList, ERR := transfer_manager.TransferMangerStatic.TransferPushTaskSharingDirsFind(dirAbs)
		if ERR.CheckFail() {
			ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
			return
		}
		fileHashs := make([][]byte, 0, len(fileList))
		for _, one := range fileList {
			fileHashs = append(fileHashs, one.Hash)
		}
		prices, ERR := db.Sharebox_server_FindGoodsMore(fileHashs)
		if ERR.CheckFail() {
			ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
			return
		}
		pricesMap := make(map[string]*object_beans.OrderShareboxGoods)
		for _, one := range prices {
			pricesMap[utils.Bytes2string(one.GetId())] = one
		}

		for _, fileInfo := range fileList {
			fileOne := go_protos.File{Name: []byte(fileInfo.Name), IsDir: fileInfo.IsDir,
				UpdateTime: fileInfo.Time, Length: uint64(fileInfo.Size)}
			fileOne.Hash = fileInfo.Hash
			pOne, ok := pricesMap[utils.Bytes2string(fileInfo.Hash)]
			if ok {
				if pOne.Name != "" {
					fileOne.Name = []byte(pOne.Name)
				}
				fileOne.Price = pOne.Price
			}
			files = append(files, &fileOne)
		}
		//
		//
		//dirNames[0] = dirAbs //替换为绝对路径
		//dirPath = filepath.Join(dirNames...)
		////utils.Log.Info().Msgf("绝对路径:%s", dirAbs)
		////检查是否文件夹
		//fileinfo, err := os.Stat(dirPath)
		//if err != nil {
		//	ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		//	return
		//}
		//if !fileinfo.IsDir() {
		//	ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_sharebox_Request_path_not_a_folder, "")
		//	//Node.SendP2pReplyMsgHE(message, config.MSGID_file_getShareboxList_recv, &bs)
		//	return
		//}
		//filepathNames, err := ioutil.ReadDir(dirPath)
		//if err != nil {
		//	ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		//	return
		//}
		//for _, fileInfo := range filepathNames {
		//	fileOne := go_protos.File{Name: []byte(fileInfo.Name()), IsDir: fileInfo.IsDir(),
		//		UpdateTime: fileInfo.ModTime().Unix(), Length: uint64(fileInfo.Size())}
		//	fileOne.Price
		//	files = append(files, &fileOne)
		//}
	}
	//fileListPro.ErrorCode = ERROR_CODE_Success
	fileListPro.FileList = files
	bs, err := fileListPro.Marshal()
	if err != nil {
		//utils.Log.Info().Msgf("文件列表")
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		//Node.SendP2pReplyMsgHE(message, config.MSGID_file_getShareboxList_recv, nil)
		return
	}
	//utils.Log.Info().Msgf("文件列表")
	ReplySuccess(Node, config.NET_protocol_version_v1, message, &bs)
	//Node.SendP2pReplyMsgHE(message, config.MSGID_file_getShareboxList_recv, &bs)
}

/*
获取共享文件列表 返回
*/
//func GetShareboxListP2P_recv(message *message_center.MessageBase) {
//	//engine.ResponseByteKey()
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
查询好友博客圈子类别
*/
func GetClassNames(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_circle_getClassNames_recv)
	_, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	names, ERR := db.GetClass(*config.DBKEY_circle_news_release)
	if ERR.CheckFail() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		//Node.SendP2pReplyMsgHE(message, config.MSGID_circle_getClassNames_recv, nil)
		return
	}
	classNames := go_protos.ClassNames{
		ClassList: make([]*go_protos.Class, 0),
	}
	for i, one := range names {
		classOne := go_protos.Class{}
		count, err := db.FindNewsCount(*config.DBKEY_circle_news_release, one)
		if err != nil {
			ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
			//Node.SendP2pReplyMsgHE(message, config.MSGID_circle_getClassNames_recv, nil)
			return
		}
		classOne.Name = []byte(names[i])
		classOne.Size_ = count
		classNames.ClassList = append(classNames.ClassList, &classOne)
	}
	bs, err := classNames.Marshal()
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		//Node.SendP2pReplyMsgHE(message, config.MSGID_circle_getClassNames_recv, nil)
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, &bs)
	//Node.SendP2pReplyMsgHE(message, config.MSGID_circle_getClassNames_recv, &bs)
	return
}

/*
查询好友博客圈子类别 返回
*/
//func GetClassNames_recv(message *message_center.MessageBase) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
用户博客广播 接收广播
*/
func MulticastCircleClass_recv(message *message_center.MessageBase) {
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		utils.Log.Info().Msgf("接收博客广播解析错误:%s", err.Error())
		return
	}
	news, err := model.ParseNews(&np.Data)
	if err != nil {
		utils.Log.Info().Msgf("接收博客广播解析错误:%s", err.Error())
		return
	}
	AddMultcastNews(message.SenderAddr.B58String(), news)
}

/*
提供离线消息存储节点信息 接收广播
*/
func MulticastImProxy_recv(message *message_center.MessageBase) {
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		utils.Log.Info().Msgf("接收博客广播解析错误:%s", err.Error())
		return
	}
	imProxy, err := model.ParseStorageServerInfo(np.Data)
	if err != nil {
		return
	}
	imProxy.Addr = *message.SenderAddr
	AddProxy(imProxy)
}

/*
广播搜索某节点的代理节点 接收广播
*/
func MulticastSearchProxy(message *message_center.MessageBase) {
	//utils.Log.Info().Msgf("收到查询用户信息广播")
	//replyMsgID := uint64(config.MSGID_IM_PROXY_multicast_search_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		utils.Log.Info().Msgf("接收博客广播解析错误:%s", err.Error())
		return
	}
	findAddr := nodeStore.NewAddressNet(np.Data) //nodeStore.AddressNet(np.Data)
	//如果搜索的节点就是自己
	if bytes.Equal(Node.GetNetId().GetAddr(), findAddr.GetAddr()) {
		//utils.Log.Info().Msgf("收到查询用户信息广播")
		userinfo, ERR := db.GetSelfInfo(*Node.GetNetId())
		if ERR.CheckFail() {
			//ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
			//Node.SendP2pReplyMsgHE(message, config.MSGID_get_friend_info_recv, &config.ERROR_byte_nomarl)
			return
		}
		//utils.Log.Info().Msgf("收到查询用户信息广播")
		// utils.Log.Info().Msgf("获取好友基本信息 22222")
		if userinfo == nil {
			userinfo = model.NewUserInfo(*Node.GetNetId())
		}
		classNames, ERR := GetClass()
		if ERR.CheckFail() {
			//ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, ERR.Code, ERR.Msg)
			//Node.SendP2pReplyMsgHE(message, config.MSGID_get_friend_info_recv, &config.ERROR_byte_nomarl)
			return
		}
		//utils.Log.Info().Msgf("收到查询用户信息广播")
		userinfo.CircleClass = make([]model.ClassCount, 0)
		for i, one := range classNames {
			count, err := db.FindNewsCount(*config.DBKEY_circle_news_release, one)
			if err != nil {
				//ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
				//Node.SendP2pReplyMsgHE(message, config.MSGID_get_friend_info_recv, &config.ERROR_byte_nomarl)
				return
			}
			userinfo.CircleClass = append(userinfo.CircleClass, model.ClassCount{
				Name:  classNames[i],
				Count: count,
			})
		}

		//版本号，方便以后升级
		dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
		if !ERR.CheckSuccess() {
			return
		}
		dhPuk := dhKey.GetPublicKey()
		dhPukInfo, ERR := config.BuildDhPukInfoV1(dhPuk[:])
		if !ERR.CheckSuccess() {
			return
		}
		//dhPukBs, ERR := utilsleveldb.LeveldbBuildKey(dhPuk[:])
		//if !ERR.CheckSuccess() {
		//	ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
		//	return
		//}
		//dhPukInfo := append(config.DHPUK_info_version_1.Byte(), dhPukBs...)
		userinfo.GroupDHPuk = dhPukInfo

		//utils.Log.Info().Msgf("收到查询用户信息广播:%+v", userinfo)
		bs, err := userinfo.Proto()
		if err != nil {
			//ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_system_error_self, err.Error())
			//Node.SendP2pReplyMsgHE(message, config.MSGID_get_friend_info_recv, nil)
			return
		}
		// utils.Log.Info().Msgf("获取好友基本信息 33333")
		//ReplySuccess(Node, config.NET_protocol_version_v1, message, replyMsgID, bs)

		nr := utils.NewNetResult(config.NET_protocol_version_v1, config.ERROR_CODE_success, "", *bs)
		bs, err = nr.Proto()
		if err != nil {
			//area.SendP2pReplyMsgHE(message, nil)
			return
		}
		Node.SendMulticastReplyMsg(message, bs)
		//area.SendP2pReplyMsgHE(message, bs)
		//utils.Log.Info().Msgf("收到查询用户信息广播")
		return
	}
	//utils.Log.Info().Msgf("收到查询用户信息广播")
	//查询是否在自己的代理节点中
	_, _, remain := StaticProxyServerManager.QueryUserSpaces(*findAddr)
	if remain == 0 {
		return
	}
	userinfo, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), *findAddr)
	if !ERR.CheckSuccess() {
		return
	}
	if userinfo == nil {
		userinfo = model.NewUserInfo(*findAddr)
	}
	//是自己的代理节点，则添加到代理列表中后返回给查询用户
	userinfo.Proxy.Store(utils.Bytes2string(Node.GetNetId().GetAddr()), Node.GetNetId())
	bs, err := userinfo.Proto()
	if err != nil {
		return
	}
	//ReplySuccess(Node, config.NET_protocol_version_v1, message, replyMsgID, bs)
	nr := utils.NewNetResult(config.NET_protocol_version_v1, config.ERROR_CODE_success, "", *bs)
	bs, err = nr.Proto()
	if err != nil {
		//area.SendP2pReplyMsgHE(message, nil)
		return
	}
	Node.SendMulticastReplyMsg(message, bs)
}

//var MulticastSearchProxy_recvLock = new(sync.Mutex)

/*
广播搜索某节点的代理节点 返回
*/
//func MulticastSearchProxy_recv(message *message_center.MessageBase) {
//	//utils.Log.Info().Msgf("返回个人信息")
//	np, err := utils.ParseNetParams(message.Content)
//	if err != nil {
//		utils.Log.Info().Msgf("接收博客广播解析错误:%s", err.Error())
//		return
//	}
//	userinfo, err := model.ParseUserInfo(&np.Data)
//	if err != nil {
//		utils.Log.Info().Msgf("解析用户信息 错误:%s", err.Error())
//		return
//	}
//
//	//解析公钥信息
//	keys, ERR := utilsleveldb.LeveldbParseKeyMore(userinfo.GroupDHPuk)
//	if !ERR.CheckSuccess() {
//		utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
//		return
//	}
//	//解析公钥版本号
//	versionBs, ERR := keys[0].BaseKey()
//	if !ERR.CheckSuccess() {
//		return
//	}
//	version := utils.BytesToUint64ByBigEndian(versionBs)
//	if version != config.DHPUK_version_1 {
//		return
//	}
//	//解析公钥内容
//	dhPuk, ERR := keys[1].BaseKey()
//	if !ERR.CheckSuccess() {
//		return
//	}
//	userinfo.GroupDHPuk = dhPuk
//
//	//utils.Log.Info().Msgf("返回个人信息")
//	//flood.ResponseItr(utils.Bytes2string(userinfo.Addr), userinfo)
//	//合并代理节点并保存
//	MulticastSearchProxy_recvLock.Lock()
//	defer MulticastSearchProxy_recvLock.Unlock()
//	userinfoDB, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), userinfo.Addr)
//	if !ERR.CheckSuccess() {
//		utils.Log.Error().Msgf("查询用户信息 错误:%s", err.Error())
//		return
//	}
//
//	//utils.Log.Info().Msgf("返回个人信息")
//	userinfo.Proxy.Range(func(k, v interface{}) bool {
//		userinfoDB.Proxy.Store(k, v)
//		return true
//	})
//	//utils.Log.Info().Msgf("返回个人信息:%+v", userinfo)
//	ERR = db.ImProxyClient_SaveUserinfo(*Node.GetNetId(), userinfo)
//	if !ERR.CheckSuccess() {
//		utils.Log.Error().Msgf("保存用户信息 错误:%s", ERR.String())
//		return
//	}
//	//utils.Log.Info().Msgf("返回个人信息")
//	//返回的有多条数据，直到成功才返回等待，不成功不返回
//	//engine.ResponseItrKey(config.FLOOD_key_file_transfer)
//	//flood.ResponseItr(utils.Bytes2string(message.Body.Hash), userinfo)
//	//utils.Log.Info().Msgf("返回个人信息")
//	//个人信息有变化
//	if userinfoDB == nil || userinfo.Nickname != userinfoDB.Nickname {
//		//推送给前端，刷新好友列表
//		//给前端发送一个通知
//		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
//		AddSubscriptionMsg(&msgInfo)
//	}
//}

/*
客户端向代理获取一个租用时间的订单
*/
func GetOrders(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_IM_PROXY_getorders_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	of, err := model.ParseOrderForm(np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	of.UserAddr = *message.SenderAddr
	//utils.Log.Info().Msgf("本次购买空间:%+v %+v", of, np.Data)
	of, ERR := StaticProxyServerManager.CreateOrders(of.UserAddr, utils.Byte(of.SpaceTotal), of.UseTime)
	//of, ERR := StServer.CreateOrders(of.UserAddr, of.SpaceTotal, of.UseTime)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	bs, err := of.Proto()
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, bs)
}

/*
客户端向代理获取一个租用时间的订单 返回
*/
//func GetOrders_recv(message *message_center.MessageBase) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
客户端向代理获取一个租用空间的订单
*/
func GetRenewalOrders(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_IM_PROXY_getRenewalOrders_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	of, err := model.ParseOrderForm(np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	of, ERR := StaticProxyServerManager.CreateRenewalOrders(of.PreNumber, of.SpaceTotal, of.UseTime)
	//of, ERR := StServer.CreateRenewalOrders(of.PreNumber, of.SpaceTotal, of.UseTime)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	bs, err := of.Proto()
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, bs)
}

/*
客户端向代理获取一个租用空间的订单 返回
*/
func GetRenewalOrders_recv(message *message_center.MessageBase) {
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
查询代理节点已经同步的数据链索引
*/
func GetDataChainIndex(message *message_center.MessageBase) {
	utils.Log.Info().Msgf("远端获取用户同步的index")
	//replyMsgID := uint64(config.MSGID_IM_PROXY_getDataChainIndex_recv)
	_, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	addr := message.SenderAddr
	_, _, remain := StaticProxyServerManager.QueryUserSpaces(*addr)
	if remain == 0 {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_IM_not_proxy, "")
		return
	}

	utils.Log.Info().Msgf("远端获取用户同步的index")
	//查询本地的数据链高度
	itr, ERR := db.ImProxyClient_FindDataChainLast(*Node.GetNetId(), *addr)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	if itr == nil {
		ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
		return
	}
	utils.Log.Info().Msgf("远端获取用户同步的index")
	index := itr.GetIndex()
	endIndex := index.Bytes()
	ReplySuccess(Node, config.NET_protocol_version_v1, message, &endIndex)
}

/*
查询代理节点已经同步的数据链索引 返回
*/
func GetDataChainIndex_recv(message *message_center.MessageBase) {
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
上传数据链
*/
func UploadDatachain(message *message_center.MessageBase) {
	utils.Log.Info().Msgf("收到代理数据")
	//replyMsgID := uint64(config.MSGID_IM_PROXY_upload_datachain_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	//接收到的是连续消息记录
	itrBss, _, err := model.ParseBytes(np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	if len(itrBss) == 0 {
		ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
		return
	}
	itrs := make([]imdatachain.DataChainProxyItr, 0, len(itrBss))
	for _, one := range itrBss {
		//utils.Log.Info().Msgf("收到代理数据:%+v", one)
		itrOne, ERR := imdatachain.ParseDataChain(one)
		if !ERR.CheckSuccess() {
			ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
			return
		}
		//utils.Log.Info().Msgf("收到代理数据:%+v", itrOne)
		itrs = append(itrs, itrOne)
	}
	//utils.Log.Info().Msgf("收到代理数据")
	proxyClientAddr := itrs[0].GetProxyClientAddr()
	//先检查用户是否在代理列表中
	_, _, remain := StaticProxyServerManager.QueryUserSpaces(proxyClientAddr)
	if remain == 0 {
		//不在代理用户列表中
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_IM_not_proxy, "")
		return
	}
	for _, one := range itrs {
		//检查并保存数据链
		ERR := StaticProxyServerManager.syncServer.CheckSaveDataChain(one)
		if !ERR.CheckSuccess() {
			ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
			return
		}
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
}

/*
上传数据链 返回
*/
func UploadDatachain_recv(message *message_center.MessageBase) {
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
从客户端同步数据链
*/
func SyncDatachainForClient(message *message_center.MessageBase) {
	utils.Log.Info().Msgf("从客户端同步数据链")
	//replyMsgID := uint64(config.MSGID_IM_PROXY_sync_datachain_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}

	proxyItrs, ERR := db.ImProxyClient_FindDataChainRange(*Node.GetNetId(), *Node.GetNetId(), np.Data, config.IMPROXY_sync_total_once)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}

	bss := make([][]byte, 0, len(proxyItrs))
	for _, one := range proxyItrs {
		bs, err := one.Proto()
		if err != nil {
			ERR = utils.NewErrorSysSelf(err)
			ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		}
		bss = append(bss, *bs)
	}
	bs, err := model.BytesProto(bss, nil)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, &bs)
}

/*
从客户端同步数据链 返回
*/
//func SyncDatachainForClient_recv(message *message_center.MessageBase) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
发送数据链
*/
func SendDatachain(message *message_center.MessageBase) {
	utils.Log.Info().Msgf("收到数据链消息")
	//replyMsgID := uint64(config.MSGID_IM_PROXY_send_datachain_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ERR := utils.NewErrorBus(config.ERROR_CODE_params_format, "")
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}

	itr, ERR := imdatachain.ParseDataChain(np.Data)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	//utils.Log.Info().Msgf("收到数据链消息:%+v", itr)
	//检查消息完整性
	if !itr.CheckHash() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_IM_check_hash_fail, "")
		return
	}
	//utils.Log.Info().Msgf("收到数据链消息")
	//检查接收者是否是自己，是自己的消息，直接保存
	if bytes.Equal(itr.GetAddrTo().GetAddr(), Node.GetNetId().GetAddr()) {
		//utils.Log.Info().Msgf("收到的消息:%+v", itr)

		//当自己有代理节点后，不接收直接发送的消息
		userinfo, ERR := db.GetSelfInfo(*Node.GetNetId())
		if ERR.CheckFail() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
			return
		}
		proxys := make([]string, 0)
		userinfo.Proxy.Range(func(k, v any) bool {
			key := k.(string)
			proxys = append(proxys, key)
			return true
		})
		//当有代理节点的时候，拒绝直接发送给自己
		if len(proxys) > 0 {
			ERR := utils.NewErrorBus(config.ERROR_CODE_IM_forward_proxy, "")
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
			return
		}

		newItr := imdatachain.NewDataChainProxyMsgLog(itr)
		//utils.Log.Info().Msgf("转换后的消息:%+v", newItr)
		////解密消息
		//ERR = itr.DecryptContent(sharekey[:])
		ERR = DecryptContent(newItr)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("解密消息失败:%s", ERR.String())
			ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
			return
		}

		//if newItr.GetClientItr() != nil && newItr.GetClientItr().GetClientCmd() == config.IMPROXY_Command_client_file {
		//	clientProxy := newItr.GetClientItr()
		//	sendFile := clientProxy.(*imdatachain.DatachainFile)
		//	utils.Log.Info().Msgf("打印文件hash:%+v", sendFile.SendTime)
		//}

		//是直接发送给自己的消息
		ERR = StaticProxyClientManager.ParserClient.SaveDataChain(newItr)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
			return
		}
		utils.Log.Info().Msgf("收到数据链消息")
		ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
		return
	}
	//是代理消息，先检查用户是否在代理列表中
	ERR = StaticProxyServerManager.CheckSaveDataChainNolink(itr)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	utils.Log.Info().Msgf("收到数据链消息")
	//触发转发
	//StaticProxyServerManager.syncServer.ForwardDataChain(itr.GetAddrTo())
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
}

/*
发送数据链 返回
*/
//func SendDatachain_recv(message *message_center.MessageBase) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
客户端从代理服务器下载数据链
包含两部分内容：
1、多端产生的数据链，和本节点不同步的消息。
2、代理节点保存的还未加入数据链的离线消息。
*/
func DownloadDatachain(message *message_center.MessageBase) {
	//replyMsgID := uint64(config.MSGID_IM_PROXY_download_datachain_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}

	//查询已经上链的消息
	proxyItrs, ERR := db.ImProxyClient_FindDataChainRange(*Node.GetNetId(), *message.SenderAddr, np.Data, config.IMPROXY_sync_total_once)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	//if proxyItrs == nil || len(proxyItrs) == 0 {
	//	ReplySuccess(Node, config.NET_protocol_version_v1, message,  nil)
	//	return
	//}
	bss := make([][]byte, 0, len(proxyItrs))
	for _, proxyItr := range proxyItrs {
		bsOne, err := proxyItr.Proto()
		if err != nil {
			ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
			return
		}
		bss = append(bss, *bsOne)
	}
	utils.Log.Info().Msgf("查询已经上链的消息:%d", len(proxyItrs))
	//查询未上链的离线消息
	proxyItrs, ERR = db.ImProxyServer_FindDatachainNoLinkRange(*Node.GetNetId(), *message.SenderAddr, config.IMPROXY_sync_total_once)
	//proxyItrs, _, ERR = db.ImProxyClient_FindDataChainSendFailRange(config.DBKEY_improxy_server_datachain_send_fail,
	//	Node.GetNetId(), *message.Head.Sender, config.IMPROXY_sync_total_once)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return
	}
	bss2 := make([][]byte, 0, len(proxyItrs))
	for _, proxyItr := range proxyItrs {
		bsOne, err := proxyItr.Proto()
		if err != nil {
			ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
			return
		}
		bss2 = append(bss2, *bsOne)
	}
	utils.Log.Info().Msgf("查询未上链的消息:%d", len(proxyItrs))

	bs, err := model.BytesProto(bss, bss2)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, &bs)
}

/*
下载数据链 返回
*/
//func DownloadDatachain_recv(message *message_center.MessageBase) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
发送群聊消息
*/
func SendGroupDatachain(message *message_center.MessageBase) {
	utils.Log.Info().Msgf("收到群聊消息")
	//replyMsgID := uint64(config.MSGID_IM_PROXY_group_send_datachain_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	itr, ERR := imdatachain.ParseDataChain(np.Data)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	utils.Log.Info().Msgf("收到群聊消息")
	//检查消息完整性
	if !itr.CheckHash() {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_IM_check_hash_fail, "")
		return
	}

	//_, _, ERR = GroupAuthMember(config.DBKEY_improxy_user_group_list_knit, itr.GetBase().GroupID, itr.GetAddrFrom())
	//if !ERR.CheckSuccess() {
	//	ReplyError(Node, config.NET_protocol_version_v1, message,  ERR.Code, ERR.Msg)
	//	return
	//}

	ERR = StaticProxyServerManager.groupKnitManager.KnitDataChain(itr)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
}

/*
发送群聊消息 返回
*/
//func SendGroupDatachain_recv(message *message_center.MessageBase) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
群内广播消息
*/
func multicaseGroupDatachain(message *message_center.MessageBase) {
	utils.Log.Info().Msgf("接收到群聊消息")
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		return
	}
	//接收到的是连续消息记录
	itrBss, _, err := model.ParseBytes(np.Data)
	if err != nil {
		return
	}
	if len(itrBss) == 0 {
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, nil)
	for _, one := range itrBss {
		itrOne, ERR := imdatachain.ParseDataChain(one)
		if !ERR.CheckSuccess() {
			return
		}
		ERR = StaticProxyServerManager.gsmm.ParseGroupDataChain(itrOne)
		ERR = StaticProxyClientManager.groupParserManager.SaveGroupDataChain(itrOne)
		if ERR.CheckSuccess() {
			continue
		}
		if ERR.Code == config.ERROR_CODE_IM_group_not_exist {
			//要考虑群成员删除了本地数据库而丢失群记录的情况
			StaticProxyClientManager.groupParserManager.DownloadBuildGroupDataChain(itrOne.GetBase().GroupID, itrOne.GetBase().AddrProxyServer)
		}
	}
}

/*
查询群聊起始消息
*/
func DownloadGroupDatachainStart(message *message_center.MessageBase) {
	utils.Log.Info().Msgf("收到查询群聊起始消息")
	//replyMsgID := uint64(config.MSGID_IM_PROXY_group_download_datachain_start_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	groupId := np.Data
	indexBs, ERR := db.ImProxyClient_FindGroupMemberStartIndex(*Node.GetNetId(), groupId, *message.SenderAddr)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	if indexBs == nil {
		ERR = utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	proxyItr, ERR := db.ImProxyClient_FindGroupDataChainByIndex(*Node.GetNetId(), groupId, *indexBs)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	bs, err := proxyItr.Proto()
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_system_error_self, err.Error())
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, bs)
}

/*
查询群聊起始消息 返回
*/
//func DownloadGroupDatachainStart_recv(message *message_center.MessageBase) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
下载群聊消息
*/
func DownloadGroupDatachain(message *message_center.MessageBase) {
	utils.Log.Info().Msgf("下载群聊消息")
	//replyMsgID := uint64(config.MSGID_IM_PROXY_group_download_datachain_recv)
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	list, _, err := model.ParseBytes(np.Data)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}
	groupId := list[0]
	startIndex := list[1]
	proxyItrs, ERR := db.ImProxyClient_FindGroupDataChainRange(*Node.GetNetId(), groupId, startIndex)
	if !ERR.CheckSuccess() {
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	list = make([][]byte, 0, len(proxyItrs))
	for _, one := range proxyItrs {
		bs, err := one.Proto()
		if err != nil {
			ERR = utils.NewErrorSysSelf(err)
			ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
			return
		}
		list = append(list, *bs)
	}
	bs, err := model.BytesProto(list, nil)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	ReplySuccess(Node, config.NET_protocol_version_v1, message, &bs)
}

/*
发送群聊消息 返回
*/
//func DownloadGroupDatachain_recv(message *message_center.MessageBase) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}

/*
获取自己的sendindex
*/
func GetSendIndex(message *message_center.MessageBase) {
	utils.Log.Info().Msgf("下载群聊消息")
	//replyMsgID := uint64(config.MSGID_IM_PROXY_get_sendindex_recv)
	_, err := utils.ParseNetParams(message.Content)
	if err != nil {
		ReplyError(Node, config.NET_protocol_version_v1, message, config.ERROR_CODE_params_format, err.Error())
		return
	}

	sendIndex, ERR := db.ImProxyClient_FindSendIndex(*config.DBKEY_improxy_user_datachain_sendIndex_parse, *Node.GetNetId(),
		*message.SenderAddr, *Node.GetNetId())
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		ReplyError(Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	bs := sendIndex.Bytes()
	ReplySuccess(Node, config.NET_protocol_version_v1, message, &bs)
}

/*
获取自己的sendindex 返回
*/
//func GetSendIndex_recv(message *message_center.MessageBase) {
//	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
//}
