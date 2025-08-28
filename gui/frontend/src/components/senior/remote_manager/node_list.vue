<script setup>
import { store } from '../../../store.js'
import PeerList from './PeerList.vue'
import Peer from './Peer.vue'
//import Test from './components/Test.vue'
const back = () => {
  window.history.back()
}
</script>

<template>
  <div><el-button @click="back()">返回</el-button></div>


  <!--    <Test/>-->
  <PeerList/>
  <el-drawer v-model="store.drawer" title="节点详情管理" :with-header="false" size="55%">
    <Peer/>
  </el-drawer>
</template>