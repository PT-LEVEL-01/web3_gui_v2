<template>
  <div style="height:100%;position: relative;">
    <el-row style="text-align:left;width:600px;position: absolute;top:20%;left:50%;transform:translate(-50%,-50%);">
      <el-col :span="50">
        <el-button @click="updateHead()" style="width:100px;height:100px;border:0;"><el-avatar shape="square" :size="100" :fit="fit" :src="head" /></el-button>
      </el-col>
      <el-col :span="500" style="margin-left: 60px;">
        <div>昵称：<input type="text" v-model = "nickname" style="border:0;" @blur="updateUserInfo"/></div>
        <div style="margin-top: 40px;user-select: text;">地址：{{ friendAddr }}</div>
        <div style="margin-top: 40px;">托盘：<el-switch v-model="traySwitch" class="ml-2" /></div>

      </el-col>
    </el-row>
  </div>
</template>

<style>
</style>
      
<script setup>
import { ElMessage } from 'element-plus'
import head_daniel from '../../assets/images/head/daniel.jpg'
import { IM_GetSelfInfo, IM_SetSelfInfo } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import { GetHeadUrl } from '../../head_image'
import {getCurrentInstance, reactive, ref, watch} from 'vue'
import { store } from '../../store.js'

const oldTraySwitch = ref(false) //托盘开关
const traySwitch = ref(false) //托盘开关
const oldNickname = ref("") //
const nickname = ref("") //
const friendAddr = ref("") //
const head = ref("") //
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

watch(
    () => traySwitch.value,
    (newVal, oldVal) => {
      console.log("改变",newVal,oldVal)
      updateUserInfo()
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

//修改用户个人信息
const updateUserInfo = () => {
  // console.log(this.oldNickname, this.nickname)
  if(oldNickname.value == nickname.value && oldTraySwitch.value == traySwitch.value){
    return
  }
  oldNickname.value = nickname.value
  oldTraySwitch.value = traySwitch.value
  Promise.all([IM_SetSelfInfo(nickname.value, 0,traySwitch.value)]).then(messages => {
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
    store.im_userinfo_self.Nickname = nickname.value
    // var messageOne = messages[0];
    // console.log(messageOne)
    ElMessage({
      showClose: true,
      message: '修改成功',
      type: 'success',
    })

  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '修改失败：'+error,
      type: 'error',
    })
  });
}

function updateHead(){
  thistemp.$router.push({path: '/index/setup/head'});
}

function init(){
  Promise.all([IM_GetSelfInfo()]).then(messages => {
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
    console.log("个人信息",messageOne)
    oldNickname.value = messageOne.info.Nickname
    nickname.value = messageOne.info.Nickname
    friendAddr.value = messageOne.info.Addr
    head.value = GetHeadUrl(messageOne.info.HeadNum)
    traySwitch.value = messageOne.info.Tray
    oldTraySwitch.value = messageOne.info.Tray
  });
}
init()

</script>