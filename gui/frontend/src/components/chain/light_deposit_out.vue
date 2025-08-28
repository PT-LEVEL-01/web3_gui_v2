<template>
<div>
      <el-page-header @back="goBack" content="取消轻节点质押">
    </el-page-header>

<el-form :model="ruleForm" status-icon :rules="rules" ref="ruleForm" label-width="100px" class="demo-ruleForm"  style="margin-top:40px;">
  <!-- <el-form-item label="见证人名称" prop="witnessName">
    {{ruleForm.witnessName}}
  </el-form-item> -->
  <!-- <el-form-item label="见证人地址" prop="witnessAddr">
    {{ruleForm.witnessAddr}}
  </el-form-item> -->
  <el-form-item label="质押地址" prop="addr">
    {{ruleForm.addr}}
  </el-form-item>
  <el-form-item label="金额" prop="amount">
    {{ruleForm.amount}}
  </el-form-item>
  <el-form-item label="手续费" prop="gas">
    <el-input v-model="ruleForm.gas"></el-input>
  </el-form-item>
  <el-form-item label="密码" prop="pass">
    <el-input type="password" v-model="ruleForm.pass"></el-input>
  </el-form-item>
  <el-form-item>
    <el-button type="primary" @click="submitForm('ruleForm')">提交</el-button>
    <el-button @click="resetForm('ruleForm')">重置</el-button>
  </el-form-item>
</el-form>
</div>
</template>
<script setup>
import { ElMessage } from 'element-plus'
import { Chain_VoteOut } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, onUnmounted, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const ruleForm = ref({
  witnessAddr: store.lightDepositOut.WitnessAddr,
  witnessName: store.lightDepositOut.Payload,
  amount:store.lightDepositOut.Value,
  addr:store.lightDepositOut.AddrSelf,
  gas:"",
  comment:"",
  pass: ''
})
const rules = ref({
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
    Promise.all([Chain_VoteOut(3,ruleForm.value.addr,amount,gas,0,ruleForm.value.pass,
        ruleForm.value.comment)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      console.log("取消轻节点质押",messageOne)
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
    // //params.data = {method:"votein",params:{votetype:1,address:ruleForm.value.addr, witness:ruleForm.value.witnessAddr,
    // amount:amount,gas:gas,pwd:ruleForm.value.pass,payload:ruleForm.value.comment}};
    // params.data = {method:"voteout",params:{witness:ruleForm.value.witnessAddr,address:ruleForm.value.addr,
    // txid:store.lightDepositOut.Txid,amount:amount,gas:gas,pwd:ruleForm.value.pass}}
    // console.log(params);
    // this.$axios(params).then((response)=> {
    //   console.log(response.data);
    //   // thistemp.$checkResultCode(response);
    //   if(thistemp.$checkResultCode(response)){
    //     // var flaginfo = (store.nameinfo == null) ? "创建域名成功" : "修改域名成功";
    //     this.$alert("取消质押成功", '成功', {
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

onUnmounted(() => {
  store.setLightDepositOut(null);
});

</script>