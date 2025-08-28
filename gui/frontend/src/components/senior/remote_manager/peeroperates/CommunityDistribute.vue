<template>
    <el-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="rules"
        label-width="120px"
        class="demo-ruleForm"
        status-icon
    >
        <el-form-item label="地址" prop="srcaddress">
            <el-input v-model="ruleForm.srcaddress" type="text" />
        </el-form-item>
        <el-form-item label="手续费" prop="gas">
            <el-input v-model="ruleForm.gas" type="number"/>
        </el-form-item>
        <el-form-item label="Password" prop="pass">
            <el-input v-model="ruleForm.pass" type="password" show-password/>
        </el-form-item>

        <el-form-item>
            <el-button type="primary" @click="submitForm(ruleFormRef)">
                确认
            </el-button>
            <el-button @click="resetForm(ruleFormRef)">重置</el-button>
        </el-form-item>
    </el-form>
</template>


<script lang="ts" setup>
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
// import {useStore} from "vuex";
import { store } from '../store.js'
import {ElMessage} from "element-plus";
import {CommunityDistribute} from '../../../../../bindings/web3_gui/gui/server_api/sdkapi'
// const store = useStore()

interface RuleForm {
    srcaddress: string
    pass: string
    gas: number
}

const ruleFormRef = ref<FormInstance>()
const ruleForm = reactive<RuleForm>({
    srcaddress: '',
    pass:'',
    gas:0.1,
})

const rules = reactive<FormRules<RuleForm>>({
    srcaddress: [
        {
            required: true,
            message: '地址不能为空',
            trigger: 'blur',
        },
        { min: 30, max: 128, message: '地址长度不正确', trigger: 'blur' },
    ],
    gas: [
        {
            required: true,
            message: '手续费不能为空',
            trigger: 'blur',
        },
        {pattern: /^(0\.\d+|[1-9]\d*(\.\d+)?)$/,"message": "必须大于0", trigger: 'blur'}
    ],
    pass: [
        {
            required: true,
            message: '密码不能为空',
            trigger: 'blur',
        },
        { min: 1, max: 256, message: '密码长度不正确', trigger: 'blur' },
    ],
})

const submitForm = async (formEl: FormInstance | undefined) => {
    if (!formEl) return
    await formEl.validate((valid, fields) => {
        if (valid) {
            CommunityDistribute(store.getPeerInfo().id,ruleForm.srcaddress,parseFloat(ruleForm.gas.toString()),ruleForm.pass).then((result) => {
                ElMessage.success("成功！")
            }).catch(error => {
                // 处理错误
                ElMessage.error(error)
            });
        } else {
            console.log('error submit!', fields)
        }
    })
}

const resetForm = (formEl: FormInstance | undefined) => {
    if (!formEl) return
    formEl.resetFields()
}
</script>