<template>
  <el-container style="height: 100%; border: 1px solid #eee">
    <el-container>
      <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
        <span style="font-size:20px;">{{nickname}}</span>
      </el-header>
      
      <el-main class="chat_content">
        <el-row :gutter="20">
          <el-col :span="16"><el-input v-model="netaddr" placeholder="请输入内容"></el-input></el-col>
          <el-col :span="8"><el-button type="primary" :icon="Search" @click="submitSearch()">搜索</el-button></el-col>
        </el-row>
        <div v-if="this.friendinfo != null">
          <el-row style="text-align: left;">
            <el-col :span="50"><el-avatar shape="square" :size="100" :fit="fit" :src="head" /></el-col>
            <el-col :span="500">昵称：{{ friendinfo.Nickname }}<br/>地址：{{ friendinfo.Addr }}</el-col>
          </el-row>
          <el-button type="success" @click="addFriend()">添加好友</el-button>
        </div>
      </el-main>
    </el-container>
  </el-container>
    
</template>
  
<style>
</style>
    
<script setup>
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { IM_SearchFriendInfo, IM_AddFriend } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance,ref} from "vue";

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const netaddr = ref("")
const nickname = ref("添加好友")
const friendinfo = ref(null)

function submitSearch() {
  // console.log("search");
  if(netaddr.value === ""){return}
  Promise.all([IM_SearchFriendInfo(netaddr.value)]).then(messages => {
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
    friendinfo.value = messageOne.info
  });
}

function addFriend() {
  // console.log("添加好友",this.friendAddr)
  Promise.all([IM_AddFriend(friendinfo.value.Addr)]).then(messages => {
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