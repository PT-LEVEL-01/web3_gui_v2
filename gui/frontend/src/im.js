import head_daniel from './assets/images/head/daniel.jpg'
import head_elliot from './assets/images/head/elliot.jpg'
import head_elyse from './assets/images/head/elyse.jpg'
import head_helen from './assets/images/head/helen.jpg'
import head_jenny from './assets/images/head/jenny.jpg'
import head_kristy from './assets/images/head/kristy.jpg'
import head_matthew from './assets/images/head/matthew.jpg'
import head_molly from './assets/images/head/molly.jpg'
import head_steve from './assets/images/head/steve.jpg'
import head_stevie from './assets/images/head/stevie.jpg'
import head_veronika from './assets/images/head/veronika.jpg'



/*
根据头像类型获取url地址
*/
export function GetHeadUrl(type) {
    switch(type){
        case 0:return head_daniel
        case 1:return head_elliot
        case 2:return head_elyse
        case 3:return head_helen
        case 4:return head_jenny
        case 5:return head_kristy
        case 6:return head_matthew
        case 7:return head_molly
        case 8:return head_steve
        case 9:return head_stevie
        case 10:return head_veronika
    }
    
}