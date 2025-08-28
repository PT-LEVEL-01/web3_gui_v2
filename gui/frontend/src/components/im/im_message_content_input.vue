<script setup>
import {Close} from "@element-plus/icons-vue";
import {ElMessage} from "element-plus";
import {
  GetScreenShot,
  IM_SendMsg, IM_SendMsgAgain, ImProxyClient_GroupSendFiles, ImProxyClient_GroupSendImage,
  ImProxyClient_GroupSendText, OpenFileDialog, SendFiles, SendImage,
  SendVoiceBase64
} from "../../../bindings/web3_gui/gui/server_api/sdkapi.js";
import {store} from "../../store.js";
import {getCurrentInstance, nextTick, onBeforeUnmount, onMounted, onUnmounted, ref, watch} from "vue";
import * as wails from "@wailsio/runtime";
import "vditor/dist/index.css";
import Vditor from "vditor";

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
let voiceStartTime = new Date();
let voiceSecend = 0;
let audioMimeType = ""
let audioContentData = ""
let mediaRecorder;
let chunks = [];

const audioPlayerDisabled = ref(false)

const vditor = ref(null);
const markdownRef = ref(null);
const editorMode = ref("wysiwyg");

//初始化聊天输入框
function initVditor(){
  vditor.value = new Vditor(markdownRef.value, {
    // lang: 'zh_CN',
    mode:editorMode.value,
    // value: props.modelValue,
    height:150,
    cache: {
      enable: false
    },
    fullscreen: {
      index: 10000
    },
    toolbar: [
      'emoji',
      {
        hotkey: '⇧⌘S',
        name: 'sendFile',
        tipPosition: 'n',
        tip: '发送文件',
        className: 'sendFile',
        icon: '<svg xmlns="http://www.w3.org/2000/svg" id="Layer_1" data-name="Layer 1" viewBox="0 0 24 24" width="512"' +
            ' height="512"><path d="m19.95,5.536l-3.485-3.485c-1.322-1.322-3.08-2.05-4.95-2.05h-4.515C4.243,0,2,2.243,2,' +
            '5v14c0,2.757,2.243,5,5,5h10c2.757,0,5-2.243,5-5v-8.515c0-1.87-.728-3.627-2.05-4.95Zm-1.414,1.414c.318.318.' +
            '587.671.805,1.05h-4.341c-.551,0-1-.449-1-1V2.659c.379.218.733.487,1.05.805l3.485,3.485Zm1.464,12.05c0,1.654' +
            '-1.346,3-3,3H7c-1.654,0-3-1.346-3-3V5c0-1.654,1.346-3,3-3h4.515c.163,0,.325.008.485.023v4.977c0,1.654,1.346,' +
            '3,3,3h4.977c.015.16.023.322.023.485v8.515Zm-4.293-4.519c.391.391.391,1.023,0,1.414-.195.195-.451.293-.707.293s-' +
            '.512-.098-.707-.293l-1.293-1.293v4.398c0,.552-.448,1-1,1s-1-.448-1-1v-4.398l-1.293,1.293c-.391.391-1.023.391-1' +
            '.414,0s-.391-1.023,0-1.414l1.614-1.614c1.154-1.154,3.032-1.154,4.187,0l1.614,1.614Z"/></svg>',
        click () {
          addSendFile()
        },
      },
      {
        hotkey: '⇧⌘S',
        name: 'sendFile',
        tipPosition: 'n',
        tip: '截图',
        className: 'sendFile',
        icon: '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" id="Icons" style="enable' +
            '-background:new 0 0 32 32;" version="1.1" viewBox="0 0 32 32" xml:space="preserve"><g><path d="M29.7,7.2C29.9,' +
            '7,30,6.7,30,6.4s-0.1-0.5-0.3-0.7c-2.3-2.3-6.1-2.3-8.4,0L9.1,18c-1.8-0.3-3.7,0.3-5.1,1.6   c-1.1,1.1-1.7,2.6-1.' +
            '7,4.2c0,1.6,0.6,3.1,1.7,4.2c1.2,1.2,2.7,1.7,4.2,1.7c1.5,0,3-0.6,4.2-1.7c1.1-1.1,1.7-2.6,1.7-4.2   c0-0.3,0-0.' +
            '6-0.1-0.9L29.7,7.2z"/><path d="M8.2,14.1c1.5,0,3-0.6,4.2-1.7c1.1-1.1,1.7-2.6,1.7-4.2c0-1.6-0.6-3.1-1.7-4.2C10,' +
            '1.7,6.3,1.7,4,4C2.9,5.1,2.2,6.6,2.2,8.2   c0,1.6,0.6,3.1,1.7,4.2C5.1,13.5,6.6,14.1,8.2,14.1z"/><path d="M30.7' +
            ',24.8L22.9,17l-4.8,4.8l4.3,4.5c1.2,1.2,2.7,1.7,4.2,1.7c1.5,0,3-0.6,4.2-1.7c0.2-0.2,0.3-0.4,0.3-0.7   S30.9,25' +
            ',30.7,24.8z"/></g></svg>',
        click () {
          cutScreen()
        },
      },
      {
        hotkey: '⇧⌘S',
        name: 'sendFile',
        tipPosition: 'n',
        tip: '隐藏窗口截图',
        className: 'sendFile',
        icon: '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" id="Icons" style="enable' +
            '-background:new 0 0 32 32;" version="1.1" viewBox="0 0 32 32" xml:space="preserve"><style type="text/css">\n' +
            '    .st0{fill:none;stroke:#000000;stroke-width:2;stroke-linecap:round;stroke-linejoin:round;stroke-miterlimit:10;}\n' +
            '</style><circle class="st0" cx="8.2" cy="23.8" r="4.9"/><path class="st0" d="M9.4,19.1L22.1,6.4c1.9-1.9,5-1.' +
            '9,6.9,0l0,0l-16,16"/><circle class="st0" cx="8.2" cy="8.2" r="4.9"/><path class="st0" d="M16.1,19.3l5.9,6.2' +
            'c1.9,1.9,5,1.9,6.9,0l0,0L19.4,16"/></svg>',
        click () {
          cutScreenMinimize()
        },
      },
      'bold','italic','headings','|','list','ordered-list','check','code','|',
      'line','quote','table','link','|','preview','fullscreen',"export", "edit-mode",
    ],
    // cdn: "https://ld246.com/js/lib/vditor",
    hint:{
      //https://raw.githubusercontent.com/88250/lute/refs/heads/master/parse/emoji_map.go
      emoji: {
        "heart_eyes":                           "😍",
        "sleeping":                             "😴",
        "sleepy":                               "😪",
        "slightly_frowning_face":               "🙁",
        "slightly_smiling_face":                "🙂",
        "smile":                                "😄",
        "smiley":                               "😃",
        "smirk":                                "😏",
        "sob":                                  "😭",
        "stuck_out_tongue":                     "😛",
        "stuck_out_tongue_closed_eyes":         "😝",
        "stuck_out_tongue_winking_eye":         "😜",
        "sunglasses":                           "😎",
        "sweat":                                "😓",
        "thinking":                             "🤔",
        "triumph":                              "😤",
        "unamused":                             "😒",
        "upside_down_face":                     "🙃",
        "weary":                                "😩",
        "v":                                    "✌️",
        "+1":                                   "👍",
        "-1":                                   "👎",
        "muscle":                               "💪",
        "tipping_hand_man":                     "💁‍♂",
        "tipping_hand_woman":                   "💁",
        "toilet":                               "🚽",
        "tada":                                 "🎉",
        'love':                                 '❤️',
        "broken_heart":                         "💔",
        "watermelon":                           "🍉",
        "wc":                                   "🚾",
        "100":                                  "💯",
        "airplane":                             "✈️",
        "bullettrain_front":                    "🚅",
        "anchor":                               "⚓️",
        "bus":                                  "🚌",
        "car":                                  "🚗",
        "motor_scooter":                        "🛵",
        "bike":                                 "🚲",
        "kick_scooter":                         "🛴",
        "dromedary_camel":                      "🐪",
        "running":                              "🏃",
        "walking":                              "🚶",
        "baseball":                             "⚾️",
        "basketball":                           "🏀",
        "bath":                                 "🛀",
        "bathtub":                              "🛁",
        "chart_with_downwards_trend":           "📉",
        "chart_with_upwards_trend":             "📈",
      },
    },
    upload: {
      // accept: 'image/*',
      handler(files) {
        console.log("粘贴文件数量",files.length);
        // 处理多个图片文件
        Array.from(files).forEach(file => {
          console.log("图片类型",files.type);

          if (!file.type.match('image/*')) return;
          const reader = new FileReader();
          reader.onload = (e) => {
            const base64 = e.target.result;
            const imgMarkdown = `![${file.name}](${base64})`;
            vditor.insertValue(imgMarkdown + '\n');
          };
          reader.readAsDataURL(file);
        });
        return false; // 阻止默认上传
      },
      // 粘贴处理配置
      // pasteLink: true,
      // 禁用其他上传方式
      // linkToImgUrl: '',
      // url: ''
    },
    // 添加粘贴事件监听（备用方案）
    after() {
      // this.element.addEventListener('paste', (e) => {
      //   console.log("粘贴图片");
      //   const items = (e.clipboardData || window.clipboardData).items;
      //
      //   for (let i = 0; i < items.length; i++) {
      //     if (items[i].type.indexOf('image') !== -1) {
      //       const blob = items[i].getAsFile();
      //       const reader = new FileReader();
      //
      //       reader.onload = (event) => {
      //         const base64 = event.target.result;
      //         vditor.insertValue(`![粘贴的图片](${base64})`);
      //       };
      //
      //       reader.readAsDataURL(blob);
      //       e.preventDefault();
      //       break;
      //     }
      //   }
      // });
    },
    input(value) {
      // emit("update:modelValue", value);
    },
    focus(value) {
      // emit("focus", value);
    },
    blur(value) {
      // emit("blur", value);
    },
    esc(value) {
      // emit("esc", value);
    },
    ctrlEnter(value) {
      // emit("ctrlEnter", value);
    },
    select(value) {
      // emit("select", value);
    }
  });

  // vditor.value.toolbar.forEach((menuItem) => {
  //   menuItem.tipPosition = "s"
  // })
}

// const drawer = ref(false);
const sendFileDisabled = ref(false);
const filePath = ref([]);//要发送的文件路径
//监听待发送文件列表
watch(filePath.value,(newVal, oldVal) => {
  if(newVal.length === 0){
    sendFileDisabled.value = false
  }else{
    sendFileDisabled.value = true
  }
})

//添加待发送文件
function addSendFile() {
  // const content = vditor.value.getValue()
  // vditor.value.insertValue('插入的文本'+content);
  Promise.all([OpenFileDialog()]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+messageOne.error,
        type: 'error',
      })
      return
    }
    for(var i=0; i<messageOne.paths.length; i++){
      var filePathOne = messageOne.paths[i]
      filePath.value.push(filePathOne)
      // if (editor == null) return;
      // editor.dangerouslyInsertHtml('<a class="sendFile" href="" target="_blank">'+filePath+'</a><br>')
    }
    if(filePath.value.length > 0){
      sendFileDisabled.value = true;
      // drawer.value = true
    }
  });
}
//删除待发送文件
function delSendFile(index){
  filePath.value.splice(index, 1)
}

//截图
function cutScreen() {
  Promise.all([GetScreenShot(true)]).then(messages => {
    if (!messages || !messages[0]) {
      return
    }
    var messageOne = messages[0];
    // console.log("获取离线服务器列表", messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if (!result.success) {
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+messageOne.error,
        type: 'error',
      })
      return
    }
    console.log("截屏长度",messageOne.info.length)
    // window.runtime.WindowUnminimise()//将窗口恢复到最小化之前的尺寸和位置
    // window.runtime.WindowFullscreen()//窗口全屏
    // cutScreenVisible.value = true
    store.im_cutScreenVisible = true//
    store.im_cutScreenBase64Str = messageOne.info//
    // bus.emit("changeBase64Str",messageOne.info)
    // const imgMarkdown = `![image](`+messageOne.info+`)`;
    // vditor.value.insertValue(imgMarkdown);
    // editor.dangerouslyInsertHtml('<img width="900" height="450" src="data:image/png;base64,'+ messageOne.info+'"/>')
    // valueHtml.value = ''
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

//窗口最小化后截图
function cutScreenMinimize() {
  Promise.all([GetScreenShot(false)]).then(messages => {
    if (!messages || !messages[0]) {
      return
    }
    var messageOne = messages[0];
    // console.log("获取离线服务器列表", messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if (!result.success) {
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+messageOne.error,
        type: 'error',
      })
      return
    }
    console.log("截图大小",messageOne.info)
    // window.runtime.WindowUnminimise()//将窗口恢复到最小化之前的尺寸和位置
    // window.runtime.WindowFullscreen()//窗口全屏
    // cutScreenVisible.value = true
    store.im_cutScreenVisible = true//
    store.im_cutScreenBase64Str = messageOne.info//
    // bus.emit("changeBase64Str",messageOne.info)
    // const imgMarkdown = `![image](`+messageOne.info+`)`;
    // vditor.value.insertValue(imgMarkdown);
    // editor.dangerouslyInsertHtml('<img width="900" height="450" src="data:image/png;base64,'+ messageOne.info+'"/>')
    // valueHtml.value = ''
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
}

//压缩图片方法
function compressImage(file, quality, callback) {
  const reader = new FileReader();
  reader.onload = (e) => {
    const img = new Image();
    img.onload = () => {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');

      // 计算压缩后的尺寸
      const maxWidth = 800;
      const maxHeight = 800;
      let width = img.width;
      let height = img.height;

      if (width > height) {
        if (width > maxWidth) {
          height *= maxWidth / width;
          width = maxWidth;
        }
      } else {
        if (height > maxHeight) {
          width *= maxHeight / height;
          height = maxHeight;
        }
      }

      canvas.width = width;
      canvas.height = height;
      ctx.drawImage(img, 0, 0, width, height);

      // 转换为Base64
      const compressedBase64 = canvas.toDataURL('image/jpeg', quality);
      callback(compressedBase64);
    };
    img.src = e.target.result;
  };
  reader.readAsDataURL(file);
}

const voiceNotDisplay = () => {
  audioPlayerDisabled.value = false
  chunks = [];
}

//按下开始录音
const voiceStart = () => {
  // console.log("开始录音")
  voiceNotDisplay()
  voiceStartTime = new Date();
  const stream = navigator.mediaDevices.getUserMedia({ audio: true });
  stream.then(mediaStream => {
    mediaRecorder = new MediaRecorder(mediaStream);
    mediaRecorder.ondataavailable = event => {
      if (event.data.size > 0) {
        chunks.push(event.data);
        // console.log("录音内容",event.data)
      }
    };
    mediaRecorder.onstop = () => {
      // console.log("录音停止---")
      if(chunks.length==0){return}
      const reader = new FileReader();
      reader.onload = function(event) {
        // console.log("录音内容编码",event.target.result); //输出: data:text/plain;base64,...
        audioMimeType = getMimeType(event.target.result)
        audioContentData = getContentData(event.target.result)
        // console.log("文件类型",mimeType,"文件内容",contentData)
        const audioBlob = base64ToBlob(audioMimeType, audioContentData)
        // const audioBlob = new Blob(contentData, { type: mimeType });
        const audioUrl = URL.createObjectURL(audioBlob);
        document.getElementById('audioPlayer').src = audioUrl;
        // document.getElementById('saveAudio').disabled = false;
      };
      reader.readAsDataURL(chunks[0]);
    };
    mediaRecorder.start();
    // document.getElementById('startRecord').disabled = true;
    // document.getElementById('stopRecord').disabled = false;
  }).catch(err => {
    // console.log("Error accessing media devices.", err)
    if(err == "NotFoundError: Requested device not found"){
      ElMessage({
        showClose: true,
        message: "未找到麦克风",
        type: 'error',
      })
    }
  });
  // return item.Nickname
}


//获取字符串"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."中的文件类型字符串
function getMimeType(dataString) {
  // 检查字符串是否以 "data:" 开头
  if (!dataString.startsWith("data:")) {
    return "";
  }
  // 查找第一个逗号的位置，MIMETYPE和编码信息位于逗号之前
  const commaIndex = dataString.indexOf("base64,");
  if (commaIndex == -1) {
    return "";
  }
  // 提取MIMETYPE部分（从 "data:" 之后到逗号之前）
  const mimeTypeAndCharset = dataString.slice(5, commaIndex-1);
  return mimeTypeAndCharset; // 或者 throw new Error("Invalid data string format");
}

//获取字符串"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."中的base64字符串内容
function getContentData(dataString) {
  // 检查字符串是否以 "data:" 开头
  if (!dataString.startsWith("data:")) {
    return "";
  }
  // 查找第一个逗号的位置，MIMETYPE和编码信息位于逗号之前
  const commaIndex = dataString.indexOf(",");
  if (commaIndex == -1) {
    return "";
  }
  // 提取MIMETYPE部分（从 "data:" 之后到逗号之前）
  const contentData = dataString.slice(commaIndex+1, dataString.length);
  return contentData; // 或者 throw new Error("Invalid data string format");
}

//把base64字符串转化为Blob对象
function base64ToBlob(mimeType, base64) {
  // 将Base64编码转换为Uint8Array
  const byteCharacters = atob(base64);
  const byteNumbers = new Array(byteCharacters.length);
  for (let i = 0; i < byteCharacters.length; i++) {
    byteNumbers[i] = byteCharacters.charCodeAt(i);
  }
  const byteArray = new Uint8Array(byteNumbers);
  // 创建Blob对象
  const blob = new Blob([byteArray], { type: mimeType });
  return blob;
}

//释放鼠标按钮发送录音
const voiceSend = () => {
  console.log("发送录音")
  if(mediaRecorder == null){return}
  mediaRecorder.stop();
  // mediaRecorder.
  const voiceEndTime = new Date();
  voiceSecend = Math.abs(voiceEndTime - voiceStartTime); // 取绝对值，以防止出现负值
  console.log("录音总时长",voiceSecend)
  voiceSecend = Math.floor(voiceSecend / 1000);
  // voiceSecend = voiceSecend/1000
  // if(voiceSecend > 0){//录音必须1秒钟以上
  //   console.log("录音内容汇总",chunks)
  //
  // }
  audioPlayerDisabled.value = true
  mediaRecorder = null
  // chunks = []; //重置chunks下次使用
  // return item.Nickname
}

//取消发送录音
const voiceStop = () => {
  if(mediaRecorder==null){return}
  console.log("停止录音")
  mediaRecorder.stop();
  mediaRecorder = null
  // chunks = []; //重置chunks下次使用
  // return item.Nickname
}

//发送语言消息
function sendMsgVoice(){
  if(chunks.length<=0){
    return false
  }
  if(voiceSecend<=0){
    ElMessage({
      showClose: true,
      message: '语音时长大于1秒才能发送',
      type: 'error',
    })
    return true
  }
  // console.log("录音内容编码",voiceCoding)
  Promise.all([SendVoiceBase64(store.im_userinfo.Addr,audioMimeType,audioContentData,voiceSecend)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("发送语音",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    voiceNotDisplay()
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
  return true
}

//发送文件消息
function sendMsgFile() {
  if(filePath.value.length<=0){
    // sendFileDisabled.value = false
    return false
  }
  var toAddr = store.im_userinfo.Addr == "multicast" ? "":store.im_userinfo.Addr
  //广播地址不能发送文件
  if(toAddr == ""){
    return false
  }

  //开始发送文件
  const filePathOne = filePath.value[0]
  //是群发送
  if(store.im_userinfo.IsGroup){
    Promise.all([ImProxyClient_GroupSendFiles(toAddr,filePathOne)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      // console.log("获取离线服务器列表",messageOne)
      var result = thistemp.$checkResultCode(messageOne.code)
      if(!result.success){
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
      filePath.value.splice(0, 1)
      sendMsgFile()
    }).catch(error => {
      ElMessage({
        showClose: true,
        message: '发送消息失败：'+error,
        type: 'error',
      })
    });
    return true
  }
  Promise.all([SendFiles(toAddr,filePathOne)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    // console.log("获取离线服务器列表",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    filePath.value.splice(0, 1)
    sendMsgFile()
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
  return true
}

//发送文本消息
function sendMsgTxt() {
  console.log("开始发送文本")
  const contentText = vditor.value.getValue()
  //发送文字消息
  if(contentText == ""){
    console.log("发送文本为空")
    return
  }
  var toAddr = store.im_userinfo.Addr == "multicast" ? "":store.im_userinfo.Addr
  //发送群文字消息
  if(store.im_userinfo.IsGroup){
    console.log("发送群消息",store.im_userinfo)
    // return
    Promise.all([ImProxyClient_GroupSendText(toAddr, contentText,"")]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      console.log("发送文字消息返回结果",messageOne)
      var result = thistemp.$checkResultCode(messageOne.code)
      if(!result.success){
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
      vditor.value.setValue("", true)
      // valueHtml.value = ''
    }).catch(error => {
      ElMessage({
        showClose: true,
        message: '发送消息失败：'+error,
        type: 'error',
      })
    });
    return
  }
  //发送给个人的消息
  Promise.all([IM_SendMsg(contentText, toAddr)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    console.log("发送文字消息返回结果",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    vditor.value.setValue("", true)
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
  return false
}

//发送消息
const submitSend = () =>{
  //有语音，先发送语音消息
  if(sendMsgVoice()){
    console.log("发送了语音")
    return
  }
  //有文件，先发送文件
  if(sendMsgFile()){
    console.log("发送了文件")
    return
  }
  if (vditor.value == null) {
    console.log("内容为空")
    return
  }
  console.log("发送了文本")
  //发送文本消息内容
  sendMsgTxt()
  return

  // if(editor.getText() == "")return;
  var filePath = []
  var imgList = []
  var contentText = editor.getText().split("\n").join('')
  var htmlArr = editor.getHtml().split('<a')
  var imgArr = editor.getHtml().split('<img')
  // console.log(editor.getText(), editor.getHtml());

  //解析图片
  if(imgArr.length == 1){
  }else{
    for(var i=0; i<imgArr.length; i++){
      var one = imgArr[i]
      var startIndex = one.indexOf('src')
      var endIndex = one.indexOf('alt="')
      if(endIndex == -1){
        continue
      }
      var filePathOne = one.substring(startIndex+5,endIndex-2)
      // console.log(filePathOne);
      imgList.push(filePathOne)
    }
  }

  //解析文件
  if(htmlArr.length == 1){
  }else{
    for(var i=0; i<htmlArr.length; i++){
      var one = htmlArr[i]
      var startIndex = one.indexOf('>')
      var endIndex = one.indexOf('</a>')
      if(endIndex == -1){
        continue
      }
      var filePathOne = one.substring(startIndex+1,endIndex)
      // console.log(filePathOne);
      filePath.push(filePathOne)
    }
    for(var i=0; i<filePath.length; i++){
      // var temp = contentText.split(filePath[i])
      // console.log(filePath[i], temp, temp.join(''))
      contentText = contentText.split(filePath[i]).join('')
    }
  }
  var toAddr = store.im_userinfo.Addr == "multicast" ? "":store.im_userinfo.Addr
  //广播地址不能发送文件
  if(toAddr == "" && filePath.length>0){
    return
  }
  //广播地址不能发送图片
  if(toAddr == "" && imgList.length>0){
    return
  }

  // console.log(filePath, contentText)
  //开始发送图片
  for(var i=0; i<imgList.length; i++){
    //是群发送
    if(store.im_userinfo.IsGroup){
      Promise.all([ImProxyClient_GroupSendImage(toAddr,imgList[i])]).then(messages => {
        if(!messages || !messages[0]){return}
        var messageOne = messages[0];
        // console.log("获取离线服务器列表",messageOne)
        var result = thistemp.$checkResultCode(messageOne.code)
        if(!result.success){
          ElMessage({
            showClose: true,
            message: "code:"+messageOne.code+" msg:"+result.error,
            type: 'error',
          })
          return
        }
        valueHtml.value = ''
      }).catch(error => {
        ElMessage({
          showClose: true,
          message: '发送消息失败：'+error,
          type: 'error',
        })
      });
      continue
    }
    Promise.all([SendImage(toAddr,imgList[i])]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      // console.log("获取离线服务器列表",messageOne)
      var result = thistemp.$checkResultCode(messageOne.code)
      if(!result.success){
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
      valueHtml.value = ''
    }).catch(error => {
      ElMessage({
        showClose: true,
        message: '发送消息失败：'+error,
        type: 'error',
      })
    });
  }


}

//重新发送消息
const sendAgain = (from,to,sendID) => {
  Promise.all([IM_SendMsgAgain(from,to,sendID)]).then(messages => {
    if(!messages || !messages[0]){return}
    var messageOne = messages[0];
    // console.log("获取离线服务器列表",messageOne)
    var result = thistemp.$checkResultCode(messageOne.code)
    if(!result.success){
      ElMessage({
        showClose: true,
        message: "code:"+messageOne.code+" msg:"+result.error,
        type: 'error',
      })
      return
    }
    // valueHtml.value = ''
  }).catch(error => {
    ElMessage({
      showClose: true,
      message: '发送消息失败：'+error,
      type: 'error',
    })
  });
};

//销毁聊天输入框
function destroyVditor(){
  const editorInstance = vditor.value;
  if (!editorInstance) return;
  try {
    editorInstance?.destroy?.();
  } catch (error) {
    console.log(error);
  }
}

//键盘事件
function handleKeyDown(event) {
  if (event.key === 'Enter' && event.ctrlKey) {
    // 用户按下了ctrl + Enter键
    submitSend()
    return
  }
  // 检查按键并处理逻辑
  if (event.key === 'Enter') {
    // 用户按下了Enter键
  }
}

//添加拖拽文件事件
wails.Events.On("dragfiles", function(event) {
  event.data.forEach(function(file) {
    // const editor = editorRef.value;
    // if (editor == null) return;
    // editor.dangerouslyInsertHtml('<a class="sendFile" href="" target="_blank">'+file+'</a><br>')
  });
})

//监听编辑后的图片
watch(
    () => store.im_ScreenBase64Str,
    (newVal, oldVal) => {
      // console.log("模态框是否显示",newVal,oldVal)
      const imgMarkdown = `![image](`+newVal+`)`;
      vditor.value.insertValue(imgMarkdown);
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

//在组件实例挂载到 DOM 后被调用
onMounted(() => {
  initVditor()
  // 添加键盘事件监听
  window.addEventListener('keydown', handleKeyDown);
  //监听截图成功后发送过来的图片
  // bus.on("newImgB64Str",(b64Str)=>{
  //   console.log("有新截图", b64Str.length);
  //   //加载新截图
  //   const imgMarkdown = `![image](`+b64Str+`)`;
  //   vditor.value.insertValue(imgMarkdown);
  //   // const editor = editorRef.value;
  //   // if (editor == null) return;
  //   // editor.dangerouslyInsertHtml('<img width="900" height="450" src="'+ b64Str +'"/>')
  // })
  //聊天窗口滚动条移动到显示最新消息
  nextTick().then(() => {
    // DOM更新完成后的操作
    // scrollbarRef.value.setScrollTop(innerRef.value.scrollHeight)
  });
});

//组件销毁时，也及时销毁编辑器，重要！
onBeforeUnmount(() => {
  //删除文件拖拽事件
  wails.Events.Off("dragfiles")
  // 移除键盘事件监听
  window.removeEventListener('keydown', handleKeyDown);
  //组件销毁时，也及时销毁编辑器，重要！
  // const editor = editorRef.value;
  // if (editor == null) return;
  // editor.destroy();
  // destroyVditor()
});

onUnmounted(() => {
  destroyVditor()
});

</script>

<template>
  <div style="height:150px;">
    <div ref="markdownRef" style="height:100%;border-bottom:1px;border-left:1px;border-right:1px;"/>
  </div>
  <div style="margin-top:8px;">
    <el-button style="float:left;margin-left:20px;" @mousedown="voiceStart()" @mouseup="voiceSend()" @mouseout="voiceStop()">按住 说话</el-button>
    <div v-show="audioPlayerDisabled" style="float: left;height: 30px;">
      <audio id="audioPlayer" controls style="height: 30px;"></audio>
      <el-button :icon="Close" @click="voiceNotDisplay()" circle style=""/>
    </div>
<!--  显示待发送文件列表  -->
    <div v-show="sendFileDisabled" style="float: left;height: 30px;overflow-y: scroll;max-width:calc(100% - 220px);">
      <el-tag v-for="(item,i) in filePath" :key="item" closable :disable-transitions="false" @close="delSendFile(i)">
        {{ item }}
      </el-tag>
    </div>
    <el-tooltip class="box-item" effect="dark" content="Send(Ctrl+Enter)" placement="top-start">
      <el-button style="float:right;margin-right:20px;" @click="submitSend()">发送</el-button>
    </el-tooltip>
  </div>
</template>

<style scoped>

</style>