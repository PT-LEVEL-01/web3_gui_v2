<template>
  <div style="border:red solid 0px;">
    <div>
      {{serverNickname}}
      <el-button type="primary" :icon="Edit" />
      <el-button type="primary" :icon="Share" />
      <el-button type="primary" :icon="Delete" />
      <el-button type="primary" :icon="Search">Search</el-button>
      <el-button type="primary" @click="uploadFile">
        上传<el-icon class="el-icon--right"><Upload /></el-icon>
      </el-button>
      <el-button type="danger" :icon="Delete" circle />
    </div>

    <!-- <el-button @click="onDelete">Delete Item</el-button> -->
    <el-scrollbar>
      <el-table :data="shareboxlist" style="width: 100%" empty-text="没有文件">
        <el-table-column type="selection" width="55" />
        <el-table-column type="index" :index="indexMethod" />
        <el-table-column prop="Nickname" label="名称" width="180" />
        <el-table-column prop="selling" label="大小" width="100" />
        <el-table-column prop="residue" label="" width="100" />
        <el-table-column prop="priceUnit" label="" width="120" />
        <el-table-column label="">
          <template #default="scope">
            <el-button @click="storageServerPage(scope.row)" style="float: right;">购买</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-scrollbar>
  </div>
</template>


<script setup>
import { ElMessage } from 'element-plus'
import {Check, Delete, Edit, Message, Search, Star, Upload} from '@element-plus/icons-vue'
import { File_GetShareboxList, File_OpenMultipleFilesDialog, File_AddSharebox, File_DelSharebox,
  Storage_client_GetStorageServiceList, Storage_client_GetSearchStorageServiceList,Storage_Client_GetFileList }
  from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, reactive, ref} from 'vue'
import { store } from '../../store.js'

const serverNickname = ref("")
const count = ref(3)
const shareboxlist = ref([])
const serverListInOrder = ref([])//订单中的服务器信息列表
const showGetStartPage = ref(true) //是否显示开始引导页面
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const storageServerPage = (row) => {
  store.storage_client_selectServerInfo = row
  thistemp.$router.push({path: '/index/files/storage_server_info'});
}

const uploadFile = () => {
  //打开选择目录对话框
  Promise.all([File_OpenMultipleFilesDialog()]).then(messages => {
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
    console.log("选中的文件路径",messageOne)
    return
  });
}
const onDelete = (dirPath) => {
  //删除共享目录列表
  Promise.all([File_DelSharebox(dirPath)]).then(messages => {
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
    console.log("删除共享目录列表",messageOne)
    //刷新共享目录列表
    Promise.all([File_GetShareboxList()]).then(messages => {
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
      console.log("刷新共享目录列表",messageOne.list)
      shareboxlist.value = new Array()
      for(var key in messageOne.list){
        shareboxlist.value.push(messageOne.list[key])
      }
    });
  });
  return
  if (count.value > 0) {
    count.value--
  }
}

function created(){
  console.log(store.storage_client_selectServerInfo)
  var sererAddr = store.storage_client_selectServerInfo.Addr
  serverNickname.value = store.storage_client_selectServerInfo.Nickname
  // return
  Promise.all([Storage_Client_GetFileList(sererAddr, "")]).then(messages => {
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
    console.log("获取订单中的云存储列表",messageOne)
    return
    serverListInOrder.value = new Array()
    for(var i=0 ; i<messageOne.list.length ; i++){
      var one = messageOne.list[i]
      one.residue = "剩余"+(one.Selling - one.Sold)+"G"
      one.selling = "总共"+one.Selling+"G"
      one.priceUnit = "单价 "+one.PriceUnit+" 1G/1天"
      // console.log("列表",messageOne.list[i])
      serverListInOrder.value.push(messageOne.list[i])
    }
    console.log(serverListInOrder.value)
    if(messageOne.list.length > 0){
      showGetStartPage.value = false
      return
    }
  });
}
created()
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
