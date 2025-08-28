<template>
<div>
      <el-page-header @back="goBack" content="注册轻节点">
    </el-page-header>
<el-form :model="ruleForm" status-icon :rules="rules" ref="ruleForm" label-width="100px" class="demo-ruleForm"  style="margin-top:40px;">

  <el-form-item label="轻节点地址" prop="addr">
    <el-button @click="openaddrs()">选择地址</el-button>{{ruleForm.addr}}
    <!-- <el-input v-model="ruleForm.addr" autocomplete="off"></el-input> -->
  </el-form-item>
  <el-form-item label="金额" prop="amount">
    <el-input v-model="ruleForm.amount" autocomplete="off"></el-input>
  </el-form-item>
  <el-form-item label="手续费" prop="gas">
    <el-input v-model="ruleForm.gas"></el-input>
  </el-form-item>
  <el-form-item label="密码" prop="pass">
    <el-input type="password" v-model="ruleForm.pass"></el-input>
  </el-form-item>
  <el-form-item>
    <el-button type="primary" @click="submitForm('ruleForm')">立即注册</el-button>
    <el-button @click="resetForm('ruleForm')">重置</el-button>
  </el-form-item>
</el-form>

<el-dialog v-model="dialogTableVisible" title="Shipping address">
  <div>
    <el-radio-group v-model="ruleForm.addr">
      <li v-for="(item, index) in ruleForm.addrs" :key="index">
        <el-radio :label="item.AddrCoin" v-bind:disabled="item.Disable" >{{item.AddrCoin}}</el-radio>
      </li>
    </el-radio-group>
  </div>
  <el-button @click="closeaddrs()">选择地址</el-button>
  <!-- <el-button @click="closeaddrs()">关闭</el-button> -->
</el-dialog>

<!-- <el-drawer title="我是标题" :visible.sync="drawer" size="50%" :with-header="false">
  <el-button @click="closeaddrs()">选择地址</el-button>
  <el-button @click="closeaddrs()">关闭</el-button>
  <el-radio-group v-model="ruleForm.addr">
    <li v-for="(item, index) in ruleForm.value.addrs" :key="index">
      <el-radio :label="item.AddrCoin">{{item.AddrCoin}}</el-radio>
    </li>
  </el-radio-group>
</el-drawer> -->
</div>
</template>
<script setup>
import { ElMessage } from 'element-plus'
import { Chain_GetCoinAddress, Chain_VoteIn } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, onUnmounted, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const checkDepositInAmount = (rule, value, callback) =>{
  if(value * store.coinCompany < store.chain_getinfo.LightNode){
    var big = new thistemp.$Calculator();
    var value = big.divide(store.chain_getinfo.LightNode, store.coinCompany);
    callback(new Error('轻节点押金不能少于 '+ value));
    return
  }
  callback();
}

const dialogTableVisible = ref(false)
const drawer = ref(false)
const ruleForm = ref({
  addrs: new Array(),
  amount:"",
  addr:"",
  gas:"",
  comment:"",
  pass: '',
  disable:false,
})
const rules = ref({
  amount: [{required: true, validator: thistemp.$checkAmountNotZero, message: '押金不能为空', trigger: 'blur' },
    {required: true, validator: checkDepositInAmount, trigger: 'blur' }],
      addr: [{required: true, message: '地址不能为空', trigger: 'blur' }],
      gas: [{required: true, validator: thistemp.$checkAmountHaveZero, trigger: 'blur' }],
      comment: [{required: true, message: '社区名称不能为空', trigger: 'blur'}],
      pass: [{required: true, message: '密码不能为空', trigger: 'blur' }]
})


function goBack() {
  history.back(-1);
}

function submit(){

}

function handleCommand(command) {
  // this.$message('click on item ' + command);
  ruleForm.value.addr = command;
}

function submitForm(formName) {
  if(ruleForm.value.addr == ""){
    ElMessage({
      showClose: true,
      message: "请选择一个轻节点地址",
      type: 'error',
    })
    return
  }
  thistemp.$refs[formName].validate((valid) => {
    if (!valid){
      return false;
    }
    var gas = 0;
    if (ruleForm.value.gas > 0){
      gas = ruleForm.value.gas * store.coinCompany;
    }
    var amount = 0;
    if (ruleForm.value.amount > 0){
      amount = ruleForm.value.amount * store.coinCompany;
    }

    Promise.all([Chain_VoteIn(3,ruleForm.value.addr,"",0,amount,gas,0,
        ruleForm.value.pass,ruleForm.value.comment)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      console.log("轻节点押金",messageOne)
      var result = thistemp.$checkResultCode(messageOne.code)
      if(!result.success){
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
      ElMessage({
        showClose: true,
        message: '成功',
        type: 'success',
      })
    });
    return

    // var params = JSON.parse(JSON.stringify(store.rpcParams));
    // //@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
    // params.data = {method:"votein",params:{votetype:3,address:ruleForm.value.addr,amount:amount,gas:gas,pwd:ruleForm.value.pass,payload:ruleForm.value.comment}};
    // this.$axios(params).then((response)=> {
    //   console.log(response.data);
    //   // thistemp.$checkResultCode(response);
    //   if(thistemp.$checkResultCode(response)){
    //     // var flaginfo = (store.nameinfo == null) ? "创建域名成功" : "修改域名成功";
    //     this.$alert("注册成功", '成功', {
    //       confirmButtonText: '确定',
    //       type: 'success ',
    //       callback: action => {
    //       }
    //     });
    //   }
    // });
  });
}

function resetForm(formName) {
  thistemp.$refs[formName].resetFields();
}

function openaddrs(){
  // console.log("获取焦点")
  // this.drawer = true;
  dialogTableVisible.value = true;
}

function closeaddrs(done) {
  // this.drawer = false;
  dialogTableVisible.value = false;
}

function created(){
  Promise.all([Chain_GetCoinAddress()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    // this.tableData = messageOne;
    // this.tableData = new Array();
    // console.log("获取的地址",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    for(var j = 0,len=messageOne.data.length; j < len; j++) {
      var one = messageOne.data[j]
      // var newOne = {};
      // var bignumber = this.$BigNumber(one.Value);
      // one.Value = bignumber.dividedBy(100000000).toNumber();
      one.Value = new thistemp.$Calculator().divide(one.Value, store.coinCompany);
      //1=见证人;2=社区节点;3=轻节点;4=什么也不是;
      if(one.Type == 1){
        // one.Type = "见证人";
        one.AddrCoin = one.AddrCoin+" 见证人"
        one.disable = true
      }else if(one.Type == 2){
        // one.Type = "社区节点";
        one.AddrCoin = one.AddrCoin+" 社区节点"
        one.disable = true
      }else if(one.Type == 3){
        // one.Type = "轻节点";
        one.AddrCoin = one.AddrCoin+" 轻节点"
        one.disable = true
      }else{
      }
      ruleForm.value.addrs.push({"AddrCoin":one.AddrCoin,"Disable":one.disable});
    }
  });
  return
  // var params = JSON.parse(JSON.stringify(store.rpcParams));
  // params.data = {method:"listaccounts"};
  // this.$axios(params).then((response)=> {
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
  //     // var bignumber = this.$BigNumber(one.Value);
  //     // one.Value = bignumber.dividedBy(100000000).toNumber();
  //     one.Value = new thistemp.$Calculator().divide(one.Value, store.coinCompany);
  //     //1=见证人;2=社区节点;3=轻节点;4=什么也不是;
  //     if(one.Type == 1){
  //       // one.Type = "见证人";
  //     }else if(one.Type == 2){
  //       // one.Type = "社区节点";
  //     }else if(one.Type == 3){
  //       // one.Type = "轻节点";
  //     }else{
  //       // one.Type = "无";
  //       ruleForm.value.addrs.push(one.AddrCoin);
  //       if(j==0){
  //         ruleForm.value.addr = one.AddrCoin;
  //       }
  //     }
  //   }
  //   // console.log(ruleForm.value.addrs)
  //   // this.HighestBlock = response.data.result.HighestBlock;
  //   // this.PulledStates = response.data.result.PulledStates;
  //   // this.$store.commit('setinfo', response.data.result);
  // });
}
created()

</script>