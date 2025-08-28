<template>
<div>
      <el-page-header @back="goBack" :model="ruleForm" content="域名信息">
    </el-page-header>

<el-button @click="addCoinAddr">新增钱包地址</el-button>
<el-button @click="addNetIds">新增网络地址</el-button>

<el-form :model="ruleForm" status-icon :rules="rules" ref="ruleForm" label-width="100px" class="demo-dynamic" style="margin-top:40px;">
  <el-form-item label="域名名称" prop="name">
    <el-input v-model="ruleForm.name" autocomplete="off"></el-input>
  </el-form-item>
  <el-form-item label="押金" prop="amount">
    <el-input v-model="ruleForm.amount"></el-input>
  </el-form-item>
  <el-form-item label="拥有者" prop="address">
    <el-input v-model="ruleForm.address" autocomplete="off"></el-input>
  </el-form-item>
  <el-form-item label="手续费" prop="gas">
    <el-input v-model="ruleForm.gas"></el-input>
  </el-form-item>

  <el-form-item v-for="(addrcoin, index) in ruleForm.addrcoins"
    :label="'钱包地址' + index" :key="addrcoin.key" :prop="'addrcoins.' + index + '.value'"
    :rules="{required: false, message: '钱包地址不能为空', trigger: 'blur'}">
    <el-input v-model="addrcoin.value"></el-input><el-button @click.prevent="removeCoinAddr(addrcoin)">删除</el-button>
  </el-form-item>
  
  <el-form-item v-for="(addrcoin, index) in ruleForm.netids"
    :label="'网络地址' + index" :key="addrcoin.key" :prop="'addrcoins.' + index + '.value'"
    :rules="{required: false, message: '网络地址不能为空', trigger: 'blur'}">
    <el-input v-model="addrcoin.value"></el-input><el-button @click.prevent="removeNetIds(addrcoin)">删除</el-button>
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
import { GetNetwork, Chain_NameIn } from '../../../bindings/web3_gui/gui/server_api/sdkapi'

import {getCurrentInstance, onUnmounted, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const ruleForm = ref({
  title:"",
  name:"",//名称
  address:"",//拥有者
  amount:"",//押金
  gas:"",
  netids:[{value:''}],
  addrcoins:[{value:''}],
  comment:"",//备注
  pass: ''
})
const rules = ref({
  name: [{required: true, message: '名称不能为空', trigger: 'blur'}],
      //   symbol: [{required: true, message: 'Token单位不能为空', trigger: 'blur'}],
  amount: [
    {required: true, validator: thistemp.$checkAmountNotZero, trigger: 'blur' }
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
      gas = ruleForm.value.gas * store.coinCompany;
    }
    var amount = 0;
    if (ruleForm.value.amount > 0){
      amount = ruleForm.value.amount * store.coinCompany;
    }
    //   var big = new thistemp.$Calculator();
    //   var amount = big.multiply(ruleForm.value.amount, store.coinCompany);

    var netids = new Array();
    for(var i=0; i<ruleForm.value.netids.length; i++){
      netids.push(ruleForm.value.netids[i].value);
    }
    var coinaddrs = new Array();
    for(var i=0; i<ruleForm.value.addrcoins.length; i++){
      coinaddrs.push(ruleForm.value.addrcoins[i].value);
    }
    // console.log(ruleForm.value.address,amount,gas,ruleForm.value.pass+"",ruleForm.value.name,netids,coinaddrs)
    Promise.all([Chain_NameIn("",ruleForm.value.address,amount,gas,0,ruleForm.value.pass+"",
        "",ruleForm.value.name,netids,coinaddrs)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      console.log("缴纳押金结果",messageOne, messageOne[0], messages.length)
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
      console.log("2222222")
      ElMessage({
        showClose: true,
        message: '注册域名失败：'+error,
        type: 'error',
      })
    });
    // var params = JSON.parse(JSON.stringify(store.rpcParams));
    // params.data = {method:"namesin",params:{address:ruleForm.value.address,amount:amount,
    //   gas:gas,pwd:ruleForm.value.pass,name:ruleForm.value.name,netids:netids,addrcoins:coinaddrs}}
    // console.log(params.data);
    // this.$axios(params).then((response)=> {
    //   console.log(response.data);
    //   // thistemp.$checkResultCode(response);
    //   if(thistemp.$checkResultCode(response)){
    //       var flaginfo = (store.nameinfo == null) ? "创建域名成功" : "修改域名成功";
    //     this.$alert(flaginfo, '成功', {
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
    value: ''
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
    value: ''
  });
}


onUnmounted(() => {
  store.setNameinfo(null);
});


function created(){
  // console.log("2222", store.nameinfo);
  if (store.nameinfo == null){
    ruleForm.value.title = "注册域名";
  }else{
    ruleForm.value.title = "修改域名";
    ruleForm.value.name = store.nameinfo.Name;
    ruleForm.value.address = store.nameinfo.Owner;
    ruleForm.value.amount = store.nameinfo.Deposit;
    ruleForm.value.addrcoins = new Array();
    for (var i=0 ; i<store.nameinfo.AddrCoins.length ; i++){
      var one = store.nameinfo.AddrCoins[i];
      ruleForm.value.addrcoins.push({value: one});
    }
    ruleForm.value.netids = new Array();
    for (var i=0 ; i<store.nameinfo.NetIds.length ; i++){
      var one = store.nameinfo.NetIds[i];
      ruleForm.value.netids.push({value: one});
    }
  }

  //如果netids为空，则自动填入自己节点的网络地址
  // console.log("11111", ruleForm.value.netids.length);
  if(ruleForm.value.netids[0].value == ""){
    Promise.all([GetNetwork()]).then(messages => {
      var messageOne = messages[0];
      ruleForm.value.netids[0].value = messageOne.NetAddr;
    }).catch(error => {
      ElMessage({
        showClose: true,
        message: '获取网络地址失败：'+error,
        type: 'error',
      })
    });
  }
}
created()

</script>