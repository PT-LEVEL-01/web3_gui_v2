<template>
    
  <div class="about">
    <el-page-header @back="goBack" content="分发奖励">
    </el-page-header>

    <!-- <el-row style="margin-top:20px;">
      <el-col :span="8"><div class="grid-content bg-purple">社区节点奖励：</div></el-col>
      <el-col :span="8"><div class="grid-content bg-purple-light">{{DistributeRewards.CommunityReward}}</div></el-col>
    </el-row>
    <el-row>
      <el-col :span="8"><div class="grid-content bg-purple">轻节点奖励：</div></el-col>
      <el-col :span="8"><div class="grid-content bg-purple-light">{{DistributeRewards.LightReward}}</div></el-col>
    </el-row>
    <el-row>
      <el-col :span="8"><div class="grid-content bg-purple">所有轻节点数量：</div></el-col>
      <el-col :span="8"><div class="grid-content bg-purple-light">{{DistributeRewards.AllLight}}</div></el-col>
    </el-row>
    <el-row>
      <el-col :span="8"><div class="grid-content bg-purple">已经奖励的轻节点数量：</div></el-col>
      <el-col :span="8"><div class="grid-content bg-purple-light">{{DistributeRewards.RewardLight}}</div></el-col>
    </el-row> -->

    
<el-form :model="ruleForm" status-icon :rules="rules" ref="ruleForm" label-width="100px" class="demo-ruleForm"  style="margin-top:40px;">
  <el-form-item label="社区名称" prop="witnessName">
    {{ruleForm.witnessName}}
  </el-form-item>
  <el-form-item label="社区地址" prop="witnessAddr">
    {{ruleForm.witnessAddr}}
  </el-form-item>
  <el-form-item label="备注" prop="comment">
    <el-input v-model="ruleForm.comment"></el-input>
  </el-form-item>
  <el-form-item label="手续费" prop="gas">
    <el-input v-model="ruleForm.gas"></el-input>
  </el-form-item>
  <el-form-item label="密码" prop="pass">
    <el-input type="password" v-model="ruleForm.pass"></el-input>
  </el-form-item>
  <el-form-item>
    <el-button type="primary" @click="submitForm('ruleForm')">分发奖励</el-button>
    <el-button @click="resetForm('ruleForm')">重置</el-button>
  </el-form-item>
</el-form>

    <!-- <el-button type="primary" @click="distributeRewards()">分配奖励</el-button> -->
  </div>
</template>


<script setup>
import { ElMessage } from 'element-plus'
import { Chain_CommunityDistribute,Chain_GetCoinAddress,Chain_GetCommunityList,Chain_GetVoteList } from '../../../bindings/web3_gui/gui/server_api/sdkapi'
import {getCurrentInstance, reactive, ref} from 'vue'
import { store } from '../../store.js'

const thistemp = getCurrentInstance().appContext.config.globalProperties; //vue3获取当前this

const DistributeRewards = ref({
  CommunityReward:0,
})
const activeNames = ref(['1'])
const ruleForm = ref({
  addrs: new Array(),
  witnessAddr: store.communityVote.addr,
  witnessName: store.communityVote.payload,
  addr:"",
  gas:"",
  comment:"",
  pass: '',
})
const rules = ref({
  addr: [{required: true, message: '地址不能为空', trigger: 'change' }],
      gas: [{required: true, validator: thistemp.$checkAmountHaveZero, trigger: 'blur' }],
      comment: [{required: false, message: '', trigger: 'blur'}],
      pass: [{required: true, message: '密码不能为空', trigger: 'blur' }]
})

function goBack() {
  history.back(-1);
}

function distributeRewards(){

}

function handleChange(val) {
  // console.log(val);
  // var c = this.$BigNumber();
}

function submitForm(formName) {
  // console.log("11111111111")
  thistemp.$refs[formName].validate((valid) => {
    // console.log("22222222222")
    if (!valid){
      // console.log("验证不通过")
      return false;
    }
    // console.log("33333333333333")
    var gas = 0;
    if (ruleForm.value.gas > 0){
      gas = ruleForm.value.gas * store.coinCompany;
    }
    var amount = 0;
    if (ruleForm.value.amount > 0){
      amount = ruleForm.value.amount * store.coinCompany;
    }

    Promise.all([Chain_CommunityDistribute(store.communityVote.Applicant,gas,0,ruleForm.value.pass,
        ruleForm.value.comment)]).then(messages => {
      if(!messages || !messages[0]){return}
      var messageOne = messages[0];
      // console.log("分发奖励返回：",messageOne)
      var result = thistemp.$checkResultCode(messageOne.code)
      if(!result.success){
        ElMessage({
          showClose: true,
          message: "code:"+messageOne.code+" msg:"+result.error,
          type: 'error',
        })
        return
      }
      ElMessage({
        showClose: true,
        message: '成功',
        type: 'success',
      })
    });
    return

    // var params = JSON.parse(JSON.stringify(store.rpcParams));
    // //@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
    // params.data = {method:"votein",params:{votetype:1,address:ruleForm.value.addr, witness:ruleForm.value.witnessAddr,
    // amount:amount,gas:gas,pwd:ruleForm.value.pass,payload:ruleForm.value.comment}};
    // this.$axios(params).then((response)=> {
    //   // console.log("444444444444")
    //   console.log(response.data);
    //   // thistemp.$checkResultCode(response);
    //   if(thistemp.$checkResultCode(response)){
    //     // var flaginfo = (store.nameinfo == null) ? "创建域名成功" : "修改域名成功";
    //     this.$alert("质押成功", '成功', {
    //       confirmButtonText: '确定',
    //       type: 'success ',
    //       callback: action => {
    //       }
    //     });
    //   }
    // });
  });
}

function getBalance() {
  return new thistemp.$Calculator().divide(store.getinfo.balance, store.coinCompany);
}


function created(){
  // Promise.all([Chain_GetCoinAddress(), Chain_GetCommunityList(),Chain_GetVoteList(2)]).then(messages => {
  //   if(!messages || !messages[0]){return}
  //   var accountList = messages[0];
  //   console.log("地址列表",accountList)
  //   var messageOne = messages[1];
  //   console.log("社区列表",messageOne)
  //   //轻节点给社区投票的地址
  //   var voteList = messages[2];
  //   console.log("轻节点投票记录",voteList)
  //   var result = thistemp.$checkResultCode(messageOne.code)
  //   if(!result.success){
  //     ElMessage({
  //       showClose: true,
  //       message: "code:"+messageOne.code+" msg:"+result.error,
  //       type: 'error',
  //     })
  //     return
  //   }
  //   //社区自己发起分奖励
  //   for(var j=0; j<accountList.length; j++){
  //     var two = accountList[j];
  //     if(two.AddrCoin == ruleForm.value.witnessAddr){
  //       console.log("社区节点分奖励:")
  //       ruleForm.value.addr = two.AddrCoin
  //       return
  //     }
  //   }
  //   //轻节点发起分奖励
  //   for(var j=0; j<voteList.list.length; j++){
  //     var two = voteList.list[j];
  //     if(two.AddrSelf == ruleForm.value.witnessAddr){
  //       console.log("轻节点分奖励:")
  //       ruleForm.value.addr = two.AddrSelf
  //       return
  //     }
  //   }
  // });
}
created()
</script>

