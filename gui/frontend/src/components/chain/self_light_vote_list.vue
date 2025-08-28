<template>
  <div class="about">
        <el-page-header @back="goBack" content="我的投票记录">
        </el-page-header>
  <el-table ref="singleTable" :data="tableData" highlight-current-row @current-change="handleCurrentChange" style="width: 100%">
    <el-table-column type="index" width="50">
    </el-table-column>
    <!-- <el-table-column property="Payload" label="社区节点名称" width="120">
    </el-table-column> -->
    <el-table-column property="WitnessAddr" label="社区节点地址" width="350">
    </el-table-column>
    <el-table-column property="AddrSelf" label="钱包地址" width="350">
    </el-table-column>
    <el-table-column property="Value" label="票数">
    </el-table-column>
    <el-table-column fixed="right" label="操作" width="120">
      <template #default="scope">
        <el-button @click.prevent="vote(scope.$index, tableData)" type="text" size="small">取消</el-button>
      </template>
    </el-table-column>
  </el-table>

  </div>
</template>


<script setup>
import { ElMessage } from 'element-plus'
import { Chain_GetVoteList } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import { store } from '../../store.js'
import {getCurrentInstance, reactive, ref} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const tableData = ref([])
const currentRow = ref(null)


function goBack() {
  history.back(-1);
}

function vote(index, rows) {
  // rows.splice(index, 1);
  store.setLightVoteOut(rows[index]);
  thistemp.$router.push({path: '/index/wallet/lightvoteout'});
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
  Promise.all([Chain_GetVoteList(2)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("投票列表",messageOne)
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
      one.Value = new thistemp.$Calculator().divide(one.Value, store.coinCompany);
    }
    //排序
    messageOne.list = messageOne.list.sort(function(a,b){
      return a.Height - b.Height;
      // return a.TokenId.localeCompare(b.TokenId);
    })
    tableData.value = messageOne.list;
  });

  // var params = JSON.parse(JSON.stringify(store.rpcParams));
  // //@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
  // params.data = {method:"getvotelist",params:{votetype:2}};
  // this.$axios(params).then((response)=> {
  //     // console.log(response.data);
  //     if(response.data.code != 2000){
  //         return
  //     }
  //     //显示处理
  //     for(var i=0; i<response.data.result.length ; i++){
  //         var one = response.data.result[i];
  //         one.Value = new thistemp.$Calculator().divide(one.Value, store.coinCompany);
  //     }
  //     //排序
  //     response.data.result = response.data.result.sort(function(a,b){
  //       return a.Height - b.Height;
  //       // return a.TokenId.localeCompare(b.TokenId);
  //     })
  //     this.tableData = response.data.result;
  // });
}
created()

</script>

