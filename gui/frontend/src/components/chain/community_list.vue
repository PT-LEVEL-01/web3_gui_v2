<template>
  <div class="about">
        <!-- <div style="margin-top: 20px">
            <router-link to="/index/wallet/selflightvotelist"><el-button>我的质押记录</el-button></router-link>
        </div> -->
  <el-table ref="singleTable" :data="tableData" highlight-current-row @current-change="handleCurrentChange" style="width: 100%"
            :row-class-name="tableRowClassName">
    <el-table-column type="index" width="50"></el-table-column>
    <el-table-column property="payload" label="社区名称" width="120"></el-table-column>
    <el-table-column property="addr" label="地址" width="360"></el-table-column>
    <el-table-column property="reward_ratio" label="分奖比例(%)"></el-table-column>
    <el-table-column property="vote" label="票数"></el-table-column>
    <el-table-column fixed="right" label="操作" width="160">
      <template #default="scope">
        <el-button @click.prevent="vote(scope.$index, tableData)" type="text" size="small">投票</el-button>
        <el-button v-if="scope.row.CanReward" @click.prevent="DistributeRewards(scope.$index, tableData)" type="text" size="small">
          分发奖励</el-button>
      </template>
    </el-table-column>
  </el-table>

  </div>
</template>

<style>
.el-table .warning-row {
  background: oldlace;
}
.el-table .success-row {
  background: #f0f9eb;
}
</style>

<script setup>
import { ElMessage } from 'element-plus'
import { Chain_GetCoinAddress,Chain_GetCommunityList,Chain_GetVoteList } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
const tableData = ref([])
const currentRow = ref(null)


function vote(index, rows) {
  // rows.splice(index, 1);
  store.setCommunityVote(rows[index]);
  thistemp.$router.push({path: '/index/wallet/communityvotein'});
}

function DistributeRewards(index, rows) {
  // rows.splice(index, 1);
  store.setCommunityVote(rows[index]);
  thistemp.$router.push({path: '/index/wallet/communitydistributerewards'});
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

function tableRowClassName({row, rowIndex}) {
  if (row.IsSelf) {
    return 'warning-row';
  } else {
    return '';
  }
  return '';
}

function created(){
  Promise.all([Chain_GetCoinAddress(), Chain_GetCommunityList(),Chain_GetVoteList(2)]).then(messages => {
    if(!messages || !messages[0]){return}
    var accountList = messages[0];
    // console.log("地址列表",accountList)
    var messageOne = messages[1];
    // console.log("社区列表",messageOne)
    //轻节点给社区投票的地址
    var voteList = messages[2];
    // console.log("轻节点投票记录",voteList)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    //显示处理，社区自己可以发起分奖励
    for(var i=0; i<messageOne.list.length ; i++){
      var one = messageOne.list[i];
      one.CanReward = false
      one.vote = new thistemp.$Calculator().divide(one.vote, store.coinCompany);
      //社区自己相关的地址，改变显示的颜色
      for(var j=0; j<accountList.length; j++){
        var two = accountList[j];
        if(two.AddrCoin == one.addr){
          one.IsSelf = true;
          one.CanReward = true;
          one.Applicant = two.AddrCoin
          break;
        }
      }
    }
    //显示处理，给社区投过票的轻节点可以发起分奖励
    for(var i=0; i<messageOne.list.length ; i++){
      var one = messageOne.list[i];
      //社区自己相关的地址，改变显示的颜色
      for(var j=0; j<voteList.list.length; j++){
        var two = voteList.list[j];
        if(two.WitnessAddr == one.addr){
          // one.IsSelf = true;
          one.CanReward = true;
          one.Applicant = two.AddrSelf
          console.log("自己投过票")
          break;
        }
      }
    }
    //排序
    messageOne.list = messageOne.list.sort(function(a,b){
      return a.vote - b.vote;
      // return a.TokenId.localeCompare(b.TokenId);
    })
    tableData.value = messageOne.list;
    console.log("整理后的社区列表:",messageOne.list)
  });
}
created()
</script>

