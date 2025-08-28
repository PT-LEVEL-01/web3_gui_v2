<template>
    <el-container style="margin:0;padding:0;height: 100%; border: 0px solid #eee">
        <el-container>
            <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
                <span style="font-size:20px;"><el-button link @click="back()"><el-icon><Back /></el-icon></el-button></span>选择头像
            </el-header>
            <el-main style="user-select: text;">
                <div class="heads" style="float: left;user-select: text;">
                  <el-radio-group v-model="radio1" size="large" fill="#6cf">
                    <el-radio-button @click="selectHead(0)" label="head_daniel" value="head_daniel">
                      <el-avatar shape="square" size="large" :src="head_daniel"/></el-radio-button>
                    <el-radio-button @click="selectHead(1)" label="head_elliot" value="head_elliot">
                      <el-avatar shape="square" size="large" :src="head_elliot"/></el-radio-button>
                    <el-radio-button @click="selectHead(2)" label="head_elyse" value="head_elyse">
                      <el-avatar shape="square" size="large" :src="head_elyse"/></el-radio-button>
                    <el-radio-button @click="selectHead(3)" label="head_helen" value="head_helen">
                      <el-avatar shape="square" size="large" :src="head_helen"/></el-radio-button>
                    <el-radio-button @click="selectHead(4)" label="head_jenny" value="head_jenny">
                      <el-avatar shape="square" size="large" :src="head_jenny"/></el-radio-button>
                    <el-radio-button @click="selectHead(5)" label="head_kristy" value="head_kristy">
                      <el-avatar shape="square" size="large" :src="head_kristy"/></el-radio-button>
                    <el-radio-button @click="selectHead(6)" label="head_matthew" value="head_matthew">
                      <el-avatar shape="square" size="large" :src="head_matthew"/></el-radio-button>
                    <el-radio-button @click="selectHead(7)" label="head_molly" value="head_molly">
                      <el-avatar shape="square" size="large" :src="head_molly"/></el-radio-button>
                    <el-radio-button @click="selectHead(8)" label="head_steve" value="head_steve">
                      <el-avatar shape="square" size="large" :src="head_steve"/></el-radio-button>
                    <el-radio-button @click="selectHead(9)" label="head_stevie" value="head_stevie">
                      <el-avatar shape="square" size="large" :src="head_stevie"/></el-radio-button>
                    <el-radio-button @click="selectHead(10)" label="head_veronika" value="head_veronika">
                      <el-avatar shape="square" size="large" :src="head_veronika"/></el-radio-button>
                  </el-radio-group>
                </div>
                <div style="margin-top: 20px;">
                    <el-button type="success" @click="submit()">确定</el-button>
                </div>
            </el-main>
        </el-container>
    </el-container>
</template>
    
<style>
.headone{
  width:80px;height:80px;border:0;padding:5px;float: left;
}
.headone::selection {
  background: blue;  /* 背景颜色 */
}
</style>
      
<script setup>
import head_daniel from '../../assets/images/head/daniel.jpg'
import head_elliot from '../../assets/images/head/elliot.jpg'
import head_elyse from '../../assets/images/head/elyse.png'
import head_helen from '../../assets/images/head/helen.jpg'
import head_jenny from '../../assets/images/head/jenny.jpg'
import head_kristy from '../../assets/images/head/kristy.png'
import head_matthew from '../../assets/images/head/matthew.png'
import head_molly from '../../assets/images/head/molly.png'
import head_steve from '../../assets/images/head/steve.jpg'
import head_stevie from '../../assets/images/head/stevie.jpg'
import head_veronika from '../../assets/images/head/veronika.jpg'
import { GetHeadUrl } from '../../head_image'
import { ElMessage } from 'element-plus'
import { IM_GetSelfInfo, IM_SetSelfInfo } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, ref } from "vue";
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const radio1 = ref('New York')

// const head_daniel = ref(head_daniel)
// const head_elliot = ref(head_elliot)
// const head_elyse = ref(head_elyse)
// const head_helen = ref(head_helen)
// const head_jenny = ref(head_jenny)
// const head_kristy = ref(head_kristy)
// const head_matthew = ref(head_matthew)
// const head_molly = ref(head_molly)
// const head_steve = ref(head_steve)
// const head_stevie = ref(head_stevie)
// const head_veronika = ref(head_veronika)
const selectNum = ref(0)


function back(){
  window.history.back()
}

function selectHead(num){
  selectNum.value = num
  // console.log(num,this.selectNum)
}

function submit() {
  // console.log("添加好友",this.friendAddr)
  Promise.all([IM_SetSelfInfo(store.im_userinfo_self.Nickname, selectNum.value,
      store.im_userinfo_self.Tray)]).then(messages => {
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
    // var messageOne = messages[0];
    // console.log(messageOne)
    store.im_userinfo_self.HeadNum = selectNum.value
    store.im_userinfo_self.HeadUrl = GetHeadUrl(selectNum.value)
    ElMessage({
      showClose: true,
      message: '修改成功',
      type: 'success',
    })
    thistemp.$router.push({path: '/index/setup/account'});
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '修改失败：'+error,
      type: 'error',
    })
  });
}


function init(){
  //获取个人信息
  Promise.all([IM_GetSelfInfo()]).then(messages => {
    if (!messages || !messages[0]) {
      return
    }
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if (!result.success) {
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    // console.log("获取个人信息", messages)
    var userinfo = messageOne.info;
    userinfo.HeadUrl = GetHeadUrl(userinfo.HeadNum)
    store.im_userinfo_self = userinfo
  });
}
init()

</script>