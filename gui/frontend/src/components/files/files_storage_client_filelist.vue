<template>
    <el-container style="height: 100%;">
      <el-header style="border:red solid 0px;text-align: left;height:70px;">
        <div>{{serverNickname}}</div>
        <el-button-group class="ml-4">
          <el-button type="primary" @click="download"><el-icon><Download /></el-icon></el-button>
          <el-button type="primary" @click="uploadFile">
            上传<el-icon class="el-icon--right"><Upload /></el-icon>
          </el-button>
          <el-button type="primary" @click="dialogVisible = true">新建文件夹</el-button>
          <el-button type="danger" :icon="Delete" @click="onDelete"/>
        </el-button-group>
        <el-breadcrumb separator="/">
          <el-breadcrumb-item><el-button @click="flashFileList(currentDir.ID)" link><el-icon><Refresh /></el-icon></el-button></el-breadcrumb-item>
          <el-breadcrumb-item><el-button @click="flashFileList('')" link><el-icon><HomeFilled /></el-icon></el-button></el-breadcrumb-item>
          <el-breadcrumb-item v-for="(item,i) in dirList"><el-button @click="storageServerPage(item)" link>{{ item.Name }}</el-button></el-breadcrumb-item>
        </el-breadcrumb>
      </el-header>
      <el-main style="border:red solid 0px;">
        <el-table :data="shareboxlist" style="width: 100%;height:100%;" empty-text="没有文件" @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="55" />
          <el-table-column type="index" :index="indexMethod" />
          <el-table-column label="名称" width="380">
            <template #default="scope">
              <el-icon v-if="scope.row.IsDir"><Folder /></el-icon>
              <el-icon v-if="!scope.row.IsDir"><Document /></el-icon>
              &nbsp;{{scope.row.Name}}
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
      </el-main>
    </el-container>

  <el-dialog v-model="dialogVisible" title="请输入新文件夹名称" width="500" draggable overflow center>
    <el-form ref="ruleFormRef" style="max-width: 600px" :model="ruleForm" status-icon :rules="rules" label-width="auto" class="demo-ruleForm">
      <el-form-item label="" prop="dirName">
        <el-input v-model.number="ruleForm.dirName"/>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createDir(ruleFormRef)">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>


<script setup>
import { ElMessage } from 'element-plus'
import {Delete, Download, Upload} from '@element-plus/icons-vue'
import { File_OpenDirectoryDialog, File_OpenMultipleFilesDialog, Storage_Client_UploadFiles, Storage_Client_CreateDir,
  Storage_Client_GetFileList, Storage_Client_download, Storage_Client_DelDirAndFile } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {onBeforeUnmount, reactive, ref, shallowRef, onMounted, watch, getCurrentInstance, nextTick} from 'vue';
import * as wails from "@wailsio/runtime";
import { store } from '../../store.js'

const serverNickname = ref("")
const count = ref(3)
const shareboxlist = ref([])
const serverListInOrder = ref([])//订单中的服务器信息列表
const showGetStartPage = ref(true) //是否显示开始引导页面
const currentDir = ref(null)//当前目录
const ruleFormRef = ref() //
const dialogVisible = ref(false)//新建文件夹对话框
const dirList = ref([])
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
const tableChoice = ref([])


const ruleForm = reactive({
  dirName: '',
})
const checkAge = (rule, value, callback) => {
  if (!value) {
    return callback(new Error('请输入文件夹名称'))
  }
  if (value.length >250){
    return callback(new Error('名称长度不超过250字'))
  }
  callback()
}
const rules = reactive({
  dirName: [{ validator: checkAge, trigger: 'blur' }],
})

//添加拖拽文件事件
wails.Events.On("dragfiles", function(event) {
  console.log("files",event.data)
  var filePaths = new Array()
  event.data.forEach(function(files) {
    files.forEach(function(file) {
      filePaths.push(file)
      // console.log("file",file)
    })
  });
  console.log("上传参数", store.storage_client_selectServerInfo.Addr,currentDir.value.ID,filePaths)
  Promise.all([Storage_Client_UploadFiles(store.storage_client_selectServerInfo.Addr,
      currentDir.value.ID,filePaths)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    // console.log("返回了啥",messageOne)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
  });
})

//创建文件夹
const createDir = (formEl) => {
  if (!formEl) return
  formEl.validate((valid) => {
    if (!valid) {
      // console.log('error submit!')
      return false
    }
    // console.log("创建文件夹",thistemp.$store.state.storage_client_selectServerInfo.Addr,currentDir.value.ID,ruleForm.dirName)
    Promise.all([Storage_Client_CreateDir(store.storage_client_selectServerInfo.Addr,
        currentDir.value.ID,ruleForm.dirName)]).then(messages => {
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
      dialogVisible.value = false
      flashFileList(currentDir.value.ID)
      ElMessage({
        showClose: true,
        message: '成功',
        type: 'success',
      })
    });
  })
}

//表格多选改变事件
const handleSelectionChange = (val) => {
  tableChoice.value = val
}

//进入子文件夹
const storageServerPage = (row) => {
  flashFileList(row.ID)
  var have = false
  for(var i=0; i<dirList.value.length; i++){
    if(dirList.value[i].ID == row.ID){
      dirList.value = dirList.value.slice(0, i+1)
      have = true
      break
    }
  }
  if(!have){
    dirList.value.push(row)
  }
}

//下载文件
const download = (row) => {
  if(tableChoice.value.length==0){
    ElMessage({
      showClose: true,
      message: '请选中要下载的文件复选框',
      type: 'error',
    })
    return
  }
  var ids = new Array()
  for(var i=0; i<tableChoice.value.length;i++){
    ids.push(tableChoice.value[i].Hash)
  }
  // console.log("选中的下载条目",ids)
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
    // return
    Promise.all([Storage_Client_download(store.storage_client_selectServerInfo.Addr,ids,
        messageOne.path)]).then(messages => {
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
    });
  });
}

//打开上传文件列表页面
const uploadFileListPage = (row) => {
  // thistemp.$store.state.storage_client_selectServerInfo = row
  thistemp.$router.push({path: '/index/files/client_uploadlist'});
}

//上传文件
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
    // console.log("选中的文件路径",messageOne)
    // return

    console.log("上传参数", store.storage_client_selectServerInfo.Addr,currentDir.value.ID,messageOne.paths)
    Promise.all([Storage_Client_UploadFiles(store.storage_client_selectServerInfo.Addr,
        currentDir.value.ID,messageOne.paths)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      var result = thistemp.$checkResultCode(messageOne.code)
      // console.log("返回了啥",messageOne)
      if(!result.success){
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
    });
  });
}

//删除多个文件和文件夹
const onDelete = () => {
  if(tableChoice.value.length==0){
    ElMessage({
      showClose: true,
      message: '请选中要删除的文件复选框',
      type: 'error',
    })
    return
  }
  var ids = new Array()
  for(var i=0; i<tableChoice.value.length;i++){
    var one = tableChoice.value[i]
    if(one.Hash==null){
      ids.push(one.ID)
      continue
    }
    ids.push(one.Hash)
  }
  // console.log("选中的删除条目",ids)
  //打开选择目录对话框
  Promise.all([Storage_Client_DelDirAndFile(store.storage_client_selectServerInfo.Addr,ids)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: messageOne.code + result.error,
        type: 'error',
      })
      return
    }
    flashFileList(currentDir.value.ID)
  });
}

//刷新当前文件列表
const flashFileList = (dirID) => {
  if(dirID == ""){
    dirList.value = new Array()
  }
  // console.log(thistemp.$store.state.storage_client_selectServerInfo)
  var sererAddr = store.storage_client_selectServerInfo.Addr
  serverNickname.value = store.storage_client_selectServerInfo.Nickname
  // return
  Promise.all([Storage_Client_GetFileList(sererAddr, dirID)]).then(messages => {
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
    // console.log("获取顶层目录",messageOne)
    currentDir.value = messageOne.dir

    shareboxlist.value = new Array()
    //文件夹排在前面
    for(var i=0 ; i<messageOne.dir.Dirs.length ; i++){
      var one = messageOne.dir.Dirs[i]
      one.IsDir = true
      // one.FileSize = this.$changeSize(one.FileSize)
      // one.residue = "剩余"+(one.Selling - one.Sold)+"G"
      // one.selling = "总共"+one.Selling+"G"
      // one.priceUnit = "单价 "+one.PriceUnit+" 1G/1天"
      // console.log("列表",messageOne.list[i])
      shareboxlist.value.push(one)
    }
    //文件排在后面
    for(var i=0 ; i<messageOne.dir.Files.length ; i++){
      var one = messageOne.dir.Files[i]
      one.FileSize = thistemp.$changeSize(one.FileSize)
      // one.residue = "剩余"+(one.Selling - one.Sold)+"G"
      // one.selling = "总共"+one.Selling+"G"
      // one.priceUnit = "单价 "+one.PriceUnit+" 1G/1天"
      // console.log("列表",messageOne.list[i])
      shareboxlist.value.push(one)
    }
  });
}
flashFileList("")

//在组件实例挂载到 DOM 后被调用
onMounted(() => {
  //聊天窗口滚动条移动到显示最新消息
  nextTick().then(() => {
    // DOM更新完成后的操作
  });
});
//组件销毁时，也及时销毁编辑器，重要！
onBeforeUnmount(() => {
  wails.Events.Off("dragfiles")
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
