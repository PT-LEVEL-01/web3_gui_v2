<template>
  <div>
    <ul class="todo-list">
      <li>
        <input type="text" value="haha"/>
      </li>
      <li>
        <input v-if = "edit" v-model = "content" @blur= "edit = false; $emit('update')" @keyup.enter = "edit=false; $emit('update')"/>
      </li>
      <li v-for = "todo in todos">
        <input v-if = "todo.edit" v-model = "todo.title"
        @blur= "todo.edit = false; $emit('update')"
        @keyup.enter = "todo.edit=false; $emit('update')">
              <div v-else>
          <label @click = "todo.edit = true;"> {{todo.title}} </label>
        </div>
      </li>
    </ul>
    
  </div>
  
  <el-scrollbar ref="scrollbarRef" height="400px" always @scroll="scroll">
    <div ref="innerRef">
      <p v-for="item in 20" :key="item" class="scrollbar-demo-item">
        {{ item }}
      </p>
    </div>
  </el-scrollbar>

  <el-slider
    v-model="value"
    :max="max"
    :format-tooltip="formatTooltip"
    @input="inputSlider"
  />
</template>

<script lang="ts" setup>
import { reactive, ref, onMounted } from 'vue'
import { ElScrollbar } from 'element-plus'

const edit = ref(false)
const content = ref("hahahahaha")

const max = ref(0)
const value = ref(0)
const innerRef = ref<HTMLDivElement>()
const scrollbarRef = ref<InstanceType<typeof ElScrollbar>>()

const todos = reactive([{'title':'one value','edit':false},
                  {'title':'one value','edit':false},
                    {'title':'otro titulo','edit':false}])


onMounted(() => {
  max.value = innerRef.value!.clientHeight - 380
})

const inputSlider = (value: number) => {
  scrollbarRef.value!.setScrollTop(value)
}
const scroll = ({ scrollTop }) => {
  value.value = scrollTop
}
const formatTooltip = (value: number) => {
  return `${value} px`
}
</script>

<style scoped>
.scrollbar-demo-item {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 50px;
  margin: 10px;
  text-align: center;
  border-radius: 4px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}
.el-slider {
  margin-top: 20px;
}
</style>
