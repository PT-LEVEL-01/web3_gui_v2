<template>
    <div>
    <el-form :model="ruleForm" status-icon :rules="rules" ref="ruleForm" label-width="100px" class="demo-dynamic" style="margin-top:40px;">
      <el-form-item label="密码" prop="pass">
        <el-input type="password" v-model="ruleForm.pass"></el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="submitForm('ruleForm')">确定</el-button>
        <el-button @click="resetForm('ruleForm')">重置</el-button>
      </el-form-item>
    </el-form>
    <div v-if="keys != ''" >
        <p style="color: red;">助记词未经加密，可以还原出完整钱包，请妥善保管</p>
        <p style="border:red solid 1px;">{{ keys }}</p>
    </div>
    </div>
</template>
<script setup>
import { ElMessage } from 'element-plus'
import { Chain_ExportKey } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, onUnmounted, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const keys = ref("")
const ruleForm = ref({
    pass: '',
  })
const rules = ref({
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
    // console.log(ruleForm.value.address,amount,gas,ruleForm.value.pass+"",ruleForm.value.name,netids,coinaddrs)
    Promise.all([Chain_ExportKey(ruleForm.value.pass+"")]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      console.log("返回结果",messageOne, messageOne[0], messages.length)
      var result = thistemp.$checkResultCode(messageOne.code)
      if(!result.success){
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
      keys.value = messageOne.keys
    }).catch(error => {
      ElMessage({
        showClose: true,
        message: '错误：'+error,
        type: 'error',
      })
    });
  });
}

function resetForm(formName) {
  thistemp.$refs[formName].resetFields();
}

onUnmounted(() => {
  store.setNameinfo(null);
});

</script>