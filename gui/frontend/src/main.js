import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import ElementPlus, {ElMessage} from 'element-plus'
import 'element-plus/dist/index.css'
// import axios from 'axios'
import BigNumber from "bignumber.js";
import Calculator from "./Calculator.min.js";
// import SuiVue from 'semantic-ui-vue'
// import 'semantic-ui-css/semantic.min.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
// import VMdEditor from '@kangc/v-md-editor';
// import '@kangc/v-md-editor/lib/style/base-editor.css';
// import githubTheme from '@kangc/v-md-editor/lib/theme/github.js';
// import '@kangc/v-md-editor/lib/theme/style/github.css';
// import vuepressTheme from '@kangc/v-md-editor/lib/theme/vuepress.js';
// 引入所有语言包
// import hljs from 'highlight.js';

const app = createApp(App)
app.use(router)
// app.use(store)
// app.use(SuiVue)

// 使用主题
// VMdEditor.use(vuepressTheme, {
//     Hljs: hljs,
// });
// app.use(VMdEditor);

//全局注册组件
// import ImMessageHistory from './components/im/im_message_history.vue'
// app.component("ImMessageHistory", ImMessageHistory)

for (const [name, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(name, component);
}
// for ([name, comp] of Object.entries(ElementPlusIconsVue)) {
//   app.component(name, comp);
// }
// app.component('el-icon', ElIcon);
// Vue.use(ElementUI);
app.use(ElementPlus);

app.config.productionTip = false
// app.prototype.$axios = axios;
// app.config.globalProperties.$axios = axios;
// app.prototype.$BigNumber = BigNumber;
app.config.globalProperties.$BigNumber = BigNumber;
// app.prototype.$Calculator = Calculator;
app.config.globalProperties.$Calculator = Calculator;


var getAssetsImages = (fileUrl) => {
    return new URL(`/src/assets/${fileUrl}`, import.meta.url).href;
}
app.config.globalProperties.$getAssetsImages = getAssetsImages;

var checkAmountNotZero = (rule, value, callback) => {
    if (value === '') {
        callback(new Error('不能为空'));
        return
    }
    if (value <= 0){
        callback(new Error('不能为0'));
        return
    }
    var a=/^[1-9]*(\.[0-9]{1,8})?$/;
    var  b=/^[0]{1}(\.[0-9]{1,8})?$/;
    var  c=/^[1-9]*(\.[0-9]{1,8})?$/;
    var  d=/^[1-9][0-9]*(\.[0-9]{1,8})?$/;
    var  e=/^\.\d{1,8}?$/;
    if((!a.test(value)&&!b.test(value)&&!c.test(value)&&!d.test(value))||e.test(value)){
        callback(new Error('转账金额只能是数字或小数点后1-2位数字,且前面不要带无效数字0'));
    }else{
        callback();
    }
};

var checkAmountHaveZero = (rule, value, callback) => {
    if (value === '') {
        callback(new Error('不能为空'));
        return
    }
    var a=/^[1-9]*(\.[0-9]{1,8})?$/;
    var  b=/^[0]{1}(\.[0-9]{1,8})?$/;
    var  c=/^[1-9]*(\.[0-9]{1,8})?$/;
    var  d=/^[1-9][0-9]*(\.[0-9]{1,8})?$/;
    var  e=/^\.\d{1,8}?$/;
    if((!a.test(value)&&!b.test(value)&&!c.test(value)&&!d.test(value))||e.test(value)){
        callback(new Error('转账金额只能是数字或小数点后1-2位数字,且前面不要带无效数字0'));
    }else{
        callback();
    }
};

app.config.globalProperties.$checkAmountNotZero = checkAmountNotZero;
app.config.globalProperties.$checkAmountHaveZero = checkAmountHaveZero;



app.config.globalProperties.$changeSize = (size) => {
    var unit = "B"
    var tempSize = size
    while(unit != "GB"){
        var newSize = tempSize/1024
        if(newSize < 1){
            break
        }
        tempSize = newSize
        switch(unit){
            case "B":
                unit = "KB"
                break
            case "KB":
                unit = "MB"
                break
            case "MB":
                unit = "GB"
                break
        }
    }
    tempSize = tempSize+""
    tempSize = tempSize.substr(0,tempSize.indexOf(".")+3);
    return tempSize+" "+unit
}
app.config.globalProperties.$checkResultCode = (code) => {
    switch(code){
        case 0: return {success:true,error:"success"}
        case 2000:
            // app.config.globalProperties.$alert('创建Token成功', '成功', {
            //   confirmButtonText: '确定',
            //   type: 'success ',
            //   callback: action => {
            //   }
            // });
            return {success:true,error:""}
        case 40002: return {success:false,error:"密码错误"}
        case 5003: return {success:false,error:"自定义错误"}
        case 5006: return {success:false,error:"密码错误"}
        case 5008:
            //余额不足
            // app.config.globalProperties.$alert('可用余额不足', '失败', {
            //   confirmButtonText: '确定',
            //   type: 'error',
            //   callback: action => {
            //   }
            // });
            return {success:false,error:"余额不足"}
        case 5009: return {success:false,error:"参数格式不正确"}
        case 5010: return {success:false,error:"转账不能为0"}
        case 5011: return {success:false,error:"地址角色不正确"}
        case 5012: return {success:false,error:"余额不足"}
        case 5013: return {success:false,error:"投票已经存在"}
        case 5014: return {success:false,error:"投票功能暂未开放"}
        case 5015: return {success:false,error:"轻节点奖励异步执行中"}
        case 5016: return {success:false,error:"备注信息字符串超过最大长度"}
        case 5017: return {success:false,error:"交易手续费太少"}
        case 5018: return {success:false,error:"有投票，不能取消轻节点奖励"}
        case 5019: return {success:false,error:"系统错误"}
        case 5020: return {success:false,error:"不是轻节点"}
        case 5021: return {success:false,error:"不是见证者节点"}
        case 5022: return {success:false,error:"不是社区节点"}
        case 5023: return {success:false,error:"金额太小"}
        case 5024: return {success:false,error:"公钥不存在"}
        case 5025: return {success:false,error:"链没有初始化"}
        case 5026: return {success:false,error:"奖励没有上链"}
        case 5027: return {success:false,error:"重复奖励"}
        case 5028: return {success:false,error:"分配奖励太早"}
        case 5029: return {success:false,error:"分配比例不能大于100"}
        case 5030: return {success:false,error:"参数错误"}
        case 5031: return {success:false,error:"nft不属于自己"}
        case 5032: return {success:false,error:"gas费用过高"}
        case 5033: return {success:false,error:"测试币领取限制"}
        case 5034: return {success:false,error:"不是测试链"}
        case 5035: return {success:false,error:"见证人质押已经存在"}
        case 5036: return {success:false,error:"见证人押金数量不对"}

        case 50001: return {success:false,error:"序列化错误"}
        case 50002: return {success:false,error:"此路径不存在"}

        // case 60000: return {success:true,error:"success"}
        case 60001: return {success:false,error:""}
        case 60002: return {success:false,error:""}

        case 61005: return {success:false,error:""}
        case 61006: return {success:false,error:""}
        case 61007: return {success:false,error:""}
        case 61008: return {success:false,error:""}

        case 61009: return {success:false,error:"密码错误"}

        case 64001: return {success:false,error:"好友在列表中"}
        case 64002: return {success:false,error:"无效的同意添加好友"}
        case 64003: return {success:false,error:"用户不存在"}
        case 64004: return {success:false,error:"未送达消息太多"}

        case 66001: return {success:false,error:""}
        case 66002: return {success:false,error:""}
        case 66003: return {success:false,error:""}
        case 66004: return {success:false,error:""}
        case 66005: return {success:false,error:"存储空间不足"}
        case 66006: return {success:false,error:""}
        case 66007: return {success:false,error:""}
        case 66008: return {success:false,error:""}
        case 66009: return {success:false,error:""}

        default:
            //余额不足
            // app.config.globalProperties.$alert(response.data.code, '失败', {
            //   confirmButtonText: '确定',
            //   type: 'error',
            //   callback: action => {
            //   }
            // });
            return {success:false,error:"未定义的错误编号:"+code}
    }
}

var formatUnixTime = (timestamp) => {
    const date = new Date(timestamp * 1000); // JavaScript中的Date对象以毫秒为单位
    const year = date.getFullYear();
    const month = date.getMonth() + 1; // 月份是从0开始的
    const day = date.getDate();
    const hours = date.getHours();
    const minutes = date.getMinutes();
    const seconds = date.getSeconds();
    return `${year}-${month.toString().padStart(2, '0')}-${day.toString().padStart(2, '0')} ${hours.toString().padStart(2, '0')}:
        ${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
};
app.config.globalProperties.$formatUnixTime = formatUnixTime;

app.mount('#app')



