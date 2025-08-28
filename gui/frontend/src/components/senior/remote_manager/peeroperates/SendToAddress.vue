<template>
    <el-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="rules"
        label-width="120px"
        class="demo-ruleForm"
        status-icon
    >
        <el-form-item label="收款地址" prop="toaddress">
            <el-cascader v-model="ruleForm.toaddress" :options="store.addressesOptions" :show-all-levels="false" :props="{emitPath:false}" clearable/>
            <el-button type="info" @click="pasteFromClipboard" style="margin-left: 1%" plain>粘贴</el-button>
        </el-form-item>
        <el-form-item label="金额" prop="amount">
            <el-input v-model="ruleForm.amount" type="number" />
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
// import { useStore } from 'vuex'
import { store } from '../store.js'
import type { FormInstance, FormRules } from 'element-plus'
import { SendToAddress} from '../../../../../bindings/web3_gui/gui/server_api/sdkapi'
import {ElMessage} from "element-plus";
//import axios from 'axios';

interface RuleForm {
    toaddress:string
    pass: string
    amount: number
    gas: number

}
// const store = useStore()
const ruleFormRef = ref<FormInstance>()
const ruleForm = reactive<RuleForm>({
    toaddress:'',
    pass: '',
    amount:0,
    gas:0.1,
})

const rules = reactive<FormRules<RuleForm>>({
    toaddress: [
        {
            required: true,
            message: '地址不能为空',
            trigger: 'change',
        },
        { min: 30, max: 128, message: '地址长度不正确', trigger: 'change' },
    ],
    amount: [
        {
            required: true,
            message: '金额不能为空',
            trigger: 'blur',
        },
        {pattern: /^(0\.\d+|[1-9]\d*(\.\d+)?)$/,"message": "必须大于0", trigger: 'blur'}
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
            SendToAddress(store.getPeerInfo().id,store.selectedAddress,ruleForm.toaddress,ruleForm.pass,parseFloat(ruleForm.amount.toString()),parseFloat(ruleForm.gas.toString())).then((result) => {
                ElMessage.success("成功！")
            }).catch(error => {
                // 处理错误
                ElMessage.error(error)
            });
            // const data = {
            //     toaddress: ruleForm.toaddress,
            //     pass: ruleForm.pass,
            //     amount: ruleForm.amount,
            //     gas: ruleForm.gas
            // };
            //
            // try {
            //     const response = await axios.post('/api/form', data, {
            //         headers: {
            //             'Content-Type': 'application/json'
            //         }
            //     });
            //
            //     // 处理返回结果
            //     console.log('Response:', response.data);
            // } catch (error) {
            //     console.error('Error:', error);
            // }
            //

        } else {
            console.log('ruleForm.toaddress：', ruleForm.toaddress)
            console.log('error submit!', fields)
        }
    })
}

const resetForm = (formEl: FormInstance | undefined) => {
    if (!formEl) return
    formEl.resetFields()
}

const pasteFromClipboard = async () => {
    try {
        const text = await navigator.clipboard.readText()
        ruleForm.toaddress = text.trim()
        store.addressesOptions.push({value: ruleForm.toaddress,label: ruleForm.toaddress,children:[]})
    } catch (error) {
        console.error('无法从剪贴板中读取文本', error)
    }
}


</script>