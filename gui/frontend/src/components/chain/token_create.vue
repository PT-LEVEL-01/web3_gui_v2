<template>
<div>
  <el-page-header @back="goBack" content="创建Token">
  </el-page-header>

<el-form :model="ruleForm" status-icon :rules="rules" ref="ruleForm" label-width="100px" class="demo-ruleForm" style="margin-top:40px;">
  <el-form-item label="名称" prop="name">
    <el-input v-model="ruleForm.name" autocomplete="off"></el-input>
  </el-form-item>
  <el-form-item label="单位" prop="symbol">
    <el-input v-model="ruleForm.symbol"></el-input>
  </el-form-item>
  <el-form-item label="总量" prop="supply">
    <el-input v-model="ruleForm.supply" autocomplete="off"></el-input>
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
import { Chain_TokenPublish } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import { store } from '../../store.js'
import {getCurrentInstance, reactive, ref} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this



const ruleForm = ref({
  name:"",//名称
  symbol:"",//单位
  supply:"",//发行总量
  gas:"",
  comment:"",//备注
  pass: ''
})
const rules = ref({
  name: [{required: true, message: 'Token名称不能为空', trigger: 'blur'}],
  symbol: [{required: true, message: 'Token单位不能为空', trigger: 'blur'}],
  supply: [
    {required: true, validator: checkAmountNotZero, trigger: 'blur' }
  ],
  gas: [
    { validator: checkAmountHaveZero, trigger: 'blur' }
  ],
  comment: [{required: false, message: '', trigger: 'blur'}],
  pass: [{required: true, message: '密码不能为空', trigger: 'blur'}]
})


function checkAmountNotZero(rule, value, callback){
  if (value === '') {
    callback(new Error('不能为空'));
    return
  }
  if (value <= 0){
    callback(new Error('不能为0'));
    return
  }
  var a=/^[1-9]*(\.[0-9]{1,8})?$/;
  var  b=/^[0]{1}(\.[0-9]{1,8})?$/;
  var  c=/^[1-9]*(\.[0-9]{1,8})?$/;
  var  d=/^[1-9][0-9]*(\.[0-9]{1,8})?$/;
  var  e=/^\.\d{1,8}?$/;
  if((!a.test(value)&&!b.test(value)&&!c.test(value)&&!d.test(value))||e.test(value)){
    callback(new Error('转账金额只能是数字或小数点后1-2位数字,且前面不要带无效数字0'));
  }else{
    callback();
  }
};
function checkAmountHaveZero(rule, value, callback){
  if (value === '') {
    callback();
    return
  }
  var a=/^[1-9]*(\.[0-9]{1,8})?$/;
  var  b=/^[0]{1}(\.[0-9]{1,8})?$/;
  var  c=/^[1-9]*(\.[0-9]{1,8})?$/;
  var  d=/^[1-9][0-9]*(\.[0-9]{1,8})?$/;
  var  e=/^\.\d{1,8}?$/;
  if((!a.test(value)&&!b.test(value)&&!c.test(value)&&!d.test(value))||e.test(value)){
    callback(new Error('转账金额只能是数字或小数点后1-2位数字,且前面不要带无效数字0'));
  }else{
    callback();
  }
};


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
    var supply = 0;
    if (ruleForm.value.supply > 0){
      supply = ruleForm.value.supply * store.coinCompany;
    }

    Promise.all([Chain_TokenPublish("","",gas,ruleForm.value.pass,ruleForm.value.comment,
        ruleForm.value.name,ruleForm.value.symbol,supply)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      console.log("创建token",messageOne, messageOne[0], messages.length)
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
    }).catch(error => {
      ElMessage({
        showClose: true,
        message: '获取网络地址失败：'+error,
        type: 'error',
      })
    });
    return

    // var params = JSON.parse(JSON.stringify(store.rpcParams));
    // params.data = {method:"tokenpublish",params:{gas:gas,pwd:ruleForm.value.pass,name:ruleForm.value.name,
    //     symbol:ruleForm.value.symbol,supply:supply,owner:"",comment:ruleForm.value.comment}};
    // console.log(params.data);
    // this.$axios(params).then((response)=> {
    //   console.log(response.data);
    //   if(response.data.code == 5008){
    //     //余额不足
    //     this.$alert('可用余额不足', '失败', {
    //       confirmButtonText: '确定',
    //       type: 'error',
    //       callback: action => {
    //         // this.$message({
    //         //   type: 'info',
    //         //   message: `action: ${ action }`
    //         // });
    //       }
    //     });
    //     return
    //   }
    //   if(response.data.code == 2000){
    //     //余额不足
    //     this.$alert('创建Token成功', '成功', {
    //       confirmButtonText: '确定',
    //       type: 'success ',
    //       callback: action => {
    //         // this.$message({
    //         //   type: 'info',
    //         //   message: `action: ${ action }`
    //         // });
    //       }
    //     });
    //     return
    //   }
    // });
  });
}

function resetForm(formName) {
  thistemp.$refs[formName].resetFields();
}

</script>