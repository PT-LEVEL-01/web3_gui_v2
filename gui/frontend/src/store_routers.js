import { reactive } from 'vue'

export const store_routers = reactive({
    currentPageKey_root:"login",//
    currentPageKey_index:"im",//

    currentPageKey_im:"index",//
    currentPageKey_im_list:["index"],
    gopage_im(pageKey){
        // console.log("前进页面",pageKey)
        this.currentPageKey_im_list.push(pageKey)
        this.currentPageKey_im = pageKey
    },
    goback_im(){
        // console.log("回退页面1",this.currentPageKey_im_list.length)
        if(this.currentPageKey_im_list.length <= 1){return}
        // console.log("回退页面2")
        this.currentPageKey_im_list.pop()
        // const lastKey = this.currentPageKey_im_list[this.currentPageKey_im_list.length-1]
        // console.log("回退页面",lastKey)
        this.currentPageKey_im = this.currentPageKey_im_list[this.currentPageKey_im_list.length-1]
    },
})