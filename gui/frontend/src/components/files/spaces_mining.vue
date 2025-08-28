<template>
  <div class="about">
        <!-- <el-page-header @back="goBack" content="我的投票记录">
        </el-page-header> -->
    <router-link to="/index/files/spacesminingadd"><el-button type="primary" style="margin-top:20px;">添加空间</el-button></router-link>
  <el-table ref="singleTable" :data="tableData" highlight-current-row @current-change="handleCurrentChange" style="width: 100%">
    <el-table-column type="index" width="50">
    </el-table-column>
    <el-table-column property="DbPath" label="存储路径" width="350">
    </el-table-column>
    <el-table-column property="SpaceNum" label="空间大小(G)" width="150">
    </el-table-column>
    <el-table-column fixed="right" label="操作" width="120">
      <template slot-scope="scope">
        <!-- <el-button @click.native.prevent="delSpaces(scope.$index, tableData)" type="text" size="small">删除</el-button> -->
      </template>
    </el-table-column>
  </el-table>
  <!-- <div style="margin-top: 20px">
    <el-button @click="setCurrent(tableData[1])">选中第二行</el-button>
    <el-button @click="setCurrent()">取消选择</el-button>
  </div> -->

  </div>
</template>


<script setup>
import {getCurrentInstance, ref, onMounted, onUnmounted } from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties //vue3获取当前this
const tableData = ref([])
const currentRow = ref(null)

function goBack() {
  history.back(-1);
}

function delSpaces(index, rows) {
  console.log(rows[index])
  // rows.splice(index, 1);
  // this.$store.commit('setLightVoteOut', rows[index]);
  // this.$router.push({path: '/index/wallet/lightvoteout'});
}

function setCurrent(row) {
  thistemp.$refs.singleTable.setCurrentRow(row);
}

function handleCurrentChange(val) {
  currentRow.value = val;
}

function handleChange(val) {
  // console.log(val);
  // var c = this.$BigNumber();
}

function created(){
  // var params = JSON.parse(JSON.stringify(store.rpcParams));
  // params.data = {method:"getminingspacelist"};
  // this.$axios(params).then((response)=> {
  //   console.log(response.data);
  //   if(response.data.code != 2000){
  //     return
  //   }
  //   //显示处理
  //   for(var i=0; i<response.data.result.SpaceList.length ; i++){
  //     var one = response.data.result.SpaceList[i];
  //     one.SpaceNum = new thistemp.$Calculator().divide(one.SpaceNum, 1024 * 1024 * 1024);
  //   }
  //   // //排序
  //   // response.data.result = response.data.result.sort(function(a,b){
  //   //   return a.Height - b.Height;
  //   //   // return a.TokenId.localeCompare(b.TokenId);
  //   // })
  //   this.tableData = response.data.result.SpaceList;
  // });
}
created()

</script>

