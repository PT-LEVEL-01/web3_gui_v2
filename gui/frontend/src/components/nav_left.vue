<template>
  <div class="common-layout" style="border:red solid 0px; height:100%;width:100%;padding:0;overflow:hidden;">
    <div class="drawing-container" v-show="store.im_cutScreenVisible">
      <!-- 绘图组件容器DOM -->
      <div id="tui-image-editor"></div>
      <div class="save">
        <el-button type="primary" @click="closeCutScreenVisible_submit">复制</el-button>
        <el-button type="primary" @click="closeCutScreenVisible_cancel">退出</el-button>
      </div>
    </div>
    <el-container style="height:100%;width:100%;border:red solid 0px;">
      <el-aside width="65px" style="padding:0;height:100%;border:red solid 0px;">
        <el-menu default-active="1" active-text-color="#ffd04b" background-color="#545c64" text-color="#fff" class="el-menu-vertical-demo"
                 :collapse="isCollapse" @select="handleSelect" @open="handleOpen" @close="handleClose" style="height:100%;">
          <el-menu-item index="1">
            <el-icon><el-badge :is-dot="store.getNavBadgeShow('im')" class="item"><ChatDotRound /></el-badge></el-icon>
            <!-- <template #title>聊天</template> -->
          </el-menu-item>
          <el-menu-item index="2">
            <el-icon><el-badge :is-dot="store.getNavBadgeShow('circle')" class="item"><Orange /></el-badge></el-icon>
            <!-- <template #title>圈子</template> -->
          </el-menu-item>
          <el-menu-item index="3">
            <el-icon><el-badge :is-dot="store.getNavBadgeShow('files')" class="item"><FolderOpened /></el-badge></el-icon>
            <!-- <template #title>存储</template> -->
          </el-menu-item>
          <el-menu-item index="4">
            <el-icon><el-badge :is-dot="store.getNavBadgeShow('wallet')" class="item"><CreditCard /></el-badge></el-icon>
            <!-- <template #title>钱包</template> -->
          </el-menu-item>
          <el-menu-item index="5" disabled>
            <el-icon><el-badge :is-dot="store.getNavBadgeShow('xxx')" class="item"><Connection /></el-badge></el-icon>
            <!-- <template #title>交易所</template> -->
          </el-menu-item>
          <el-menu-item index="8">
            <el-icon><el-badge :is-dot="store.getNavBadgeShow('xxx')" class="item"><Setting /></el-badge></el-icon>
            <!-- <template #title>设置</template> -->
          </el-menu-item>
          <!-- <el-menu-item index="9">
            <el-icon><el-badge :is-dot="store.getNavBadgeShow('xxx')" class="item"><InfoFilled /></el-badge></el-icon>
            <template #title>设置</template>
          </el-menu-item> -->
        </el-menu>
      </el-aside>
      <el-main style="padding:0;border:red solid 0px;">
        <component :is="currentView"/>
<!--        <router-view/>-->
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { GetHeadUrl } from '../head_image'
import {
  IM_StartIm_wait,
  IM_GetInfo,
  IM_GetFriendList,
  IM_GetMsg,
  IM_GetSelfInfo,
  GetFileDownloadList,
  Chain_GetInfo,
  IM_NoticeCancelFlicker
} from '../../bindings/web3_gui/gui/server_api/sdkapi'
import { SdkApi } from '../../bindings/web3_gui/gui/server_api'
import {ElMessage} from "element-plus";
import {getCurrentInstance, watch, nextTick, onMounted, ref, computed} from "vue";
import im_layout from "./im/im_layout.vue";
import circle_nav from "./circle/circle_nav.vue";
import files_nav from "./files/files_nav.vue";
import wallet_nav from "./chain/wallet_nav.vue";
import setup_nav from "./setup/setup_nav.vue";
import { store } from '../store.js'
import 'tui-image-editor/dist/tui-image-editor.css';
import 'tui-color-picker/dist/tui-color-picker.css';
import ImageEditor from 'tui-image-editor';
import * as wails from "@wailsio/runtime";
import login from "./login.vue";
import {store_routers} from "../store_routers.js";

// const store = useStore()
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const isCollapse = ref(true)
const HighestBlock = ref(0)
const PulledStates = ref(0)
const module_im = ref(false)
const module_chain = ref(false)
const instance = ref(null);

const routes = {
  "im":im_layout,
  "circle":circle_nav,
  "files":files_nav,
  "chain":wallet_nav,
  "setup":setup_nav,
}
// const currentPath = ref("login")
const currentView = computed(() => {
  return routes[store_routers.currentPageKey_index]
})

watch(
    () => store.im_cutScreenVisible,
    (newVal, oldVal) => {
      console.log("模态框是否显示",newVal,oldVal)
      if(newVal){
        console.log("全屏窗口")
        wails.Window.Fullscreen()
        // window.runtime.WindowFullscreen()//窗口全屏
      }else{
        console.log("取消全屏窗口")
        // wails.Window.UnFullscreen() //目前版本此方法还存在问题，未生效
        SdkApi.SetFullscreen(false)
        // window.runtime.WindowUnfullscreen()//取消窗口全屏，恢复全屏之前的先前窗口尺寸和位置
      }
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

//截屏窗口关闭事件
const closeCutScreenVisible_submit = () => {
  const base64String = instance.value.toDataURL(); // base64 文件
  // bus.emit("newImgB64Str",base64String)
  store.im_ScreenBase64Str = base64String
  store.im_cutScreenVisible = false
};
//截屏窗口关闭事件
const closeCutScreenVisible_cancel = () => {
  store.im_cutScreenVisible = false
};


// 中文菜单
const locale_zh = {
  ZoomIn: '放大',
  ZoomOut: '缩小',
  Hand: '手掌',
  History: '历史',
  Resize: '调整宽高',
  Crop: '裁剪',
  DeleteAll: '全部删除',
  Delete: '删除',
  Undo: '撤销',
  Redo: '反撤销',
  Reset: '重置',
  Flip: '镜像',
  Rotate: '旋转',
  Draw: '画',
  Shape: '形状标注',
  Icon: '图标标注',
  Text: '文字标注',
  Mask: '遮罩',
  Filter: '滤镜',
  Bold: '加粗',
  Italic: '斜体',
  Underline: '下划线',
  Left: '左对齐',
  Center: '居中',
  Right: '右对齐',
  Color: '颜色',
  'Text size': '字体大小',
  Custom: '自定义',
  Square: '正方形',
  Apply: '应用',
  Cancel: '取消',
  'Flip X': 'X 轴',
  'Flip Y': 'Y 轴',
  Range: '粗细',
  Stroke: '描边',
  Fill: '填充',
  Circle: '圆',
  Triangle: '三角',
  Rectangle: '矩形',
  Free: '曲线',
  Straight: '直线',
  Arrow: '箭头',
  'Arrow-2': '箭头2',
  'Arrow-3': '箭头3',
  'Star-1': '星星1',
  'Star-2': '星星2',
  Polygon: '多边形',
  Location: '定位',
  Heart: '心形',
  Bubble: '气泡',
  'Custom icon': '自定义图标',
  'Load Mask Image': '加载蒙层图片',
  Grayscale: '灰度',
  Blur: '模糊',
  Sharpen: '锐化',
  Emboss: '浮雕',
  'Remove White': '除去白色',
  Distance: '距离',
  Brightness: '亮度',
  Noise: '噪音',
  'Color Filter': '彩色滤镜',
  Sepia: '棕色',
  Sepia2: '棕色2',
  Invert: '负片',
  Pixelate: '像素化',
  Threshold: '阈值',
  Tint: '色调',
  Multiply: '正片叠底',
  Blend: '混合色',
  Width: '宽度',
  Height: '高度',
  'Lock Aspect Ratio': '锁定宽高比例',
};

// 画布组件自定义样式
const customTheme = {
  'common.bi.image': '', // 左上角logo图片
  'common.bisize.width': '0px',
  'common.bisize.height': '0px',
  'common.backgroundImage': 'none',
  // 'common.backgroundColor': '#f3f4f6',
  'common.border': '1px solid #333',

  // header
  'header.backgroundImage': 'none',
  'header.backgroundColor': '#f3f4f6',
  'header.border': '0px',

  // load button
  'loadButton.backgroundColor': '#fff',
  'loadButton.border': '1px solid #ddd',
  'loadButton.color': '#222',
  'loadButton.fontFamily': 'NotoSans, sans-serif',
  'loadButton.fontSize': '12px',
  'loadButton.display': 'none', // 可以直接隐藏掉

  // download button
  'downloadButton.backgroundColor': '#fdba3b',
  'downloadButton.border': '1px solid #fdba3b',
  'downloadButton.color': '#fff',
  'downloadButton.fontFamily': 'NotoSans, sans-serif',
  'downloadButton.fontSize': '12px',
  'downloadButton.display': 'none', // 可以直接隐藏掉

  // icons default
  'menu.normalIcon.color': '#8a8a8a',
  'menu.activeIcon.color': '#555555',
  'menu.disabledIcon.color': '#ccc',
  'menu.hoverIcon.color': '#e9e9e9',
  'submenu.normalIcon.color': '#8a8a8a',
  'submenu.activeIcon.color': '#e9e9e9',

  'menu.iconSize.width': '24px',
  'menu.iconSize.height': '24px',
  'submenu.iconSize.width': '32px',
  'submenu.iconSize.height': '32px',

  // submenu primary color
  'submenu.backgroundColor': '#1e1e1e',
  'submenu.partition.color': '#858585',

  // submenu labels
  'submenu.normalLabel.color': '#858585',
  'submenu.normalLabel.fontWeight': 'lighter',
  'submenu.activeLabel.color': '#fff',
  'submenu.activeLabel.fontWeight': 'lighter',

  // checkbox style
  'checkbox.border': '1px solid #ccc',
  'checkbox.backgroundColor': '#fff',

  // rango style
  'range.pointer.color': '#fff',
  'range.bar.color': '#666',
  'range.subbar.color': '#d1d1d1',

  'range.disabledPointer.color': '#414141',
  'range.disabledBar.color': '#282828',
  'range.disabledSubbar.color': '#414141',

  'range.value.color': '#fff',
  'range.value.fontWeight': 'lighter',
  'range.value.fontSize': '11px',
  'range.value.border': '1px solid #353535',
  'range.value.backgroundColor': '#151515',
  'range.title.color': '#fff',
  'range.title.fontWeight': 'lighter',

  // colorpicker style
  'colorpicker.button.border': '1px solid #1e1e1e',
  'colorpicker.title.color': '#fff',
};


onMounted(() => {
  nextTick(() => {
    init(); // 页面创建完成后调用
  });
  // bus.on("changeBase64Str",(number)=>{
  //   //加载新截图
  //   instance.value.loadImageFromURL(store.im_cutScreenBase64Str, 'screenshot').then((x) => {
  //     console.log(x)
  //   });
  // })
});

//监听
watch(
    () => store.im_cutScreenBase64Str,
    (newVal, oldVal) => {
      // console.log("截图长度11",newVal.length)
      nextTick().then(() => {
        //加载新截图
        instance.value.loadImageFromURL(store.im_cutScreenBase64Str, 'screenshot').then((x) => {
          console.log(x)
        });
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);


//初始化截图程序
const init = () => {
  //文档地址
  //https://ui.toast.com/tui-image-editor
  //https://nhn.github.io/tui.image-editor/latest/ImageEditor
  instance.value = new ImageEditor(document.querySelector('#tui-image-editor'), {
    includeUI: {
      loadImage: {
        // path: props.imgUrl,
        // path: 'https://fuss10.elemecdn.com/e/5d/4a731a90594a4af544c0c25941171jpeg.jpeg', // 饿了么图片
        // path: store.im_cutScreenBase64Str,
        path: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAcAAAAMCAYAAACulacQAAAAAXNSR0IArs4c6QAAABhJREFUKFNj/" +
            "P///38GHIBxVJKBgfxAAADJUC/dSvvIGQAAAABJRU5ErkJggg==",
        name: 'image',
      },
      // menu: ['resize', 'crop', 'rotate', 'draw', 'shape', 'icon', 'text', 'filter'], // 底部菜单按钮列表 隐藏镜像flip和遮罩mask
      initMenu: 'crop', // 默认打开的菜单项
      menuBarPosition: 'bottom', // 菜单所在的位置
      locale: locale_zh, // 本地化语言为中文
      theme: customTheme, // 自定义样式
    },
    // cssMaxWidth: 400, // canvas 最大宽度
    // cssMaxHeight: 500, // canvas 最大高度
  });
  document.getElementsByClassName('tui-image-editor-main')[0].style.top = '0px'; // 调整图片显示位置
  // document.getElementsByClassName('tie-btn-reset tui-image-editor-item help')[0].style.display = 'none'; // 隐藏顶部重置按钮
};

// init()


// 保存图片，并上传
const save = () => {
  const base64String = instance.value.toDataURL(); // base64 文件
  const data = window.atob(base64String.split(',')[1]);
  const ia = new Uint8Array(data.length);
  for (let i = 0; i < data.length; i++) {
    ia[i] = data.charCodeAt(i);
  }
  const file = new File([ia], "打标.png", {type: "image/png"});
  // emit('getNewImg', file);
};

function handleOpen(key, keyPath) {
  // console.log(key, keyPath)
}
function handleClose(key, keyPath) {
  // console.log(key, keyPath)
}
function handleSelect(key, keyPath) {
  // console.log(key, keyPath);
  //设置未选中任何好友，当选中好友会解除托盘闪烁
  store.setImUserinfo({Nickname: "", Addr: ""})
  // this.$store.commit('setImUserinfo', {Nickname: "", Addr: ""});
  switch (key) {
    case "1":
      store.nav_showModules = "im"
      store.setNavBadgeShow({name: "im", show: false})
      // store.commit('setNavBadgeShow', {name: "im", show: false});
      // thistemp.$router.push({path: '/index/im/index'});
      store_routers.currentPageKey_index = "im"
      break;
    case "2":
      store.nav_showModules = "circle"
      store.im_userinfo = null
      store.setNavBadgeShow({name: "circle", show: false});
      // thistemp.$router.push({path: '/index/circle/index'});
      store_routers.currentPageKey_index = "circle"
      break;
    case "3":
      store.nav_showModules = "files"
      store.im_userinfo = null
      store.setNavBadgeShow({name: "files", show: false});
      // thistemp.$router.push({path: '/index/files'});
      store_routers.currentPageKey_index = "files"
      break;
    case "4":
      store.nav_showModules = "wallet"
      store.im_userinfo = null
      store.setNavBadgeShow({name: "wallet", show: false});
      // thistemp.$router.push({path: '/index/wallet/info'});
      store_routers.currentPageKey_index = "chain"
      break;
    case "7":
      break;
    case "8":
      store.nav_showModules = "about"
      store.im_userinfo = null
      store.setNavBadgeShow({name: "about", show: false});
      // thistemp.$router.push({path: '/index/setup/account'});
      store_routers.currentPageKey_index = "setup"
      break;
    case "8-1":
      thistemp.$router.push({path: '/index/about'});
      break;
    case "9":
      thistemp.$router.push({path: '/index/test'});
      break;
    default:
  }
}

function getNavBadgeShow(badgeType) {
  return store.getNavBadgeShow(badgeType)
}
function getHighestBlock() {
  return thistemp.HighestBlock;
}


//新消息推送
function getNewMsg() {
  console.log("开始接收新消息推送")
  Promise.all([IM_GetMsg()]).then(messages => {
    console.log("有新消息推送")
    if (!messages || !messages[0]) {
      return
    }
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if (!result.success) {
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    var msgOne = messageOne.info
    console.log("有新消息通知", msgOne)
    switch (msgOne && msgOne.Subscription) {
      case 1: {//聊天消息
        if (!msgOne.FromIsSelf) {
          msgOne.HeadUrl = store.findFriendHeadNum(msgOne.From)
        }
        // var msgOne = {Type:msgOne.Type,FromIsSelf:msgOne.FromIsSelf,From:msgOne.From,To:msgOne.To,Content:msgOne.Content};
        store.pushMsgList(msgOne);
        //传文件过程中不提示
        if(msgOne.Type == 4 && msgOne.TransProgress != msgOne.progress && msgOne.progress != 0){
        }else{
          store.im_msg_content_change++
        }
        //如果是广播消息，则不在导航栏显示徽章
        if (msgOne.From == "" || msgOne.To == "") {
          break
        }
        store.setNavBadgeShow({name: "im", show: true});
        break
      }
      case 2: {//申请添加好友
        store.im_friend_apply_list_change++
        // console.log("有申请添加好友")
        if (store.im_userinfo && store.im_userinfo.Addr == "newFriend") {
          //刷新好友列表
          Promise.all([IM_GetFriendList()]).then(messages => {
            if (!messages || !messages[0]) {
              return
            }
            var messageOne = messages[0];
            var result = thistemp.$checkResultCode(messageOne.code)
            if (!result.success) {
              ElMessage({
                showClose: true,
                message: "code:"+messageOne.code+" msg:"+result.error,
                type: 'error',
              })
              return
            }
            // console.log(messageOne)
            console.log("好友列表",messageOne.info.UserList)
            for (var i = 0; i < messageOne.info.UserList.length; i++) {
              var userOne = messageOne.info.UserList[i]
              messageOne.info.UserList[i].HeadUrl = GetHeadUrl(userOne.HeadNum)
            }
            store.im_friendList = messageOne.info.UserList
          }).catch(error => {
            console.log("刷新好友列表错误:", error)
          });
          break
        }
        store.setFriendHeadBadgeShow({name: "newFriend", show: true});
        store.setNavBadgeShow({name: "im", show: true});
        break
      }
      case 3: {//同意添加好友
        store.im_friend_apply_list_change++
        store.setNavBadgeShow({name: "im", show: true});
        //刷新好友列表
        Promise.all([IM_GetFriendList()]).then(messages => {
          if (!messages || !messages[0]) {
            return
          }
          var messageOne = messages[0];
          var result = thistemp.$checkResultCode(messageOne.code)
          if (!result.success) {
            ElMessage({
              showClose: true,
              message: "code:"+messageOne.code+" msg:"+result.error,
              type: 'error',
            })
            return
          }
          // console.log("刷新好友列表", messageOne)
          for (var i = 0; i < messageOne.info.UserList.length; i++) {
            var userOne = messageOne.info.UserList[i]
            messageOne.info.UserList[i].HeadUrl = GetHeadUrl(userOne.HeadNum)
          }
          store.im_friendList = messageOne.info.UserList
          // console.log(thistemp.friendList)

        }).catch(error => {
          console.log("刷新好友列表错误:", error)
        });
        break
      }
      case 4: {//更新好友基本信息
        // console.log()
        //刷新好友列表
        Promise.all([IM_GetFriendList()]).then(messages => {
          if (!messages || !messages[0]) {
            return
          }
          var messageOne = messages[0];
          var result = thistemp.$checkResultCode(messageOne.code)
          if (!result.success) {
            ElMessage({
              showClose: true,
              message: "code:"+messageOne.code+" msg:"+result.error,
              type: 'error',
            })
            return
          }
          // console.log(messageOne)
          for (var i = 0; i < messageOne.info.UserList.length; i++) {
            var userOne = messageOne.info.UserList[i]
            messageOne.info.UserList[i].HeadUrl = GetHeadUrl(userOne.HeadNum)
          }
          store.im_friendList = messageOne.info.UserList
          // console.log(thistemp.friendList)

        }).catch(error => {
          console.log("刷新好友列表错误:", error)
        });
        break
      }
      case 5: {//群成员变动
        store.im_group_members_change++
      }
      case 11: {//服务端订单支付成功
        store.chain_payorder_server_orderid = msgOne.Content
      }
      case 12: {//客户端订单支付成功
        store.chain_payorder_client_orderid = msgOne.Content
      }
    }
    getNewMsg()
  });
}

function start(){
  //处理im模块
  Promise.all([IM_GetInfo()]).then(messages => {
    var messageOne = messages[0];
    // console.log(messageOne, !messageOne)
    if (!messageOne) {
      //没有这个模块
      return
    }
    //获取个人信息
    Promise.all([IM_GetSelfInfo()]).then(messages => {
      if (!messages || !messages[0]) {
        return
      }
      var messageOne = messages[0];
      var result = thistemp.$checkResultCode(messageOne.code)
      if (!result.success) {
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
      // console.log("获取个人信息", messages)
      var userinfo = messageOne.info;
      userinfo.HeadUrl = GetHeadUrl(userinfo.HeadNum)
      store.im_userinfo_self = userinfo
    });
    //处理im模块
    Promise.all([IM_StartIm_wait()]).then(messages => {
      // console.log("开始获取消息",messages)
      var messageOne = messages[0];
      store.im_addrself = messageOne
      getNewMsg()
    });
    //获取好友列表
    Promise.all([IM_GetFriendList()]).then(messages => {
      if (!messages || !messages[0]) {
        return
      }
      var messageOne = messages[0];
      var result = thistemp.$checkResultCode(messageOne.code)
      if (!result.success) {
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
      // console.log(messageOne)
      for (var i = 0; i < messageOne.info.UserList.length; i++) {
        var userOne = messageOne.info.UserList[i]
        messageOne.info.UserList[i].HeadUrl = GetHeadUrl(userOne.HeadNum)
      }
      store.im_friendList = messageOne.info.UserList
      // console.log(thistemp.friendList)

    }).catch(error => {
      console.log("拉取好友列表错误:", error)
    });
  });

  //区块链钱包
  Promise.all([Chain_GetInfo()]).then(messages => {
    if (!messages || !messages[0]) {
      return
    }
    var messageOne = messages[0];
    thistemp.module_chain = true
    store.setinfo(messageOne);
  })
  if (!thistemp.module_chain) {
    return
  }
}
start()

</script>

<style>
/* .el-menu-vertical-demo:not(.el-menu--collapse) {
  width: 200px;
  height:500px;
} */

.drawing-container {
  border:red solid 0px;
  height: 100%;
  width: 100%;
  position: relative;
}

.drawing-container .save {
  position: absolute;
  right: 50px;
  top: 15px;
}
</style>