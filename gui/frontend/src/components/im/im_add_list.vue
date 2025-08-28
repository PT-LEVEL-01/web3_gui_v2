<template>
  <el-container style="height: 100%; border: 1px solid #eee" @mouseup="mouseupSendNoticeTray">
    
    <el-container>
      <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
        <span style="font-size:20px;">新的朋友</span>
      </el-header>
      
      <el-main class="chat_content">
        <el-row v-for="(item,i) in friendList">
          <el-col :span="50">
            <el-image style="width: 50px; height: 50px" :src="head" fit="cover" />
          </el-col>
          <el-col v-if="item.IsGroup" :span="500" style="padding-left: 10px;text-align: left;">
            <div>{{ item.Nickname }}<span style="color: #9a6e3a;margin:0 5px;">邀请您入群</span>{{item.RemarksName}}</div>
            <div>{{ item.Addr }}</div>
          </el-col>
          <el-col v-if="!item.IsGroup" :span="500" style="padding-left: 10px;text-align: left;">
            <div>{{ item.Nickname }}</div>
            <div>{{ item.Addr }}</div>
          </el-col>
          <el-col :span="50">
            <el-button v-if="item.Status == 0 && !item.IsGroup" type="success" @click="agreeFriend(item.Token)">接受</el-button>
            <el-button v-if="item.Status == 0 && item.IsGroup && item.GroupSign==''" type="success" @click="groupAccept(item)">同意入群</el-button>
            <el-button v-if="item.Status == 0 && item.IsGroup && item.GroupSign!=''" type="success" @click="groupAddMember(item)">添加成员</el-button>
          </el-col>
        </el-row>
      </el-main>
    </el-container>
  </el-container>
      
</template>

<style>
</style>

<script setup>
import { ElMessage } from 'element-plus'
import head_daniel from '../../assets/images/head/daniel.jpg'
import { IM_GetNewFriend, IM_AgreeApplyFriend, IM_NoticeCancelFlicker, ImProxyClient_GroupAccept,
  ImProxyClient_GroupAddMember } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
// import {useStore} from "vuex";
import { store } from '../../store.js'
import {onBeforeUnmount, ref, shallowRef, onMounted, watch, getCurrentInstance, nextTick } from 'vue';

// const store = useStore()
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
const friendList = ref([])
const head = ref(head_daniel)

//获取新好友申请列表
const getNewFriendList = () => {
  Promise.all([IM_GetNewFriend()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("新朋友列表",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    friendList.value = messageOne.info.UserList
    // console.log(friendList.value)
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '获取添加好友列表失败：'+error,
      type: 'error',
    })
  });
}
getNewFriendList()

//有新好友申请时，和状态变化时
watch(
    () => store.im_friend_apply_list_change,
    (newVal, oldVal) => {
      nextTick().then(() => {
        //刷新好友申请列表
        getNewFriendList()
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

//同意入群
const groupAccept = (userinfo) => {
  // console.log(userinfo)
  // return
  Promise.all([ImProxyClient_GroupAccept(userinfo.Token)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("邀请好友",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    getNewFriendList()
    mouseupSendNoticeTray()
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

//添加群成员
const groupAddMember = (userinfo) => {
  console.log("添加的成员信息",userinfo)
  // return
  Promise.all([ImProxyClient_GroupAddMember(userinfo.Token)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("添加群成员",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    mouseupSendNoticeTray()
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

//同意好友申请
const agreeFriend = (addr) => {
  Promise.all([IM_AgreeApplyFriend(addr)]).then(messages => {
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
    mouseupSendNoticeTray()
    ElMessage({
      showClose: true,
      message: '添加成功',
      type: 'success',
    })
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '添加失败：'+error,
      type: 'error',
    })
  });
}

const mouseupSendNoticeTray = () => {
  // console.log("鼠标事件")
  var userInfo = store.im_userinfo
  //如果当前选中的是添加好友页面，则消除托盘闪烁
  IM_NoticeCancelFlicker("2")
  //徽章提醒消失
  store.setFriendHeadBadgeShow({name:userInfo.Addr, show:false});
};
</script>