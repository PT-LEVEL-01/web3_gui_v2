<template>
  <div style="text-align: left;border:red solid 0px;">
    <el-form label-width="120px">
      <el-form-item label="昵称">
        <el-text>{{storageServerInfo.Nickname}}</el-text>
      </el-form-item>
      <el-form-item label="可以出售的空间">
        <el-text>{{storageServerInfo.Selling}}</el-text>
      </el-form-item>
      <el-form-item label="单价">
        <el-text>{{storageServerInfo.PriceUnit}} TEST/1G</el-text>
      </el-form-item>
      <el-form-item label="最长租用时间">
        <el-text>{{storageServerInfo.UseTimeMax}}天</el-text>
      </el-form-item>
      <el-form-item label="购买容量">
        <el-input v-model="OrderSize" style="width: 100px;"></el-input>G
      </el-form-item>
      <el-form-item label="购买时间">
        <el-input v-model="OrderTime" style="width: 100px;"></el-input>天
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="createOrder">创建订单</el-button>
        <el-button @click="back()">返回</el-button>
      </el-form-item>
    </el-form>

  </div>
</template>


<script setup>
import { ElMessage } from 'element-plus'
import {
  Storage_Client_GetOrders,
} from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, reactive, ref, watch} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const storageServerInfo = ref(null)
storageServerInfo.value = store.storage_client_selectServerInfo
const OrderSize = ref(0)//购买容量
const OrderTime = ref(0)//购买时间

const back = () => {
  window.history.back()
}

const createOrder = () => {
  Promise.all([Storage_Client_GetOrders(storageServerInfo.value.Addr, parseInt(OrderSize.value), parseInt(OrderTime.value))]).then(messages => {
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
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '修改失败：'+error,
      type: 'error',
    })
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
