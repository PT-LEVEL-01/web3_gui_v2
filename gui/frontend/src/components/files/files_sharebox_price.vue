<template>
  <el-table :data="shareboxlist" border show-summary style="height:100%;width: 100%">
    <el-table-column type="selection" width="55" />
    <el-table-column label="名称" width="220">
      <template #default="scope">
        <div>{{ scope.row.Name }}</div>
        <div v-loading="!scope.row.IsDir&&scope.row.Hash==''">{{ scope.row.Hash }}</div>
      </template>
    </el-table-column>
<!--    <el-table-column property="name" label="修改时间" width="120"/>-->
    <el-table-column label="价格">
      <template #default="scope">
        <el-input v-show="!scope.row.IsDir" v-loading="!scope.row.IsDir&&scope.row.Hash==''" v-model="scope.row.Price"
                  style="width: 240px" placeholder="0.1" @blur="SetFilePrice(scope.row)"/>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { Sharebox_GetFileHash, Sharebox_SetFilePrice } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, reactive, ref, onMounted, onUnmounted } from 'vue'
import { store } from '../../store.js'

// const store = useStore()
const shareboxlist = ref([])
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
var loopFileHash = store.sharebox_filePrice_process_id != null

// 定义一个ref来持有定时器
const timer = ref(null);
// 创建定时器
const createTimer = () => {
  timer.value = setInterval(() => {
    // 定时器的逻辑
    getFileHash()
  }, 1000);
};

//在组件挂载时创建定时器
onMounted(() => {
  createTimer();
});

//在组件卸载时清除定时器
onUnmounted(() => {
  if (timer.value) {
    clearInterval(timer.value);
  }
});


//初始化页面
const init = () => {
  if(store.sharebox_filePrice_process_id == null){
    //查看已经设置的价格列表
  }else{
  }
}
init()

const getFileHash = () => {
  if(store.sharebox_filePrice_process_id == null){return}
  // var thistemp = this
  //打开选择目录对话框
  Promise.all([Sharebox_GetFileHash(store.sharebox_filePrice_process_id)]).then(messages => {
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
    // console.log("返回的数据",messageOne)
    //任务删除了就停止
    if(messageOne.list == null){
      store.sharebox_filePrice_process_id = null
      return
    }
    //任务完成了也停止
    var finish = true
    for(var i=0;i<messageOne.list.length;i++){
      var one = messageOne.list[i]
      if(one.IsDir){
        continue
      }
      if(one.Hash == ""){
        finish = false
        continue
      }
      if(one.Price!=0){
        var bignumber = thistemp.$BigNumber(one.Price);
        messageOne.list[i].Price = bignumber.dividedBy(100000000).toNumber();
        // console.log("价格",one.Price,bignumber.toNumber(),bignumber.dividedBy(100000000).toNumber(),messageOne.list[i].Price)
      }
    }
    shareboxlist.value = messageOne.list
    if(finish){
      store.sharebox_filePrice_process_id = null
      return
    }
  });
}
//设置文件价格
const SetFilePrice = (fileInfo) => {
  // console.log(fileInfo)
  if(fileInfo.Hash == ""){return}
  var price = parseFloat(fileInfo.Price)
  price = price * 100000000
  // return
  //打开选择目录对话框
  Promise.all([Sharebox_SetFilePrice(fileInfo.Hash,Math.floor(price))]).then(messages => {
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
    // console.log("返回的数据",messageOne)
  });
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
