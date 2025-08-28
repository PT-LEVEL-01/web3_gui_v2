<template>
  <el-container style="height: 100%; border: 1px solid #eee">
    <el-container>
      <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
        <span style="font-size:20px;"><el-button link @click="back()"><el-icon><Back /></el-icon></el-button></span>
      </el-header>

      <el-main class="chat_content">
        <el-col :span="500" style="margin-left: 60px;">
          <div>群名称：<el-input v-model="nickname" style="width: 240px" placeholder="" /></div>
          <div style="margin-top: 40px;">禁言：<el-switch v-model="shoutUpSwitch" class="ml-2" /></div>
          <el-button type="primary" @click="submitForm()">{{btName}}</el-button>
        </el-col>
      </el-main>
    </el-container>
  </el-container>

</template>

<style>
</style>

<script setup>
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { IM_SearchFriendInfo, IM_AddFriend, ImProxyClient_CreateGroup, ImProxyClient_UpdateGroup }
  from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import { store } from '../../store.js'
import {store_routers} from "../../store_routers.js";
import {getCurrentInstance, ref} from "vue";

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const nickname = store.im_groupinfo == null ? ref("") : ref(store.im_groupinfo.Nickname)
//禁言开关
const shoutUpSwitch = store.im_groupinfo == null ? ref(false) : ref(store.im_groupinfo.ShutUp)

const btName = store.im_groupinfo == null ? ref("创建") : ref("修改")

//创建或修改群信息
const submitForm = () => {
  if(nickname.value == ""){
    ElMessage({
      showClose: true,
      message: '昵称需要至少1个字符',
      type: 'error',
    })
    return
  }
  //创建群
  if(store.im_groupinfo == null){
    Promise.all([ImProxyClient_CreateGroup(nickname.value,"", shoutUpSwitch.value)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      // console.log("创建群",messageOne)
      var result = thistemp.$checkResultCode(messageOne.code)
      if(!result.success){
        console.log("添加失败:"+error)
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
      ElMessage({
        showClose: true,
        message: '添加成功',
        type: 'success',
      })
      // window.history.back()
      store_routers.goback_im()
    }).catch(error => {
      console.log("添加失败:"+error)
      ElMessage({
        showClose: true,
        message: '添加失败：'+error,
        type: 'error',
      })
    });
    return
  }
  //是修改群信息
  Promise.all([ImProxyClient_UpdateGroup(store.im_groupinfo.GroupID, "",
      nickname.value, shoutUpSwitch.value, false)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    // console.log("修改群",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    ElMessage({
      showClose: true,
      message: '添加成功',
      type: 'success',
    })
    // window.history.back()
    store_routers.goback_im()
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '添加失败：'+error,
      type: 'error',
    })
  });
}

const back = () => {
  // window.history.back()
  store_routers.goback_im()
}

function submitSearch() {
  // console.log("search");
  if(this.netaddr == ""){return}
  Promise.all([IM_SearchFriendInfo(this.netaddr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("用户信息",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    this.friendinfo = messageOne.info

  });
}

function addFriend() {
  // console.log("添加好友",this.friendAddr)
  Promise.all([IM_AddFriend(this.friendinfo.Addr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("用户信息",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
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
</script>