<template>
    <div style="margin-top:20px;">
    <router-link to="/index/wallet/namereg"><el-button type="primary" plain style="width:100%;"><i class="el-icon-plus"></i> 注册域名</el-button></router-link>

    <el-table :data="nameDatas" :border="parentBorder" style="width: 100%">
      <el-table-column type="expand">
        <template #default="props">
          <div m="4">
            <p m="t-0 b-2">注册时间: {{ props.row.Height }}</p>
            <p m="t-0 b-2">到期时间: {{ props.row.NameOfValidity }}</p>
            <p m="t-0 b-2">缴纳押金: {{ props.row.Deposit }}</p>
            <p m="t-0 b-2">钱包地址: {{ props.row.AddrCoins }}</p>
            <p m="t-0 b-2">网络地址: {{ props.row.NetIds }}</p>
            <el-button type="text" @click="destroyName(props.row)">注销域名</el-button>
            <el-button type="text" @click="renewName(props.row)">续费</el-button>
            <el-button type="text" @click="renewName(props.row)">修改</el-button>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="域名名称" prop="Name" />
      <el-table-column label="有效期" prop="NameOfValidity" />
    </el-table>

    </div>
</template>

<style>
  .demo-table-expand {
    font-size: 0;
  }
  .demo-table-expand label {
    width: 90px;
    color: #99a9bf;
  }
  .demo-table-expand .el-form-item {
    margin-right: 0;
    margin-bottom: 0;
    width: 50%;
  }
</style>

<script setup>
import { ElMessage } from 'element-plus'
import { Chain_GetNames } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, onUnmounted, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
const nameDatas = ref(null)

function destroyName(nameinfo) {
  // console.log(nameinfo)
  store.setNamedestroy(nameinfo);
  thistemp.$router.push({path: '/index/wallet/namedestroy'});
}

function renewName(nameinfo) {
  // console.log(nameinfo)
  store.setNameinfo(nameinfo);
  thistemp.$router.push({path: '/index/wallet/namereg'});
}

function created(){
  Promise.all([Chain_GetNames()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("获取域名列表",messageOne, messageOne[0], messages.length)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    //显示处理
    for(var i=0; i<messageOne.List.length ; i++){
      var one = messageOne.List[i];
      var bignumber = thistemp.$BigNumber(one.Deposit);
      one.Deposit = bignumber.dividedBy(100000000).toNumber();
    }
    //排序
    messageOne.List = messageOne.List.sort(function(a,b){
      return a.NameOfValidity - b.NameOfValidity;
      // return a.TokenId.localeCompare(b.TokenId);
    })
    nameDatas.value = messageOne.List;
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '获取域名列表失败：'+error,
      type: 'error',
    })
  });
  // var params = JSON.parse(JSON.stringify(store.rpcParams));
  // params.data = {method:"getnames"};
  // this.$axios(params).then((response)=> {
  //     console.log(response.data);
  //     if(response.data.code != 2000){
  //         return
  //     }
  //     //显示处理
  //     for(var i=0; i<response.data.result.length ; i++){
  //         var one = response.data.result[i];
  //         var bignumber = this.$BigNumber(one.Deposit);
  //         one.Deposit = bignumber.dividedBy(100000000).toNumber();
  //     }
  //     //排序
  //     response.data.result = response.data.result.sort(function(a,b){
  //       return a.NameOfValidity - b.NameOfValidity;
  //       // return a.TokenId.localeCompare(b.TokenId);
  //     })
  //     this.nameDatas = response.data.result;
  // });
}
created()
</script>
