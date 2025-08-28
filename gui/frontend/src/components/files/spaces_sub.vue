<template>
  <div class="about">
    <el-page-header @back="goBack" content="添加存储空间">
    </el-page-header>
    <router-link to="/index/wallet/addressadd"><el-button type="primary" style="margin-top:20px;">添加新地址</el-button></router-link>
    <el-table :data="tableData" style="width: 100%;margin-top:20px;">
      <el-table-column type="index" width="50">
      </el-table-column>
      <el-table-column prop="AddrCoin" label="钱包地址" width="380">
      </el-table-column>
      <el-table-column prop="Value" label="余额" width="180">
      </el-table-column>
      <el-table-column prop="Type" label="角色">
      </el-table-column>
    </el-table>
  </div>
</template>


<script setup>
import { store } from '../../store.js'
import {getCurrentInstance, reactive, ref} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties //vue3获取当前this
const input = ref('')
const tableData = ref(null)

function goBack() {
  history.back(-1);
}

function created(){
  // var params = JSON.parse(JSON.stringify(store.rpcParams)); //Object.create(store.rpcParams);
  // params.data = {method:"listaccounts",params:{token_id:""}};
  // this.$axios(params).then((response)=> {
  //   // console.log("123");
  //   if(response.data.code != 2000){
  //     return
  //   }
  //   console.log(response.data);
  //   this.tableData = response.data.result;
  //   this.tableData = new Array();
  //
  //   for(var j = 0,len=response.data.result.length; j < len; j++) {
  //     var one = response.data.result[j]
  //     // var newOne = {};
  //     var bignumber = this.$BigNumber(one.Value);
  //     one.Value = bignumber.dividedBy(100000000).toNumber();
  //     //1=见证人;2=社区节点;3=轻节点;4=什么也不是;
  //     if(one.Type == 1){
  //       one.Type = "见证人";
  //     }else if(one.Type == 2){
  //       one.Type = "社区节点";
  //     }else if(one.Type == 3){
  //       one.Type = "轻节点";
  //     }else{
  //       one.Type = "无";
  //     }
  //     this.tableData.push(one);
  //   }
  //   // this.HighestBlock = response.data.result.HighestBlock;
  //   // this.PulledStates = response.data.result.PulledStates;
  //   // this.$store.commit('setinfo', response.data.result);
  // });
}
created()

</script>