<template>
    <el-button-group class="ml-4">
        <el-button type="primary" @click="groupDialogFormVisible = true"><el-icon><Grid /></el-icon>添加分组</el-button>
        <el-button type="primary" @click="showFormDialog(null)"><el-icon><Place /></el-icon>添加节点</el-button>
        <el-button type="primary" @click="showTagDialog()"><el-icon><Management /></el-icon>管理分组</el-button>
    </el-button-group>
    <el-table ref="tableRef" :data="store.peerList" style="width: 100%"  @filter-change="filterChange">
        <el-table-column prop="status" :min-width="30">
            <template  #default="scope">
                <el-switch size="small" v-model="scope.row.status" :active-value="true" :inactive-value="false" @change="switchChange(scope.row)"/>
            </template>
        </el-table-column>
        <el-table-column prop="group" label="分组" :filters="groupfilters" :filter-method="filterTag" filter-placement="bottom-end"
                         :min-width="40" column-key="groups"/>
        <el-table-column prop="name" label="节点名" :min-width="40"/>
        <el-table-column prop="highest_block" label="高度">
            <template #default="scope">
                <el-text class="mx-1" :type="getCurrentBlockType(scope.row)">{{scope.row.current_block}}</el-text>/
              <el-text class="mx-1" :type="scope.row.status?'success':'info'" tag="b">{{scope.row.highest_block}}</el-text>
            </template>
        </el-table-column>
        <el-table-column prop="total_balance" label="余额"/>
        <el-table-column prop="types" label="角色" :min-width="40">
            <template #default="scope">
                <el-tag v-for="(num, type) in scope.row.types" :type="getTagType(scope.row.status?type:-1)" size="small">
                  {{ getTagText(type) }}{{num>1?num:''}}</el-tag>
            </template>
        </el-table-column>
        <el-table-column prop="used_time" label="延迟" :min-width="30">
            <template #default="scope">
                <el-text class="mx-1" :type="getUsedTimeType(scope.row)">{{scope.row.used_time}}ms</el-text>
            </template>
        </el-table-column>
        <el-table-column fixed="right" label="Operations" :min-width="55">
            <template #default="scope">
                <el-button link type="primary" @click="showDrawer(scope.row)" ><el-icon><Expand /></el-icon></el-button>
                <el-button link type="primary" @click.prevent="showFormDialog(scope.row)"><el-icon><Edit /></el-icon></el-button>
                <el-popconfirm title="是否删除该节点?" confirm-button-text="是" cancel-button-text="否" @confirm="deleteRow(scope.row)">
                    <template #reference>
                        <el-button link type="primary"><el-icon><Delete /></el-icon></el-button>
                    </template>
                </el-popconfirm>

            </template>
        </el-table-column>
    </el-table>



    <el-dialog v-model="dialogFormVisible" title="节点配置" destroy-on-close center draggable>
        <el-form  :model="peerConfForm" ref="peerConfFormRef" :rules="rules">

            <el-form-item label="分组" :label-width="formLabelWidth"  prop="group">
                <el-select v-model="peerConfForm.group" placeholder="Please select a group">
                    <el-option
                        v-for="(item) in groupfilters"
                        :key="item.value"
                        :label="item.value"
                        :value="item.value"
                    />
                </el-select>
            </el-form-item>
            <el-form-item label="节点名" :label-width="formLabelWidth" prop="name">
                <el-input v-model="peerConfForm.name" autocomplete="off" />
            </el-form-item>
            <el-form-item label="IP" :label-width="formLabelWidth"  prop="ip">
                <el-input v-model="peerConfForm.ip" autocomplete="off" />
            </el-form-item>
            <el-form-item label="Port" :label-width="formLabelWidth" prop="port">
                <el-input v-model="peerConfForm.port" autocomplete="off" type="number"/>
            </el-form-item>
            <el-form-item label="RPC用户名" :label-width="formLabelWidth" prop="username">
                <el-input v-model="peerConfForm.username" autocomplete="off" />
            </el-form-item>
            <el-form-item label="RPC密码" :label-width="formLabelWidth" prop="password">
                <el-input v-model="peerConfForm.password" autocomplete="off" />
            </el-form-item>
        </el-form>
        <template #footer>
      <span class="dialog-footer">
        <el-button @click="closeFormDialog">取消</el-button>
        <el-button type="primary" @click="peerConfFormSubmit">
          保存
        </el-button>
      </span>
        </template>
    </el-dialog>




    <el-dialog v-model="groupDialogFormVisible" width="40%" title="添加分组" destroy-on-close  center draggable>
        <el-form :model="groupForm">
            <el-form-item label="分组名" :label-width="formLabelWidth" prop="group">
                <el-input v-model="groupForm.group" autocomplete="off" />
            </el-form-item>
        </el-form>
        <template #footer>
      <span class="dialog-footer">
        <el-button @click="groupDialogFormVisible = false">取消</el-button>
        <el-button type="primary" @click="groupFormSubmit">
          保存
        </el-button>
      </span>
        </template>
    </el-dialog>

    <el-dialog v-model="tagDialogVisible" title="分组管理" destroy-on-close center draggable>
            <div>
                <el-tag :key="index" v-for="(tag, index) in dynamicTags" closable :disable-transitions="false" @close="showPopover(tag)">
                    <span v-if="!tag.editable" @click="tagHandleEdit(index)">{{ tag.value }}</span>
                    <el-input v-else v-model="tag.value" size="small" ref="editInput" @keyup.enter.native="$event.target.blur()"
                        @blur="tagHandleInputConfirm(index)"></el-input>

                    <el-popover :visible="tag.popover" placement="bottom" :width="160"  trigger="manual">
                        <p>删除该分组和该分组下所有节点?</p>
                        <div style="text-align: right; margin: 0">
                            <el-button size="small" text @click="tag.popover = false">否</el-button>
                            <el-button size="small" type="primary" @click="tagHandleClose(tag)"
                            >是</el-button>
                        </div>
                        <template #reference><span></span></template>
                    </el-popover>

                </el-tag>
                <el-input
                    class="input-new-tag"
                    v-if="newTagInputVisible"
                    v-model="newTagInputValue"
                    ref="saveTagInput"
                    size="small"
                    @keyup.enter.native="$event.target.blur()"
                    @blur="newTagHandleInputConfirm"
                ></el-input>
                <el-button v-else class="button-new-tag" size="small" @click="showNewTagInput">+ 新增分组</el-button>
            </div>
    </el-dialog>
</template>
<script setup>
import {getCurrentInstance, unref, ref, onMounted, onUnmounted} from 'vue'
import { ElMessage } from 'element-plus'
import {AddGroup,AddPeer,DelPeer,DelGroup,EditGroup, EditPeer, GetPeerList,GetPeerGroup, SetStatus}
  from '../../../../bindings/web3_gui/gui/server_api/sdkapi'
import { store } from '../../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

// const timer = ref(0)
const peerConfFormRef = ref(null)
const tableRef = ref(null)
const dialogFormVisible = ref(false)
const groupDialogFormVisible = ref(false)
const tagDialogVisible = ref(false)
const peerConfForm = ref({
  id: '',
  group: '',
  name: '',
  ip: '',
  port: 0,
  username: '',
  password: '',
  status: false,
  highest_block: 0,
  current_block: 0,
  snapshot_height: 0,
  total_balance: 0,
  types: null,
  used_time: 0,
  updated_at:0,
  error:"",
  default_address: null,
  addresses:[],
})
const editingItem = ref(false)
const formLabelWidth = ref('100px')
const groupfilters = ref([])
const rules = ref({
  group: [
    { required: true, message: '请输入组名', trigger: 'blur' },
  ],
  name: [
    { required: true, message: '请输入节点名', trigger: 'blur' },
  ],
  ip: [
    { required: true, message: '请输入ip', trigger: 'blur' },
  ],
  port: [
    { required: true, message: '请输入端口', trigger: 'blur' },
    {pattern:  /^[1-9]\d*$/,"message": "只能输入大于0的正整数", trigger: 'blur'}
  ],
  username: [
    { required: true, message: '请输入RPC用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入RPC密码', trigger: 'blur' },
  ]
})
const selectedGroups = ref([])
const groupForm = ref({group:''})
const dynamicTags = ref([
  // { value: 'Tag 1', editable: false,tmp:'',popover:false  },
  // { value: 'Tag 2', editable: false,tmp:'',popover:false  },
  // { value: 'Tag 3', editable: false ,tmp:'',popover:false }
])
const newTagInputVisible = ref(false)
const newTagInputValue = ref('')
const peerCurrentBlocks = ref({})

function loadPeerList(){
  GetPeerList().then((result) => {
    if (result != null){
      if (selectedGroups.value.length > 0) {
        // 过滤出 selectedGroups 中存在的 group
        const selected = result.filter(item => selectedGroups.value.includes(item.group));
        // 将相同 group 的节点对象放在一起
        const grouped = selected.reduce((obj, item) => {
          obj[item.group] = [...(obj[item.group] || []), item];
          return obj;
        }, {});
        // 将 grouped 的值放在数组前面
        store.peerList = [].concat(...Object.values(grouped), ...result.filter(item => !selectedGroups.value.includes(item.group)));
      } else {
        store.peerList = result;
      }
    }
  }).catch(error => {
    // 处理错误
    console.log(error)
  });
}

function loadGroups(){
  GetPeerGroup().then((result) => {
    console.log("获取的分组",result);
    const groups = []
    for (var item in result) {
      groups.push({
        text: result[item],value: result[item]
      })
      // const index = this.groupfilters.findIndex((data) => data.text === item);
      // if (index === -1) {
      //     this.groupfilters.push({
      //         text: result[item],value: result[item]
      //     })
      // }
    }
    groupfilters.value = groups
  }).catch(error => {
    // 处理错误
    console.log(error)
  });
}

function switchChange(item){
  SetStatus(item.id,item.status).then((result) => {
    console.log('id :'+item.id+' status:'+item.status)
  }).catch(error => {
    // 处理错误
    ElMessage.error(error)
  });
}

function showFormDialog(item) {
  dialogFormVisible.value = true;
  if (item) {
    editingItem.value = true;
    peerConfForm.value = item;
  } else {
    editingItem.value = false;
    peerConfForm.value.id='';
    peerConfForm.value.group='';
    peerConfForm.value.name='';
    peerConfForm.value.ip='';
    peerConfForm.value.port=0;
    peerConfForm.value.username='';
    peerConfForm.value.password='';
    peerConfForm.value.status=false;
    peerConfForm.value.highest_block=0;
    peerConfForm.value.current_block=0;
    peerConfForm.value.snapshot_height=0;
    peerConfForm.value.total_balance=0;
    peerConfForm.value.types=null;
    peerConfForm.value.used_time=0;
    peerConfForm.value.updated_at=0;
    peerConfForm.value.error='';
    peerConfForm.value.default_address=null;
    peerConfForm.value.addresses=[];
  }
}

function closeFormDialog() {
  dialogFormVisible.value = false;
}

function peerConfFormSubmit() {
  const formRef = unref(thistemp.$refs.peerConfFormRef);
  if (!formRef) return
  //通过ref的值触发验证
  formRef.validate((valid) => {
    if (valid) {
      const newItem = {
        id: peerConfForm.value.id,
        group: peerConfForm.value.group,
        name: peerConfForm.value.name,
        ip: peerConfForm.value.ip,
        port: parseInt(peerConfForm.value.port.toString()),
        username: peerConfForm.value.username,
        password: peerConfForm.value.password,
        // status: this.peerConfForm.status,
        // highest_block: this.peerConfForm.highest_block,
        // current_block: this.peerConfForm.current_block,
        // balance: this.peerConfForm.balance,
        // types: this.peerConfForm.types,
        // delay: this.peerConfForm.delay,
        // is_del: this.peerConfForm.is_del,
      };
      if (editingItem.value) {
        EditPeer(newItem).then((result) => {
          peerConfForm.value = result;
          loadPeerList();
          closeFormDialog();
        }).catch(error => {
          // 处理错误
          ElMessage.error(error)
        });
      } else {
        AddPeer(newItem).then((result) => {
          //store.peerList.push(result)
          loadPeerList();
          closeFormDialog();
        }).catch(error => {
          // 处理错误
          ElMessage.error(error)
        });
      }
    } else {
      console.log("未通过");
    }
  });
}

function deleteRow(item) {
  const index = store.peerList.findIndex((data) => data.id === item.id);
  if (index !== -1) {
    store.peerList.splice(index, 1);
    DelPeer(item.id).then((result) => {
      if (peerCurrentBlocks.value[item.id]){
        delete peerCurrentBlocks.value[item.id]
      }
      console.log('删除id :'+item.id+' result:'+result)
    }).catch(error => {
      // 处理错误
      ElMessage.error(error)
    });
  }
}

function filterTag(value,row,column){
  selectedGroups.value = column.filteredValue //记录选中的项
  return true //让筛选不起作用
}

function getTagType(type){
  if (type == 1) {
    return 'success';
  } else if (type == 2) {
    return 'danger';
  } else if (type == 3) {
    return 'warning';
  } else {
    return 'info';
  }
}

function getTagText(type){
  if (type == 1) {
    return '见证人';
  } else if (type == 2) {
    return '社区节点';
  } else if (type == 3) {
    return '轻节点';
  } else {
    return '普通节点';
  }
}

function getUsedTimeType(row){
  if (row.status=== false || row.used_time === 0) {
    return 'info';
  } else if (0 < row.used_time && row.used_time <= 50) {
    return 'success';
  } else if (50 <row.used_time && row.used_time <= 200) {
    return 'warning';
  } else {
    return 'danger';
  }
}

function showDrawer(row){
  const index = store.peerList.findIndex((data) => data.id === row.id);
  if (index !== -1) {
    store.peerInfoIndex = index
    store.selectedAddress = row.default_address?row.default_address.address:''
    store.drawer = true
    store.setAddressesOptions()
  }
}

function filterChange(filters) {
  selectedGroups.value = filters['groups']
}

function groupFormSubmit() {
  const group = groupForm.value.group.trim()

  if (group === '') {
    ElMessage.error("请填写分组名！")
    return false
  }
  AddGroup(group).then((result) => {
    groupfilters.value.push({text: group,value:group})
    groupDialogFormVisible.value = false
    loadGroups()
    groupForm.value.group = ''
  }).catch(error => {
    // 处理错误
    ElMessage.error(error)
    return false
  });
}

function showTagDialog(){
  const tags = [];
  groupfilters.value.forEach(function (item, index) {
    tags.push({ value: item.value, editable: false,tmp:item.value,popover:false });
  });
  dynamicTags.value = tags
  tagDialogVisible.value = true
}

function showPopover(tag) {
  tag.popover = true;
}

function tagHandleClose(tag) {
  const index = dynamicTags.value.findIndex(t => t.value === tag.value);
  if (index !== -1) {
    tag.popover = false;
    //删除
    DelGroup(tag.value.trim()).then((result) => {
      this.loadGroups()
    }).catch(error => {
      // 处理错误
      ElMessage.error(error)
      return false
    });
    dynamicTags.value.splice(index, 1);
  }
}

function showNewTagInput() {
  newTagInputVisible.value = true;
  this.$nextTick(() => {
    thistemp.$refs.saveTagInput.focus();
  });
}

function tagHandleEdit(index) {
  dynamicTags.value[index].editable = true;
  dynamicTags.value[index].tmp = dynamicTags.value[index].value;
  thistemp.$nextTick(() => {
    const input = thistemp.$refs.editInput[index];
    if (input) {
      input.$el.querySelector('input').focus();
    }
  });
}

function tagHandleInputConfirm(index) {
  const tag = dynamicTags.value[index];
  tag.editable = false;
  if (tag.value.trim()) {
    if (tag.value.trim() !== tag.tmp.trim()) {
      console.log(tag)
      //修改
      EditGroup(tag.value.trim(),tag.tmp.trim()).then((result) => {
        loadGroups()
        tag.tmp = tag.value.trim()
      }).catch(error => {
        // 处理错误
        ElMessage.error(error)
        return false
      });
    }
  } else {
    tag.value = tag.tmp.trim()
  }
}

function newTagHandleInputConfirm() {
  const newGorup = newTagInputValue.value.trim();
  if (newGorup) {
    AddGroup(newGorup).then((result) => {
      dynamicTags.value.push({ value: newGorup, editable: false,tmp:newGorup,popover:false});
      loadGroups()
    }).catch(error => {
      // 处理错误
      ElMessage.error(error)
      return false
    });
  }
  newTagInputValue.value = '';
  newTagInputVisible.value = false;
}

function getCurrentBlockType(row){
  if (!row.status) {
    return 'info'
  }
  //console.log(this.peerCurrentBlocks)
  if (peerCurrentBlocks.value[row.id] == null){
    peerCurrentBlocks.value[row.id] = [row.current_block]
  }else {
    const time =new Date().getTime()
    if (time%30==0){
      peerCurrentBlocks.value[row.id].push(row.current_block)
      if (peerCurrentBlocks.value[row.id].length>3) {
        peerCurrentBlocks.value[row.id].splice(0,1)
      }
    }
    const sum = peerCurrentBlocks.value[row.id].reduce((total,num)=>total+num);
    if ((sum / peerCurrentBlocks.value[row.id].length) == row.current_block){
      return 'danger'
    }
  }
  return 'primary'
}

loadPeerList();
loadGroups()

// 定义一个ref来持有定时器
const timer = ref(null);
// 创建定时器
const createTimer = () => {
  timer.value = setInterval(() => {
    loadPeerList();
    // 定时器的逻辑
  }, 500);
};

//在组件挂载时创建定时器
onMounted(() => {
  createTimer();
});

//在组件卸载时清除定时器
onUnmounted(() => {
  if (timer.value) {
    clearInterval(timer.value);
  }
});

</script>

<style scoped>
.el-tag + .el-tag {
    margin-left: 10px;
}
.button-new-tag {
    margin-left: 10px;
    height: 32px;
    line-height: 30px;
    padding-top: 0;
    padding-bottom: 0;
}
.input-new-tag {
    width: 90px;
    margin-left: 10px;
    vertical-align: bottom;
}
</style>
