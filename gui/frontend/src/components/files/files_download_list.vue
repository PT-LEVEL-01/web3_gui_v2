<template>
  <div style="text-align: left;">
    <div>正在下载的文件列表</div>
    <el-scrollbar>
      <el-table :data="shareboxlist" style="width: 100%" empty-text="没有文件" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" />
        <el-table-column type="index" :index="indexMethod" />
        <el-table-column label="名称" width="380">
          <template #default="scope">
            <div>
              <el-icon v-if="scope.row.IsDir"><Folder /></el-icon>
              <el-icon v-if="!scope.row.IsDir"><Document /></el-icon>
              &nbsp;{{scope.row.Name}}
            </div>
            <div>
              <el-progress :percentage="percentage(scope.row)" :format="format" />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="FileSize" label="大小" width="80" />
        <!--        <el-table-column prop="residue" label="" width="100" />-->
        <!--        <el-table-column prop="priceUnit" label="" width="120" />-->
        <el-table-column label="">
          <template #default="scope">
            <!--            <el-button v-if="!scope.row.IsDir" @click="download(scope.row)" style="float: right;">下载</el-button>-->
            <el-button v-if="scope.row.IsDir" @click="storageServerPage(scope.row)" style="float: right;">进入</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-scrollbar>

  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { File_StopDownload, File_StartDownload, File_DelDownload, Storage_Client_downloadList, Storage_Client_uploadList }
  from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, ref, onMounted, onUnmounted } from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties //vue3获取当前this
const shareboxlist = ref([])
// 定义一个ref来持有定时器
const timer = ref(null)

const format = (percentage) => (percentage === 100 ? 'success' : `${percentage}%`)
const percentage = (row) => {
  return Math.floor(row.PullSize/row.FileSize*100)
}

//处理文件传输列表进度刷新
const flashDownloadList = () => {
  Promise.all([Storage_Client_downloadList()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      // ElMessage({
      //   showClose: true,
      //   message: "code:"+messageOne.code+" msg:"+result.error,
      //   type: 'error',
      // })
      return
    }
    console.log("下载列表",messageOne)
    shareboxlist.value = messageOne.list
    // ElMessage({
    //   showClose: true,
    //   message: '成功',
    //   type: 'success',
    // })
  });
}

flashDownloadList()
// 创建定时器
const createTimer = () => {
  timer.value = setInterval(() => {
    flashDownloadList()
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


//暂停下载
const stopDownload = (pullTaskID) => {
  console.log(pullTaskID)
  Promise.all([File_StopDownload(pullTaskID)]).then(messages => {
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
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '暂停下载任务失败：'+error,
      type: 'error',
    })
  });
}
//开始继续下载
const startDownload = (pullTaskID) => {
  console.log(pullTaskID)
  Promise.all([File_StartDownload(pullTaskID)]).then(messages => {
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
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '继续下载任务失败：'+error,
      type: 'error',
    })
  });
}
//删除下载任务
const delDownload = (pullTaskID) => {
  console.log(pullTaskID)
  Promise.all([File_DelDownload(pullTaskID)]).then(messages => {
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
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '删除下载任务失败：'+error,
      type: 'error',
    })
  });
}
</script>

<style>
.infinite-list {
  /* height: 300px; */
  padding: 0;
  margin: 0;
  list-style: none;
}
.infinite-list .infinite-list-item {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 50px;
  background: var(--el-color-primary-light-9);
  margin: 10px;
  color: var(--el-color-primary);
}
.infinite-list .infinite-list-item + .list-item {
  margin-top: 10px;
}
</style>
