<template>
<div>
    <el-page-header @back="goBack" content="取消见证人资格">
    </el-page-header>

<el-form :model="ruleForm" status-icon :rules="rules" ref="ruleForm" label-width="100px" class="demo-ruleForm"  style="margin-top:40px;">
  <el-form-item label="见证人名称" prop="name">
    {{ruleForm.name}}
  </el-form-item>
  <el-form-item label="见证人地址" prop="addr">
    {{ruleForm.addr}}
  </el-form-item>
  <el-form-item label="押金" prop="amount">
    {{ruleForm.amount}}
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
    <el-button type="primary" @click="submitForm('ruleForm')">立即取消见证人资格</el-button>
    <!-- <el-button @click="resetForm('ruleForm')">重置</el-button> -->
  </el-form-item>
</el-form>
</div>
</template>
<script setup>
import { store } from '../../store.js'
import {getCurrentInstance, reactive, ref} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const ruleForm = ref({
  name:"",
  addr:"",
  amount: "",
  gas:"",
  comment:"",
  pass: ''
})
const rules = ref({
  name: [{ message: '名称不能为空', trigger: 'blur' }],
  amount: [{ validator: thistemp.$checkAmountNotZero, message: '押金不能为空', trigger: 'blur' }],
  gas: [{required: true, validator: thistemp.$checkAmountHaveZero, trigger: 'blur' }],
  comment: [{required: false, message: '', trigger: 'blur'}],
  pass: [{required: true, message: '密码不能为空', trigger: 'blur' }]
})

ruleForm.value.name = store.witnessinfo.Payload;
ruleForm.value.addr = store.witnessinfo.Addr;
ruleForm.value.amount = new thistemp.$Calculator().divide(store.witnessinfo.Value, store.coinCompany);

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
    // var params = JSON.parse(JSON.stringify(store.rpcParams));
    // params.data = {method:"depositout",params:{witness:ruleForm.value.addr,amount:amount,gas:gas,pwd:ruleForm.value.pass}};
    // this.$axios(params).then((response)=> {
    //   console.log(response.data);
    //   // thistemp.$checkResultCode(response);
    //   if(thistemp.$checkResultCode(response)){
    //     // var flaginfo = (store.nameinfo == null) ? "创建域名成功" : "修改域名成功";
    //     this.$alert("取消见证人资格成功", '成功', {
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

</script>