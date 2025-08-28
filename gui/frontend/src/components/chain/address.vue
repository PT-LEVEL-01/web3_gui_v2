<template>
  <div class="about" style="user-select: text;">
    <el-page-header @back="goBack" content="钱包地址">
    </el-page-header>
    <router-link to="/index/wallet/addressadd"><el-button type="primary" style="margin-top:20px;">添加新地址</el-button></router-link>
    <el-table :data="tableData" style="width: 100%;margin-top:20px;user-select: text;">
      <el-table-column type="index" width="50">
      </el-table-column>
      <el-table-column prop="AddrCoin" label="钱包地址" width="490" style="user-select: text;">
      </el-table-column>
      <el-table-column prop="Value" label="余额" width="180">
      </el-table-column>
      <el-table-column prop="Type" label="角色">
      </el-table-column>
    </el-table>
  </div>
</template>


<script setup>
import { Chain_GetCoinAddress } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {ElMessage} from "element-plus";
import {getCurrentInstance, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
const input = ref("")
const tableData = ref([])

function goBack() {
  history.back(-1);
}

function created(){
  Promise.all([Chain_GetCoinAddress()]).then(messages => {
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
    tableData.value = new Array();
    console.log("地址列表",messageOne, messageOne[0], messages.length)
    for(var j = 0,len=messageOne.data.length; j < len; j++) {
      var one = messageOne.data[j]
      // var newOne = {};
      var bignumber = thistemp.$BigNumber(one.Value);
      one.Value = bignumber.dividedBy(100000000).toNumber();
      //1=见证人;2=社区节点;3=轻节点;4=什么也不是;
      if(one.Type == 1){
        one.Type = "见证人";
      }else if(one.Type == 2){
        one.Type = "社区节点";
      }else if(one.Type == 3){
        one.Type = "轻节点";
      }else{
        one.Type = "无";
      }
      tableData.value.push(one);
    }
  });
}
created()
</script>