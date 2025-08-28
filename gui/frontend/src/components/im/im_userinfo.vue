<template>
        
      <el-container>
        <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
          <span style="font-size:20px;"><el-button link @click="back()"><el-icon><Back /></el-icon></el-button></span>
        </el-header>
        
        <el-main class="chat_content">
          <el-row style="text-align: left;">
            <el-col :span="50"><el-avatar shape="square" :size="100" :fit="fit" :src="head" /></el-col>
            <el-col :span="500">昵称：{{ nickname }}<br/>地址：{{ friendAddr }}</el-col>
          </el-row>
          <el-button type="success" @click="addFriend()" :loading="loading" :disabled="disabled">添加好友</el-button>
          <el-button type="success" @click="getSharebox()" :loading="loading" :disabled="disabled">共享文件</el-button>
          <div style="margin-top: 20px;">已加入圈子</div>
          <div style="">
              <el-button-group v-for="tag in classNames" @click="showNewsTemp(tag.name)" style="margin:5px 10px;">
                  <el-button type="primary" :icon="ArrowLeft">{{ tag.Name }}</el-button>
                  <el-button type="primary">{{ tag.Count }}</el-button>
              </el-button-group>
          </div>
        </el-main>
      </el-container>
</template>
    
<style>
  .float_left{
    text-align: left;
  }
  .float_right{
    text-align: right;
  }
  .el-table__header{
    height:0px;
    display:none;
  }
  .el-textarea__inner{
    border:0;
    line-height:50px;
  }

  .el-table .el-table__cell{
    padding-right: 5px;
    /* position: absolute; */
    /* right:0px; */
    /* border:1px solid red; */
    /* text-align: right; */
  }
  ::-webkit-scrollbar{display:none;}
  .chat_content{
    border-top: 1px solid rgb(224, 224, 224);
    height: 300px;
  }
  .el-header{
    height:40px;
  }
</style>
      
<script setup>
import { ElMessage } from 'element-plus'
import { GetHeadUrl } from '../../head_image'
import { IM_GetFriendInfo, IM_SearchFriendInfo, IM_AddFriend, Circle_FindClassNames }
  from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, ref} from "vue";
import { store } from '../../store.js'
import {store_routers} from "../../store_routers.js";

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const isFriend = ref(false)//是不是好友
const nickname = ref("")
const friendAddr = ref("")
const head = ref("")
const classNames = ref([])
const loading = ref(true)
const disabled = ref(false)


Promise.all([IM_GetFriendInfo(store.im_show_userinfo.netaddr)]).then(messages => {
  if(!messages || !messages[0]){return}
  var messageOne = messages[0];
  var result = thistemp.$checkResultCode(messageOne.code)
  // console.log("IM_GetFriendInfo",messageOne)
  if(!result.success){
    ElMessage({
      showClose: true,
      message: "code:"+messageOne.code+" msg:"+result.error,
      type: 'error',
    })
    return
  }
  if(messageOne.have){
    //是好友
    isFriend.value = true
    nickname.value = messageOne.info.Nickname
    friendAddr.value = messageOne.info.Addr
    head.value = GetHeadUrl(messageOne.info.HeadNum)
    loading.value = false
    disabled.value = false
    //查对方已加入的圈子列表
    Promise.all([Circle_FindClassNames(store.im_show_userinfo.netaddr)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      var result = thistemp.$checkResultCode(messageOne.code)
      // console.log("Circle_FindClassNames",messageOne)
      if(!result.success){
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
      classNames.value = messageOne.ClassNames

    });
    return
  }
  //不是好友，去远端查询
  Promise.all([IM_SearchFriendInfo(store.im_show_userinfo.netaddr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    // console.log("IM_SearchFriendInfo",messageOne)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      loading.value = false
      disabled.value = true
      return
    }
    var userinfo = messageOne.info
    // console.log(messageOne)
    nickname.value = userinfo.Nickname
    friendAddr.value = userinfo.Addr
    head.value = GetHeadUrl(userinfo.HeadNum)
    classNames.value = userinfo.ClassCount
    loading.value = false
    disabled.value = false
  });
});

function addFriend() {
  // console.log("添加好友",this.friendAddr)
  Promise.all([IM_AddFriend(friendAddr.value)]).then(messages => {
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

function getSharebox() {
  // thistemp.$router.push({path: '/index/im/sharebox'});
  store_routers.gopage_im("sharebox")
}
const back = () => {
  // window.history.back()
  store_routers.goback_im()
}
</script>