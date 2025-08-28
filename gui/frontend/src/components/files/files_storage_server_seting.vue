<template>
  <div style="text-align: left;border:red solid 0px;">
    <el-form label-width="120px">
      <el-form-item label="昵称">
        <input type="text" v-model = "nickname" style="border:0;" @blur="updateNickname"/>
      </el-form-item>
      <el-form-item label="云服务器状态">
        <el-switch inline-prompt active-text="开" inactive-text="关" v-model="storageServerSwitch"/>
      </el-form-item>
      <el-form-item label="可以出售的空间">
        <el-input-number v-model="Selling" @change="updateNickname" :min="SellingMin" :max="SellingMax" label="描述文字"></el-input-number>G
      </el-form-item>
      <el-form-item label="单价">
        <input type="text" v-model = "PriceUnit" style="border:0;" @blur="updateNickname"/>TEST/1G
      </el-form-item>
      <el-form-item label="最长租用时间">
        <input type="text" v-model = "UseTimeMax" style="border:0;" @blur="updateNickname"/>天
      </el-form-item>
    </el-form>

    <div style="height:50px; padding:5px 10px; border: rgb(238, 236, 236) solid 1px;">
      <div style="width:400px;border:red solid 0px;float:left;text-align: left;">
        添加空闲位置
      </div>
      <div style="width:100px;border:red solid 0px;float: right;">
        <el-button @click="add">选择文件夹</el-button>
      </div>
    </div>
    <div style="clear: both;"></div>
    <el-table :data="shareboxlist" style="width: 100%">
      <el-table-column type="index" :index="indexMethod" />
      <el-table-column prop="size" label="" width="80" />
      <el-table-column prop="path" label="" />
      <el-table-column label="">
        <template #default="scope">
          <el-button @click="onDelete(scope.row.path)" style="float: right;">删除</el-button>
        </template>
      </el-table-column>
    </el-table>


  </div>
</template>


<script setup>
import { ElMessage } from 'element-plus'
import {
  File_GetShareboxList,
  File_OpenDirectoryDialog,
  IM_SetSelfInfo,
  Storage_server_SetPriceUnit,
  Storage_server_AddDirectory,
  Storage_server_DelDirectory,
  Storage_server_GetStatus,
  Storage_server_SetOpen
} from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, reactive, ref, watch} from 'vue'

import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
const count = ref(3)
const shareboxlist = ref([])
const storageServerSwitch = ref(false)
const nickname = ref("")
const oldnickname = ref("")
const serverInfo = ref(null)

const oldPriceUnit = ref(0)//单价
const PriceUnit = ref(0)//单价
const oldSelling = ref(0)//可以出售的空间
const Selling = ref(0)//可以出售的空间
const SellingMin = ref(0)//最小可以出售的空间
const SellingMax = ref(0)//最大可以出售的空间
const oldUseTimeMax = ref(0)//最大可以租用空间的时间
const UseTimeMax = ref(0)//最大可以租用空间的时间


const getstatus = () => {
  Promise.all([Storage_server_GetStatus()]).then(messages => {
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
    serverInfo.value = messageOne.storageServerInfo
    console.log("存储服务器信息",messageOne.storageServerInfo)
    storageServerSwitch.value = messageOne.storageServerInfo.IsOpen
    nickname.value = messageOne.storageServerInfo.Nickname
    oldnickname.value = messageOne.storageServerInfo.Nickname
    oldPriceUnit.value = messageOne.storageServerInfo.PriceUnit
    PriceUnit.value = messageOne.storageServerInfo.PriceUnit
    oldSelling.value = messageOne.storageServerInfo.Selling
    Selling.value = messageOne.storageServerInfo.Selling
    oldUseTimeMax.value = messageOne.storageServerInfo.UseTimeMax
    UseTimeMax.value = messageOne.storageServerInfo.UseTimeMax
    SellingMax.value = 0
    var dirs = messageOne.storageServerInfo.Directory
    shareboxlist.value = new Array()
    for(var i=0; dirs !=null && i<dirs.length; i++){
      var size = messageOne.storageServerInfo.DirectoryFreeSize[i]
      var one = {size:size+" G",path:dirs[i]}
      SellingMax.value += size
      shareboxlist.value.push(one)
    }
  });
}
getstatus()



const updateNickname = () => {
  var haveChange = false
  if(oldnickname.value != nickname.value){
    haveChange = true
  }
  if(oldPriceUnit.value != PriceUnit.value){
    haveChange = true
  }
  if(oldSelling.value != Selling.value){
    haveChange = true
  }
  if(oldUseTimeMax.value != UseTimeMax.value){
    haveChange = true
  }
  if(!haveChange){
    return
  }

  Promise.all([Storage_server_SetPriceUnit(nickname.value, parseInt(PriceUnit.value),Selling.value,Selling.value,
      Selling.value,parseInt(UseTimeMax.value),serverInfo.value.RenewalTime)]).then(messages => {
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
      message: '成功',
      type: 'success',
    })
    oldnickname.value = nickname.value
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '修改失败：'+error,
      type: 'error',
    })
  });
}

//监听服务器切换按钮开关
watch(
    () => storageServerSwitch.value,
    (newVal, oldVal) => {
      console.log(newVal,oldVal)
      Promise.all([Storage_server_SetOpen(newVal)]).then(messages => {
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
        // storageServerSwitch.value = messageOne.storageServerInfo.IsOpen
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

const add = () => {
  //打开选择目录对话框
  Promise.all([File_OpenDirectoryDialog()]).then(messages => {
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
    console.log("选中的文件路径",messageOne.path)
    if(messageOne.path == ""){return}
    //添加目录
    Promise.all([Storage_server_AddDirectory(messageOne.path)]).then(messages => {
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
        message: '成功',
        type: 'success',
      })
      getstatus()
    });
  });
  count.value++
}
const onDelete = (dirPath) => {
  //删除共享目录列表
  Promise.all([Storage_server_DelDirectory(dirPath)]).then(messages => {
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
    // console.log("删除共享目录列表",messageOne)
    //刷新共享目录列表
    getstatus()
  });
  return
  if (count.value > 0) {
    count.value--
  }
}

</script>

<style scoped>
.scrollbar-demo-item {
  /* display: flex; */
  /* align-items: center; */
  /* justify-content: left; */
  height: 50px;
  margin: 10px 0;
  padding:10px;
  /* text-align: left; */
  border-radius: 4px;
  border:rgb(238, 236, 236) solid 1px;
  /* background: var(--el-color-primary-light-9); */
  /* color: var(--el-color-primary); */
}
</style>
