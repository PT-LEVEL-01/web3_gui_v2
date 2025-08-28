<script setup>
import {getCurrentInstance} from "vue";
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
function gotoStartPriNet(event) {
  thistemp.$router.push({ path: '/senior/login_private_network' });
}
function gotoNodeManager(event) {
  thistemp.$router.push({ path: '/senior/node_manager' });
}
const back = () => {
  window.history.back()
}
</script>

<template>
  <div>
    <el-link v-on:click="back" type="primary">返回</el-link>
  </div>
<!--  <router-link to="/senior/login_private_network">启动私有网络</router-link>-->
  <el-link v-on:click="gotoStartPriNet" type="primary">启动私有网络</el-link>
  | <el-link v-on:click="gotoNodeManager" type="primary">节点管理</el-link>
</template>

<style scoped>

</style>