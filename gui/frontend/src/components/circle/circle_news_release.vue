<template>
  <el-container>
      <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
          <span style="font-size:20px;"><el-button link @click="back()"><el-icon><Back /></el-icon></el-button></span>
          {{ store.circle_select_class }}
          <el-button type="" @click="showDraftBox()" style="">草稿箱</el-button>
          <el-button type="" @click="showNewsEditor()" style="">创建</el-button>
      </el-header>
      
      <el-main class="chat_content" style="text-align: left;">
          <div v-for="news in newsList"><el-button @click="EditorNewsOne(news)" link>{{ news.Title }}</el-button></div>
      </el-main>
  </el-container>
</template>

<script>
// import '@wangeditor/editor/dist/css/style.css' // 引入 css
import { ElMessage } from 'element-plus'
import { getCurrentInstance, onBeforeUnmount, ref, shallowRef, onMounted } from 'vue'
// import { useStore } from 'vuex'
import { store } from '../../store.js'
import { Circle_FindNewsRelease } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
export default {
//   components: { Editor, Toolbar },
setup() {
  const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
  // const store = useStore()

  const newsList = ref([])

  const back = () => {
      window.history.back()
  }

  const showDraftBox = () => {
  //   store.state.circle_editor_html = valueHtml.value
    thistemp.$router.push({path: '/index/circle/newsDraft'});
  }

  const EditorNewsOne = (news) => {
    store.circle_news_index_type = 1
    store.circle_news_index = news.Index
    store.circle_news_title = news.Title
    store.circle_editor_html = news.Content
    thistemp.$router.push({path: '/index/circle/editor'});
  }

  const showNewsEditor = () => {
      store.circle_news_index_type = 0
      store.circle_news_index = 0
      store.circle_news_title = ""
      store.circle_editor_html = ""
      thistemp.$router.push({path: '/index/circle/editor'});
  }

  Promise.all([Circle_FindNewsRelease(store.circle_select_class)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    newsList.value = messageOne.News

  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '获取草稿箱失败：'+error,
      type: 'error',
    })
  });

  return {
      back,
      showDraftBox,
      EditorNewsOne,
      showNewsEditor,
      newsList,
  }
}
}
</script>