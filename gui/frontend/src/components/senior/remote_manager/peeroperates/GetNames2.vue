

<template>
  <el-divider direction="vertical"></el-divider>
  <el-button type="text" @click="openNameInfo">sdfsa Message Box</el-button>
</template>

<script setup>
import {getCurrentInstance, ref} from "vue";
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

function openNameInfo() {
  console.log()
  thistemp.$alert('<p><strong>Name:</strong> John Doe</p><p><strong>Email:</strong> john@example.com</p><p><strong>Message:' +
      '</strong> This is a long message that will demonstrate how text automatically wraps in an HTML paragraph element ' +
      'when it exceeds the width of the container. The text will automatically break to a new line when necessary.</p>', 'HTML 鐗囨', {
    dangerouslyUseHTMLString: true
  });
}



</script>