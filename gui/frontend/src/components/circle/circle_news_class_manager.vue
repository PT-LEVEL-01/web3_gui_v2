<template>
    <el-container>
        <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
            <span style="font-size:20px;"><el-button link @click="back()"><el-icon><Back /></el-icon></el-button></span>
            <span>类别管理</span>
        </el-header>
        
        <el-main class="chat_content" style="text-align: left;">
          <div>
            <el-input v-model="classNameInput" placeholder="" style="width:200px;"/>
            <el-button type="" @click="SaveRelease()" style="">添加类别</el-button>
          </div>
          <div style="margin-top: 20px;">
              <el-button-group v-for="tag in classNames" @click="showNewsTemp(tag.name)" style="margin:5px 10px;">
                  <el-button type="primary" :icon="ArrowLeft">{{ tag.name }}</el-button>
                  <el-button type="primary"><el-icon><Delete /></el-icon></el-button>
              </el-button-group>
          </div>
        </el-main>
    </el-container>
</template>

<script>
// import '@wangeditor/editor/dist/css/style.css' // 引入 css
import { ElMessage } from 'element-plus'
import { getCurrentInstance, onBeforeUnmount, ref, shallowRef, onMounted } from 'vue'
// import { useStore } from 'vuex'
import { store } from '../../store.js'
import { Circle_FindNewsClass, Circle_SaveNewsClass } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
export default {
//   components: { Editor, Toolbar },
  setup() {
    const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
    // const store = useStore()

    const classNameInput = ref('')

    const classNames = ref([])

    const back = () => {
        window.history.back()
    }
    const SaveRelease = () => {
      Promise.all([Circle_SaveNewsClass(classNameInput.value)]).then(messages => {
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
        flashClassList()

      }).catch(error => {
        ElMessage({
          showClose: true,
          message: '添加类别失败：'+error,
          type: 'error',
        })
      });
    }

    const flashClassList = () => {
      classNames.value = new Array()
      Promise.all([Circle_FindNewsClass()]).then(messages => {
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
        for(var i=0; i<messageOne.Class.length; i++){
          classNames.value.push({name:messageOne.Class[i]})
        }

      }).catch(error => {
        ElMessage({
          showClose: true,
          message: '获取类别列表失败：'+error,
          type: 'error',
        })
      });
    }

    flashClassList()

    return {
        back,
        SaveRelease,
        classNames,
        classNameInput,
    }
  },
  created(){
    // var thistemp = this
    // Promise.all([Circle_FindNewsClass()]).then(messages => {
    //   if(messages){
    //     // console.log(messages)
    //     var messageOne = messages[0];
    //     // console.log(messageOne)
    //     thistemp.friendList = messageOne.UserList
    //   }
    // }).catch(error => {
    //   ElMessage({
    //     showClose: true,
    //     message: '获取添加好友列表失败：'+error,
    //     type: 'error',
    //   })
    // });
  },
}
</script>