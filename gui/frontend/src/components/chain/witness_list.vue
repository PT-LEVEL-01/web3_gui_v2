<template>
  <div class="about">
        <div style="margin-top: 20px">
            <router-link to="/index/wallet/selfcommunitydepositlist"><el-button>我的质押记录</el-button></router-link>
        </div>
  <el-table ref="singleTable" :data="tableData" highlight-current-row @current-change="handleCurrentChange" style="width: 100%">
    <el-table-column type="index" width="50"></el-table-column>
    <el-table-column property="payload" label="见证人名称" width="120"></el-table-column>
    <el-table-column property="addr" label="地址" width="350"></el-table-column>
    <el-table-column property="ratio" label="分奖比例(%)"></el-table-column>
    <el-table-column property="vote" label="票数"></el-table-column>
    <el-table-column fixed="right" label="Operations" width="120">
      <template #default="scope">
        <el-button @click.prevent="vote(scope.$index, tableData)" type="text" size="small">质押</el-button>
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
import { ElMessage } from 'element-plus'
import { Chain_GetCandidateList } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import { store } from '../../store.js'
import {getCurrentInstance, reactive, ref} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const tableData = ref([])
const currentRow = ref(null)


function vote(index, rows) {
  // console.log("点击质押:",index,rows)
  // rows.splice(index, 1);
  store.setWitnessScore(rows[index]);
  thistemp.$router.push({path: '/index/wallet/witnessscorein'});
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
  Promise.all([Chain_GetCandidateList()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("见证人列表",messageOne)
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
    for(var i=0; i<messageOne.list.length ; i++){
      var one = messageOne.list[i];
      one.vote = new thistemp.$Calculator().divide(one.vote, store.coinCompany);
    }
    //排序
    messageOne.list = messageOne.list.sort(function(a,b){
      return a.vote - b.vote;
      // return a.TokenId.localeCompare(b.TokenId);
    })
    tableData.value = messageOne.list;
  });
  // var params = JSON.parse(JSON.stringify(store.rpcParams));
  // params.data = {method:"getcandidatelist"};
  // this.$axios(params).then((response)=> {
  //     // console.log(response.data);
  //     if(response.data.code != 2000){
  //         return
  //     }
  //     //显示处理
  //     for(var i=0; i<response.data.result.length ; i++){
  //         var one = response.data.result[i];
  //         one.Vote = new thistemp.$Calculator().divide(one.Vote, store.coinCompany);
  //     }
  //     //排序
  //     response.data.result = response.data.result.sort(function(a,b){
  //       return a.Vote - b.Vote;
  //       // return a.TokenId.localeCompare(b.TokenId);
  //     })
  //     this.tableData = response.data.result;
  // });
}
created()
</script>

