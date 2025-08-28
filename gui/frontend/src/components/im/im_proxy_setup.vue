<template>
  <div style="text-align: left;border:red solid 0px;">
    <el-page-header @back="back" content="设置服务器">
    </el-page-header>
    <el-form label-width="120px">
      <el-form-item label="昵称">
        <input type="text" v-model = "nickname" style="border:0;" @blur="updateNickname"/>
      </el-form-item>
      <el-form-item label="代理节点状态">
        <el-switch inline-prompt active-text="开" inactive-text="关" v-model="storageServerSwitch"/>
      </el-form-item>
      <el-form-item label="可以出售的空间">
        <el-input-number v-model="Selling" @change="updateNickname" :min="SellingMin" :max="SellingMax" label="描述文字"></el-input-number>G
      </el-form-item>
      <el-form-item label="单价">
        <input type="text" v-model = "PriceUnit" style="border:0;" @blur="updateNickname"/>TEST/1G/天
      </el-form-item>
      <el-form-item label="最长租用时间">
        <input type="text" v-model = "UseTimeMax" style="border:0;" @blur="updateNickname"/>天
      </el-form-item>
    </el-form>

  </div>
</template>


<script>
import { ElMessage } from 'element-plus'
import { IMProxyServer_SetProxyInfo, IMProxyServer_GetProxyInfo, IMProxyServer_SetOpen } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, reactive, ref, watch} from 'vue'
import {store_routers} from "../../store_routers.js";
export default {
  components: { },
  setup() {
    const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
    const count = ref(3)
    const shareboxlist = ref([])
    const storageServerSwitch = ref(false)
    const nickname = ref("")
    const oldnickname = ref("")
    const serverInfo = ref(null)

    const oldPriceUnit = ref(0)//单价
    const PriceUnit = ref(0)//单价
    const oldSelling = ref(0)//可以出售的空间
    const Selling = ref(0)//可以出售的空间
    const SellingMin = ref(0)//最小可以出售的空间
    const SellingMax = ref(0)//最大可以出售的空间
    const oldUseTimeMax = ref(0)//最大可以租用空间的时间
    const UseTimeMax = ref(0)//最大可以租用空间的时间

    const back = () => {
      // window.history.back()
      store_routers.goback_im()
    }

    const getstatus = () => {
      Promise.all([IMProxyServer_GetProxyInfo()]).then(messages => {
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
        if(messageOne.info.PriceUnit != 0){
          var bignumber = thistemp.$BigNumber(messageOne.info.PriceUnit);
          messageOne.info.PriceUnit = bignumber.dividedBy(100000000).toNumber();
        }
        serverInfo.value = messageOne.info

        console.log("存储服务器信息",messageOne.info)
        storageServerSwitch.value = messageOne.info.IsOpen
        nickname.value = messageOne.info.Nickname
        oldnickname.value = messageOne.info.Nickname
        oldPriceUnit.value = messageOne.info.PriceUnit
        PriceUnit.value = messageOne.info.PriceUnit
        oldSelling.value = messageOne.info.Selling
        Selling.value = messageOne.info.Selling
        oldUseTimeMax.value = messageOne.info.UseTimeMax
        UseTimeMax.value = messageOne.info.UseTimeMax
        SellingMax.value = 999999
        // var dirs = messageOne.storageServerInfo.Directory
        // shareboxlist.value = new Array()
        // for(var i=0; dirs !=null && i<dirs.length; i++){
        //   var size = messageOne.storageServerInfo.DirectoryFreeSize[i]
        //   var one = {size:size+" G",path:dirs[i]}
        //   SellingMax.value += size
        //   shareboxlist.value.push(one)
        // }
      });
    }
    getstatus()

    const updateNickname = () => {
      var haveChange = false
      if(oldnickname.value != nickname.value){
        haveChange = true
      }
      if(oldPriceUnit.value != PriceUnit.value){
        haveChange = true
      }
      if(oldSelling.value != Selling.value){
        haveChange = true
      }
      if(oldUseTimeMax.value != UseTimeMax.value){
        haveChange = true
      }
      if(!haveChange){
        return
      }
      var priceUnit = 0
      priceUnit = parseFloat(PriceUnit.value) * 100000000
      Promise.all([IMProxyServer_SetProxyInfo(nickname.value, priceUnit,Selling.value,Selling.value,
          Selling.value, parseInt(UseTimeMax.value),serverInfo.value.RenewalTime)]).then(messages => {
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
        ElMessage({
          showClose: true,
          message: '成功',
          type: 'success',
        })
        oldnickname.value = nickname.value
      }).catch(error => {
        ElMessage({
          showClose: true,
          message: '修改失败：'+error,
          type: 'error',
        })
      });
    }

    //监听服务器切换按钮开关
    watch(
        () => storageServerSwitch.value,
        (newVal, oldVal) => {
          console.log(newVal,oldVal)
          if(newVal==undefined){return}
          Promise.all([IMProxyServer_SetOpen(newVal)]).then(messages => {
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
            // storageServerSwitch.value = messageOne.storageServerInfo.IsOpen
          });
        },
        {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
    );

    return {
      back,
      nickname,
      PriceUnit,
      Selling,
      SellingMin,
      SellingMax,
      UseTimeMax,
      updateNickname,
      storageServerSwitch,
      count,
      shareboxlist,
    };
  },
  data() {
    return {
      // shareboxlist:this.shareboxlist,
    }
  },
  created(){
  },
}
</script>
<style scoped>
.scrollbar-demo-item {
  /* display: flex; */
  /* align-items: center; */
  /* justify-content: left; */
  height: 50px;
  margin: 10px 0;
  padding:10px;
  /* text-align: left; */
  border-radius: 4px;
  border:rgb(238, 236, 236) solid 1px;
  /* background: var(--el-color-primary-light-9); */
  /* color: var(--el-color-primary); */
}
</style>
