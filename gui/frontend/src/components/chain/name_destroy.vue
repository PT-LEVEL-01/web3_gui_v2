<template>
<div>
      <el-page-header @back="goBack" content="注销域名">
    </el-page-header>


<el-form :model="ruleForm" status-icon :rules="rules" ref="ruleForm" label-width="100px" class="demo-dynamic" style="margin-top:40px;">
  <el-form-item label="域名名称" prop="name">
    {{ruleForm.name}}
  </el-form-item>
  <el-form-item label="押金" prop="amount">
    {{ruleForm.amount}}
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
    <el-button type="primary" @click="submitForm('ruleForm')">立即注销</el-button>
    <el-button @click="resetForm('ruleForm')">重置</el-button>
  </el-form-item>
</el-form>
</div>
</template>
<script setup>
import {getCurrentInstance, onUnmounted, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const ruleForm = ref({
  name:store.namedestroy.Name,//名称
  amount:store.namedestroy.Deposit,//押金
  gas:"",
  comment:"",//备注
  pass: ''
})
const rules = ref({
  name: [{ message: '名称不能为空', trigger: 'blur'}],
  amount: [
    { validator: thistemp.$checkAmountNotZero, trigger: 'blur' }
  ],
  gas: [
    {required: true, validator: thistemp.$checkAmountHaveZero, trigger: 'blur' }
  ],
  comment: [{required: false, message: '', trigger: 'blur'}],
  pass: [{required: true, message: '密码不能为空', trigger: 'blur'}]
})

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
      // var big = new thistemp.$Calculator();
      // gas = big.multiply(ruleForm.value.gas, store.coinCompany);
      gas = ruleForm.value.gas * store.coinCompany;
    }

    // var params = JSON.parse(JSON.stringify(store.rpcParams));
    // params.data = {method:"namesout",params:{address:"",gas:gas,pwd:ruleForm.value.pass,name:ruleForm.value.name,comment:ruleForm.value.comment}}
    // console.log(params.data);
    // this.$axios(params).then((response)=> {
    //   console.log(response.data);
    //   if(thistemp.$checkResultCode(response)){
    //     this.$alert('注销成功', '成功', {
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

function removeCoinAddr(item) {
  var index = ruleForm.value.addrcoins.indexOf(item)
  if (index !== -1) {
    ruleForm.value.addrcoins.splice(index, 1)
  }
}

function addCoinAddr() {
  ruleForm.value.addrcoins.push({
    value: '',
    key: Date.now()
  });
}

function removeNetIds(item) {
  var index = ruleForm.value.netids.indexOf(item)
  if (index !== -1) {
    ruleForm.value.netids.splice(index, 1)
  }
}

function addNetIds() {
  ruleForm.value.netids.push({
    value: '',
    key: Date.now()
  });
}

</script>