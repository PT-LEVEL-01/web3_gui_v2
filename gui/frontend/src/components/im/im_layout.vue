<template>
  <div class="common-layout" style="text-align: left; height:100%; margin:0; border:red solid 0px;">
    <el-container style="height:100%;">
      <el-aside width="200px" style="height:100%;border:red solid 0px;background-color:#eee;">
        <el-scrollbar>
          <el-menu class="el-menu-vertical-demo" @select="meetingListSelect" style="background-color:#eee;">
            <el-menu-item index="3">
              <el-icon><Setting /></el-icon>
              <span>离线服务</span>
            </el-menu-item>
            <el-menu-item index="0">
             <div style="display: flex;align-items: center;">
              <el-icon><Search /></el-icon>
              <span>添加好友</span>
             </div>
            </el-menu-item>
            <el-menu-item index="1">
              <el-icon><el-badge :is-dot="store.getFriendHeadBadgeShow('newFriend')" class="item"><CirclePlus /></el-badge></el-icon>
              <span>新的朋友</span>
            </el-menu-item>
            <el-menu-item index="2">
              <el-icon><el-badge :is-dot="store.getFriendHeadBadgeShow('multicast')" class="item"><Phone /></el-badge></el-icon>
              <span>广播消息</span>
            </el-menu-item>
            <el-menu-item v-for="(item,i) in store.im_friendList" :index="i+10">
              <!-- <img src="../../assets/head/daniel.jpg" alt="" style="height:50px;width:50px;"> -->
              <el-badge :is-dot="store.getFriendHeadBadgeShow(item.Addr)" class="item">
              <el-image style="width: 50px; height: 50px" :src="item.HeadUrl"></el-image>
              </el-badge>
              <span v-if="item.RemarksName!=''" style="margin-left:10px;width:120px;overflow-x: hidden;text-overflow: ellipsis;">{{ item.RemarksName }}</span>
              <span v-else style="margin-left:10px;width:120px;overflow-x: hidden;text-overflow: ellipsis;">{{ item.Nickname }}</span>
            </el-menu-item>
          </el-menu>
        </el-scrollbar>
      </el-aside>
      <el-main style="padding:0;">
        <component :is="currentView"/>
<!--        <router-view/>-->
      </el-main>
    </el-container>
  </div>
</template>

<style>
</style>
  
<script setup>
import { GetHeadUrl } from '../../head_image'
import { IM_GetFriendList, GetChatHistory, IM_NoticeCancelFlicker } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {ElMessage} from "element-plus";
import {getCurrentInstance, ref, onMounted, onUnmounted, computed} from 'vue'
import { store } from '../../store.js'
import im_add_list from "./im_add_list.vue";
import im_group_create from "./im_group_create.vue";
import im_index from "./im_index.vue";
import im_proxy_setup from "./im_proxy_setup.vue";
import im_message_content from "./im_message_content.vue";
import im_message_content_input from "./im_message_content_input.vue";
import im_message_content_preview from "./im_message_content_preview.vue";
import im_proxy_list from "./im_proxy_list.vue";
import im_proxy_order from "./im_proxy_order.vue";
import im_proxy_order_list from "./im_proxy_order_list.vue";
import im_screen_shot from "./im_screen_shot.vue";
import im_search from "./im_search.vue";
import im_sharebox from "./im_sharebox.vue";
import im_userinfo from "./im_userinfo.vue";
import {store_routers} from "../../store_routers.js";

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const routes = {
  "index":im_index,
  "createGroup":im_group_create,
  "imProxyList":im_proxy_list,
  "imProxySetup":im_proxy_setup,
  "imProxyOrder":im_proxy_order,
  "imProxyOrderList":im_proxy_order_list,
  "addlist":im_add_list,
  "message":im_message_content,
  "message_preview":im_message_content_preview,
  "search":im_search,
  "userinfo":im_userinfo,
  "sharebox":im_sharebox,
}
// const currentPath = ref("login")
const currentView = computed(() => {
  return routes[store_routers.currentPageKey_im]
})

const friendList = ref(null)

// 定义一个ref来持有定时器
const timer = ref(null);
// 创建定时器
const createTimer = () => {
  timer.value = setInterval(() => {
    // flashDownloadList()
    // flashUploadList()
    // 定时器的逻辑
  }, 1000);
};

//在组件挂载时创建定时器
onMounted(() => {
  createTimer();
});

//在组件卸载时清除定时器
onUnmounted(() => {
  if (timer.value) {
    clearInterval(timer.value);
  }
});

function getHeadUrl(type){
  return GetHeadUrl(type)
}

function meetingListSelect(key, keyPath) {
  // console.log("select",key, keyPath);
  if (key == 0) {
    // thistemp.$router.push({path: '/index/im/search'});
    store_routers.gopage_im("search")
    return
  }
  if (key == 1) {
    var userInfo = {Nickname:"新的朋友",Addr:"newFriend"};
    store.setImUserinfo(userInfo);
    store.setFriendHeadBadgeShow({name:"newFriend", show:false});
    IM_NoticeCancelFlicker("2")
    // thistemp.$router.push({path: '/index/im/addlist'});
    store_routers.gopage_im("addlist")
    return
  }
  if (key == 2) {
    var userInfo = {Nickname:"广播消息",Addr:"multicast"};
    store.setImUserinfo(userInfo);
    store.setFriendHeadBadgeShow({name:"multicast", show:false});
    store.im_now_msgList = []
    // thistemp.$router.push({path: '/index/im/message'});
    store_routers.gopage_im("message")
    return
  }
  if (key == 3) {
    // thistemp.$router.push({path: '/index/im/imProxyList'});
    store_routers.gopage_im("imProxyList")
    return
  }
  //设置选中的好友
  var userInfo = store.im_friendList[key-10];
  // console.log(userInfo)
  store.setImUserinfo(userInfo);
  // var thistemp = this
  //拉取历史记录
  Promise.all([GetChatHistory("", 10,userInfo.Addr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    var msgList = messageOne.list.MessageList
    console.log("拉取历史记录",msgList)
    // messageOne.HeadUrl = GetHeadUrl(messageOne.HeadNum)
    // store.im_userinfo_self = messageOne
    var headUrl = GetHeadUrl(userInfo.HeadNum)
    // console.log(headUrl)
    var newArray = []
    for(var i=msgList.length; i>0; i--){
      var msgOne = msgList[i-1]
      msgOne.HeadUrl = headUrl
      // if(store.im_userinfo != null){msgOne.From = ""}
      newArray.push(msgOne)
    }
    // console.log("拉取历史记录 1111111111", newArray)
    // thistemp.$store.state.im_now_msgList = messageOne
    store.SetMsgList(newArray);
    store.setFriendHeadBadgeShow({name:userInfo.Addr, show:false});
    // console.log("选中好友",userInfo.Addr)
    IM_NoticeCancelFlicker("1"+userInfo.Addr)
    // console.log("拉取历史记录 2222222222")
    // thistemp.$router.push({path: '/index/im/message'});
    store_routers.gopage_im("message")
    // console.log("拉取历史记录 3333333333")
  });
}

function init(){
  Promise.all([IM_GetFriendList()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    for(var i=0; i<messageOne.info.UserList.length; i++){
      var userOne = messageOne.info.UserList[i]
      messageOne.info.UserList[i].HeadUrl = GetHeadUrl(userOne.HeadNum)
    }
    store.im_friendList = messageOne.info.UserList
    console.log("好友列表",messageOne.info.UserList)
  }).catch(error => {
    console.log("拉取好友列表错误:",error)
  });
}
init()
</script>