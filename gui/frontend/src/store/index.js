import { createStore } from 'vuex'
import BigNumber from "bignumber.js";
// import { IM_PrintLog } from '../../bindings/web3_gui/gui/server_api/sdkapi'
export default createStore({
  state: {
    nav_badge:new Map(),//左侧导航栏是否显示徽章提示
    nav_showModules:"",//左侧导航栏目前显示的模块

    count: 0,
    peerInfo: null,
    peerConfig:null,
    isRPC:false,
    rpcUser:"",
    rpcPwd:"",
    getinfo:null,
    rpcParams:null,
    coinCompany:100000000,
    namedestroy:null,
    nameinfo:null,
    payTokeninfo:null,
    
    witnessScore:null,//见证人质押信息
    witnessScoreOut:null,//取消见证人质押信息
    communityVote:null,//社区节点投票信息
    communityVoteOut:null,//取消社区节点投票信息
    lightVoteOut:null,//轻节点取消投票信息
    lightDepositOut:null,//轻节点取消质押信息

    selectFileInfo:null,//文件存储中，被选中的文件信息
    spacesTotalSize: 0, //云存储
    spacesUseSize: 0,   //空间已经使用容量

    im_addrself:"",
    im_userinfo_self:null,//自己的个人信息
    im_friendList:null,//好友列表
    im_userinfo:null,//已经选中的聊天好友个人信息
    im_msgMap:new Map(),//保存好友聊天记录列表;key:=好友地址;value:Array=聊天记录;
    im_now_msgList:[],//当前页面正在显示的
    im_now_msgList_multicast:[],//当前页面正在显示的
    im_show_userinfo:null,//个人信息页面
    im_badge_userlist:new Map(),//好友列表中的头像是否显示新消息提醒
    im_DownloadListProgress:new Map(),//文件下载进度条
    im_testrate:0,//
    im_msg_content_change:0,//消息内容区域有变化，让消息内容窗口滑动到最底部
    im_cutScreenVisible:false,//
    im_cutScreenBase64Str:"",//
    im_groupinfo:null,//选中的群信息，修改群信息用
    im_group_members_change:0,//群成员变动
    im_friend_apply_list_change:0,//好友申请列表状态变动

    file_download_list:[],//文件下载列表

    chain_getinfo:null,//
    chain_witnessinfo:null,//见证人状态信息
    chain_payorder_result:null,//支付订单返回结果
    chain_payorder_server_orderid:null,//服务端支付订单成功推送
    chain_payorder_client_orderid:null,//客户端支付订单成功推送

    circle_select_class:"",//选中的类别
    circle_news_index_type:0,//0=无类型;1=草稿箱;2=发布;
    circle_news_index:0,//修改的新闻数据库index
    circle_news_title:"",//编辑中的标题
    circle_editor_html:"",//保存编辑器

    storage_client_selectServerInfo:null,//客户端购买云空间，选中的服务器信息
    sharebox_filePrice_process_id:null,    //异步计算文件hash的id
  },
  mutations: {
    setRpcUser(state, user){
      state.rpcUser=user;
    },
    setRpcPwd(state, pwd){
      state.rpcPwd=pwd;
    },
    setIsRpc(state, isRPC){
      state.isRPC = isRPC;
      
    },
    setPeerInfo(state, peerinfo){
      state.peerInfo = peerinfo;
    },
    setPeerConfig(state, peerConfig){
      state.peerConfig = peerConfig;
    },
    setNamedestroy(state, nameinfo){
      state.namedestroy = nameinfo;
    },
    setNameinfo(state, nameinfo){
      state.nameinfo = nameinfo;
    },
    setPayTokeninfo(state, tokeninfo){
      state.payTokeninfo = tokeninfo;
    },
    setWitnessinfo(state, witnessinfo){
      var bignumber = BigNumber(witnessinfo.Value);
      witnessinfo.Value = bignumber.dividedBy(100000000).toNumber();
      state.chain_witnessinfo = witnessinfo;
    },
    setWitnessScore(state, info){
      state.witnessScore = info;
    },
    setWitnessScoreOut(state, info){
      state.witnessScoreOut = info;
    },
    setCommunityVote(state, info){
      state.communityVote = info;
    },
    setCommunityVoteOut(state, info){
      state.communityVoteOut = info;
    },
    setLightVoteOut(state, info){
      state.lightVoteOut = info;
    },
    setLightDepositOut(state, info){
      state.lightDepositOut = info;
    },
    setSelectFileInfo(state, info){state.selectFileInfo = info;},
    setinfo(state, getinfo){
      // console.log("设置基本信息",getinfo)
      //将getinfo数据排序，格式化
      //排序
      if(getinfo.TokenBalance != null){
        getinfo.TokenBalance = getinfo.TokenBalance.sort(function(a,b){
          return a.TokenId.localeCompare(b.TokenId);
        })
        for(var i=0 ; i< getinfo.TokenBalance.length ; i++){
          // console.log(getinfo.TokenBalance[i].TokenId);
          var one = getinfo.TokenBalance[i];
          var bignumber = BigNumber(one.Balance);
          one.Balance = bignumber.dividedBy(100000000).toNumber();
          bignumber = BigNumber(one.BalanceFrozen);
          one.BalanceFrozen = bignumber.dividedBy(100000000).toNumber();
          bignumber = BigNumber(one.BalanceLockup);
          one.BalanceLockup = bignumber.dividedBy(100000000).toNumber();
        }
      }
      state.chain_getinfo = getinfo;
    },
    setRpcParams(state, rpcParams){
      state.rpcParams = rpcParams;
    },
    setImUserinfo(state, info){
      state.im_userinfo = info;
    },
    RefreshFriendList(state){
      
    },
    SetMsgList(state, msgList){
      state.im_now_msgList = msgList
    },
    //添加一条新聊天消息
    pushMsgList(state, msgOne){
      // console.log("新消息内容",msgOne);
      //处理文件下载进度
      if(msgOne.FileBlockTotal>0){
        msgOne.progress = ((msgOne.FileBlockIndex+1)/msgOne.FileBlockTotal)*100
        msgOne.progress = parseInt(msgOne.progress.toFixed(0))
      }
      if(msgOne.IsGroup){
        //是群消息
        //是选中的群
        var have = false
        //先判断消息是否存在，存在则是修改状态
        for(var i=state.im_now_msgList.length; i>0; i--){
          if(state.im_now_msgList[i-1].SendID && state.im_now_msgList[i-1].SendID == msgOne.SendID){
            state.im_now_msgList[i-1].State = msgOne.State
            //如果是图片，则更新图片内容
            state.im_now_msgList[i-1].Content = msgOne.Content
            //下载文件有进度，则更新进度
            state.im_now_msgList[i-1].TransProgress = msgOne.TransProgress
            if(msgOne.progress){
              state.im_now_msgList[i-1].progress = msgOne.progress
            }
            have = true
            break
          }
        }
        // console.log("是新消息")
        //消息不存在，是新消息
        if(!have){
          state.im_now_msgList.push(msgOne)
        }
        //自己发的消息不提醒
        if(msgOne.From == state.im_userinfo_self.Addr){
          return
        }
        //文件头和尾提醒，文件体不提醒
        if(msgOne.FileSize>0 && msgOne.FileBlockIndex != 0 && msgOne.FileBlockTotal != msgOne.FileBlockIndex+1){return}
        //群头像显示徽章提醒
        state.im_badge_userlist.set(msgOne.To, true)
        return
      }
      var from = msgOne.From;
      if(msgOne.FromIsSelf == true){
        from = msgOne.To
      }
      //判断是否是广播消息，from和to中有一个字段为""，都认为是广播消息
      if(msgOne.From == "" || msgOne.To == ""){
        state.im_now_msgList_multicast.push(msgOne)
        //判断是否打开广播窗口
        if(state.im_userinfo && state.im_userinfo.Addr == "multicast"){
          // return
        }
        //自己发送的广播消息不提醒
        if(msgOne.From == state.im_userinfo_self.Addr){
          // console.log("自己发送的广播消息不提醒");
          return
        }
        state.im_badge_userlist.set("multicast", true)
        return
      }
      // console.log("判断是否选中好友",state.im_userinfo,from)
      //判断这条消息是不是当前选中的好友
      if(state.im_userinfo && state.im_userinfo.Addr == from){
        var have = false
        //先判断消息是否存在，存在则是修改状态
        for(var i=state.im_now_msgList.length; i>0; i--){
          if(state.im_now_msgList[i-1].SendID && state.im_now_msgList[i-1].SendID == msgOne.SendID){
            state.im_now_msgList[i-1].State = msgOne.State
            //如果是图片，则更新图片内容
            state.im_now_msgList[i-1].Content = msgOne.Content
            //下载文件有进度，则更新进度
            state.im_now_msgList[i-1].TransProgress = msgOne.TransProgress
            if(msgOne.progress){
              state.im_now_msgList[i-1].progress = msgOne.progress
            }
            have = true
            break
          }
        }
        // console.log("是新消息")
        //消息不存在，是新消息
        if(!have){
          state.im_now_msgList.push(msgOne)
        }
      }
      //自己发的消息不提醒
      if(msgOne.From == state.im_userinfo_self.Addr){
        return
      }
      //文件头和尾提醒，文件体不提醒
      if(msgOne.FileSize>0 && msgOne.FileBlockIndex != 0 && msgOne.FileBlockTotal != msgOne.FileBlockIndex+1){return}
      //好友头像显示徽章提醒
      state.im_badge_userlist.set(from, true)
    },
    //添加聊天历史记录
    pushMsgListHistory(state, msgList){
      // console.log(msgOne);
      var list = new Array()

    },
    setNavBadgeShow(state, params){//设置左侧导航徽章是否显示
      // console.log(params.show, state.nav_showModules, params.name )
      if(params.show && state.nav_showModules == params.name){
        return
      }
      state.nav_badge.set(params.name, params.show)
      // console.log("修改成功:", params.name, state.nav_badge.get(params.name))
    },
    setFriendHeadBadgeShow(state, params){//设置好友头像是否显示徽章
      state.im_badge_userlist.set(params.name, params.show)
    },
    setDownloadListProgress(state, params){//设置下载列表进度
      state.file_download_list = params
      // console.log("设置进度条",params)
      var newMap = new Map()
      for(var i=0; i<params.length; i++){
        var one = params[i]
        newMap.set(one.From+one.PushTaskID, one.Rate)
      }
      // state.im_DownloadListProgress.set(params.From+params.PushTaskID, params.Rate)
      state.im_DownloadListProgress = newMap
      // state.file_download_list.push({PushTaskID:99999, PullTaskID:99999, Size:10240,Name:"测试数据",Rate:16,Status:true})
    },
    increment(state) {
      state.count++
    }
  },
  actions: {
    increment(context) {
      context.commit('increment')
    }
  },
  getters: {
    //查询一个好友的头像
    findFriendHeadNum:(state) => (addr) =>{
      for(var i=0; i<state.im_friendList.length; i++){
        var one = state.im_friendList[i]
        if(one.Addr == addr){
          return one.HeadUrl
        }
      }
      return ''
    },
    //获取选中用户的昵称
    getFriendNickname:(state) => () =>{
      if(state.im_userinfo){
        if(state.im_userinfo.RemarksName == ""){
          return state.im_userinfo.Nickname
        }
        return state.im_userinfo.RemarksName
      }
      return ""
    },
    //获取选中用户的地址
    getFriendAddr:(state) => () =>{
      if(state.im_userinfo){
        return state.im_userinfo.Addr == "multicast" ? "":state.im_userinfo.Addr
      }
      return ""
    },
    getNavBadgeShow:(state) => (badgeType) =>{
      // console.log(badgeType,state.nav_badge.get(badgeType))
      var value = state.nav_badge.get(badgeType)
      if(value == undefined){
        // console.log("1212121212")
        return false
      }
      // console.log(value)
      if(value){
        return true
      }
      return false
    },
    getFriendHeadBadgeShow:(state) => (badgeType) =>{
      // console.log(badgeType,state.im_badge_userlist.get(badgeType))
      var value = state.im_badge_userlist.get(badgeType)
      if(value == undefined){
        // console.log("1212121212")
        return false
      }
      // console.log(value)
      if(value){
        return true
      }
      return false
    },
    getMsgByAddr:(state) => () =>{
      // IM_PrintLog("getMsgByAddr 111111111111")
      if(state.im_userinfo == null){
        // IM_PrintLog("getMsgByAddr 2222222222222")
        return
      }
      if(state.im_userinfo.Addr == "multicast"){
        // IM_PrintLog("getMsgByAddr 333333333333333")
        return state.im_now_msgList_multicast
      }
      // IM_PrintLog("getMsgByAddr 44444444444")
      return state.im_now_msgList
    },
    //获取下载列表进度
    getDownloadListProgress: (state) => (params) =>{
      // console.log("获取进度条",params, params.From+params.PullAndPushID)
      var value = state.im_DownloadListProgress.get(params.From+params.PullAndPushID)
      // console.log("获取进度条",params, params.From+params.PullAndPushID,value)
      if(value == undefined){
        return 100
      }
      return value
    },
    getCount: state => state.count
  }
})