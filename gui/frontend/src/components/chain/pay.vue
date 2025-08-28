<template>
<div>
      <el-page-header @back="goBack" content="转账">
    </el-page-header>

<el-form :model="ruleForm" status-icon :rules="rules" ref="ruleFormRef" label-width="100px" class="demo-ruleForm"  style="margin-top:40px;">
  <el-form-item v-if="ruleForm.txid != ''" label="Token全称" prop="txid">
    {{ruleForm.name}}
  </el-form-item>
  <el-form-item v-if="ruleForm.txid != ''" label="单位" prop="txid">
    {{ruleForm.symbol}}
  </el-form-item>
  <el-form-item v-if="ruleForm.txid != ''" label="合约地址" prop="txid">
    {{ruleForm.txid}}
  </el-form-item>
  <el-form-item label="地址" prop="addr">
    <el-input v-model="ruleForm.addr" autocomplete="off"></el-input>
  </el-form-item>
  <el-form-item label="金额" prop="amount">
    <el-input v-model="ruleForm.amount" autocomplete="off"></el-input>
  </el-form-item>
  <el-form-item label="手续费" prop="gas">
    <el-input v-model="ruleForm.gas"></el-input>
  </el-form-item>
  <el-form-item label="备注" prop="comment">
    <el-input v-model="ruleForm.comment"></el-input>
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
import { Chain_pay } from '../../../bindings/web3_gui/gui/server_api/sdkapi'

import {getCurrentInstance, onUnmounted, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
const ruleFormRef = ref()
const ruleForm = ref({
  name:"",//token名称全称
  symbol:"",//token单位
  txid:"",//token合约地址
  amount:"",
  addr:"",
  gas:"",
  comment:"",
  pass: ''
})
const rules = ref({
  amount: [{required: true, validator: thistemp.$checkAmountNotZero, message: '名称不能为空', trigger: 'blur' }],
  addr: [{required: true, message: '地址不能为空', trigger: 'blur' }],
  gas: [{required: true, validator: thistemp.$checkAmountHaveZero, trigger: 'blur' }],
  comment: [{required: false, message: '', trigger: 'blur'}],
  pass: [{required: true, message: '密码不能为空', trigger: 'blur' }]
})

function goBack() {
  history.back(-1);
}

function submit(){

}

function submitForm(formName) {
  if (!ruleFormRef.value) return
  ruleFormRef.value.validate((valid, fields) => {
    if (!valid) {
      console.log('error submit!', fields)
      return
    }
  })
  var gas = 0;
  if (ruleForm.value.gas > 0){
    gas = ruleForm.value.gas * store.coinCompany;
  }
  var amount = 0;
  if (ruleForm.value.amount > 0){
    amount = ruleForm.value.amount * store.coinCompany;
  }
  Promise.all([Chain_pay("", ruleForm.value.addr, amount, gas, 0, ruleForm.value.pass,
      ruleForm.value.comment)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("转账结果",messageOne)
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
  })
}

function resetForm(formName) {
  thistemp.$refs[formName].resetFields();
}

onUnmounted(() => {
  store.setPayTokeninfo(null);
});

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