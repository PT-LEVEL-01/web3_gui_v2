import head_daniel from './assets/images/head/daniel.jpg'
import head_elliot from './assets/images/head/elliot.jpg'
import head_elyse from './assets/images/head/elyse.png'
import head_helen from './assets/images/head/helen.jpg'
import head_jenny from './assets/images/head/jenny.jpg'
import head_kristy from './assets/images/head/kristy.png'
import head_matthew from './assets/images/head/matthew.png'
import head_molly from './assets/images/head/molly.png'
import head_steve from './assets/images/head/steve.jpg'
import head_stevie from './assets/images/head/stevie.jpg'
import head_veronika from './assets/images/head/veronika.jpg'

/*
根据url地址获取头像类型
*/
export function GetHeadType(url) {

}

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