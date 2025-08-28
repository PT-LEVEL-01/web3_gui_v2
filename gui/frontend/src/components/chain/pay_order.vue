<template>
  <div>支付订单</div>
  <div>商品名称：{{goodsName}}</div>
  <div>订单号：{{orderId}}</div>
  <div>总价：{{price}}</div>
  <div>收款地址：{{serverAddr}}</div>

  支付密码<el-input v-model="pwd" type="password" autocomplete="off"/>
  <el-button type="primary" @click="submitForm()">支付</el-button>
</template>

<script setup>
import {ElMessage} from "element-plus";
import { PayOrder } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import { store } from '../../store.js'
import {getCurrentInstance, reactive, ref} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const props = defineProps({
  goodsName: String,
  orderId:String,
  price: Number,
  serverAddr:String,
})
// const formRef = ref()

const pwd = ref("")

const submitForm = () => {
  // console.log("订单参数",props)
  Promise.all([PayOrder(props.serverAddr,props.orderId,props.price*100000000,pwd.value)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    // console.log("支付订单",messageOne)
    store.chain_payorder_result = messageOne
    // var result = thistemp.$checkResultCode(messageOne.Code)
    // if(!result.success){
    //   ElMessage({
    //     showClose: true,
    //     message: "code:"+messageOne.Code+" msg:"+result.error,
    //     type: 'error',
    //   })
    //   return
    // }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '支付失败：'+error,
      type: 'error',
    })
  });
}
</script>