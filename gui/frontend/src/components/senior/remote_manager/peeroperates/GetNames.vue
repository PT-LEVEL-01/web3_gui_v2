
<template>
  <el-space :size="'large'" wrap>
    <div v-for="(vv, index) in store.names" :key="vv.name">
      <el-button plain @click="open(index)">{{vv.name}}</el-button>
    </div>
  </el-space>

</template>


<script lang="ts" setup>
import { ElMessageBox } from 'element-plus'
// import { useStore} from 'vuex'
import { store } from '../store.js'
import { computed,watch } from 'vue';
// const store = useStore()

const drawer = computed(()=>store.drawer);
// 监控变量 value 的变化,当抽屉关闭时清空内容
watch(drawer, (newValue, oldValue) => {
  if (newValue == false) {
    store.names = []
  }
});

const open = (index: number) => {
  const netIds = store.names[index].netIds
  const addrCoins = store.names[index].addrCoins

 const html = '<p><strong>拥有者:</strong> '+store.names[index].owner+'</p>' +
     '<p><strong>注册区块高度:</strong> '+store.names[index].height+'</p>' +
     '<p><strong>有效块数量:</strong> '+store.names[index].nameOfValidity+'</p>' +
     '<p><strong>冻结金额:</strong> '+store.names[index].deposit+'</p>' +
     '<p><strong>钱包收款地址:</strong> '+addrCoins.join("<br/>")+' </p>'+
     '<p><strong>节点地址:</strong> '+netIds.join("<br/>")+'</p>'

  ElMessageBox.alert(
      html,
      store.names[index].name,
      {
        dangerouslyUseHTMLString: true,
        draggable:true,
        customStyle:{
          "max-width":"45%",
        }
      }
  )
}






</script>
