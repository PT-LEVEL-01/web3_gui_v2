<template>
  <div style="border:red solid 0px;">
    <el-page-header @back="back" content="订单列表">
    </el-page-header>
    <el-scrollbar>
      <el-table :data="shareboxlist" style="width: 100%">
        <el-table-column type="index" :index="indexMethod" />
        <el-table-column prop="Nickname" label="" width="180" />
        <el-table-column prop="SpaceTotal" label="" width="180" />
        <el-table-column prop="CreateTime" label="" width="180" />
        <el-table-column label="">
          <template #default="scope">
            <el-button v-if="scope.row.Status==1" @click="storageServerPage(scope.row)" style="float: right;">支付</el-button>
            <p v-if="scope.row.Status==2">等待付款上链中</p>
            <p v-if="scope.row.Status==3">支付成功</p>
            <p v-if="scope.row.Status==4&&scope.row.ChainTx==''">过期未支付</p>
            <p v-if="scope.row.Status==4&&scope.row.ChainTx!=''">服务到期</p>
          </template>
        </el-table-column>
      </el-table>
    </el-scrollbar>
  </div>
  <el-drawer v-model="payOrderDrawer" title="支付订单" size="600" :with-header="false">
    <PayOrder :goodsName="goodsName" :orderId="orderId" :price="price" :serverAddr="serverAddr"/>
  </el-drawer>
</template>

<script setup>
import { ElMessage } from 'element-plus'
// element-plus 集成的 dayjs 默认也安装了 dayjs 插件，所以相关插件可以直接使用
import { dayjs } from 'element-plus'
import { File_GetShareboxList, File_OpenDirectoryDialog, File_AddSharebox, File_DelSharebox, ImProxyClient_GetOrderList,
  ImProxyClient_SetOrderWaitOnChain } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, nextTick, reactive, ref, watch} from 'vue'
// import {useStore} from "vuex";
import { store } from '../../store.js'
import {store_routers} from "../../store_routers.js";
import PayOrder from "../chain/pay_order.vue";

const count = ref(3)
const shareboxlist = ref([])
const showGetStartPage = ref(true) //是否显示开始引导页面
// const store = useStore()
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const back = () => {
  // window.history.back()
  store_routers.goback_im()
}

const payOrderDrawer = ref(false)
const goodsName = ref("")
const orderId = ref("")
const price = ref(0)
const serverAddr = ref("")

//监听订单返回状态
//监听服务器切换按钮开关
watch(
    () => store.chain_payorder_result,
    (newVal, oldVal) => {
      nextTick().then(() => {
        console.log("支付订单返回",newVal)
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
        //修改订单状态为，已支付，但未上链
        Promise.all([ImProxyClient_SetOrderWaitOnChain(newVal.Data.orderId16,newVal.Data.data.lock_height)]).then(messages => {
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
        });
        //支付成功
        for(var i=0; i<shareboxlist.value.length; i++){
          var fileOne = shareboxlist.value[i]
          if(fileOne.Number == newVal.Data.orderId16){
            shareboxlist.value[i].Status = 2
            break
          }
        }
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
        for(var i=0; i<shareboxlist.value.length; i++){
          var fileOne = shareboxlist.value[i]
          if(fileOne.Number == newVal){
            shareboxlist.value[i].Status = 3
            break
          }
        }
        // console.log("文件列表更改后",list.value)
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

const storageServerPage = (row) => {
  //给订单页面设置信息
  goodsName.value = "购买IM存储空间"
  orderId.value = row.Number
  var bignumber = thistemp.$BigNumber(row.TotalPrice);
  price.value = bignumber.dividedBy(100000000).toNumber();
  // price.value = 2
  serverAddr.value = row.ServerAddr
  payOrderDrawer.value = true
  // thistemp.$store.state.storage_client_selectServerInfo = row
  // thistemp.$router.push({path: '/index/files/storage_server_info'});
}

//获取订单列表
const getOrdersList = () => {
  Promise.all([ImProxyClient_GetOrderList()]).then(messages => {
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
    console.log("获取订单列表",messageOne)
    shareboxlist.value = new Array()
    if(messageOne.list == null){
      return
    }
    for(var key in messageOne.list){
      var one = messageOne.list[key]
      const createTime = dayjs.unix(one.CreateTime)
      one.CreateTime = createTime.format('YYYY-MM-DD HH:mm:ss')
      // var date = new Date(one.CreateTime * 1000);
      // one.CreateTime = date.toISOString()
      var date = new Date(one.TimeOut * 1000);
      one.TimeOut = date.toISOString()
      shareboxlist.value.push(one)
    }
    // console.log("列表",shareboxlist.value)
    if(messageOne.list.length > 0){
      showGetStartPage.value = false
      return
    }
  });
}
getOrdersList()

const add = () => {
  // var thistemp = this
  //打开选择目录对话框
  Promise.all([File_OpenDirectoryDialog()]).then(messages => {
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
    console.log("选中的文件路径",messageOne.path)
    if(messageOne.path.length == 3 && messageOne.path.match(/[A-Z]:\\/) != null){
      ElMessage({
        showClose: true,
        message: '不能选择根目录',
        type: 'error',
      })
      return
    }
    // for(var i=0; i<messageOne.length; i++){
    //   var filePath = messageOne[i]
    //   if (editor == null) return;
    //   editor.dangerouslyInsertHtml('<a class="sendFile" href="" target="_blank">'+filePath+'</a><br>')
    // }
    //添加目录
    Promise.all([File_AddSharebox(messageOne.path)]).then(messages => {
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
      //刷新共享目录列表
      Promise.all([File_GetShareboxList()]).then(messages => {
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
        console.log("刷新共享目录列表",messageOne.list)
        shareboxlist.value = new Array()
        for(var key in messageOne.list){
          shareboxlist.value.push(messageOne.list[key])
        }
      });
    });
  });
  count.value++
}
const onDelete = (dirPath) => {
  //删除共享目录列表
  Promise.all([File_DelSharebox(dirPath)]).then(messages => {
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
    // console.log("删除共享目录列表",messageOne)
    //刷新共享目录列表
    Promise.all([File_GetShareboxList()]).then(messages => {
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
      console.log("刷新共享目录列表",messageOne.list)
      shareboxlist.value = new Array()
      for(var key in messageOne.list){
        shareboxlist.value.push(messageOne.list[key])
      }
    });
  });
  return
  if (count.value > 0) {
    count.value--
  }
}

</script>
