<template>
  <div class="about">
    <el-row style="margin-top:20px;">
      <el-col :span="8"><div class="grid-content bg-purple">可用余额：</div></el-col>
      <el-col :span="8"><div class="grid-content bg-purple-light">{{getBalance()}}</div></el-col>
      <el-col :span="8"><div class="grid-content bg-purple">
        <router-link to="/index/wallet/pay"><el-button type="text">转账</el-button></router-link>
      </div></el-col>
    </el-row>
    <el-row>
      <el-col :span="8"><div class="grid-content bg-purple">冻结：</div></el-col>
      <el-col :span="8"><div class="grid-content bg-purple-light">{{getBalanceFrozen()}}</div></el-col>
      <el-col :span="8"><div class="grid-content bg-purple">
        <router-link to="/index/wallet/address"><el-button type="text">地址列表</el-button></router-link>
      </div></el-col>
    </el-row>
    <el-row>
      <el-col :span="8"><div class="grid-content bg-purple">锁定：</div></el-col>
      <el-col :span="8"><div class="grid-content bg-purple-light">{{getBalanceLockup()}}</div></el-col>
      <el-col :span="8"><div class="grid-content bg-purple">
        <router-link to="/index/wallet/createtoken"><el-button type="text">创建Token</el-button></router-link>
      </div></el-col>
    </el-row>

   <el-collapse @change="handleChange">
    <el-collapse-item v-for="(item, index) in store.chain_getinfo.TokenBalance" :key="index" :title="item.Symbol" :name="index">
        <div>{{item.TokenId}}</div>
        <p>可用余额：{{item.Balance}} <el-button type="text" @click="paytokeninfo(item)">转账</el-button>
          | 冻结：{{item.BalanceFrozen}} | 锁定：{{item.BalanceLockup}}</p>
    </el-collapse-item>
    </el-collapse>
  </div>
</template>


<script setup>
import { store } from '../../store.js'
import {getCurrentInstance, reactive, ref} from 'vue'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const activeNames = ref(['1'])

function handleChange(val) {
  // console.log(val);
  // var c = this.$BigNumber();
}

function paytokeninfo(tokeninfo) {
  // console.log(nameinfo)
  store.setPayTokeninfo(tokeninfo);
  thistemp.$router.push({path: '/index/wallet/pay'});
}

// 计算属性的 getter
function getBalance() {
  // var big = new thistemp.$Calculator();
  return new thistemp.$Calculator().divide(store.chain_getinfo.balance, store.coinCompany);

  // var bignumber = this.$BigNumber(store.chain_getinfo.balance);
  // return bignumber.dividedBy(100000000).toNumber();
}

function getBalanceFrozen() {
  // var big = new thistemp.$Calculator();
  return new thistemp.$Calculator().divide(store.chain_getinfo.BalanceFrozen, store.coinCompany);
  // var bignumber = this.$BigNumber(store.chain_getinfo.BalanceFrozen);
  // return bignumber.dividedBy(100000000).toNumber();
}

function getBalanceLockup() {
  // var big = new thistemp.$Calculator();
  return new thistemp.$Calculator().divide(store.chain_getinfo.BalanceLockup, store.coinCompany);
  // var bignumber = this.$BigNumber(store.chain_getinfo.BalanceLockup);
  // return bignumber.dividedBy(100000000).toNumber();
}


</script>

