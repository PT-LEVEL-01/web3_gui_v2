<template>
    <div>
        <el-button type="primary" @click="showFormDialog">添加</el-button>
        <el-table :data="tableData">
            <el-table-column label="名称" prop="name"></el-table-column>
            <el-table-column label="操作">
                <template #default="{ row }">
                    <el-button @click="showFormDialog(row)">修改</el-button>
                    <el-button @click="deleteItem(row)">删除</el-button>
                </template>
            </el-table-column>
        </el-table>
        <el-dialog v-model="formDialogVisible" @close="closeFormDialog" title="表单">
            <el-form :model="formData" label-width="80px">
                <el-form-item label="名称">
                    <el-input v-model="formData.name"></el-input>
                </el-form-item>
            </el-form>
            <div slot="footer">
                <el-button @click="closeFormDialog">取消</el-button>
                <el-button type="primary" @click="submitForm">提交</el-button>
            </div>
        </el-dialog>
    </div>
</template>

<script setup>
import {getCurrentInstance, ref} from "vue";
const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this
const tableData = ref([
  { id: 1, name: '数据1' },
  { id: 2, name: '数据2' },
  { id: 3, name: '数据3' },
])
const formDialogVisible = ref(false);
const formData = ref({name: ''})
const editingItem = ref(null)

function showFormDialog(item) {
  formDialogVisible.value = true;
  if (item) {
    editingItem.value = item;
    formData.value.name = item.name;
  } else {
    editingItem.value = null;
    formData.value.name = '';
  }
}

function closeFormDialog() {
  formDialogVisible.value = false;
  formData.value.name = '';
}

function submitForm() {
  if (editingItem.value) {
    editingItem.value.name = formData.value.name;
  } else {
    const newItem = {
      id: tableData.value.length + 1,
      name: formData.value.name,
    };
    tableData.value.push(newItem);
  }
  closeFormDialog();
}

function deleteItem(item) {
  const index = tableData.value.findIndex((data) => data.id === item.id);
  if (index !== -1) {
    tableData.value.splice(index, 1);
  }
}
</script>