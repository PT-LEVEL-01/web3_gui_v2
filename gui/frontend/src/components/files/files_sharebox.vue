<template>
  <div style="border:red solid 0px;">
    <div style="height:50px; padding:5px 10px; border: rgb(238, 236, 236) solid 1px;">
      <div style="width:400px;border:red solid 0px;float:left;text-align: left;">
        价格设置<br/>给文件单独设置价格，付费下载
      </div>
      <div style="width:100px;border:red solid 0px;float: right;">
        <el-button @click="addPriceSettingDir">打开文件夹</el-button>
      </div>
      <div style="width:100px;border:red solid 0px;float: right;">
        <el-button @click="addPriceSettingFile">打开文件</el-button>
      </div>
      <div style="width:100px;border:red solid 0px;float: right;">
        <el-button @click="addPriceSettingDir">修改价格</el-button>
      </div>
    </div>

    <div style="height:50px; padding:5px 10px; border: rgb(238, 236, 236) solid 1px;">
      <div style="width:400px;border:red solid 0px;float:left;text-align: left;">
        选择文件夹<br/>共享你选择的文件夹中的文件，其他人可以下载
      </div>
      <div style="width:100px;border:red solid 0px;float: right;">
        <el-button @click="add">选择文件夹</el-button>
      </div>
    </div>
    <div style="clear: both;"></div>
    
    <!-- <el-button @click="onDelete">Delete Item</el-button> -->
    <el-scrollbar>
      <div v-for="item in shareboxlist" :key="item" class="scrollbar-demo-item">
        <p style="float: left;">{{ item }}</p>
        <el-button @click="onDelete(item)" style="float: right;">删除</el-button>
      </div>
    </el-scrollbar>
  </div>
</template>


<script setup>
import { ElMessage } from 'element-plus'
import { File_GetShareboxList, File_OpenDirectoryDialog, File_OpenMultipleFilesDialog, File_AddSharebox,
  File_DelSharebox, Sharebox_GetFileInfo } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, reactive, ref} from 'vue'
import { store } from '../../store.js'

const count = ref(3)
const shareboxlist = ref([])
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

//打开要设置价格的文件
const addPriceSettingFile = () => {
  // var thistemp = this
  //打开选择目录对话框
  Promise.all([File_OpenMultipleFilesDialog()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+messageOne.error,
        type: 'error',
      })
      return
    }
    // console.log("选中的文件路径",messageOne.paths)
    if(messageOne.paths.length == 0)return
    //添加目录
    Promise.all([Sharebox_GetFileInfo(null,messageOne.paths)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      var result = thistemp.$checkResultCode(messageOne.code)
      if(!result.success){
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+messageOne.error,
          type: 'error',
        })
        return
      }
      // console.log("文件信息",messageOne)
      store.sharebox_filePrice_process_id = messageOne.pid
      thistemp.$router.push({path: '/index/files/sharebox_price'});
    });
  });
  count.value++
}

//打开要设置价格的文件夹
const addPriceSettingDir = () => {
  // var thistemp = this
  //打开选择目录对话框
  Promise.all([File_OpenDirectoryDialog()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+messageOne.error,
        type: 'error',
      })
      return
    }
    // console.log("选中的文件夹路径",messageOne.path)
    if(messageOne.path == "")return
    //添加目录
    Promise.all([Sharebox_GetFileInfo([messageOne.path])]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      var result = thistemp.$checkResultCode(messageOne.code)
      if(!result.success){
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+messageOne.error,
          type: 'error',
        })
        return
      }
      // console.log("文件信息",messageOne)
      store.sharebox_filePrice_process_id = messageOne.pid
      thistemp.$router.push({path: '/index/files/sharebox_price'});
    });
  });
  count.value++
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
        message: "code:"+messageOne.code+" msg:"+messageOne.error,
        type: 'error',
      })
      return
    }
    // console.log("选中的文件路径",messageOne.path)
    if(messageOne.path == "")return
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
          message: "code:"+messageOne.code+" msg:"+messageOne.error,
          type: 'error',
        })
        return
      }
      //刷新共享目录列表
      Promise.all([File_GetShareboxList()]).then(messages => {
        if(!messages || !messages[0]){return}
        var messageOne = messages[0];
        var result = thistemp.$checkResultCode(messageOne.code)
        if(!result.success){
          ElMessage({
            showClose: true,
            message: "code:"+messageOne.code+" msg:"+messageOne.error,
            type: 'error',
          })
          return
        }
        // console.log("刷新共享目录列表",messageOne.list)
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
        message: "code:"+messageOne.code+" msg:"+messageOne.error,
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
          message: "code:"+messageOne.code+" msg:"+messageOne.error,
          type: 'error',
        })
        return
      }
      // console.log("刷新共享目录列表",messageOne.list)
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
  Promise.all([File_GetShareboxList()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    // console.log("返回错误:",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+messageOne.error,
        type: 'error',
      })
      return
    }
    // console.log("选中的文件路径",messageOne.list)
    shareboxlist.value = new Array()
    for(var key in messageOne.list){
      shareboxlist.value.push(messageOne.list[key])
    }
    console.log(shareboxlist.value)

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
  