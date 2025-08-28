package sol

const SOURCE_SOL = `
/*
 * Copyright 2014-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * */
 
pragma solidity ^0.4.25;


library LibAddressSet {

    struct AddressSet {
        mapping (address => uint256) indexMapping;
        address[] values;
    }

    function add(AddressSet storage self, address value) internal {
        require(value != address(0x0), "LibAddressSet: value can't be 0x0");
        require(!contains(self, value), "LibAddressSet: value already exists in the set.");
        self.values.push(value);
        self.indexMapping[value] = self.values.length;
    }

    function contains(AddressSet storage self, address value) internal view returns (bool) {
        return self.indexMapping[value] != 0;
    }

    function remove(AddressSet storage self, address value) internal {
        require(contains(self, value), "LibAddressSet: value doesn't exist.");
        uint256 toDeleteindexMapping = self.indexMapping[value] - 1;
        uint256 lastindexMapping = self.values.length - 1;
        address lastValue = self.values[lastindexMapping];
        self.values[toDeleteindexMapping] = lastValue;
        self.indexMapping[lastValue] = toDeleteindexMapping + 1;
        delete self.indexMapping[value];
        self.values.length--;
    }

    function getSize(AddressSet storage self) internal view returns (uint256) {
        return self.values.length;
    }

    function get(AddressSet storage self, uint256 index) internal view returns (address){
        return self.values[index];
    }

    function getAll(AddressSet storage self) internal view returns(address[] memory) {
    	address[] memory output = new address[](self.values.length);
    	for (uint256 i; i < self.values.length; i++){
            output[i] = self.values[i];
        }
        return output;
    }


}

// File: contracts/reward_contract/br.sol

pragma solidity >=0.4.24 <0.6.11;
pragma experimental ABIEncoderV2;

import "./LibAddressSet.sol";

contract DistributeReward {
    event LogRewardHistory(
        address indexed into,
        address indexed from,
        uint256 reward,
        string name,
        uint8 utype
    );
    event WithDrawReward(address indexed addr, uint256 reward);
    using LibAddressSet for LibAddressSet.AddressSet;

    //中间变量记录地址集合
    LibAddressSet.AddressSet tempSet;

    LibAddressSet.AddressSet allCommunity;
    LibAddressSet.AddressSet allLight;

    struct Light {
        address addr;
        address c;
        uint256 vote;
        uint256 blockHeight;
        string name;
    }
    struct LightOut {
        address addr;
        address c;
        uint256 vote;
        uint256 score;
        string cname;
        uint256 blockHeight;
        string name;
    }
    //轻节点映射，用addr是否等于零地址来判断是否是轻节点
    mapping(address => Light) public lightMap;
    struct Community {
        address addr;
        address wit;
        uint256 vote;
        string name;
        uint256 blockHeight;
        LibAddressSet.AddressSet lights;
    }
    struct CommunityOut {
        address addr;
        address wit;
        uint256 vote;
        string name;
        uint256 score;
        uint256 blockHeight;
    }
    struct CommunityRewardOut {
        address addr;
        address wit;
        uint256 vote;
        string name;
        uint256 score;
        uint256 blockHeight;
        uint256 reward;
        uint8 rate;
        uint256 lightCount;
    }
    struct WitnessInfoOut {
        address addr;
        address wit;
        uint8 rate;
        uint256 reward;
        uint256 remainReward;
        CommunityRewardOut[] communitys;
    }
    //社区节点映射，用addr等于零地址判断是否曾经是社区节点。用wit等于零地址判断是否质押社区节点
    mapping(address => Community) communityMap;
    struct Witness {
        address addr;
        uint256 vote;
        LibAddressSet.AddressSet communities;
    }
    //见证人映射
    mapping(address => Witness) witnessMap;
    //历史奖励总记录
    mapping(address => uint256) public allReward;
    //剩余未提现的奖励
    mapping(address => uint256) public remainReward;
    //见证人节点奖励
    mapping(address => uint256) public allWReward;
    mapping(address => uint256) public remainWReward;
    //社区节点奖励
    mapping(address => uint256) public allCReward;
    mapping(address => uint256) public remainCReward;
    //轻节点奖励
    mapping(address => uint256) public allLReward;
    mapping(address => uint256) public remainLReward;
    //见证人节点和社区节点的分配比例
    mapping(address => uint8) public addrRate;
    //节点押金
    mapping(address => uint256) public deposit;

    //质押成功社区节点
    function addCommunity(
        address w,
        uint8 rate,
        string name
    ) public payable {
        require(lightMap[msg.sender].addr == address(0x0), "address  is light");
        require(
            communityMap[msg.sender].wit == address(0x0),
            "address has a witness"
        );
        witnessMap[w].vote += communityMap[msg.sender].vote;
        if (!witnessMap[w].communities.contains(msg.sender)) {
            witnessMap[w].communities.add(msg.sender);
        }
        communityMap[msg.sender].wit = w;
        communityMap[msg.sender].addr = msg.sender;
        communityMap[msg.sender].name = name;
        communityMap[msg.sender].blockHeight = block.number;
        addrRate[msg.sender] = rate;
        deposit[msg.sender] = msg.value;
        allCommunity.add(msg.sender);
    }

    //取消质押
    function delCommunity() public {
        require(
            communityMap[msg.sender].wit != address(0x0),
            "address must be community"
        );
        require(
            communityMap[msg.sender].lights.getSize() > 0,
            "community must be remove all lights"
        );
        require(
            remainCReward[msg.sender] != 0,
            "community reward must be 0"
        );
        if (communityMap[msg.sender].wit != address(0x0)) {
            address w = communityMap[msg.sender].wit;
            //解除映射关系，减去票数
            witnessMap[w].vote -= communityMap[msg.sender].vote;
            witnessMap[w].communities.remove(msg.sender);
            //设置社区节点的见证人为空。
            communityMap[msg.sender].wit = address(0x0);
            uint256 t = deposit[msg.sender];
            deposit[msg.sender] = 0;
            allCommunity.remove(msg.sender);
            //退钱
            msg.sender.transfer(t);
        }
    }

    function setRate(uint8 rate) public {
        addrRate[msg.sender] = rate;
    }

    function getRate() public view returns (uint8) {
        return addrRate[msg.sender];
    }

    //添加投票
    function addVote(address c) public payable {
        require(
            lightMap[msg.sender].addr != address(0x0),
            "address must be is light"
        );
        require(
            communityMap[msg.sender].wit == address(0x0),
            "address is community"
        );
        require(
            lightMap[msg.sender].c == address(0x0) ||
                lightMap[msg.sender].c == c,
            "cannot vote for  multiple community nodes"
        );
        //先更新社区节点数据
        communityMap[c].vote += msg.value;
        //将轻节点记录下来
        if (!communityMap[c].lights.contains(msg.sender)) {
            communityMap[c].lights.add(msg.sender);
        }
        lightMap[msg.sender].c = c;
        lightMap[msg.sender].vote += msg.value;
        //见证人节点的票数更新
        address w = communityMap[c].wit;
        if (w != address(0x0)) {
            if (witnessMap[w].communities.contains(c)) {
                witnessMap[w].vote += msg.value;
            }
        }
        deposit[msg.sender] += msg.value;
    }

    //取消投票
    function delVote(uint256 vote) public {
        //4666666666666666666666666666
        require(
            lightMap[msg.sender].vote >= vote,
            "cancel vote must be less your vote"
        );
        address c = lightMap[msg.sender].c;
        require(c != address(0x0), "address must has vote community");
        if (c != address(0x0)) {
            //communityMap[c].vote -= lightMap[msg.sender].vote;
            communityMap[c].vote -= vote; //0
            //票数全部取消完
            if (vote == lightMap[msg.sender].vote) {
                communityMap[c].lights.remove(msg.sender);
                lightMap[msg.sender].c = address(0x0);
            }

            if (communityMap[c].wit != address(0x0)) {
                address w = communityMap[c].wit;
                witnessMap[w].vote -= vote;
            }
            deposit[msg.sender] -= vote;
            lightMap[msg.sender].vote -= vote;
            msg.sender.transfer(vote);
        }
    }

    //添加轻节点
    function addLight(string name) public payable {
        require(
            communityMap[msg.sender].wit == address(0x0),
            "address is community"
        );
        require(
            lightMap[msg.sender].addr == address(0x0),
            "address already a light"
        );
        lightMap[msg.sender].addr = msg.sender;
        lightMap[msg.sender].blockHeight = block.number;
        lightMap[msg.sender].name = name;
        deposit[msg.sender] += msg.value;
        allLight.add(msg.sender);
    }

    //移除轻节点
    function delLight() public {
        require(lightMap[msg.sender].vote == 0, "address must be no vote");
        require(
            lightMap[msg.sender].addr != address(0x0),
            "address must be light"
        );
        if (lightMap[msg.sender].addr != address(0x0)) {
            uint256 t = deposit[msg.sender];
            deposit[msg.sender] = 0;
            lightMap[msg.sender].addr = address(0x0);
            allLight.remove(msg.sender);
            msg.sender.transfer(t);
        }
    }

    //给轻节点发奖励
    function _disLight(Community c, uint256 totalReward) internal {
        tempSet = c.lights;
        address[] memory list = tempSet.getAll();
        for (uint256 i = 0; i < list.length; i++) {
            if (lightMap[list[i]].vote > 0) {
                uint256 reward = (totalReward * lightMap[list[i]].vote) /
                    c.vote;
                //记录下来即可
                allReward[list[i]] += reward;
                remainReward[list[i]] += reward;
                allLReward[list[i]] += reward;
                remainLReward[list[i]] += reward;
                emit LogRewardHistory(
                    list[i],
                    c.addr,
                    reward,
                    lightMap[list[i]].name,
                    3
                );
            }
        }
    }

    //给社区节点发奖励
    function _disCommunity(Witness w, uint256 totalReward) internal {
        tempSet = w.communities;
        address[] memory list = tempSet.getAll();
        for (uint256 i = 0; i < list.length; i++) {
            if (communityMap[list[i]].vote > 0) {
                uint256 reward = (totalReward * communityMap[list[i]].vote) /
                    w.vote;
                uint256 disReward = (reward * addrRate[list[i]]) / 100;
                _disLight(communityMap[list[i]], disReward);
                allReward[list[i]] += reward - disReward;
                remainReward[list[i]] += reward - disReward;
                allCReward[list[i]] += reward - disReward;
                remainCReward[list[i]] += reward - disReward;
                emit LogRewardHistory(
                    list[i],
                    communityMap[list[i]].wit,
                    reward - disReward,
                    communityMap[list[i]].name,
                    2
                );
            }
        }
    }

    //分发奖励
    function distribute(
        address[] list,
        uint256 index,
        uint256 totalReward
    ) public {
        //总票数
        uint256 totalVote = 0;
        //记录有投票的见证人最大索引
        uint256 maxIndex;
        for (uint256 i = 0; i < list.length; i++) {
            if (witnessMap[list[i]].vote > 0) {
                totalVote += witnessMap[list[i]].vote;
                maxIndex = i;
            }
        }

        //30%平分,70%按票数权重分配
        uint256 reward30 = (totalReward * 30) / 100;
        uint256 reward70 = totalReward - reward30;

        //70%分配进度
        uint256 reward70use = 0;
        for (uint256 j = 0; j < list.length; j++) {
            //平分的奖励
            uint256 reward = reward30 / list.length;
            //如果是最后一个
            if (j == list.length - 1) {
                reward += reward30 % list.length;
            }
            //如果是属于出块的人
            uint256 reward2 = 0;

            if (j < index) {
                if (totalVote > 0) {
                    reward2 = (reward70 * witnessMap[list[j]].vote) / totalVote;
                    reward70use += reward2;
                    if (j == maxIndex) {
                        reward2 += reward70 - reward70use;
                    }
                } else {
                    //平分这些奖励
                    reward2 = reward70 / index;
                    if (j == index - 1) {
                        reward2 += reward70 % index;
                    }
                }
            }
            reward += reward2;
            //计算我要分出去的提成
            //判断是否有投票，如果有则计算分配的比例，没有则全部给当前见证人
            uint256 disReward = (reward * addrRate[list[j]]) / 100;
            if (witnessMap[list[j]].vote == 0) {
                disReward = 0;
            }
            if (disReward > 0) {
                _disCommunity(witnessMap[list[j]], disReward);
            }
            if ((reward - disReward) > 0) {
                //list[j].transfer(reward-disReward);

                allReward[list[j]] += reward - disReward;
                remainReward[list[j]] += reward - disReward;
                allWReward[list[j]] += reward - disReward;
                remainWReward[list[j]] += reward - disReward;
                emit LogRewardHistory(
                    list[j],
                    address(0x0),
                    reward - disReward,
                    "",
                    1
                );
            }
        }
    }

    //查询总奖励
    function queryAllReward(address addr) public view returns (uint256) {
        return allReward[addr];
    }

    function getWiteVote(address addr) public view returns (uint256) {
        return witnessMap[addr].vote;
    }

    function getCommVote(address addr) public view returns (uint256) {
        return communityMap[addr].vote;
    }

    function getAddrState(address addr) public view returns (uint256) {
        if (lightMap[addr].addr != address(0x0)) {
            return 3;
        }
        if (communityMap[addr].wit != address(0x0)) {
            return 2;
        }
        return 4;
    }

    //获取所有的社区节点
    function getAllCommunity() public view returns (CommunityOut[]) {
        address[] memory list = allCommunity.getAll();
        return getCommunityList(list);
    }

    function getCommunityList(address[] list)
        public
        view
        returns (CommunityOut[])
    {
        CommunityOut[] memory r = new CommunityOut[](list.length);
        for (uint256 i = 0; i < list.length; i++) {
            r[i] = CommunityOut(
                communityMap[list[i]].addr,
                communityMap[list[i]].wit,
                communityMap[list[i]].vote,
                communityMap[list[i]].name,
                deposit[list[i]],
                communityMap[list[i]].blockHeight
            );
        }
        return r;
    }

    function getCommunityRewardList()
        public
        view
        returns (CommunityRewardOut[])
    {
        address[] memory list = allCommunity.getAll();
        CommunityRewardOut[] memory r = new CommunityRewardOut[](list.length);
        for (uint256 i = 0; i < list.length; i++) {
            r[i] = CommunityRewardOut(
                communityMap[list[i]].addr,
                communityMap[list[i]].wit,
                communityMap[list[i]].vote,
                communityMap[list[i]].name,
                deposit[list[i]],
                communityMap[list[i]].blockHeight,
                allCReward[list[i]],
                addrRate[list[i]],
                communityMap[list[i]].lights.getSize()
            );
        }
        return r;
    }

    function getLightList(address[] list) public view returns (LightOut[]) {
        LightOut[] memory r = new LightOut[](list.length);
        for (uint256 i = 0; i < list.length; i++) {
            r[i] = LightOut(
                lightMap[list[i]].addr,
                lightMap[list[i]].c,
                lightMap[list[i]].vote,
                deposit[list[i]],
                communityMap[lightMap[list[i]].c].name,
                lightMap[list[i]].blockHeight,
                lightMap[list[i]].name
            );
        }
        return r;
    }

    function getCommunityListByWit(address w)
        public
        view
        returns (CommunityOut[])
    {
        tempSet = witnessMap[w].communities;
        address[] memory list = tempSet.getAll();
        return getCommunityList(list);
    }

    function getLightListByC(address c) public view returns (LightOut[]) {
        tempSet = communityMap[c].lights;
        address[] memory list = tempSet.getAll();
        return getLightList(list);
    }

    //获取所有的轻节点
    function getAllLight(uint256 start, uint256 end)
        public
        view
        returns (LightOut[])
    {
        address[] memory list = allLight.getAll();
        require(start <= list.length, "start is too big");
        if (end > list.length) {
            end = list.length;
        }
        address[] resList;
        for (uint256 i = start; i < end; i++) {
            // resList[i-start] = list[i];
            resList.push(list[i]);
        }
        return getLightList(resList);
    }

    //提取社区节点奖励
    function withDrawC(uint256 _value) public {
        //判断是否有余额，判断余额是否充足
        require(
            remainCReward[msg.sender] >= _value,
            "amount must be less or equal your remain reward"
        );
        require(address(this).balance >= _value, "amount is illegal");
        remainCReward[msg.sender] -= _value;
        msg.sender.transfer(_value);
        emit WithDrawReward(msg.sender, _value);
    }

    //提取轻节点奖励
    function withDrawL(uint256 _value) public {
        //判断是否有余额，判断余额是否充足
        require(
            remainLReward[msg.sender] >= _value,
            "amount must be less or equal your remain reward"
        );
        require(address(this).balance >= _value, "amount is illegal");
        remainLReward[msg.sender] -= _value;
        msg.sender.transfer(_value);
        emit WithDrawReward(msg.sender, _value);
    }

    //提取见证人奖励
    function withDrawW(uint256 _value) public {
        //判断是否有余额，判断余额是否充足
        require(
            remainWReward[msg.sender] >= _value,
            "amount must be less or equal your remain reward"
        );
        require(address(this).balance >= _value, "amount is illegal");
        remainWReward[msg.sender] -= _value;
        msg.sender.transfer(_value);
        emit WithDrawReward(msg.sender, _value);
    }

    function getWitByL(address l) public view returns (address) {
        address c = lightMap[l].c;
        return communityMap[c].wit;
    }

    function getWitByC(address c) public view returns (address) {
        return communityMap[c].wit;
    }

    function getLightTotal() public view returns (uint256) {
        return allLight.getSize();
    }

    function getRoleTotal() public view returns (uint256, uint256) {
        return (allCommunity.getSize(), allLight.getSize());
    }

    function getRateAndVoteByAddrs(address[] addrs)
        public
        view
        returns (uint8[], uint256[])
    {
        uint8[] memory rates = new uint8[](addrs.length);
        uint256[] memory votes = new uint256[](addrs.length);

        for (uint256 i = 0; i < addrs.length; i++) {
            rates[i] = addrRate[addrs[i]];
            votes[i] = witnessMap[addrs[i]].vote;
        }
        return (rates, votes);
    }

    function getWitnessInfo(address w) public view returns (WitnessInfoOut) {
        uint8 ratio = addrRate[w];
        uint256 reward = allWReward[w];
        uint256 remainReward = remainWReward[w];

        tempSet = witnessMap[w].communities;

        address[] memory list = tempSet.getAll();

        CommunityRewardOut[] memory communitys = new CommunityRewardOut[](
            list.length
        );

        for (uint256 i = 0; i < list.length; i++) {
            communitys[i] = CommunityRewardOut(
                communityMap[list[i]].addr,
                communityMap[list[i]].wit,
                communityMap[list[i]].vote,
                communityMap[list[i]].name,
                deposit[list[i]],
                communityMap[list[i]].blockHeight,
                allReward[list[i]],
                addrRate[list[i]],
                communityMap[list[i]].lights.getSize()
            );
        }
        return WitnessInfoOut(w, w, ratio, reward, remainReward, communitys);
    }
}`
