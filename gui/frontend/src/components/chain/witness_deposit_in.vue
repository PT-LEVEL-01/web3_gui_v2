<template>
<div>
    

<el-form :model="ruleForm" status-icon :rules="rules" ref="ruleForm" label-width="100px" class="demo-ruleForm"  style="margin-top:40px;">
  <el-form-item label="见证人名称" prop="name">
    <el-input v-model="ruleForm.name" autocomplete="off"></el-input>
  </el-form-item>
  <el-form-item label="押金" prop="amount">
    <el-input v-model="ruleForm.amount" autocomplete="off"></el-input>
  </el-form-item>
  <el-form-item label="手续费" prop="gas">
    <el-input v-model="ruleForm.gas"></el-input>
  </el-form-item>
  <!-- <el-form-item label="备注" prop="comment">
    <el-input v-model="ruleForm.comment"></el-input>
  </el-form-item> -->
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
import { Chain_WitnessDepositIn, Chain_GetWitnessInfo } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import { store } from '../../store.js'
import {getCurrentInstance, reactive, ref} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const ruleForm = ref({
  name:"",
  amount:"",
  gas:"",
  comment:"",
  pass: ''
})
const rules = ref({
  name: [{required: true, message: '名称不能为空', trigger: 'blur' }],
  amount: [{required: true, validator: thistemp.$checkAmountNotZero, message: '押金不能为空', trigger: 'blur' },
    {required: true, validator: checkDepositInAmount, trigger: 'blur' }],
  gas: [{required: true, validator: thistemp.$checkAmountHaveZero, trigger: 'blur' }],
  comment: [{required: false, message: '', trigger: 'blur'}],
  pass: [{required: true, message: '密码不能为空', trigger: 'blur' }]
})



function checkDepositInAmount(rule, value, callback){
  if(value * store.coinCompany < store.chain_getinfo.WitnessNode){
    var big = new thistemp.$Calculator();
    var value = big.divide(store.chain_getinfo.WitnessNode, store.coinCompany);
    callback(new Error('见证人押金不能少于 '+ value));
    return
  }
  callback();
}

function goBack() {
  history.back(-1);
}

function submit(){

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
    // console.log("填写的参数")


    Promise.all([Chain_WitnessDepositIn(amount, gas, ruleForm.value.pass, ruleForm.value.name)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      // console.log("缴纳押金结果",messageOne, messageOne[0], messages.length)

      var result = thistemp.$checkResultCode(messageOne)
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
      // return
      // store.chain_getinfo = messageOne
      // this.$store.commit('setinfo', messageOne);
      // this.$store.commit('setDownloadListProgress', messageOne);
      //查询见证人信息
      // Promise.all([Chain_GetWitnessInfo()]).then(messages => {
      //   if(!messages || !messages[0]){return}
      //   var messageOne = messages[0];
      //   // console.log("开始获取文件下载列表",messageOne)
      //   if(messageOne.IsCandidate){
      //     this.$store.commit('setWitnessinfo', messageOne);
      //     // this.$router.push({path: '/index/wallet/witnessdepositin'});
      //     this.$router.push({path: '/index/wallet/witnessinfo'});
      //     return;
      //   }
      // });
    });
    return
    // var params = JSON.parse(JSON.stringify(store.rpcParams));
    // params.data = {method:"depositin",params:{amount:amount,gas:gas,pwd:ruleForm.value.pass,payload:ruleForm.value.name}};
    // this.$axios(params).then((response)=> {
    //   console.log(response.data);
    //   // thistemp.$checkResultCode(response);
    //   if(thistemp.$checkResultCode(response)){
    //     // var flaginfo = (store.nameinfo == null) ? "创建域名成功" : "修改域名成功";
    //     this.$alert("缴纳押金成功", '成功', {
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

function created(){
  if (store.payTokeninfo == null){
    return;
  }
  ruleForm.value.name = store.payTokeninfo.Name;
  ruleForm.value.symbol = store.payTokeninfo.Symbol;
  ruleForm.value.txid = store.payTokeninfo.TokenId;
}
created()

</script>