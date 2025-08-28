<template>
<div class="demo-image__lazy" style="">
    <!-- <el-image v-for="url in urls" :key="url" :src="url" lazy /> -->
    <el-button key="plain" type="" link @click="showEditor()" style="color: white;float:right;"><el-icon><Edit /></el-icon></el-button>
</div>
<div style="margin-top: 20px;">
    <el-button key="plain" type="" @click="showClassManager()" style="">类别管理</el-button>
</div>
<div style="margin-top: 20px;">
    已保存的类别
</div>
<div style="">
    <el-button-group v-for="tag in classNames" @click="showNewsRelease(tag.name)" style="margin:5px 10px;">
        <el-button type="primary" :icon="ArrowLeft">{{ tag.name }}</el-button>
    </el-button-group>
</div>

<div style="margin-top: 20px;">
    等待搜索到的类别...
</div>
<div style="">
    <el-button-group v-for="tag in classNamesMuticast" @click="showNewsTemp(tag.name)" style="margin:5px 10px;">
        <el-button type="primary" :icon="ArrowLeft">{{ classButtomText(tag) }}</el-button>
        <el-button type="primary"><el-icon><Plus /></el-icon></el-button>
    </el-button-group>
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
import backgroundImage from '../../assets/images/background.jpg'
import { ElMessage } from 'element-plus'
import { getCurrentInstance, onBeforeUnmount, ref, shallowRef, onMounted } from 'vue'
// import { useStore } from 'vuex'
import { store } from '../../store.js'
import { Circle_FindNewsClass, Circle_FindClassNamesMulticast } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
// const urls = [
//     'https://fuss10.elemecdn.com/2/11/6535bcfb26e4c79b48ddde44f4b6fjpeg.jpeg',
// ]

export default {
components: { },
setup() {
    const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
    // const store = useStore()
    
    const showNewsRelease = (className) => {
      store.circle_select_class = className
      thistemp.$router.push({path: '/index/circle/newsRelease'});
    }

    const showNewsTemp = (className) => {
      store.circle_select_class = className
      thistemp.$router.push({path: '/index/circle/newsTemp'});
    }

    const showClassManager = () => {
      thistemp.$router.push({path: '/index/circle/classManager'});
    }

    const classNames = ref([])
    const classNamesMuticast = ref([])
    
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
      classNamesMuticast.value = new Array()
      Promise.all([Circle_FindClassNamesMulticast()]).then(messages => {
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
        for(var i=0; i<messageOne.ClassNames.length; i++){
          classNamesMuticast.value.push({name:messageOne.ClassNames[i].Name,count:messageOne.ClassNames[i].Count})
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
    //处理文件传输列表进度刷新
    window.setInterval(()=> {
      flashClassList()
    },5000);

    const classButtomText = (tag) => {
      return tag.name + "(" + tag.count + ")"
    }

    return {
        showNewsTemp,
        showNewsRelease,
        classNames,
        classNamesMuticast,
        showClassManager,
        classButtomText,
    // test,
    };
},
data() {
    return {
    temlrate:0,
    textarea: '',
    }
},
created(){
    

},
computed: {
    count() {
    return this.$store.getters.getCount
    },
},
watch:{
    // 监听值的变化
    test:function(newValue,oldValue){
    },
},
methods:{
    showEditor(){
        // console.log("show editor")
        this.$router.push({path: '/index/circle/editor'});
    },
},
mounted() {
}
};
</script>