<template>
    <div style="text-align: left;">
        <div class="demo-image__lazy" style="">
            <!-- <el-image v-for="url in urls" :key="url" :src="url" lazy /> -->
            <!-- <el-button key="plain" type="" link @click="showEditor()" style="color: white;float:right;"><el-icon><Edit /></el-icon></el-button> -->
        </div>
        <div>
            <span style="font-size:20px;"><el-button link @click="back()"><el-icon><Back /></el-icon></el-button></span>
            <el-button type="" @click="showDraftBox()" style="">草稿箱</el-button>
            <el-button type="" @click="showReleaseBox()" style="">我发布的</el-button>
            <el-button type="" @click="showNewsEditor()" style="">创建</el-button>
        </div>
        <div style="margin-top: 20px;">
            <span style="font-size: 30px;">{{ store.circle_select_class }}</span>
        </div>
        <div>
            <div v-for="news in newsList"><el-button @click="ShowNewsOne(news)" link>{{ news.Title }}</el-button></div>
            <!-- <div>一套餐椅</div>
            <div>一套餐椅</div>
            <div>一套餐椅</div> -->
        </div>
        
        <!-- <el-container>
            <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
                <span style="font-size:20px;"><el-button link @click="back()"><el-icon><Back /></el-icon></el-button></span>
                <el-button type="" @click="SaveDraftBox()" style="">存草稿</el-button>
                <el-button type="" @click="SaveRelease()" style="">发布</el-button>
            </el-header>
            
            <el-main class="chat_content" style="text-align: left;">
                <span v-html="store.circle_editor_html"></span>
            </el-main>
        </el-container> -->
    </div>
</template>

<style scoped>
.demo-image__lazy {
    height: 200px;
    overflow-y: hidden;
    border:red solid 0px;
    margin: 0;
    background:url('../../assets/images/background.jpg');
    background-position: center;
}
.demo-image__lazy .el-image {
    display: block;
    min-height: 200px;
    margin-bottom: 10px;
}
.demo-image__lazy .el-image:last-child {
    margin-bottom: 0;
}
</style>

<script>
// import '@wangeditor/editor/dist/css/style.css' // 引入 css
import { getCurrentInstance, onBeforeUnmount, ref, shallowRef, onMounted } from 'vue'
// import { useStore } from 'vuex'
import { store } from '../../store.js'
import { Circle_FindClassNamesMulticastNewsList } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {ElMessage} from "element-plus";
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
    const showReleaseBox = () => {
    //   store.state.circle_editor_html = valueHtml.value
      thistemp.$router.push({path: '/index/circle/newsRelease'});
    }

    const ShowNewsOne = (news) => {
        // store.state.circle_news_index_type = 0
        // store.state.circle_news_index = 0
        store.circle_news_title = news.Title
        store.circle_editor_html = news.Content
        thistemp.$router.push({path: '/index/circle/newsContent'});
    }
    
    const showNewsEditor = () => {
        store.circle_news_index_type = 0
        store.circle_news_index = 0
        store.circle_news_title = ""
        store.circle_editor_html = ""
      thistemp.$router.push({path: '/index/circle/editor'});
    }

    Promise.all([Circle_FindClassNamesMulticastNewsList(store.circle_select_class)]).then(messages => {
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
      // console.log("获取广播的云存储列表",messageOne)
      newsList.value = messageOne.ClassNames

    }).catch(error => {
        ElMessage({
            showClose: true,
            message: '获取类别列表失败：'+error,
            type: 'error',
        })
    });
    return {
        back,
        newsList,
        showDraftBox,
        showReleaseBox,
        showNewsEditor,
        ShowNewsOne,
    }
  }
}
</script>