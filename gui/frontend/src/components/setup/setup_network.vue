<template>
  <div class="about" style="user-select: text;">
  <el-table ref="singleTable" :data="tableData" highlight-current-row @current-change="handleCurrentChange" style="width: 100%;user-select: text;">
    <el-table-column property="key" label="" width="120">
    </el-table-column>
    <el-table-column property="value" label="" width="500" style="user-select: text;">
    </el-table-column>
  </el-table>
  <el-button v-if="!isRestar" @click="checkAndUpdate()" style="">检查并更新版本</el-button>
  <el-button v-if="isRestar" @click="updateVersion()" style="">重启更新</el-button>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { GetNetwork, GetVersion, CheckUpdateVersion, UpdateRestar } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, ref } from "vue";

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
const tableData = ref([])
const currentRow = ref(null)
const isRestar = ref(true)//是否需要重启


function handleCurrentChange(val) {
  currentRow.value = val;
}

function checkAndUpdate() {
  Promise.all([CheckUpdateVersion()]).then(messages => {
    console.log("检查版本信息",messages)
    var messageOne = messages[0];
    if(messageOne.IsRester){
      ElMessage({
        showClose: true,
        message: '重启后应用新版本'+messageOne.VersionName,
        type: 'success',
      })
      return
    }
    if(messageOne.IsNew){
      ElMessage({
        showClose: true,
        message: '有新的版本，正在下载新版本'+messageOne.VersionName,
        type: 'success',
      })
      return
    }
    ElMessage({
      showClose: true,
      message: '已经是最新版本',
      type: 'success',
    })
  })
}

function updateVersion(){
  Promise.all([UpdateRestar()]).then(messages => {
    console.log(messages)
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: '重启失败',
        type: 'error',
      })
    }
  })
}


function created(){
  Promise.all([GetVersion()]).then(messages => {
    var messageOne = messages[0];
    console.log(messageOne)
    tableData.value.push({"key":"版本号", "value":messageOne});
  })
  Promise.all([GetNetwork()]).then(messages => {
    var messageOne = messages[0];
    console.log(messageOne)
    tableData.value.push({"key":"网络地址", "value":messageOne.NetAddr});
    tableData.value.push({"key":"是否超级节点", "value":messageOne.Issuper?"是":"否"});
    tableData.value.push({"key":"web地址", "value":messageOne.WebAddr});
    tableData.value.push({"key":"TCP地址", "value":messageOne.TCPAddr});
    for(var i=0; i<messageOne.LogicAddr.length; i++){
      var one = messageOne.LogicAddr[i]
      tableData.value.push({"key":"连接" + (i+1), "value":one});
    }
  })
  // this.tableData.push({"key":"网络地址", "value":store.im_addrself});
}
created()

</script>
  
  