<script setup>
import login from './components/login.vue'
import nav_left from './components/nav_left.vue'
import {getCurrentInstance, watch, nextTick, onMounted, ref, computed} from "vue";
import {store_routers} from "./store_routers.js";
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const routes = {
  'login': login,
  'index': nav_left
}
// const currentPath = ref("login")
const currentView = computed(() => {
  return routes[store_routers.currentPageKey_root]
})

// thistemp.$router.push({ path: '/' });
</script>

<template>
  <div style="width: 100%;height: 100%;">
    <component :is="currentView" />
<!--    <router-view style="width: 100%;height: 100%;border:red solid 0px;"></router-view>-->
  </div>
</template>

<style scoped>
</style>
