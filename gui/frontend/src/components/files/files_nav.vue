<template>
  <div class="common-layout" style="height: 100%;">
    <el-container style="height: 100%;">
      <el-aside width="160px">
        <el-menu default-active="1" class="el-menu-vertical-demo" @select="handleSelect">
          <el-menu-item index="1">
            <span>共享文件</span>
          </el-menu-item>
          <el-menu-item index="2">
            <span>云服务器</span>
          </el-menu-item>
          <el-menu-item index="3">
            <span>云客户端</span>
          </el-menu-item>
          <el-menu-item index="4">
            <span>下载中</span>
          </el-menu-item>
          <el-menu-item index="5">
            <span>下载完成</span>
          </el-menu-item>
          <el-menu-item index="6">
            <span>上传中</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      <el-main style="padding: 0;">
        <router-view/>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import {
  Document,
  Menu as IconMenu,
  Location,
  Setting,
} from '@element-plus/icons-vue'
import {getCurrentInstance,onMounted } from 'vue'
// const instance  = getCurrentInstance()
const thistemp = getCurrentInstance().appContext.config.globalProperties //vue3获取当前this

const handleSelect = (key, keyPath) => {
  // console.log(key, keyPath)
  switch(key){
    case "1":
      thistemp.$router.push({path: '/index/files/sharebox'});
      break;
    case "2":
      thistemp.$router.push({path: '/index/files/storage_server'});
      break;
    case "3":
      thistemp.$router.push({path: '/index/files/storage_client'});
      break;
    case "4":
      thistemp.$router.push({path: '/index/files/download'});
      break;
    case "5":
      thistemp.$router.push({path: '/index/files/download_finish'});
      break;
    case "6":
      thistemp.$router.push({path: '/index/files/upload_list'});
      break;
    default:
  }
}

</script>