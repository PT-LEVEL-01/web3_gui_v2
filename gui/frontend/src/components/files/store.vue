<template>
<div>
        <div class="ui menu" style="margin:0;">
            <a id="upload_bt" class="item" @click="addUpdateFile()"><i class="upload icon"></i> 上传 </a>
            <a id="download_bt" class="item"><i class="download icon"></i> 下载 </a>
            <a id="remove_bt" class="item"><i class="trash alternate icon"></i> 删除 </a>
            <div class="right icon menu">
                <a class="item"><i class="th list icon"></i></a>
                <a class="item"><i class="th icon"></i></a>
            </div>
        </div>

    <div id="content" class="ui six doubling cards" style="overflow-y: scroll; overflow-x: hidden; height: calc(100vh - 100px);border:0px red solid;padding:5px;">
        

        <div class="card newcard" style="height:100px;">
            <h2 class="ui icon header">
                <i class="file archive icon"></i>
                <div class="content">
                    <div class="sub header card_filename" data="fhduiafd">nihao</div>
                </div>
            </h2>
        </div>
        
    </div>
</div>
</template>


<script setup>
import { store } from '../../store.js'
import {getCurrentInstance, reactive, ref} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties //vue3获取当前this

const filesAll = ref(null)
const filesShow = ref(null)
const isCollapse = ref(true)
const HighestBlock = ref(0)
const PulledStates = ref(0)

function handleOpen (key, keyPath) {
  console.log(key, keyPath)
}

function handleClose (key, keyPath) {
  console.log(key, keyPath)
}

function addUpdateFile () {
  // var dirpath = ipcRenderer.sendSync('open_file_dialog');
  // console.log(dirpath);
  //
  // var params = JSON.parse(JSON.stringify(store.rpcParams));
  // params.url = "http://localhost:"+store.peerConfig.WebPort+"/store/addfile";
  // var params = new FormData(); // 创建form对象
  // params.append("file", dirpath); // 通过append向form对象添加数据
  // // param.append("chunk", "别的数据"); // 添加form表单中其他数据
  // console.log(params.get("file")); // FormData私有类对象，访问不到，可以通过get判断值是否传进去
  //
  // let config = {
  //   headers: { "Content-Type": "multipart/form-data" }
  // };

  // this.$axios.post("http://localhost:"+store.peerConfig.WebPort+"/store/addfile", params, config);

  // var params = JSON.parse(JSON.stringify(store.rpcParams)); //Object.create(store.rpcParams);
  // params.url = "http://localhost:"+store.peerConfig.WebPort+"/store/addfile";
  // params.filepath = dirpath;
  // // params.data = {method:"addlocalsharefolder",params:{path:dirpath}};
  // this.$axios(params).then((response)=> {
  // // console.log("123");
  //     if(response.data.code != 2000){
  //         return
  //     }
  //     // this.reloadData();
  // });
}

function getHighestBlock(){
  return HighestBlock.value;
}

function created(){
  // var params = JSON.parse(JSON.stringify(store.rpcParams)); //Object.create(store.rpcParams);
  // params.data = {method:"getsharefolderlist"};
  // this.$axios(params).then((response)=> {
  //   // console.log("123");
  //   if(response.data.code != 2000){
  //     return
  //   }
  //   console.log(response.data);
  //   this.filesAll = response.data.result;
  //   this.filesShow = response.data.result;
  //
  //   // this.HighestBlock = response.data.result.HighestBlock;
  //   // this.PulledStates = response.data.result.PulledStates;
  //   // this.$store.commit('setinfo', response.data.result);
  // });
}
created()
</script>
