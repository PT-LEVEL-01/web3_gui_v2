<template>
<div>
      <el-page-header @back="goBack" content="添加地址">
    </el-page-header>

<el-form :model="ruleForm" status-icon :rules="rules" ref="ruleFormRef" label-width="100px" class="demo-ruleForm" style="margin-top:40px;">

  <el-form-item label="密码" prop="pass">
    <el-input type="password" v-model="ruleForm.pass"></el-input>
  </el-form-item>
  <el-form-item>
    <el-button type="primary" @click="submitForm('ruleForm')">立即添加</el-button>
  </el-form-item>
</el-form>
</div>
</template>
<script setup>
import { ElMessage } from 'element-plus'
import { Chain_NewCoinAddress } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
const ruleFormRef = ref()
const ruleForm = ref({
  pass: ''
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
  if (!ruleFormRef.value) return
  ruleFormRef.value.validate((valid, fields) => {
    if (!valid) {
      console.log('error submit!', fields)
      return
    } else {
      console.log('submit!')
      Promise.all([Chain_NewCoinAddress(ruleForm.value.pass)]).then(messages => {
        if(!messages || !messages[0]){return}
        var messageOne = messages[0];
        console.log("添加地址 结果",messageOne, messageOne[0], messages.length)
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
      })
    }
  })
}

function resetForm(formName) {
  thistemp.$refs[formName].resetFields();
}
</script>