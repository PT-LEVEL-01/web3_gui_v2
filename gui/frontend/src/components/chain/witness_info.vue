<template>
  <div class="about">

    <el-table :data="tableData" style="width: 100%">
      <el-table-column prop="date" label="" width="180">
      </el-table-column>
      <el-table-column prop="value" label="">
      </el-table-column>
    </el-table>
    <router-link to="/index/wallet/witnessdepositout"><el-button type="primary" style="margin-top:20px;">取消见证人资格</el-button></router-link>
  </div>
</template>


<script setup>
import { store } from '../../store.js'
import {getCurrentInstance, reactive, ref} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const tableData = ref([{
  date: '见证人地址',
  value: 'ZHCFpXbEF6DdhvYs8WUZEF5ztDtzDmRqW1jD4',
}, {
  date: '见证人名称',
  value: 'first_witness',
}, {
  date: '押金',
  value: 0,
}, {
  date: '备用见证人',
  value: '是',
}, {
  date: '候选见证人',
  value: '是',
}, {
  date: '是否出块',
  value: '是',
}])

function handleChange(val) {
  // console.log(val);
  // var c = this.$BigNumber();
}

function paytokeninfo(tokeninfo) {
  // console.log(nameinfo)
  store.setPayTokeninfo(tokeninfo);
  thistemp.$router.push({path: '/index/wallet/pay'});
}

function created(){
  var data = store.chain_witnessinfo;
  tableData.value[0].value = data.Addr;
  tableData.value[1].value = data.Payload;
  tableData.value[2].value = data.Value;
  tableData.value[3].value = data.IsBackup ? "是":"否";
  tableData.value[4].value = data.IsCandidate ? "是":"否";
  tableData.value[5].value = data.IsKickOut ? "否":"是";
}
created()

</script>

