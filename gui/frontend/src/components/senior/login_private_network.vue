<template>
  <div class="login-container">
    <el-form status-icon ref="ruleForm2" label-position="left" label-width="0px" class="demo-ruleForm login-page">
      <div style="text-align: left;"><el-link v-on:click="back" type="primary">返回</el-link></div>
      <h3 class="title">启动私有网络</h3>
      <el-input v-model="inputValue1" type="password" show-password :placeholder="inputPlaceholder"/>
      <el-input v-if="getKeystoreExit" v-model="inputValue2" type="password" show-password :placeholder="inputPlaceholder2" style="margin-top: 20px;"/>
      <!--      <el-form-item prop="password">-->
      <!--          <el-input type="password" v-model="ruleForm2.password1" placeholder="密码"></el-input>-->
      <!--      </el-form-item>-->
      <!--      <el-form-item prop="password" v-if="getKeystoreExit">-->
      <!--          <el-input type="password" v-model="ruleForm2.password2" auto-complete="off" placeholder="重复输入密码"></el-input>-->
      <!--      </el-form-item>-->
      <div><el-checkbox v-model="checked" label="我明白无法为我找回此密码。" size="large" /></div>
<!--       <el-link v-on:click="back" type="primary">返回</el-link>-->
      <!--      <el-link v-on:click="initNewNetwork" type="primary">启动私有网络</el-link> |-->
      <!--      <el-link v-on:click="importPrivateKey" type="primary">导入密钥</el-link> |-->
      <!--      <el-link v-on:click="importPrivateKey" type="primary">高级</el-link>-->
      <el-form-item style="width:100%;margin-top:10px;">
        <el-button type="primary" style="width:100%;" @click="initNewNetwork" :loading="logining">启动私有网络</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script lang="ts" setup>
import { getCurrentInstance, onBeforeUnmount, ref,reactive, shallowRef, onMounted } from 'vue'
import {SdkApi} from '../../../bindings/web3_gui/gui/server_api'
import {ElMessage} from "element-plus";
// import {useStore} from "vuex";
import { store } from '../../store.js'

// const store = useStore()
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const peerInfo = ref(null)
const logining = ref(false)
const getKeystoreExit = ref(false)
const checked = ref(false)
const inputPlaceholder = ref('password')
const inputValue1 = ref('123456789')
const inputPlaceholder2 = ref('Retype new password')
const inputValue2 = ref('')

// const ruleForm2 = reactive({
//   password1: '123456789',
//   password2: '123456789',
// })
// const rules = reactive({
//   password1: [{ required: true, message: 'please enter your password', trigger: 'blur' }],
//   password2: [{ required: true, message: 'enter your password', trigger: 'blur' }]
// })

function importPrivateKey(event) {
  thistemp.$router.push({ path: '/importprivatekey' });
}

function gotoRemoteConn(event) {
  thistemp.$router.push({ path: '/remoteconn' });
}

const back = () => {
  window.history.back()
}

// function change(event) {
//   console.log(event);
// }

//登录
function handleSubmit(event) {
  startNode(false)
}

//启动私有节点
function initNewNetwork() {
  startNode(true)
}

function startNode(init){
  if(logining.value){
    return
  }
  if(!checked.value){
    ElMessage({
      showClose: true,
      message: "请勾选“我明白无法为我恢复此密码”",
      type: 'error',
    })
    return
  }
  logining.value = true
  //检查密码输入框长度
  if(inputValue1.value===""){
    ElMessage({
      showClose: true,
      message: "Please input a password",
      type: 'error',
    })
    logining.value = false
    return
  }
  if(getKeystoreExit.value){
    //是注册，判断两次密码是否一样
    if(inputValue1.value != inputValue2.value){
      ElMessage({
        showClose: true,
        message: "The two password inputs are inconsistent",
        type: 'error',
      })
      logining.value = false
      return
    }
  }
  Promise.all([SdkApi.IM_StartIm(inputValue1.value,init)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    logining.value = false
    console.log("登录返回", messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if (!result.success) {
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    thistemp.$router.push({ path: '/index' });
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '登录失败：'+error,
      type: 'error',
    })
  });
}

//检查是否有钱包文件
function getWallet() {
  Promise.all([SdkApi.HaveWalletKey()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("是否存在钱包文件", messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if (!result.success) {
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    if(!messageOne.data["have"]){
      inputPlaceholder.value = "New password";
      getKeystoreExit.value = true;
      checked.value = false
    }else{
      checked.value = true
    }
    // this.$router.push({ path: '/index' });
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '登录失败：'+error,
      type: 'error',
    })
  });
}
getWallet()

</script>

<style scoped>
.login-container {
  border: 1px solid #eaeaea;
  width: 100%;
  height: 100%;
}

.login-page {
  -webkit-border-radius: 5px;
  border-radius: 5px;
  margin: 180px auto;
  width: 350px;
  padding: 35px 35px 15px;
  background: #fff;
  border: 1px solid #eaeaea;
  box-shadow: 0 0 25px #cac6c6;
}

label.el-checkbox.rememberme {
  margin: 0px 0px 15px;
  text-align: left;
}</style>