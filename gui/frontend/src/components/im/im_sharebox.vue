<template>
    <el-container style="height: 100%; border: 1px solid #eee">
        <el-container>
          <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
            <span style="font-size:20px;"><el-button link @click="back()"><el-icon><Back /></el-icon></el-button></span>
          </el-header>
          
          <el-main class="chat_content">
            <div>
              <!-- <el-button size="small" @click="home()"><el-icon><HomeFilled /></el-icon></el-button> -->
              <el-breadcrumb separator="/">
                <el-breadcrumb-item><el-button @click="home('')" link><el-icon><HomeFilled /></el-icon></el-button></el-breadcrumb-item>
                <el-breadcrumb-item v-for="(item,i) in dirList"><el-button @click="home(item.path)" link>{{ item.name }}</el-button></el-breadcrumb-item>
              </el-breadcrumb>
            </div>
            <el-table :data="list" border highlight-current-row style="width: 100%">
              <el-table-column prop="name" label="名称"/>
              <el-table-column prop="time" label="修改日期"/>
              <el-table-column prop="size" label="大小"/>
              <el-table-column prop="price" label="价格">
                <template #default="scope">
                  <span v-if="scope.row.isPay" style="text-decoration:line-through;">{{scope.row.price}}</span>
                  <span v-if="!scope.row.isPay">{{scope.row.price}}</span>
                </template>
              </el-table-column>
              <el-table-column label="Operations" align="right">
                <template #default="scope">
                  <el-button v-if="!scope.row.isDir&&!scope.row.isPay&&scope.row.price!=0" size="small"
                             @click="getOrder(scope.$index, scope.row)">购买</el-button>
                  <el-button v-if="!scope.row.isDir&&scope.row.price==0||scope.row.isPay" size="small"
                             @click="download(scope.$index, scope.row)"><el-icon><Download /></el-icon></el-button>
                  <el-button v-if="scope.row.isDir" size="small" @click="goin(scope.$index, scope.row)">进入</el-button>
                </template>
              </el-table-column>
            </el-table>
            <!-- <el-button v-for="item in list" @click="back()">{{ item }}</el-button> -->
          </el-main>
        </el-container>
      </el-container>
  <el-drawer v-model="payOrderDrawer" title="支付订单" size="600" :with-header="false">
    <PayOrder :goodsName="goodsName" :orderId="orderId" :price="price" :serverAddr="serverAddr"/>
  </el-drawer>
</template>
    
<style>
  .float_left{
    text-align: left;
  }
  .float_right{
    text-align: right;
  }
  .el-table__header{
    height:0px;
    display:none;
  }
  .el-textarea__inner{
    border:0;
    line-height:50px;
  }

  .el-table .el-table__cell{
    padding-right: 5px;
    /* position: absolute; */
    /* right:0px; */
    /* border:1px solid red; */
    /* text-align: right; */
  }
  ::-webkit-scrollbar{display:none;}
  .chat_content{
    border-top: 1px solid rgb(224, 224, 224);
    height: 200px;
  }
  .el-header{
    height:40px;
  }
</style>
      
<script setup>
import { ElMessage, ElMessageBox  } from 'element-plus'
import { IM_GetShareboxList, File_download, Sharebox_GetFileOrder } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, nextTick, ref, watch} from 'vue'
import PayOrder from '../chain/pay_order.vue'
// import {useStore} from "vuex";
import { store } from '../../store.js'
import {store_routers} from "../../store_routers.js";

const shareboxlist = ref([])
const dirPath = ref("")
const dirList = ref([])
const list = ref([])
const payOrderDrawer = ref(false)
const count = ref(5)

const goodsName = ref("")
const orderId = ref("")
const price = ref(0)
const serverAddr = ref("")

// const store = useStore()
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

//监听订单返回状态
//监听服务器切换按钮开关
watch(
    () => store.chain_payorder_result,
    (newVal, oldVal) => {
      nextTick().then(() => {
        //判断支付结果
        var result = thistemp.$checkResultCode(newVal.Code)
        if(!result.success){
          //支付失败
          ElMessage({
            showClose: true,
            message: "code:"+newVal.Code+" msg:"+result.error,
            type: 'error',
          })
          return
        }
        //支付成功
        payOrderDrawer.value = false
        ElMessage({
          showClose: true,
          message: '支付成功',
          type: 'success',
        })
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

//监听订单支付成功的推送
watch(
    () => store.chain_payorder_client_orderid,
    (newVal, oldVal) => {
      nextTick().then(() => {
        console.log("有订单上链",newVal)
        //设置订单为已经支付
        for(var i=0; i<list.value.length; i++){
          var fileOne = list.value[i]
          if(fileOne.orderId == newVal){
            list.value[i].isPay = true
            break
          }
        }
        console.log("文件列表更改后",list.value)
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

function Init(){
  Promise.all([IM_GetShareboxList(store.im_show_userinfo.netaddr,"")]).then(messages => {
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
    console.log("共享文件夹列表:",messageOne)
    list.value = new Array()
    for(var i=0; i<messageOne.list.List.length; i++){
      var one = messageOne.list.List[i]
      // console.log("文件列表",one)
      var bignumber = thistemp.$BigNumber(one.Price);
      one.Price = bignumber.dividedBy(100000000).toNumber();
      var fileOne = {name: one.Name,hash:one.Hash, time:one.UpdateTime, size:one.Size, isDir:one.IsDir,price:one.Price,isPay:false}
      // console.log("文件列表",one)
      if(fileOne.size ==0 ){fileOne.size = ""}else{fileOne.size = thistemp.$changeSize(fileOne.size)}
      // console.log("文件列表",one)
      list.value.push(fileOne)
    }
  });
}

Init()

function back(){
  // window.history.back()
  store_routers.goback_im()
}

function getOrder(index, file) {
  console.log("获取订单参数",file)
  // return
  Promise.all([Sharebox_GetFileOrder(store.im_show_userinfo.netaddr,file.hash,file.price*100000000)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("获取订单",messageOne)
    var result = thistemp.$checkResultCode(messageOne.Code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.Code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    //给文件添加订单
    for(var i=0; i<list.value.length; i++){
      var fileOne = list.value[i]
      if(fileOne.hash == file.hash){
        list.value[i].orderId = messageOne.Data.data.Number
        break
      }
    }
    //给订单页面设置信息
    goodsName.value = "共享文件:"+messageOne.Data.data.GoodsId
    orderId.value = messageOne.Data.data.Number
    var bignumber = thistemp.$BigNumber(messageOne.Data.data.TotalPrice);
    price.value = bignumber.dividedBy(100000000).toNumber();
    serverAddr.value = messageOne.Data.data.ServerAddr
    payOrderDrawer.value = true
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '添加下载任务失败：'+error,
      type: 'error',
    })
  });
}

function download(index, file) {
  var dirPathTemp = dirPath.value+"/"+file.name
  // console.log("下载文件",file, dirPathTemp)
  // return
  Promise.all([File_download(store.im_show_userinfo.netaddr,dirPathTemp)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: '添加一个下载任务',
        type: 'success',
      })
      return
    }
    // if(messages){
    //   // var messageOne = messages[0];
    //   // console.log(messageOne)
    //   ElMessage({
    //     showClose: true,
    //     message: '添加成功',
    //     type: 'success',
    //   })
    // }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '添加下载任务失败：'+error,
      type: 'error',
    })
  });
}

function goin(index, dir) {
  var dirPathTemp = dirPath.value+"/"+dir.name
  Promise.all([IM_GetShareboxList(store.im_show_userinfo.netaddr,dirPathTemp)]).then(messages => {
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
    // console.log("共享文件夹列表:",messageOne)
    list.value = new Array()
    for(var i=0; i<messageOne.list.List.length; i++){
      var one = messageOne.list.List[i]
      var bignumber = thistemp.$BigNumber(one.Price);
      one.Price = bignumber.dividedBy(100000000).toNumber();
      var fileOne = {name: one.Name, hash:one.Hash, time:one.UpdateTime, size:one.Size, isDir:one.IsDir,price:one.Price,isPay:false}
      if(fileOne.size ==0 ){fileOne.size = ""}else{fileOne.size = thistemp.$changeSize(fileOne.size)}
      list.value.push(fileOne)
    }
    dirPath.value = dirPathTemp
    dirList.value.push({index:dirList.value.length, name:dir.name, path:dirPathTemp})
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '失败：'+error,
      type: 'error',
    })
  });
}

function home(dirPathTemp) {
  // console.log("添加好友",thistemp.friendAddr)
  Promise.all([IM_GetShareboxList(store.im_show_userinfo.netaddr,dirPathTemp)]).then(messages => {
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
    list.value = new Array()
    for(var i=0; i<messageOne.list.List.length; i++){
      var one = messageOne.list.List[i]
      var bignumber = thistemp.$BigNumber(one.Price);
      one.Price = bignumber.dividedBy(100000000).toNumber();
      var fileOne = {name: one.Name, hash:one.Hash, time:one.UpdateTime, size:one.Size, isDir:one.IsDir,price:one.Price,isPay:false}
      if(fileOne.size ==0 ){fileOne.size = ""}else{fileOne.size = thistemp.$changeSize(fileOne.size)}
      list.value.push(fileOne)
    }
    if(dirPathTemp == ""){
      dirPath.value = ""
      dirList.value = new Array()
      return
    }
    dirPath.value = dirPathTemp
    for(var i=0; i<dirList.value.length; i++){
      if(dirList.value[i].path == dirPathTemp){
        dirList.value = dirList.value.slice(0, dirList.value[i].index+1)
        // console.log(dirList.value)
        break
      }
    }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '失败：'+error,
      type: 'error',
    })
  });
}

function showPayPage(){
  payOrderDrawer.value = true
  // ElMessageBox.alert(
  //     '<PayOrder/>',
  //     '支付订单',
  //     {
  //       dangerouslyUseHTMLString: true,
  //     }
  // )
}
</script>