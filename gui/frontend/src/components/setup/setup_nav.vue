<template>
  <div style="height: 100%;">
    <el-container style="height: 100%;">
      <el-aside width="160px">
            <el-menu default-active="1" class="el-menu-vertical-demo" @select="handleSelect">
                <el-menu-item index="1">
                    <span>账号设置</span>
                </el-menu-item>
                <el-menu-item index="2">
                    <span>网络信息</span>
                </el-menu-item>
            </el-menu>
      </el-aside>
      <el-main style="padding:0;height:100%;">
        <router-view/>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import {getCurrentInstance, ref } from "vue";

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

function handleSelect(key, keyPath) {
  // console.log(key, keyPath);
  switch(key){
    case "1":
      thistemp.$router.push({path: '/index/setup/account'});
      break;
    case "2":
      thistemp.$router.push({path: '/index/setup/network'});
      break;
    case "3":
      break;
    case "4":
      break;
    case "5":
      break;
    case "6":
      break;
    case "7":
      break;
    case "8":
      break;
    case "9":
      break;
    default:
  }
}

</script>