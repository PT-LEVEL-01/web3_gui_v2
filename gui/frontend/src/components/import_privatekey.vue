<template>
    <div class="login-container">
        <el-form :model="ruleForm2" :rules="rules2" status-icon ref="ruleForm2" label-position="left" label-width="0px" class="demo-ruleForm login-page">
            <h3 class="title">导入私钥</h3>
            <!-- <el-form-item prop="username">
                <el-input type="text" v-model="ruleForm2.username" auto-complete="off" placeholder="用户名"></el-input>
            </el-form-item> -->
            <el-form-item prop="desc" style="height:100px;">
                <el-input type="textarea" v-model="ruleForm2.seed"  auto-complete="off" placeholder="密钥种子" style="height:100px;"></el-input>
            </el-form-item>
            <el-form-item prop="password">
                <el-input type="password" v-model="ruleForm2.password" auto-complete="off" placeholder="密码"></el-input>
            </el-form-item>
            <el-form-item prop="password" v-if="getKeystoreExit">
                <el-input type="password" v-model="ruleForm2.password" auto-complete="off" placeholder="重复输入密码"></el-input>
            </el-form-item>
            <!-- <el-checkbox v-model="checked" class="rememberme">记住密码</el-checkbox> -->
            <el-link v-on:click="goBack" type="primary">返回</el-link>
            <el-form-item style="width:100%;">
                <el-button type="primary" style="width:100%;" @click="handleSubmit" :loading="logining">立即导入</el-button>
            </el-form-item>
            导入会覆盖原密钥文件，造成原密钥丢失，请备份原密钥。
        </el-form>
    </div>
</template>

<script>
// const ipc = require('electron').ipcRenderer;
// const { ipcRenderer } = require('electron')
// console.log(peerInfo);

import { ElMessage } from 'element-plus'
import { Chain_ImportKey } from '../../bindings/web3_gui/gui/server_api/sdkapi'
export default {
    data() {
        return {
            peerInfo: null,
            logining: false,
            ruleForm2: {
                seed:"",
                username: 'admin',
                password: '123456789',
            },
            rules2: {
                username: [{required: true, message: 'please enter your account', trigger: 'blur'}],
                password: [{required: true, message: 'enter your password', trigger: 'blur'}]
            },
            checked: false
        }
    },
    methods: {
        goBack() {
            history.back(-1);
        },
        initNewNetwork: function (event) {
            // this.$store.commit('setIsRpc', false);
            var params = {init:"init",pwd: this.ruleForm2.password};
            console.log("init", params);
            // ipcRenderer.send('send_password', params);
            // var peerConfig = ipcRenderer.sendSync('send_config');
            // this.$store.commit('setPeerConfig', peerConfig);
            // this.$router.push({path: '/index'});
        },
        gotoRemoteConn: function (event) {
            this.$router.push({path: '/remoteconn'});
        },
        handleSubmit(event){
            console.log(this.ruleForm2)
            Promise.all([Chain_ImportKey(this.ruleForm2.seed,this.ruleForm2.password)]).then(messages => {
                if(!messages || !messages[0]){return}
                var messageOne = messages[0];
                // console.log("见证人列表",messageOne)
                var result = thistemp.$checkResultCode(messageOne.code)
                if(!result.success){
                    ElMessage({
                        showClose: true,
                        message: "code:"+messageOne.code+" msg:"+result.error,
                        type: 'error',
                    })
                    return
                }
            });
            history.back(-1);
            // var params = {pwd:this.ruleForm2.password, seed:this.ruleForm2.seed}
            // var code = ipcRenderer.sendSync('import_privatekey', params);
            // console.log(code);
            // if(code == 0){
            //     //导入成功
            //     this.$alert(response.data.code, '导入成功', {
            //         confirmButtonText: '确定',
            //         type: 'error',
            //         callback: action => {
            //         }
            //     });
            // }else{
            //     //导入失败
            //     this.$alert(response.data.code, '导入失败', {
            //         confirmButtonText: '确定',
            //         type: 'error',
            //         callback: action => {
            //         }
            //     });
            // }
            // return
        }
    },
    computed: {
        // 计算属性的 getter
        getKeystoreExit() {
            // var peerInfo = ipcRenderer.sendSync('get_peerInfo', 'ping');
            // // console.log(peerInfo.haveKeystory);
            // return !(peerInfo.haveKeystory);
        }
    },
    created(){
        // var peerInfo = ipcRenderer.sendSync('get_peerInfo', 'ping');
        // console.log(peerInfo);
        // this.$store.commit('setPeerInfo', peerInfo);
        // this.data.peerInfo = peerInfo;
    }
};
</script>

<style scoped>
.login-container {
    width: 100%;
    height: 100%;
    border:#fff solid 1px;
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
}
</style>