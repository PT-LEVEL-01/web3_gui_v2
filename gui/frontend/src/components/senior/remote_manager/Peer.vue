<template>
    <el-descriptions
            title="节点信息"
            :column="2"
            border
    >
        <el-descriptions-item>
            <template #label>
                <div class="cell-item">
                    节点：
                </div>
            </template>
            <el-text tag="b">{{store.getters.getPeerInfo.group}}:</el-text> {{store.getters.getPeerInfo.name}}
        </el-descriptions-item>
        <el-descriptions-item>
            <template #label>
                <div class="cell-item">
                    角色：
                </div>
            </template>
            <el-tag :type="getTagType(store.getters.getSelectedAddressInfo.peer_type)">{{ getTagText(store.getters.getSelectedAddressInfo.peer_type)}}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item>
            <template #label>
                <div class="cell-item">
                    当前块高度：
                </div>
            </template>
            <el-text class="mx-1" type="primary">{{store.getters.getPeerInfo.current_block}}</el-text>/<el-text class="mx-1" type="success" tag="b">{{store.getters.getPeerInfo.highest_block}}</el-text>
        </el-descriptions-item>
        <el-descriptions-item>
            <template #label>
                <div class="cell-item">
                    可用余额：
                </div>
            </template>
            <el-text class="mx-1" type="success">{{store.getters.getSelectedAddressInfo.balance}}</el-text>
        </el-descriptions-item>
        <el-descriptions-item>
            <template #label>
                <div class="cell-item">
                    冻结金额：
                </div>
            </template>
            <el-text class="mx-1" type="info">{{store.getters.getSelectedAddressInfo.balance_frozen}}</el-text>
        </el-descriptions-item>
        <el-descriptions-item>
            <template #label>
                <div class="cell-item">
                    锁仓金额：
                </div>
            </template>
            <el-text class="mx-1" type="info">{{store.getters.getSelectedAddressInfo.balance_lockup}}</el-text>
        </el-descriptions-item>
        <el-descriptions-item>
            <template #label>
                <div class="cell-item">
                    Address:
                </div>
            </template>
            <el-select v-model="store.state.selectedAddress" placeholder="Select">
                <el-option
                    v-for="(item,index) in store.getters.getPeerInfo.addresses"
                    :key="index"
                    :label="item.address"
                    :value="item.address"
                />
            </el-select>
            <el-button type="info" @click="copyAddressToClipboard" style="margin-left: 1%" plain>复制</el-button>
        </el-descriptions-item>
    </el-descriptions>

    <div class="demo-collapse">
        <el-collapse v-model="activeName" accordion @change="getNamesList">
            <el-collapse-item title="新建地址" name="1">
                <NewAddr/>
            </el-collapse-item>
            <el-collapse-item title="转帐" name="2">
                <SendToAddress/>
            </el-collapse-item>
            <el-collapse-item title="见证人质押" name="3">
                <DepositIn/>
            </el-collapse-item>
            <el-collapse-item title="取消见证人质押" name="4">
                <DepositOut/>
            </el-collapse-item>
            <el-collapse-item title="社区/轻节点质押，投票" name="5">
                <VoteIn/>
            </el-collapse-item>
            <el-collapse-item title="取消社区/轻节点质押，投票" name="6">
                <VoteOut/>
            </el-collapse-item>
            <el-collapse-item title="社区手动分账" name="7">
                <CommunityDistribute/>
            </el-collapse-item>
          <el-collapse-item title="注册域名" name="8">
                <NameIn/>
            </el-collapse-item>
          <el-collapse-item title="域名注销" name="9">
                <NameOut/>
          </el-collapse-item>
          <el-collapse-item title="域名列表" name="10" >
                <GetName/>
            </el-collapse-item>
        </el-collapse>
    </div>

</template>
<script lang="ts" setup>
import { ref } from 'vue'
// import { useStore } from 'vuex'
import { store } from './store.js'
import NewAddr from "./peeroperates/NewAddr.vue";
import SendToAddress from "./peeroperates/SendToAddress.vue";
import DepositIn from "./peeroperates/DepositIn.vue";
import DepositOut from "./peeroperates/DepositOut.vue";
import VoteIn from "./peeroperates/VoteIn.vue";
import VoteOut from "./peeroperates/VoteOut.vue";
import CommunityDistribute from "./peeroperates/CommunityDistribute.vue";
import {ElMessage} from "element-plus";
import NameIn from "./peeroperates/NameIn.vue";
import GetName from "./peeroperates/GetNames.vue";
import NameOut from "./peeroperates/NameOut.vue";
import { GetNames} from '../../../../bindings/web3_gui/gui/server_api/sdkapi'
const activeName = ref('1')

// const store = useStore()
const copyAddressToClipboard=()=>{
    if (store.selectedAddress) {
        navigator.clipboard.writeText(store.selectedAddress)
        ElMessage.success('已复制到剪贴板')
    } else {
        ElMessage.warning('请选择一个选项')
    }
}

const getTagType = (type: number) => {
    if (type === 1) {
        return 'success';
    } else if (type === 2) {
        return 'danger';
    } else if (type === 3) {
        return 'warning';
    } else {
        return 'info';
    }
}
const getTagText = (type: number) => {
    if (type === 1) {
        return '见证人';
    } else if (type === 2) {
        return '社区节点';
    } else if (type === 3) {
        return '轻节点';
    } else {
        return '普通节点';
    }
}

let submit = false
const getNamesList = (type) => {
  if (type == 10){
    store.names = []

    if (store.getPeerInfo().status == false){
      return
    }

    if (submit == true){
      console.log("请求频繁")
      return
    }
    submit = true
    GetNames(store.getPeerInfo().id).then((result) => {
      store.names = result
      submit = false
    }).catch(error => {
      // 处理错误
      ElMessage.error(error)
      submit = false
    });


  }

}
</script>
<style scoped>
.cell-item {
    display: flex;
    align-items: center;
}
</style>
