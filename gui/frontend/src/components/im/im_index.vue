<template>
  <el-container style="height: 100%; border: 1px solid #eee">
    <el-container>
      <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;">
        <el-button type="primary" @click="showCreateGroupPage()">创建群聊</el-button>
        <el-button type="primary" @click="showProxyServer()">成为代理节点</el-button>
      </el-header>

      <el-main class="chat_content">
        <el-table ref="singleTable" :data="tableData" style="width: 100%">
          <el-table-column property="Nickname" label="" width="120">
          </el-table-column>
          <el-table-column label="" width="450">
            <template #default="scope">
              <el-button size="small" @click="handleEdit(scope.row)">修改</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-main>
    </el-container>
  </el-container>

</template>

<style>
</style>

<script setup>
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { IM_SearchFriendInfo, IM_AddFriend, ImProxyClient_GetCreateGroupList } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
// import {useStore} from "vuex";
import { store } from '../../store.js'
import {getCurrentInstance, ref} from "vue";
import {store_routers} from "../../store_routers.js";

// const store = useStore()
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const tableData = ref([])
const netaddr = ref("")
const nickname = ref("添加好友")
const friendinfo = ref(null)

//去创建群聊页面
const showCreateGroupPage = () => {
  store.im_groupinfo = null
  // thistemp.$router.push({path: '/index/im/createGroup'});
  store_routers.gopage_im("createGroup")
  return
}
//去设置服务器页面
const showProxyServer = () => {
  // store.im_groupinfo = null
  // thistemp.$router.push({path: '/index/im/imProxySetup'});
  store_routers.gopage_im("imProxySetup")
  return
}

const handleEdit = (groupInfo) => {
  store.im_groupinfo = groupInfo
  // thistemp.$router.push({path: '/index/im/createGroup'});
  store_routers.gopage_im("createGroup")
  return
}

//获取自己创建的群列表
const getGroupList = () => {
  Promise.all([ImProxyClient_GetCreateGroupList()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("创建的群列表",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    tableData.value = messageOne.list
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '添加失败：'+error,
      type: 'error',
    })
  });
}
getGroupList()

const back = () => {
  store_routers.goback_im()
}


function submitSearch() {
  // console.log("search");
  if(this.netaddr == ""){return}
  Promise.all([IM_SearchFriendInfo(this.netaddr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("用户信息",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    this.friendinfo = messageOne.info

  });
}

function addFriend() {
  // console.log("添加好友",this.friendAddr)
  Promise.all([IM_AddFriend(this.friendinfo.Addr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("用户信息",messageOne)
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
      message: '添加成功',
      type: 'success',
    })

  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '添加失败：'+error,
      type: 'error',
    })
  });
}
</script>