<script setup>
import HelloWorld from '../HelloWorld.vue'
import im_screen_shot from './im_screen_shot.vue'
import { GetHeadUrl } from '../../head_image'
import {onBeforeUnmount, ref, shallowRef, onMounted,onUnmounted, watch, getCurrentInstance, nextTick, toRaw, computed } from 'vue';
import { ElScrollbar, ElMessage } from 'element-plus'
import { store } from '../../store.js'
import { Check, Delete, Edit, Message, CloseBold, Warning, WarningFilled, Close, MoreFilled } from '@element-plus/icons-vue'
import {
  IM_SendMsg, OpenFileDialog, SendFiles, SendImage, IM_NoticeCancelFlicker, SendVoiceBase64,
  IM_SendMsgAgain, GetScreenShot, GetChatHistory, ImProxyClient_GetGroupMembers, IM_DelFriend,
  ImProxyClient_GroupInvitation, ImProxyClient_GroupSendText, JoinBuildFile, ImProxyClient_GroupSendFiles,
  ImProxyClient_QuitGroup, ImProxyClient_DissolveGroup, ImProxyClient_GroupDelMember,
  IM_SetFriendRemarksname, IM_SetSelfInfo, ImProxyClient_UpdateGroup, ImProxyClient_GroupSendImage,
} from '../../../bindings/web3_gui/gui/server_api/sdkapi'

import * as wails from "@wailsio/runtime";
import "vditor/dist/index.css";
import Vditor from "vditor";
import {store_routers} from "../../store_routers.js";

import { marked } from "marked";
import prism from "prismjs";
// Add numbering to the Code blocks
import "prismjs/plugins/line-numbers/prism-line-numbers.js";
import "prismjs/plugins/line-numbers/prism-line-numbers.css";
import "prismjs/plugins/toolbar/prism-toolbar.js"; // required for the following plugins
import "prismjs/plugins/toolbar/prism-toolbar.css"; // required for the following plugins
import "prismjs/plugins/copy-to-clipboard/prism-copy-to-clipboard.js"; // show copy button
import "prismjs/plugins/show-language/prism-show-language.js"; // display the language of the code block
// This is needed for a conflict with other CSS files being used (i.e. Bulma).
import "prismjs/plugins/custom-class/prism-custom-class";
import Im_message_content_input from "./im_message_content_input.vue";
prism.plugins.customClass.map({ number: "prism-number", tag: "prism-tag" });
marked.use({
  highlight: (code, lang) => {
    if (prism.languages[lang]) {
      return prism.highlight(code, prism.languages[lang], lang);
    } else {
      return code;
    }
  },
});
prism.highlightAll();


const showImgPreview = ref(false) //图片预览显示开关
const imgSrcList = ref([
  'https://fuss10.elemecdn.com/a/3f/3302e58f9a181d2509f3dc0fa68b0jpeg.jpeg',
  'https://fuss10.elemecdn.com/1/34/19aa98b1fcb2781c4fba33d850549jpeg.jpeg',
  'https://fuss10.elemecdn.com/0/6f/e35ff375812e6b0020b6b4e8f9583jpeg.jpeg',
  'https://fuss10.elemecdn.com/9/bb/e27858e973f5d7d3904835f46abbdjpeg.jpeg',
  'https://fuss10.elemecdn.com/d/e6/c4d93a3805b3ce3f323f7974e6f78jpeg.jpeg',
  'https://fuss10.elemecdn.com/3/28/bbf893f792f03a54408b3b7a7ebf0jpeg.jpeg',
  'https://fuss10.elemecdn.com/2/11/6535bcfb26e4c79b48ddde44f4b6fjpeg.jpeg',
])

//图片点击预览
function msgImgClick(url){
  console.log("点击图片")
  imgSrcList.value = [url]
  showImgPreview.value = true
}

// function mdToHtml(text){
//   const mdhtml = marked.parse(text);
//   // return mdhtml;
//   //把img标签替换为<el-image style="width:200px;" :src="item.Content" :preview-src-list="[item.Content]" :fit="fit"/>
//   var imgArr = mdhtml.split('<img')
//   if(imgArr.length <= 1){
//     return mdhtml
//   }
//   const newMdHtml = mdhtml.replace(/<p><img\s+([^>]*?)><\/p>/gi, (match, p1) => {
//     // console.log(match,p1)
//     const altMatch = p1.match(/alt=["']?([^"'>]*)["']?/i);
//     const alt = altMatch ? ` alt="${altMatch[1]}"` : '';
//
//     var startIndex = match.indexOf('src')
//     var endIndex = match.indexOf('alt="')
//     var imgB64Str = match.substring(startIndex+5,endIndex-2)
//     return `<el-image style="width:200px;" ${alt} src="${imgB64Str}" preview-src-list="${imgB64Str}" fit="fit"/>`;
//   });
//   return newMdHtml;
// }

const msgDivIdCount = ref(0)
function GetMsgDivId(){
  msgDivIdCount.value ++
  return msgDivIdCount.value
}

// const mdToHtmlV2 = async (text) => {
//   const html = await Vditor.md2html(text)
//   return html
// };
// function mdToHtmlV3(text,i){
//   Vditor.preview(document.getElementById("msg"+i), text,{
//     cdn:"/Vditor",
//   })
//   return true
// }

function mdToHtml(text, fullImg){
  const mdhtml = marked.parse(text);
  // return mdhtml;
  //把img标签替换为<el-image style="width:200px;" :src="item.Content" :preview-src-list="[item.Content]" :fit="fit"/>
  var imgArr = mdhtml.split('<img')
  if(imgArr.length <= 1){
    return mdhtml
  }
  const newMdHtml = mdhtml.replace(/<p><img\s+([^>]*?)><\/p>/gi, (match, p1) => {
    // console.log(match,p1)
    const altMatch = p1.match(/alt=["']?([^"'>]*)["']?/i);
    const alt = altMatch ? ` alt="${altMatch[1]}"` : '';

    var startIndex = match.indexOf('src')
    var endIndex = match.indexOf('alt="')
    var imgB64Str = match.substring(startIndex+5,endIndex-2)
    if(fullImg){
      return `<img ${alt} src="${imgB64Str}" preview-src-list="${imgB64Str}" fit="fit"/>`;
    }
    return `<img style="width:200px;" ${alt} src="${imgB64Str}" preview-src-list="${imgB64Str}" fit="fit"/>`;
  });
  return newMdHtml;
}


// import { ServiceApi } from '../../../bindings/file_explorer/service_api'
// import {GreetService} from "../../bindings/changeme";
// import {ElMessage} from "element-plus";
// const props =  defineProps(['tabId'])

// const store = useStore()
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const temlrate = ref(0)
const head_self = ref(store.im_userinfo_self.HeadUrl)
const netaddr = ref(store.im_addrself)
const nickname = ref(store.im_userinfo.Nickname)
const textarea = ref("")
const broadcastContentList = ref([])

const cutScreenVisible = ref(false)//截屏模态框是否显示
const showDrawer = ref(false)//是否显示群或好友操作台
const dialogVisible_addFriend = ref(false)//是否显示群添加好友对话框
const groupMembers = ref([])//保存群成员
const groupOtherFriend = ref([])//保存群中除开好友，剩下可以添加的好友
const dialogTitle = ref("")//邀请群成员或者删除群成员对话框标题
const invitationTable = ref()//
const dialogType = ref(true)//false=删除群成员；true=添加群成员；
const groupShoutupOld = ref(false)//群是否禁言
const groupShoutup = ref(false)//群是否禁言
const remarksname = ref("")//好友昵称
const remarksnameOld = ref("")//好友昵称
remarksname.value = store.im_userinfo.RemarksName
remarksnameOld.value = store.im_userinfo.RemarksName





//在组件实例挂载到 DOM 后被调用
onMounted(() => {
  // initVditor()
  // 添加键盘事件监听
  // window.addEventListener('keydown', handleKeyDown);
  //监听截图成功后发送过来的图片
  // bus.on("newImgB64Str",(b64Str)=>{
  //   //加载新截图
  //   // const editor = editorRef.value;
  //   // if (editor == null) return;
  //   // editor.dangerouslyInsertHtml('<img width="900" height="450" src="'+ b64Str +'"/>')
  // })
  //聊天窗口滚动条移动到显示最新消息
  nextTick().then(() => {
    // DOM更新完成后的操作
    scrollbarRef.value.setScrollTop(innerRef.value.scrollHeight)
  });
});

//组件销毁时，也及时销毁编辑器，重要！
onBeforeUnmount(() => {
  //删除文件拖拽事件
  // wails.Events.Off("dragfiles")
  // 移除键盘事件监听
  // window.removeEventListener('keydown', handleKeyDown);
  //组件销毁时，也及时销毁编辑器，重要！
  // const editor = editorRef.value;
  // if (editor == null) return;
  // editor.destroy();
  // destroyVditor()
});

onUnmounted(() => {
  // destroyVditor()
});

const format = (percentage) => (percentage === 100 ? 'success' : `${percentage}%`)
const getSuccess = (progress) => {
  // var rate = store.getDownloadListProgress(item)
  // console.log("百分比",rate)
  return progress == 100 ? "success" : ""
}

//获取消息昵称
const getMsgNickname = (item) => {
  return item.Nickname
}

//监听禁言开关按钮
watch(
    () => groupShoutup.value,
    (newVal, oldVal) => {
      nextTick().then(() => {
        console.log("改变禁言",newVal,oldVal)
        if(newVal == groupShoutupOld.value){return}
        updateGroupShoutUp(newVal)
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

//修改群禁言
const updateGroupShoutUp = (shoutup) => {
  // return
  Promise.all([ImProxyClient_UpdateGroup(store.im_userinfo.GroupId, store.im_userinfo.ProxyAddr,
      store.im_userinfo.Nickname, shoutup, false)]).then(messages => {
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
    groupShoutup.value = shoutup
    groupShoutupOld.value = shoutup
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '修改失败：'+error,
      type: 'error',
    })
  });
}

//显示对话框，添加群成员
const showDialogAddMembers = () =>{
  countMembers(true)
  dialogTitle.value = "添加群成员"
  dialogType.value = true
  dialogVisible_addFriend.value = true
}

//显示对话框，删除群成员
const showDialogDelMembers = () =>{
  countMembers(false)
  dialogTitle.value = "删除群成员"
  dialogType.value = false
  dialogVisible_addFriend.value = true
}

//提交对话框按钮
const submitDialogMembers = () =>{
  if(dialogType.value){
    invitationFriend()
  }else{
    delGroupMembers()
  }
}

//修改好友备注昵称
const updateUserRemarksname = () => {
  if(remarksnameOld.value == remarksname.value){
    return
  }
  Promise.all([IM_SetFriendRemarksname(store.im_userinfo.Addr,remarksname.value)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("修改好友昵称",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

//解散群聊
const dissolveGroup = () => {
  // console.log("百分比",rate)
  Promise.all([ImProxyClient_DissolveGroup(store.im_userinfo.Addr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("删除好友",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

//退出群聊
const quitGroup = () => {
  // console.log("百分比",rate)
  Promise.all([ImProxyClient_QuitGroup(store.im_userinfo.Addr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("删除好友",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

//删除群成员
const delGroupMembers = () => {
  dialogVisible_addFriend.value = false
  var rows = invitationTable.value.getSelectionRows()
  // console.log("选中的行:",rows)
  var addrs = new Array()
  for (var i = 0; i < rows.length; i++) {
    var row = rows[i]
    addrs.push(row.Addr)
  }
  if(addrs.length == 0) {return}
  console.log("选中的成员地址",addrs)
  console.log("群成员地址",store.im_userinfo)
  Promise.all([ImProxyClient_GroupDelMember(store.im_userinfo.Addr,addrs)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("删除群成员",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

//删除好友
const delFriend = () => {
  // console.log("百分比",rate)
  Promise.all([IM_DelFriend(store.im_userinfo.Addr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("删除好友",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

//对下载完成的文件，生成完整的文件到指定目录
const joinBuildFile = (item) => {
  if(item.TransProgress != 100){
    // console.log("未下载完成")
    return
  }
  // console.log("下载完成")
  Promise.all([JoinBuildFile(store.im_userinfo.Addr, item.SendID, "")]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("邀请好友",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

const getPercentage = (friendAddr, pullID) => {
  var params = {friendAddr:friendAddr, pullID:pullID}
  var value = store.getDownloadListProgress(params)
  // console.log("刷新进度条",value)
  if(value == undefined){
    return 100
  }
  return value
}

//判断是不是广播聊天窗口
const checkMulticastChat = () => {
  return store.im_userinfo.Addr == "multicast"
}

//邀请好友入群
const invitationFriend = () => {
  dialogVisible_addFriend.value = false
  var rows = invitationTable.value.getSelectionRows()
  // console.log("选中的行:",rows)
  var addrs = new Array()
  for (var i = 0; i < rows.length; i++) {
    var row = rows[i]
    addrs.push(row.Addr)
  }
  if(addrs.length == 0) {return}
  console.log("选中的好友地址",addrs)
  console.log("好友地址",store.im_userinfo)
  // return
  Promise.all([ImProxyClient_GroupInvitation(store.im_userinfo.Addr, addrs)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("邀请好友",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

//根据邀请群成员和删除群成员，统计群成员列表
const countMembers = (isAdd) =>{
  if(isAdd){
    //好友列表中去除群成员列表，剩下的就是可以添加到群的好友
    var otherFriend = new Array()
    var list = store.im_friendList
    for(var i=0; i<list.length; i++){
      var one = list[i]
      //排除群
      if(one.IsGroup){continue}
      var have = false
      for(var j=0; j<groupMembers.value.length; j++){
        if(one.Addr == groupMembers.value[j].Addr){
          have = true
          break
        }
      }
      if(have){
        continue
      }
      otherFriend.push(one)
    }
    groupOtherFriend.value = otherFriend
    console.log("可以添加的好友列表",groupOtherFriend.value)
  }else{
    groupOtherFriend.value = groupMembers.value
  }
}

//是群聊，获取群成员
const getGroupMembers = () =>{
  console.log("获取群成员列表",store.im_userinfo.IsGroup)
  if(!store.im_userinfo.IsGroup){
    return
  }
  Promise.all([ImProxyClient_GetGroupMembers(store.im_userinfo.Addr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("获取群成员列表",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    groupMembers.value = messageOne.info.UserList
    store.im_userinfo = messageOne.groupinfo
    groupShoutup.value = messageOne.groupinfo.ShoutUp
    for(var i=0; i<groupMembers.value.length; i++){
      var userOne = groupMembers.value[i]
      groupMembers.value[i].HeadUrl = GetHeadUrl(userOne.HeadNum)
    }
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}
//刷新群成员列表
getGroupMembers()

//群成员变动时
watch(
    () => store.im_group_members_change,
    (newVal, oldVal) => {
      nextTick().then(() => {
        //刷新群成员列表
        getGroupMembers()
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);


//鼠标点击聊天窗口后，托盘闪烁消失
const mouseupSendNoticeTray = () => {
  // console.log("鼠标事件")
  var userInfo = store.im_userinfo
  //不是广播，则是单聊
  if(userInfo.Addr != "multicast"){
    //如果当前选中的是添加好友页面，则消除托盘闪烁
    IM_NoticeCancelFlicker("1"+userInfo.Addr)
  }
  //徽章提醒消失
  store.setFriendHeadBadgeShow({name:userInfo.Addr, show:false})
  // store.commit('setFriendHeadBadgeShow', {name:userInfo.Addr, show:false});
};

const scrollbarRef = ref()
const innerRef = ref()
const oldScrollHeight = ref(0)//上次滚动的位置
const scrollValue = ref(0)
const scroll = ({ scrollTop }) => {
  scrollValue.value = scrollTop
}

//当滚动聊天记录到最顶部时，加载历史消息
watch(
    () => scrollValue.value,
    (newVal, oldVal) => {
      if(newVal != 0){
        return
      }
      //加载历史消息
      //保存已有记录
      var oldList = store.getMsgByAddr()
      if(oldList == null || oldList.length == 0)return
      var index = oldList[0].Index
      // index--
      // console.log(list[0],list[list.length-1])
      if(index == ""){return}
      //拉取历史记录
      Promise.all([GetChatHistory(index, 10,store.im_userinfo.Addr)]).then(messages => {
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
        var msgList = messageOne.list.MessageList
        // console.log("拉取历史记录",msgList)
        if(msgList.length == 0){return}
        var headUrl = GetHeadUrl(store.im_userinfo.HeadNum)
        // console.log(headUrl)
        var newArray = []
        for(var i=msgList.length; i>0; i--){
          var msgOne = msgList[i-1]
          msgOne.HeadUrl = headUrl
          // if(store.im_userinfo != null){msgOne.From = ""}
          newArray.push(msgOne)
        }
        newArray = newArray.concat(oldList)
        store.SetMsgList( newArray);
      });
      // console.log("移动到新位置",newVal)
      // nextTick().then(() => {
      //   // DOM更新完成后的操作
      //   var father = document.getElementById("scrollbar");
      //   if(oldScrollHeight.value - scrollValue.value - father.clientHeight <= 200){
      //     scrollbarRef.value.setScrollTop(innerRef.value.scrollHeight)
      //   }
      //   oldScrollHeight.value = innerRef.value.scrollHeight
      // });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);


//当前好友对话框新增消息时，触发滚动条
watch(
    () => store.im_msg_content_change,
    (newVal, oldVal) => {
      nextTick().then(() => {
        // console.log("新增消息，触发滚动条")
        // DOM更新完成后的操作
        var father = document.getElementById("scrollbar");
        if(oldScrollHeight.value - scrollValue.value - father.clientHeight <= 200){
          scrollbarRef.value.setScrollTop(innerRef.value.scrollHeight)
        }
        oldScrollHeight.value = innerRef.value.scrollHeight
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

//重新打开聊天窗口时，触发
watch(
    () => store.im_userinfo,
    (newVal, oldVal) => {
      nextTick().then(() => {
        // DOM更新完成后的操作
        if(newVal == null){return}
        if(oldVal != null && newVal.Addr == oldVal.Addr){return}
        //刷新群成员列表
        getGroupMembers()
        // flashToolbarKeys()
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

const oldFriendAddr = ref('')
oldFriendAddr.value = store.getFriendAddr()
//当对话的好友改变了，触发滚动条移到最下面，显示最新消息
watch(
    () => store.getMsgByAddr(),
    (newVal, oldVal) => {
      var newFriendAddr = store.getFriendAddr()
      //没有新消息，不触发
      if(oldFriendAddr == newFriendAddr){return}
      if(newVal == null){return}
      //拉取历史记录时，不产生新消息，最后一条消息Index相等，不触发
      if(newVal.length>0 && oldVal.length>0 && newVal[newVal.length-1].Index==oldVal[oldVal.length-1].Index)return
      oldFriendAddr.value = newFriendAddr
      nextTick().then(() => {
        console.log("新消息，触发滚动条")
        // DOM更新完成后的操作
        scrollbarRef.value.setScrollTop(innerRef.value.scrollHeight)
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);



import { store_preview } from '../../store/store_im_content_preview.js'
// import Im_message_content_preview from "./im_message_content_preview.vue";
// const msgPreviewContentShow = ref(false)
// const msgPreviewContent = ref("")
function previewMsgContent(text){
  // console.log("previewMsg")
  // Vditor.preview(document.getElementById('previewDiv'),text)
  // msgPreviewContent.value = mdToHtml(text,true)
  store_preview.preview_content = mdToHtml(text,true)
  // thistemp.$router.push({path: '/index/im/message_preview'});
  store_routers.gopage_im("message_preview")
  // msgPreviewContentShow.value = true
}

function formatUnixTime(time){
  return thistemp.$formatUnixTime(time)
}

function showUserinfo(netaddr) {
  // console.log(netaddr)
  store.im_show_userinfo = {netaddr:netaddr}
  // thistemp.$router.push({path: '/index/im/userinfo'});
  store_routers.gopage_im("userinfo")
}

function getSharebox() {
  store.im_show_userinfo = {netaddr:store.im_userinfo.Addr}
  // this.$router.push({path: '/index/im/sharebox'});
  store_routers.gopage_im("sharebox")
}

</script>

<template>
  <el-container style="height:100%; margin:0; border:0px solid #eee;background-color:#fcfcfc;" @mouseup="mouseupSendNoticeTray">
    <el-header style="text-align:left; font-size: 12px;height:40px; line-height: 40px;border-bottom:1px solid #eee;">
      <span style="font-size:20px;">{{store.getFriendNickname()}} </span>
      <el-button :icon="MoreFilled" circle @click="showDrawer = true" style="border:0;font-size: 20px;margin-left: 20px;float: right;"/>
    </el-header>
    <el-main style="text-align: left;">
      <el-scrollbar id="scrollbar" ref="scrollbarRef" @scroll="scroll">
        <div ref="innerRef" style="overflow-x: hidden;">
          <div v-for="(item, i) in store.getMsgByAddr()" :key="i" class="infinite-list-item" style="border:red solid 0px;margin:5px 0;">
            <el-row :gutter="20">
              <el-col :span="3">
                <div class="grid-content ep-bg-purple" >
                  <el-button v-if="item.FromIsSelf==false" @click="showUserinfo(item.From)" style="width:40px;height:40px;border:0;">
                    <el-avatar shape="square" :size="40" :src="item.HeadUrl" />
                  </el-button>
                </div>
              </el-col>
              <el-col :span="16" style="">
                <!--发送文字消息-->
                <div v-if="item.Type == 1" class="grid-content ep-bg-purple" >
                  <div v-if="!item.FromIsSelf && (item.To == '' || item.From == '')">
                    <el-link :underline="false" @click="showUserinfo(item.From)">{{item.From}}</el-link>
                    <span style="margin-left:5px;color: #9a6e3a;">{{formatUnixTime(item.Time)}}</span>
                  </div>
                  <div v-if="!item.FromIsSelf && item.To != '' && item.From != ''">
                    <el-link :underline="false" @click="showUserinfo(item.From)">{{getMsgNickname(item)}}</el-link>
                    <span style="margin-left:5px;color: #9a6e3a;">{{formatUnixTime(item.Time)}}</span>
                  </div>
<!--                  <div v-if="item.FromIsSelf == false" class="msg_content" style="float:left;text-align:left;background-color:#faecd8;">{{item.Content}}</div>-->
<!--                  <div v-if="item.FromIsSelf == true" class="msg_content" style="float:right;text-align:right;background-color:#e1f3d8;max-width:90%;">{{item.Content}}</div>-->
                  <div v-if="item.FromIsSelf == false" class="msg_content" style="float:left;text-align:left;background-color:#faecd8;">
<!--                    <v-md-editor v-model="item.Content" mode="preview" />-->
                    <div class="line-numbers language-markup msg_content_text" @dblclick="previewMsgContent(item.Content)" v-html="mdToHtml(item.Content)"></div>
<!--                    <div :id="'msg'+i" v-if="mdToHtmlV3(item.Content,i)"></div>-->
                  </div>
                  <div v-if="item.FromIsSelf == true" class="msg_content" style="float:right;text-align:left;background-color:#e1f3d8;max-width:90%;">
<!--                    <v-md-editor v-model="item.Content" mode="preview" />-->
                    <div class="line-numbers language-markup msg_content_text" @dblclick="previewMsgContent(item.Content)" v-html="mdToHtml(item.Content)"></div>
<!--                    <div :id="'msg'+i" v-if="mdToHtmlV3(item.Content,i)"></div>-->
                  </div>
                  <el-button v-if="checkMulticastChat && item.FromIsSelf == true && item.State == 3" type="danger"
                             icon="CloseBold" circle style="float:right;" @click="sendAgain(item.From,item.To,item.SendID)"></el-button>
                  <el-button v-if="checkMulticastChat && item.FromIsSelf == true && item.State == 1" type="primary"
                             loading circle style="float:right;"></el-button>
                </div>
                <!--发送文件 旧版-->
                <div v-if="item.Type == 2" class="grid-content ep-bg-purple" >
                  <el-link v-if="!item.FromIsSelf && (item.To == '' || item.From == '')" :underline="false"
                           @click="showUserinfo(item.From)">{{item.From}}</el-link>
                  <div v-if="!item.FromIsSelf && (item.To != '' && item.From != '')">
                    <el-link :underline="false" @click="showUserinfo(item.From)">{{getMsgNickname(item)}}</el-link>
                    <span style="margin-left:5px;color: #9a6e3a;">{{formatUnixTime(item.Time)}}</span>
                  </div>
                  <div v-if="item.FromIsSelf == false" style="text-align:left;width:240px;overflow:hidden;border:red solid 0px;background-color:#ccc;padding:5px;">
                    <el-icon style="font-size:200%;"><Tickets /></el-icon>{{item.Content}}
                    <el-progress :percentage="store.getDownloadListProgress(item)" :format="format" :status="getSuccess(item)" />
                  </div>
                  <div v-if="item.FromIsSelf == true" style="float:right;text-align:right;width:240px;overflow:hidden;border:red solid 0px;background-color:#ccc;padding:5px;">
                    <el-icon style="font-size:200%;"><Tickets /></el-icon>{{item.Content}}
                    <el-progress :percentage="store.getDownloadListProgress(item)" :format="format" :status="getSuccess(item)"/>
                  </div>
                </div>
                <!--发送图片-->
                <div v-if="item.Type == 3" class="grid-content ep-bg-purple" >
                  <el-link v-if="!item.FromIsSelf && (item.To == '' || item.From == '')" :underline="false"
                           @click="showUserinfo(item.From)">{{item.From}}</el-link>
                  <div v-if="!item.FromIsSelf && (item.To != '' && item.From != '')">
                    <el-link :underline="false" @click="showUserinfo(item.From)">{{getMsgNickname(item)}}</el-link>
                    <span style="margin-left:5px;color: #9a6e3a;">{{formatUnixTime(item.Time)}}</span>
                  </div>
                  <div v-if="item.FromIsSelf == false" style="text-align:left;"><el-image style="width:200px;" :src="item.Content"
                                                                                          :preview-src-list="[item.Content]" :fit="fit"/></div>
                  <div v-if="item.FromIsSelf == true" style="text-align:right;"><el-image style="width:200px;" :src="item.Content"
                                                                                          :preview-src-list="[item.Content]" :fit="fit"/></div>
                </div>
                <!--发送文件-->
                <div v-if="item.Type == 4" class="grid-content ep-bg-purple" @click="joinBuildFile(item)">
                  <el-link v-if="!item.FromIsSelf && (item.To == '' || item.From == '')" :underline="false"
                           @click="showUserinfo(item.From)">{{item.From}}</el-link>
                  <div v-if="!item.FromIsSelf && (item.To != '' && item.From != '')">
                    <el-link :underline="false" @click="showUserinfo(item.From)">{{getMsgNickname(item)}}</el-link>
                    <span style="margin-left:5px;color: #9a6e3a;">{{formatUnixTime(item.Time)}}</span>
                  </div>
                  <div v-if="item.FromIsSelf == false" style="text-align:left;width:240px;overflow:hidden;border:red solid 0px;background-color:#ccc;padding:5px;">
                    <el-icon style="font-size:200%;"><Tickets /></el-icon>{{item.FileName}}
                    <el-progress :percentage="item.TransProgress" :format="format" :status="getSuccess(item.TransProgress)" />
                  </div>
                  <div v-if="item.FromIsSelf == true" style="float:right;text-align:right;width:240px;overflow:hidden;border:red solid 0px;background-color:#ccc;padding:5px;">
                    <el-icon style="font-size:200%;"><Tickets /></el-icon>{{item.FileName}}
                    <el-progress :percentage="item.TransProgress" :format="format" :status="getSuccess(item.TransProgress)"/>
                  </div>
                </div>
                <!--发送语音-->
                <div v-if="item.Type == 5" class="grid-content ep-bg-purple" @click="joinBuildFile(item)">
                  <el-link v-if="!item.FromIsSelf && (item.To == '' || item.From == '')" :underline="false"
                           @click="showUserinfo(item.From)">{{item.From}}</el-link>
                  <div v-if="!item.FromIsSelf && (item.To != '' && item.From != '')">
                    <el-link :underline="false" @click="showUserinfo(item.From)">{{getMsgNickname(item)}}</el-link>
                    <span style="margin-left:5px;color: #9a6e3a;">{{formatUnixTime(item.Time)}}</span>
                  </div>
                  <div v-if="item.FromIsSelf == false" style="text-align:left;width:310px;overflow:hidden;padding:5px;">
                    <audio :src="'data:'+item.FileMimeType+';base64,'+item.Content" controls style="height: 20px;"></audio>
                  </div>
                  <div v-if="item.FromIsSelf == true" style="float:right;text-align:right;width:310px;overflow:hidden;padding:5px;">
                    <audio :src="'data:'+item.FileMimeType+';base64,'+item.Content" controls style="height: 20px;"></audio>
                  </div>
                </div>
              </el-col>
              <el-col :span="3">
                <div class="grid-content ep-bg-purple" ><el-avatar v-if="item.FromIsSelf==true" shape="square" :size="40" :src="head_self" /></div>
              </el-col>
            </el-row>
          </div>
        </div>
      </el-scrollbar>

    </el-main>
    <el-footer style="height:200px;padding: 0;">
      <im_message_content_input/>
    </el-footer>
  </el-container>

  <el-image-viewer v-if="showImgPreview" :url-list="imgSrcList" show-progress :initial-index="4" @close="showImgPreview = false"/>

  <el-drawer v-model="showDrawer" :title="store.im_userinfo?store.im_userinfo.Nickname:''">
    <div v-show="!store.im_userinfo.IsGroup">
      <!--      <span style="font-size:20px;">{{store.getFriendNickname()}} </span>-->
      <div>备注昵称：<el-input v-model="remarksname" style="width: 240px" placeholder="" @blur="updateUserRemarksname"/></div>
      <div><el-button type="primary" @click="delFriend()">删除好友</el-button></div>
      <div><el-button type="primary">清空聊天记录</el-button></div>
      <div><el-button type="primary" @click="getSharebox()">查看共享文件</el-button></div>
    </div>
    <div v-show="store.im_userinfo.IsGroup">
      <div><el-button type="primary" @click="showDialogAddMembers()">邀请成员</el-button></div>
      <div v-show="store.im_userinfo.AddrAdmin == store.im_addrself">
        <div><el-button type="primary" @click="showDialogDelMembers()">删除成员</el-button></div>
        <div><el-button type="primary" @click="dissolveGroup()">解散</el-button></div>
        <div>禁言：<el-switch v-model="store.im_userinfo.ShoutUp" class="ml-2" /></div>
      </div>
      <div v-show="store.im_userinfo.AddrAdmin != store.im_addrself">
        <div><el-button type="primary" @click="quitGroup()">退出</el-button></div>
      </div>
      <div>
        <el-table ref="invitationTable" :data="groupMembers" style="width: 100%">
          <el-table-column label="" width="50">
            <template #default="scope">
              <el-image style="width: 26px; height: 26px" :src="scope.row.HeadUrl"></el-image>
            </template>
          </el-table-column>
          <el-table-column property="Nickname" label="" />
        </el-table>
      </div>
    </div>
  </el-drawer>

  <el-dialog v-model="dialogVisible_addFriend" :title="dialogTitle" width="500">
    <el-row>
      <el-col>
        <el-table ref="invitationTable" :data="groupOtherFriend" style="width: 100%">
          <el-table-column type="selection" />
          <el-table-column property="Nickname" label="" />
        </el-table>
      </el-col>
      <!--      <el-col :span="12"><div class="grid-content ep-bg-purple-light" /></el-col>-->
    </el-row>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible_addFriend=false">取消</el-button>
        <el-button type="primary" @click="submitDialogMembers()">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
ul{
  margin: 0;
  padding: 0;
}
ul li{
  list-style-type:none;
  border:0px solid red;
}
::-webkit-scrollbar{display:none;}
.msg_content{
  user-select: text;
  width:max-content;
  max-width:100%;
  word-wrap:break-word;
  background-color:#fff;
  padding:5px 10px;
  border-radius:5px;
}
.msg_content_text{
  border:0px solid red;
}
</style>