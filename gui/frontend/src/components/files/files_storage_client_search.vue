<template>
  <div v-if="!showGetStartPage" style="border:red solid 0px;">
    <div style="height:50px; padding:5px 10px; border: rgb(238, 236, 236) solid 0px;">
      <div style="width:400px;border:red solid 0px;float:left;text-align: left;">
        已购买的云存储提供商
      </div>
      <el-button type="text">我的订单</el-button>
    </div>
    <div style="clear: both;"></div>

    <!-- <el-button @click="onDelete">Delete Item</el-button> -->
    <el-scrollbar>
      <el-table :data="serverListInOrder" style="width: 100%">
        <el-table-column type="index" :index="indexMethod" />
        <el-table-column prop="Nickname" label="" width="180" />
        <el-table-column label="" width="100">
          <template #default="scope">
            <div>{{scope.row.selling}}</div>
            <div>{{scope.row.residue}}</div>
          </template>
        </el-table-column>
        <el-table-column label="">
          <template #default="scope">
            <el-button @click="storageClientFilelistPage(scope.row)" style="float: right;">进入</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-scrollbar>
  </div>
  <div v-if="showGetStartPage" style="border:red solid 0px;">
    <div style="height:50px; padding:5px 10px; border: rgb(238, 236, 236) solid 0px;">
      <div style="width:400px;border:red solid 0px;float:left;text-align: left;">
        选择一个云存储提供商
      </div>
      <el-button type="text">我的订单</el-button>
    </div>
    <div style="clear: both;"></div>

    <!-- <el-button @click="onDelete">Delete Item</el-button> -->
    <el-scrollbar>
      <el-table :data="shareboxlist" style="width: 100%">
        <el-table-column type="index" :index="indexMethod" />
        <el-table-column prop="Nickname" label="" width="180" />
        <el-table-column prop="selling" label="" width="100" />
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
import {
  File_GetShareboxList,
  File_OpenDirectoryDialog,
  File_AddSharebox,
  File_DelSharebox,
  Storage_client_GetStorageServiceList,
  Storage_client_GetSearchStorageServiceList,
  Storage_Client_uploadList
} from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, ref, onMounted, onUnmounted } from 'vue'
import { store } from '../../store.js'

const count = ref(3)
const shareboxlist = ref([])
const serverListInOrder = ref([])//订单中的服务器信息列表
const showGetStartPage = ref(true) //是否显示开始引导页面
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const storageClientFilelistPage = (row) => {
  store.storage_client_selectServerInfo = row
  thistemp.$router.push({path: '/index/files/client_filelist'});
}
const storageServerPage = (row) => {
  store.storage_client_selectServerInfo = row
  thistemp.$router.push({path: '/index/files/storage_server_info'});
}

const add = () => {
  // var thistemp = this
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
    if(messageOne.path.length == 3 && messageOne.path.match(/[A-Z]:\\/) != null){
      ElMessage({
        showClose: true,
        message: '不能选择根目录',
        type: 'error',
      })
      return
    }
    // for(var i=0; i<messageOne.length; i++){
    //   var filePath = messageOne[i]
    //   if (editor == null) return;
    //   editor.dangerouslyInsertHtml('<a class="sendFile" href="" target="_blank">'+filePath+'</a><br>')
    // }
    //添加目录
    Promise.all([File_AddSharebox(messageOne.path)]).then(messages => {
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
      console.log("添加目录",messageOne)
      // for(var i=0; i<messageOne.length; i++){
      //   var filePath = messageOne[i]
      //   if (editor == null) return;
      //   editor.dangerouslyInsertHtml('<a class="sendFile" href="" target="_blank">'+filePath+'</a><br>')
      // }
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
  });
  count.value++
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
    // console.log("删除共享目录列表",messageOne)
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
      console.log("刷新共享目录列表",messageOne)
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


//获取并刷新搜索的服务器列表
const getSearchStorageList = () => {
  Promise.all([Storage_client_GetSearchStorageServiceList()]).then(messages => {
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
    console.log("获取广播的云存储列表",messageOne.list)
    shareboxlist.value = new Array()
    for(var i=0 ; i<messageOne.list.length ; i++){
      var one = messageOne.list[i]
      one.residue = "剩余"+(one.Selling - one.Sold)+"G"
      one.selling = "总共"+one.Selling+"G"
      one.priceUnit = "单价 "+one.PriceUnit+" 1G/1天"
      // console.log("列表",messageOne.list[i])
      shareboxlist.value.push(messageOne.list[i])
    }
    // console.log(this.shareboxlist)
  });
}
getSearchStorageList()

// 创建定时器
// 定义一个ref来持有定时器
const timer = ref(null);
const createTimer = () => {
  timer.value = setInterval(() => {
    getSearchStorageList()
    // flashUploadList()
    // 定时器的逻辑
  }, 1000);
};

// 在组件挂载时创建定时器
onMounted(() => {
  createTimer();
});

// 在组件卸载时清除定时器
onUnmounted(() => {
  if (timer.value) {
    clearInterval(timer.value);
  }
});

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
