<template>
  <div class="about">
  <el-table ref="singleTable" :data="tableData" highlight-current-row @current-change="handleCurrentChange" style="width: 100%">
    <!-- <el-table-column type="index" width="50">
    </el-table-column> -->
    <el-table-column property="key" label="" width="120">
    </el-table-column>
    <el-table-column property="value" label="" width="450">
    </el-table-column>
  </el-table>
  <!-- <div style="margin-top: 20px">
    <el-button @click="setCurrent(tableData[1])">选中第二行</el-button>
    <el-button @click="setCurrent()">取消选择</el-button>
  </div> -->

  </div>
</template>


<script>
import {IM_GetInfo} from '../../bindings/web3_gui/gui/server_api/sdkapi'

export default {
  data() {
    return {
      tableData: [],
      currentRow: null
    };
  },
  created(){
    // var params = JSON.parse(JSON.stringify(store.rpcParams));
    // params.data = {method:"getnetworkinfo"};
    // this.$axios(params).then((response)=> {
    //     // console.log(response.data);
    //     if(response.data.code != 2000){
    //         return
    //     }

    //     this.tableData.push({"key":"网络地址", "value":response.data.result.netaddr});
    //     this.tableData.push({"key":"是否超级节点", "value":response.data.result.issuper?"是":"否"});
    //     this.tableData.push({"key":"web地址", "value":response.data.result.webaddr});
    //     this.tableData.push({"key":"TCP地址", "value":response.data.result.tcpaddr});
    // });
    
    // var netaddr = IM_GetImfo()
    // console.log(netaddr)
    this.tableData.push({"key":"网络地址", "value":store.im_addrself});
    // params = JSON.parse(JSON.stringify(store.rpcParams));
    // params.data = {method:"imp2p_getinfo"};
    // this.$axios(params).then((response)=> {
    //     // console.log(response.data);
    //     if(response.data.code != 2000){
    //         return
    //     }

    //     this.tableData.push({"key":"网络地址", "value":response.data.result.netaddr});
    // });


  },
  methods: {
      vote(index, rows) {
          // rows.splice(index, 1);
          this.$store.commit('setWitnessScore', rows[index]);
          this.$router.push({path: '/index/wallet/witnessscorein'});
      },
      setCurrent(row) {
        thistemp.$refs.singleTable.setCurrentRow(row);
      },
      handleCurrentChange(val) {
        this.currentRow = val;
      },
      handleChange(val) {
      // console.log(val);
      // var c = this.$BigNumber();
      }
  },
  computed: {
  },
}
</script>

