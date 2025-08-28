<script setup>
import { store_preview } from '../../store/store_im_content_preview.js'
import {onMounted, onUnmounted, ref, toRaw, watch, nextTick } from "vue";
import "vditor/dist/index.css";
import Vditor from "vditor";
import {store_routers} from "../../store_routers.js";

const editor = ref(null);
const markdownRef = ref(null);
const editorMode = ref("sv");



onMounted(() => {
  editor.value = new Vditor(markdownRef.value, {
    // cdn:"/Vditor",
    // lang: 'zh_CN',
    mode:editorMode.value,
    value: store_preview.preview_content,
    cache: {
      enable: false
    },
    fullscreen: {
      index: 20000000
    },
    toolbar: [
      'emoji','bold','italic','headings','|','list','ordered-list','check','code','|',
      'line','quote','table','link','|','preview','fullscreen',"export", "edit-mode",
    ],
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
    after() {
      // emit("after", toRaw(editor.value));
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
});

onUnmounted(() => {
  const editorInstance = editor.value;
  if (!editorInstance) return;
  try {
    editorInstance?.destroy?.();
  } catch (error) {
    console.log(error);
  }
});

watch(
    () => store_preview.preview_content,
    (newVal, oldVal) => {
      nextTick().then(() => {
        console.log("新消息")
        //
        editor.value.setValue(newVal,true)
      });
    },
    {flush: "post"}//DOM更新之后再执行，需要设置flush: "post"
);

const back = () => {
  // window.history.back()
  store_routers.goback_im()
}
</script>

<template>
  <el-page-header @back="back" content="预览">
  </el-page-header>
  <div style="height:100%;border:red solid 0px;">
    <div ref="markdownRef" style="height:100%;border:red solid 0px;"/>
  </div>
</template>

<style scoped>

</style>