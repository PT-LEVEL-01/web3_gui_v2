<template>
  <div class="about">
    <el-page-header @back="goBack" content="添加存储空间">
    </el-page-header>
    
    <el-form :model="numberValidateForm" ref="numberValidateForm" label-width="100px" style="margin-top:20px;" class="demo-ruleForm">
      
      <el-form-item label="存储路径" prop="spacesPath">
          <el-input v-model="numberValidateForm.spacesPath"></el-input>
          <el-button type="primary" @click="addShareDir()">选择路径</el-button>
      </el-form-item>
      <el-form-item label="数量" prop="spacesNum" :rules="[{ required: true, message: '数量不能为空'},{ type: 'number', message: '数量必须为数字值'}]">
        <el-input type="spacesNum" v-model.number="numberValidateForm.spacesNum" autocomplete="off"></el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="submitForm('numberValidateForm')">提交</el-button>
        <el-button @click="resetForm('numberValidateForm')">重置</el-button>
      </el-form-item>
    </el-form>

  </div>
</template>


<script setup>
import {getCurrentInstance, ref, onMounted, onUnmounted } from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties //vue3获取当前this

const numberValidateForm = ref({
  spacesPath: "",
  spacesNum: 1,
  age: ''
})

function goBack() {
  history.back(-1);
}

function addShareDir () {
  var dirpath = ipcRenderer.sendSync('open_directory_dialog');
  // console.log(dirpath);
  numberValidateForm.value.spacesPath = dirpath;
}

function submitForm(formName) {
  thistemp.$refs[formName].validate((valid) => {
    if (valid) {
      // alert('submit!');
      // var params = JSON.parse(JSON.stringify(store.rpcParams)); //Object.create(store.rpcParams);
      // params.data = {method:"addminingspacesize", params:{n: this.numberValidateForm.spacesNum, absPath: this.numberValidateForm.spacesPath,}};
      // this.$axios(params).then((response)=> {
      //   if(thistemp.$checkResultCode(response)){
      //     // var flaginfo = (store.nameinfo == null) ? "创建域名成功" : "修改域名成功";
      //     this.$alert("成功", '成功', {
      //       confirmButtonText: '确定',
      //       type: 'success ',
      //       callback: action => {
      //       }
      //     });
      //   }
      // });
    } else {
      console.log('error submit!!');
      return false;
    }
  });
}

function resetForm(formName) {
  thistemp.$refs[formName].resetFields();
}

</script>