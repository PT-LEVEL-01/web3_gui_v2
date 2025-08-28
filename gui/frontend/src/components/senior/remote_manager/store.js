import { reactive } from 'vue'

export const store = reactive({
    // nav_badge:new Map(),//左侧导航栏是否显示徽章提示
    // nav_showModules:"",//左侧导航栏目前显示的模块
    //
    // count: 0,
    peerInfo:null,
    peerInfoIndex:-1,//选中的节点的index
    selectedAddress:"",//选中的地址
    peerList:[],
    addressesOptions:[],
    drawer:false,
    names:[],
    getPeerInfo() {
        return this.peerList[this.peerInfoIndex]
    },
    getSelectedAddressInfo(){
        if (this.getPeerInfo.addresses && this.selectedAddress) {
            return this.getPeerInfo.addresses.find(item => item.address === this.selectedAddress)
        }
        return {
            index: -1,
            address: '',
            balance: 0,
            balance_frozen: 0,
            balance_lockup: 0,
            peer_type: 0,
        }
    },
    setAddressesOptions() {
        const options = [];

        // 创建一个 Map 对象，用于存储分组及其对应的节点
        const groupMap = new Map();

        this.peerList.forEach(peer => {
            if (peer.addresses && peer.addresses.length > 0) {
                const group = peer.group;

                // 如果当前分组不存在于 groupMap 中，则新建一个分组
                if (!groupMap.has(group)) {
                    groupMap.set(group, {
                        value: group,
                        label: group,
                        children: []
                    });
                }

                const node = {
                    value: peer.name,
                    label: peer.name,
                    children: []
                };

                const addresses = peer.addresses.map(address => {
                    return {
                        value: address.address,
                        label: address.address,
                        children: []
                    };
                });

                node.children = addresses;
                groupMap.get(group).children.push(node);
            }
        });

        // 将 groupMap 中的分组添加到 options 数组中
        groupMap.forEach(group => {
            options.push(group);
        });

        this.addressesOptions = options;
    }
})