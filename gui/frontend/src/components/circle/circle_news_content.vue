<template>
    <el-container>
        <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
            <span style="font-size:20px;"><el-button link @click="back()"><el-icon><Back /></el-icon></el-button></span>
            {{ store.circle_select_class }}
            <!-- <el-button type="" v-bind:disabled="isShow" @click="SaveDraftBox()" style="">存草稿</el-button>
            <el-button type="" v-bind:disabled="isShow" @click="SaveRelease()" style="">发布</el-button> -->
        </el-header>
        <el-main class="chat_content" style="text-align: left;">
          <div style="font-weight: bold; font-size: large; text-align: center;">
            {{ store.circle_news_title }}
          </div>
          <div>
            <span v-html="store.circle_editor_html"></span>
          </div>
        </el-main>
    </el-container>
</template>

<script>
// import '@wangeditor/editor/dist/css/style.css' // 引入 css
import { ElMessage } from 'element-plus'
import { getCurrentInstance, onBeforeUnmount, ref, shallowRef, onMounted } from 'vue'
// import { useStore } from 'vuex'
import { Circle_SaveNewsDraft, Circle_SaveNewsRelease } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
export default {
//   components: { Editor, Toolbar },
  setup() {
    const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
    // const store = useStore()
    const isShow = ref(false)
    const back = () => {
        window.history.back()
    }
    
    
    return {
        back,
        isShow,
    }
  }
}
</script>