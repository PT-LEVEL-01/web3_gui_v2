<template>
  <div class="drawing-container">
    <div id="tui-image-editor"></div>
    <el-button class="save" type="primary" size="small" @click="save">保存</el-button>
  </div>
</template>
<script setup>
import 'tui-image-editor/dist/tui-image-editor.css'
import 'tui-color-picker/dist/tui-color-picker.css'
import ImageEditor from 'tui-image-editor'
// import {GetScreenShot} from '../../wailsjs/go/main/App'
// import img from "@/assets/img/dg.jpg";
// import img from "https://fuss10.elemecdn.com/e/5d/4a731a90594a4af544c0c25941171jpeg.jpeg";
import {onBeforeUnmount, ref, shallowRef, onMounted, watch, getCurrentInstance} from 'vue';

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this



const instance = ref(null)//
const imgURL = ref("")//
const imgBase64Str = ref("")//

const locale_zh = ref({
  ZoomIn: "放大",
  ZoomOut: "缩小",
  Resize: "调整宽高",
  Width: '宽度',
  Height: '高度',
  'Lock Aspect Ratio': '锁定宽高比例',
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
  Range: '区间',
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
  Blend: '混合色'
});

const customTheme = ref({
  "common.bi.image": "", // 左上角logo图片
  "common.bisize.width": "0px",
  "common.bisize.height": "0px",
  "common.backgroundImage": "none",
  "common.backgroundColor": "#f3f4f6",
  "common.border": "1px solid #333",

  // header
  "header.backgroundImage": "none",
  "header.backgroundColor": "#f3f4f6",
  "header.border": "0px",

  // load button
  "loadButton.backgroundColor": "#fff",
  "loadButton.border": "1px solid #ddd",
  "loadButton.color": "#222",
  "loadButton.fontFamily": "NotoSans, sans-serif",
  "loadButton.fontSize": "12px",
  "loadButton.display": "none", // 隐藏

  // download button
  "downloadButton.backgroundColor": "#fdba3b",
  "downloadButton.border": "1px solid #fdba3b",
  "downloadButton.color": "#fff",
  "downloadButton.fontFamily": "NotoSans, sans-serif",
  "downloadButton.fontSize": "12px",
  "downloadButton.display": "none", // 隐藏

  // icons default
  "menu.normalIcon.color": "#8a8a8a",
  "menu.activeIcon.color": "#555555",
  "menu.disabledIcon.color": "#ccc",
  "menu.hoverIcon.color": "#e9e9e9",
  "submenu.normalIcon.color": "#8a8a8a",
  "submenu.activeIcon.color": "#e9e9e9",

  "menu.iconSize.width": "24px",
  "menu.iconSize.height": "24px",
  "submenu.iconSize.width": "32px",
  "submenu.iconSize.height": "32px",

  // submenu primary color
  "submenu.backgroundColor": "#1e1e1e",
  "submenu.partition.color": "#858585",

  // submenu labels
  "submenu.normalLabel.color": "#858585",
  "submenu.normalLabel.fontWeight": "lighter",
  "submenu.activeLabel.color": "#fff",
  "submenu.activeLabel.fontWeight": "lighter",

  // checkbox style
  "checkbox.border": "1px solid #ccc",
  "checkbox.backgroundColor": "#fff",

  // rango style
  "range.pointer.color": "#fff",
  "range.bar.color": "#666",
  "range.subbar.color": "#d1d1d1",

  "range.disabledPointer.color": "#414141",
  "range.disabledBar.color": "#282828",
  "range.disabledSubbar.color": "#414141",

  "range.value.color": "#fff",
  "range.value.fontWeight": "lighter",
  "range.value.fontSize": "11px",
  "range.value.border": "1px solid #353535",
  "range.value.backgroundColor": "#151515",
  "range.title.color": "#fff",
  "range.title.fontWeight": "lighter",

  // colorpicker style
  "colorpicker.button.border": "1px solid #1e1e1e",
  "colorpicker.title.color": "#fff",
});

//初始化图片编辑器
const initTuiImageEditor = () => {
  instance.value = new ImageEditor(
      document.querySelector("#tui-image-editor"),
      {
        includeUI: {
          loadImage: {
            // path:"data:image/png;base64,"+imgBase64Str.value,
            // path: 'https://pics7.baidu.com/feed/10dfa9ec8a136327684e46c100197fe40afac7c4.jpeg', // 饿了么图片
            path: 'https://fuss10.elemecdn.com/e/5d/4a731a90594a4af544c0c25941171jpeg.jpeg', // 饿了么图片
            // path: 'https://aqsk.hbgs.cloud:8888/static/2022/09/30/b0c03de6-1a87-44d6-9b38-aa9c5f31d0ef.jpg', // 公司图片
            // path: store.user.imgValue,
            //   path:img,
            name: "image",
          },
          // menu: ["resize", "rotate"], // 底部菜单按钮列表 隐藏镜像flip和遮罩mask
          // initMenu: "rotate", // 默认打开的菜单项
          // menuBarPosition: "bottom", // 菜单所在的位置
          locale: locale_zh.value, // 本地化语言为中文
          theme: customTheme.value, // 自定义样式
        },
        // cssMaxWidth: 240, // canvas 最大宽度
        // cssMaxHeight: 560, // canvas 最大高度
      }
  );
  document.getElementsByClassName("tui-image-editor-main")[0].style.top = "45px"; // 调整图片显示位置
  document.querySelector(".tui-image-editor-help-menu").style.display = "none";
  // document.querySelector('[tooltip-content="ZoomIn"]').style.display = 'none' // 放大
  document.querySelector('[tooltip-content="ZoomOut"]').style.display = 'none' // 缩小
  document.querySelector('[tooltip-content="Hand"]').style.display = 'none' // 拖动界面

  document.querySelector('[tooltip-content="Undo"]').style.display = "none"; // 上一步
  document.querySelector('[tooltip-content="Redo"]').style.display = "none"; // 下一步
  // document.querySelector('[tooltip-content="Reset"]').style.display = "none"; // 完全重新编辑
  // document.querySelector('[tooltip-content="History"]').style.display = "none";
  // document.querySelector('[tooltip-content="Delete"]').style.display = "none"; // 删除选中编辑内容
  // document.querySelector('[tooltip-content="DeleteAll"]').style.display = "none"; // 清空
  // 隐藏分割线
  document.querySelectorAll(".tui-image-editor-icpartition").forEach((item) => {
    item.style.display = "none";
  });
  // document.getElementsByClassName("tie-btn-reset tui-image-editor-item help")[0].style.display = "none"; // 隐藏顶部重置按钮
};

initTuiImageEditor()
// Promise.all([GetScreenShot()]).then(messages => {
//   console.log(messages[0])
//   imgBase64Str.value = messages[0]
//   initTuiImageEditor()
// }).catch(error => {
//   console.log("拉取好友列表错误:",error)
// });

// 图片工具初始化

// 保存图片，并上传
function save() {
  console.log("保存")
  const base64String = instance.value.toDataURL();
  const data = window.atob(base64String.split(",")[1]);
  const ia = new Uint8Array(data.length);
  for (let i = 0; i < data.length; i++) {
    ia[i] = data.charCodeAt(i);
  }
  const file = new File([ia], "gis图片", {type: "image/png"});
  // const fd = new FormData();
  // fd.append("file", file);
  // fd.append("groupId", this.editGroupId);
  // fd.append("type", 4);


  imgURL.value = URL.createObjectURL(file);
  //   console.log("map", this.layerImg);
  // this.layerImg.hide();
  //   this.mapValue.removeLayer(this.layerImg);
  //   if (this.layerImage === null) {
  // this.rotateImg(this.imgURL, 1);
  //   } else {
  //     this.layerImage.hide();
  // this.rotateImg(this.imgURL, Date.now());
  //   }


  //  this.layerImage = new maptalks.ImageLayer("images", [{
  //     url : this.imgURL ,
  //     // url : '',
  //     extent: this.mapValue.getExtent(),
  //     opacity : 0.5
  // }]);
  // console.log('新图层',this.layerImage);
  // this.layerImage.addTo(this.mapValue);
  //  map.removeLayer(this.layerImg);
  //  this.$nextTick(() => {
  // this.setMap(this.imgURL)
  //  });
  console.log("file结构体", file);
  console.log("file图片数据", imgURL.value);
}
// 图片旋转
// rotateImg(img, sign) {
//   this.layerImage = new maptalks.ImageLayer(sign, [
//     {
//       url: img,
//       // url : '',
//       extent: this.mapValue.getExtent(),
//       opacity: 0.5,
//     },
//   ]);
//   this.layerImage.addTo(this.mapValue);
// },

</script>

<style>
.drawing-container {
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
