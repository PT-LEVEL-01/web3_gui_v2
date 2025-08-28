<template>
  <el-container style="height: 100%; border:red solid 0px;">
      <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
          <span style="font-size:20px;"><el-button link @click="back()"><el-icon><Back /></el-icon></el-button></span>
          {{ store.circle_select_class }}
          <el-button type="" @click="showPreview()" style="">预览</el-button>
      </el-header>
      
      <el-main class="chat_content" style="text-align: left;border: red solid 0px; padding:0;">
        <div>
          标题:<el-input v-model="store.circle_news_title" placeholder="" style="width:600px;"/>
        </div>
        <div style="border: 1px solid #ccc">
<!--          <Toolbar-->
<!--            style="border-bottom: 1px solid #ccc"-->
<!--            :editor="editorRef"-->
<!--            :defaultConfig="toolbarConfig"-->
<!--            :mode="mode"-->
<!--          />-->
<!--          <Editor-->
<!--            style="height:500px; overflow-y: hidden;text-align: left;"-->
<!--            v-model="valueHtml"-->
<!--            :defaultConfig="editorConfig"-->
<!--            :mode="mode"-->
<!--            @onCreated="handleCreated"-->
<!--          />-->
        </div>
      </el-main>
      <!-- <el-footer></el-footer> -->
  </el-container>
    

</template>

<script>
import { getCurrentInstance, onBeforeUnmount, ref, shallowRef, onMounted } from 'vue'
// import '@wangeditor/editor/dist/css/style.css' // 引入 css
// import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import { store } from '../../store.js'
export default {
  // components: { Editor, Toolbar },
  setup() {
    const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
    // const store = useStore()

    const back = () => {
        window.history.back()
    }

    // 编辑器实例，必须用 shallowRef
    const editorRef = shallowRef()

    // 内容 HTML
    const valueHtml = ref(store.circle_editor_html)

    // 模拟 ajax 异步获取内容
    onMounted(() => {
      // console.log("onMounted")
        // setTimeout(() => {
        //     valueHtml.value = '<p>模拟 Ajax 异步设置内容</p>'
        // }, 1500)
    })

   

    const toolbarConfig = {}
    const editorConfig = { MENU_CONF: {}, placeholder: '请输入内容...' }
    editorConfig.MENU_CONF['uploadImage'] = {
      // 小于该值就插入 base64 格式（而不上传），默认为 0
      base64LimitSize: 50 * 1024 * 1024 // 5kb
    }

    // 组件销毁时，也及时销毁编辑器
    onBeforeUnmount(() => {
        const editor = editorRef.value
        if (editor == null) return
        editor.destroy()
    })

    const handleCreated = (editor) => {
      editorRef.value = editor // 记录 editor 实例，重要！
    }

    const showPreview = () => {
      store.circle_editor_html = valueHtml.value
      thistemp.$router.push({path: '/index/circle/show'});
    }


    return {
      back,
      editorRef,
      valueHtml,
      mode: 'default', // 或 'simple'
      toolbarConfig,
      editorConfig,
      handleCreated,
      showPreview,
    };
  }
}
</script>