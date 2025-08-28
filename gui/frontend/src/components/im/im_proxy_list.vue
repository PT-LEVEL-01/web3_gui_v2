<template>
  <div v-if="!showGetStartPage" style="border:red solid 0px;">
    <div style="height:50px; padding:5px 10px; border: rgb(238, 236, 236) solid 0px;">
      <div style="width:400px;border:red solid 0px;float:left;text-align: left;">
        已购买的代理服务提供商
      </div>
      <el-space size="10" spacer="|">
<!--        <el-button @click="storageServerSearchPage()" type="text">服务列表</el-button>-->
        <el-button @click="storageOrderList()" type="text">我的订单</el-button>
<!--        <el-button @click="goSetupPage()" type="text">设置</el-button>-->
      </el-space>
    </div>
    <div style="clear: both;"></div>

    <!-- <el-button @click="onDelete">Delete Item</el-button> -->
    <el-scrollbar>
      <el-table :data="serverListInOrder" style="width: 100%">
        <el-table-column type="index" />
        <el-table-column prop="Nickname" label="" width="180" />
        <el-table-column prop="priceUnit" label="" width="180" />
        <el-table-column label="" width="100">
          <template #default="scope">
            <div>{{scope.row.selling}}</div>
            <div>{{scope.row.residue}}</div>
          </template>
        </el-table-column>
        <el-table-column label="">
          <template #default="scope">
<!--            <el-button @click="storageClientFilelistPage(scope.row)" style="float: right;">进入</el-button>-->
          </template>
        </el-table-column>
      </el-table>
    </el-scrollbar>
  </div>
  <div v-if="showGetStartPage" style="border:red solid 0px;">
    <div style="height:50px; padding:5px 10px; border: rgb(238, 236, 236) solid 0px;">
      <div style="width:400px;border:red solid 0px;float:left;text-align: left;">
        选择一个代理商
      </div>
      <el-space size="10" spacer=" ">
        <div>
          <el-button @click="storageOrderList()" type="text">我的订单</el-button>
        </div>
        <div>
          <el-button @click="goSetupPage()" type="text">设置成为服务器</el-button>
        </div>
      </el-space>
    </div>
    <div style="clear: both;"></div>

    <!-- <el-button @click="onDelete">Delete Item</el-button> -->
    <el-scrollbar>
      <el-table :data="shareboxlist" style="width: 100%">
        <el-table-column type="index" />
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
import { IMProxyClient_GetProxyList, ImProxyClient_GetOrderList, IMProxyServer_GetStorageServiceList }
  from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, onMounted, onUnmounted, reactive, ref} from 'vue'
import {store_routers} from "../../store_routers.js";
const count = ref(3)
const shareboxlist = ref([])

const showGetStartPage = ref(true) //是否显示开始引导页面
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

//进入某一存储服务器，查看文件列表页面
// const storageClientFilelistPage = (row) => {
//   thistemp.$store.state.storage_client_selectServerInfo = row
//   thistemp.$router.push({path: '/index/files/client_filelist'});
// }
//去设置代理页面
const goSetupPage = () => {
  // thistemp.$router.push({path: '/index/im/imProxySetup'});
  store_routers.gopage_im("imProxySetup")
}
//显示订单列表页面
const storageOrderList = () => {
  // thistemp.$router.push({path: '/index/im/imProxyOrderList'});
  store_routers.gopage_im("imProxyOrderList")
}
//显示订单页面
const storageServerPage = (row) => {
  thistemp.$store.state.storage_client_selectServerInfo = row
  // thistemp.$router.push({path: '/index/im/imProxyOrder'});
  store_routers.gopage_im("imProxyOrder")
}
//显示提供存储的服务器列表页面
const storageServerSearchPage = () => {
  // thistemp.$router.push({path: '/index/im/imProxyList'});
  store_routers.gopage_im("imProxyList")
}


//获取并刷新搜索的服务器列表
const getSearchStorageList = () => {
  Promise.all([IMProxyClient_GetProxyList()]).then(messages => {
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
      var bignumber = thistemp.$BigNumber(one.PriceUnit);
      one.PriceUnit = bignumber.dividedBy(100000000).toNumber();
      one.priceUnit = "单价 "+one.PriceUnit+" /G/1天"
      // console.log("列表",messageOne.list[i])
      shareboxlist.value.push(messageOne.list[i])
    }
    // console.log(this.shareboxlist)
  });
}
getSearchStorageList()

//获取并刷新已租用的服务器列表
const serverListInOrder = ref([])//订单中的服务器信息列表
const getStorageList = () => {
  Promise.all([IMProxyServer_GetStorageServiceList()]).then(messages => {
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
    console.log("获取订单中的云存储列表",messageOne.list)
    serverListInOrder.value = new Array()
    for(var i=0 ; i<messageOne.list.length ; i++){
      var one = messageOne.list[i]
      // one.FileSize = this.$changeSize(one.Selling)
      one.residue = "剩余 "+thistemp.$changeSize(one.Selling - one.Sold)
      one.selling = "总共 "+thistemp.$changeSize(one.Selling)
      one.priceUnit = "单价 "+one.PriceUnit+" TEST/G/天"
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
getStorageList()


// 创建定时器
// 定义一个ref来持有定时器
const timer = ref(null);
const createTimer = () => {
  timer.value = setInterval(() => {
    // console.log("列表长度:",serverListInOrder.value.length)
    if(serverListInOrder.value.length > 0){
      return
    }
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
