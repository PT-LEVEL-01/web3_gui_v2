import { createRouter, createWebHistory } from 'vue-router'
// import Home from './views/login.vue'
// import About from './views/about.vue'

const routes = [
    {
        path: '/',
        name: 'home',
        component:  () => import('./components/login.vue')
    },
    // {
    //     path: '/about',
    //     name: 'About',
    //     component:  () => import('./components/HelloWorld.vue')
    // },
    // {
    //   path: '/importprivatekey',
    //   name: 'importprivatekey',
    //   component: () => import('./components/import_privatekey.vue')
    // },
    // {
    //   path: '/remoteconn',
    //   name: 'RemoteConn',
    //   component: () => import('./components/remote_conn.vue')
    // },
    // {
    //   path: '/senior',
    //   component: () => import('./components/senior/senior_page.vue'),
    //   children: [
    //     {
    //       path: 'nav',
    //       component: () => import('./components/senior/senior_nav.vue'),
    //     },
    //     {
    //       path: 'login_private_network',
    //       component: () => import('./components/senior/login_private_network.vue'),
    //     },
    //     {
    //       path: 'node_manager',
    //       component: () => import('./components/senior/remote_manager/node_list.vue'),
    //     },
    //   ]
    // },
    // {
    //   path: '/index',
    //   name: 'Index',
    //   component: () => import('./components/nav_left.vue'),
    //   children: [
    //     {
    //       path: 'wallet',
    //       component: () => import('./components/chain/wallet_nav.vue'),
    //       children: [
    //         {
    //           path: 'info',
    //           component: () => import('./components/chain/wallet_info.vue')
    //         },
    //         {
    //           path: 'pay',
    //           component: () => import('./components/chain/pay.vue')
    //         },
    //         {
    //           path: 'address',
    //           component: () => import('./components/chain/address.vue')
    //         },
    //         {
    //           path: 'addressadd',
    //           component: () => import('./components/chain/address_add.vue')
    //         },
    //         {
    //           path: 'paylog',
    //           component: () => import('./components/chain/paylog.vue')
    //         },
    //         {
    //           path: 'createtoken',
    //           component: () => import('./components/chain/token_create.vue')
    //         },
    //         {
    //           path: 'name',
    //           component: () => import('./components/chain/name.vue')
    //         },
    //         {
    //           path: 'namereg',
    //           component: () => import('./components/chain/name_reg.vue')
    //         },
    //         {
    //           path: 'namedestroy',
    //           component: () => import('./components/chain/name_destroy.vue')
    //         },
    //         {
    //           path: 'witnessinfo',
    //           component: () => import('./components/chain/witness_info.vue')
    //         },
    //         {
    //           path: 'witnessdepositin',
    //           component: () => import('./components/chain/witness_deposit_in.vue')
    //         },
    //         {
    //           path: 'witnessdepositout',
    //           component: () => import('./components/chain/witness_deposit_out.vue')
    //         },
    //         {
    //           path: 'witnesslist',
    //           component: () => import('./components/chain/witness_list.vue')
    //         },
    //         {
    //           path: 'witnessscorein',
    //           component: () => import('./components/chain/witness_score_in.vue')
    //         },
    //         {
    //           path: 'witnessscoreout',
    //           component: () => import('./components/chain/witness_score_out.vue')
    //         },
    //         {
    //           path: 'selfcommunitydepositlist',
    //           component: () => import('./components/chain/self_community_deposit_list.vue')
    //         },
    //         {
    //           path: 'communitylist',
    //           component: () => import('./components/chain/community_list.vue')
    //         },
    //         {
    //           path: 'communityvotein',
    //           component: () => import('./components/chain/community_vote_in.vue')
    //         },
    //         {
    //           path: 'communityvoteout',
    //           component: () => import('./components/chain/community_vote_out.vue')
    //         },
    //         {
    //           path: 'selflightvotelist',
    //           component: () => import('./components/chain/self_light_vote_list.vue')
    //         },
    //         {
    //           path: 'selflightlist',
    //           component: () => import('./components/chain/self_light_list.vue')
    //         },
    //         {
    //           path: 'lightdepositin',
    //           component: () => import('./components/chain/light_deposit_in.vue')
    //         },
    //         {
    //           path: 'lightdepositout',
    //           component: () => import('./components/chain/light_deposit_out.vue')
    //         },
    //         {
    //           path: 'lightvoteout',
    //           component: () => import('./components/chain/light_vote_out.vue')
    //         },
    //         {
    //           path: 'communitydistributerewards',
    //           component: () => import('./components/chain/community_distribute_rewards.vue')
    //         },
    //         {
    //           path: 'exportkey',
    //           component: () => import('./components/chain/key_export.vue')
    //         },
    //       ]
    //     },
    //     {
    //       path: 'im',
    //       component: () => import('./components/im/im_layout.vue'),
    //       children: [
    //         {
    //           path: 'index',
    //           component: () => import('./components/im/im_index.vue')
    //         },
    //         {
    //           path: 'createGroup',
    //           component: () => import('./components/im/im_group_create.vue')
    //         },
    //         {
    //           path: 'imProxyList',
    //           component: () => import('./components/im/im_proxy_list.vue')
    //         },
    //         {
    //           path: 'imProxySetup',
    //           component: () => import('./components/im/im_proxy_setup.vue')
    //         },
    //         {
    //           path: 'imProxyOrder',
    //           component: () => import('./components/im/im_proxy_order.vue')
    //         },
    //         {
    //           path: 'imProxyOrderList',
    //           component: () => import('./components/im/im_proxy_order_list.vue')
    //         },
    //         {
    //           path: 'addlist',
    //           component: () => import('./components/im/im_add_list.vue')
    //         },
    //         {
    //           path: 'message',
    //           component: () => import('./components/im/im_message_content.vue')
    //         },
    //         {
    //           path: 'message_preview',
    //           component: () => import('./components/im/im_message_content_preview.vue')
    //         },
    //         {
    //           path: 'search',
    //           component: () => import('./components/im/im_search.vue')
    //         },
    //         {
    //           path: 'userinfo',
    //           component: () => import('./components/im/im_userinfo.vue')
    //         },
    //         {
    //           path: 'sharebox',
    //           component: () => import('./components/im/im_sharebox.vue')
    //         }
    //       ]
    //     },
    //     {
    //       path: 'circle',
    //       component: () => import('./components/circle/circle_nav.vue'),
    //       children: [
    //         {
    //           path: 'index',
    //           component: () => import('./components/circle/circle_index.vue')
    //         },
    //         {
    //           path: 'editor',
    //           component: () => import('./components/circle/circle_editor.vue')
    //         },
    //         {
    //           path: 'show',
    //           component: () => import('./components/circle/circle_show.vue')
    //         },
    //         {
    //           path: 'newsContent',
    //           component: () => import('./components/circle/circle_news_content.vue')
    //         },
    //         {
    //           path: 'newsTemp',
    //           component: () => import('./components/circle/circle_news_temp.vue')
    //         },
    //         {
    //           path: 'classManager',
    //           component: () => import('./components/circle/circle_news_class_manager.vue')
    //         },
    //         {
    //           path: 'newsRelease',
    //           component: () => import('./components/circle/circle_news_release.vue')
    //         },
    //         {
    //           path: 'newsDraft',
    //           component: () => import('./components/circle/circle_news_draft.vue')
    //         },
    //       ]
    //     },
    //     {
    //       path: 'files',
    //       component: () => import('./components/files/files_nav.vue'),
    //       children: [
    //         {
    //           path: 'sharebox',
    //           component: () => import('./components/files/files_sharebox.vue')
    //         },
    //         {
    //           path: 'sharebox_price',
    //           component: () => import('./components/files/files_sharebox_price.vue')
    //         },
    //         {
    //           path: 'spacesmining',
    //           component: () => import('./components/files/spaces_mining.vue')
    //         },
    //         {
    //           path: 'spacesminingadd',
    //           component: () => import('./components/files/spaces_mining_add.vue')
    //         },
    //         {
    //           path: 'spacessub',
    //           component: () => import('./components/files/spaces_sub.vue')
    //         },
    //         {
    //           path: 'store',
    //           component: () => import('./components/files/store.vue')
    //         },
    //         {
    //           path: 'download',
    //           component: () => import('./components/files/files_download_list.vue')
    //         },
    //         {
    //           path: 'download_finish',
    //           component: () => import('./components/files/files_download_finish.vue')
    //         },
    //         {
    //           path: 'upload_list',
    //           component: () => import('./components/files/files_upload_list.vue')
    //         },
    //         {
    //           path: 'storage_server',
    //           component: () => import('./components/files/files_storage_server_seting.vue')
    //         },
    //         {
    //           path: 'storage_client',
    //           component: () => import('./components/files/files_storage_client.vue')
    //         },
    //         {
    //           path: 'storage_client_search',
    //           component: () => import('./components/files/files_storage_client_search.vue')
    //         },
    //         {
    //           path: 'storage_server_info',
    //           component: () => import('./components/files/files_storage_server_info.vue')
    //         },
    //         {
    //           path: 'orders',
    //           component: () => import('./components/files/files_orders.vue')
    //         },
    //         {
    //           path: 'client_filelist',
    //           component: () => import('./components/files/files_storage_client_filelist.vue')
    //         },
    //         {
    //           path: 'client_uploadlist',
    //           component: () => import('./components/files/files_storage_client_upload_list.vue')
    //         },
    //       ]
    //     },
    //     {
    //       path: 'setup',
    //       component: () => import('./components/setup/setup_nav.vue'),
    //       children: [
    //         {
    //           path: 'account',
    //           component: () => import('./components/setup/setup_account.vue')
    //         },
    //         {
    //           path: 'head',
    //           component: () => import('./components/setup/setup_head.vue')
    //         },
    //         {
    //           path: 'network',
    //           component: () => import('./components/setup/setup_network.vue')
    //         }
    //       ]
    //     },
    //     {
    //       path: 'about',
    //       component: () => import('./components/about.vue')
    //     },
    //     {
    //       path: 'test',
    //       component: () => import('./components/Test.vue')
    //     },
    //   ]
    // },
    // {
    //   path: '/test',
    //   name: 'Test',
    //   // route level code-splitting
    //   // this generates a separate chunk (about.[hash].js) for this route
    //   // which is lazy-loaded when the route is visited.
    //   component: () => import(/* webpackChunkName: "about" */ './components/Test.vue')
    // }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

export default router