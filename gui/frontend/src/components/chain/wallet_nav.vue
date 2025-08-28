

<template>
  <div class="common-layout" style="height:100%;margin: 0;padding: 0;">
    <el-container style="height: 100%;margin: 0;padding: 0;border:red solid 0px;">
      <el-header style="margin: 0;padding: 0;">
        <el-menu :default-active="activeIndex" class="el-menu-demo" mode="horizontal" @select="handleSelect">
          <el-menu-item index="1">钱包</el-menu-item>
          <el-sub-menu index="2">
            <template #title>投票</template>
            <el-menu-item index="2-1">见证人</el-menu-item>
            <el-menu-item index="2-2">社区节点</el-menu-item>
            <el-menu-item index="2-3">轻节点</el-menu-item>
          </el-sub-menu>
          <el-menu-item index="3">域名</el-menu-item>
          <el-menu-item index="4">交易记录</el-menu-item>
          <el-menu-item index="5">见证人</el-menu-item>
          <el-menu-item index="6">密钥</el-menu-item>
        </el-menu>
      </el-header>
      <el-main><router-view/></el-main>
      <el-footer style="height: 30px;border-top:1px solid #ccc;padding:5px 0;">
        同步高度: {{ store.chain_getinfo.CurrentBlock }} / 最新高度:{{ store.chain_getinfo.HighestBlock }}
      </el-footer>
    </el-container>
  </div>
</template>


<script setup>
import { Chain_GetInfo, Chain_GetWitnessInfo } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import { store } from '../../store.js'
import {getCurrentInstance,onMounted , onUnmounted, reactive, ref, nextTick} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this


const activeIndex = ref(['1'])
const activeIndex2 = ref(['1'])


function handleSelect(key, keyPath) {
  // console.log(key, keyPath);
  switch(key){
    case "1":
      thistemp.$router.push({path: '/index/wallet/info'});
      break;
    case "2-1":
      thistemp.$router.push({path: '/index/wallet/witnesslist'});
      break;
    case "2-2":
      thistemp.$router.push({path: '/index/wallet/communitylist'});
      break;
    case "2-3":
      thistemp.$router.push({path: '/index/wallet/selflightlist'});
      break;
    case "3":
      thistemp.$router.push({path: '/index/wallet/name'});
      break;
    case "4":
      thistemp.$router.push({path: '/index/wallet/paylog'});
      break;
    case "5":
      //查询见证人信息
      Promise.all([Chain_GetWitnessInfo()]).then(messages => {
        if(!messages || !messages[0]){return}
        var messageOne = messages[0];
        // console.log("开始获取文件下载列表",messageOne)
        if(messageOne.IsCandidate){
          store.setWitnessinfo(messageOne);
          // this.$router.push({path: '/index/wallet/witnessdepositin'});
          thistemp.$router.push({path: '/index/wallet/witnessinfo'});
          return;
        }
        thistemp.$router.push({path: '/index/wallet/witnessdepositin'});
      });
      break;
    case "6":
      thistemp.$router.push({path: '/index/wallet/exportkey'});
      break;
    default:
      console.log("default",key, keyPath);
  }
}
// 定义一个ref来持有定时器
const timer = ref(null);
// 创建定时器
const createTimer = () => {
  timer.value = setInterval(() => {
    // 定时器的逻辑
    Promise.all([Chain_GetInfo()]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      // console.log("链端基本信息",messageOne)
      // store.chain_getinfo = messageOne
      store.setinfo(messageOne);
      // this.$store.commit('setDownloadListProgress', messageOne);
    });
  }, 1000);
};

//在组件实例挂载到 DOM 后被调用
onMounted(() => {
  //聊天窗口滚动条移动到显示最新消息
  nextTick().then(() => {
    // DOM更新完成后的操作
    createTimer()
  });
});
onUnmounted(() => {
  if (timer.value) {
    clearInterval(timer.value);
  }
});

function created(){
}
created()

</script>