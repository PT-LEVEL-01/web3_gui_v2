package config

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"web3_gui/utils"
)

// 扩展配置项
var (
	Witness_backup_max        = 21 //备用见证人排名靠前的最多数量，之后的人依然是选举中的候选见证人。31
	Witness_backup_reward_max = 40 //有奖励的最大见证人数量
	Witness_Ave_Ratio         = 30 //出块见证人+候选见证人平均分奖励
	//Witness_Weight_Ratio      = 100 - Witness_Ave_Ratio //出块见证人加权分奖励
	Mining_Reward_Interval = uint64(600) //主链发奖励间隔,默认600块高度(等于1及时到账),真实交易中是含有每个见证人的奖励的,但奖励会首先放入各自见证人的奖励池中.待时间间隔到了累计的奖励全部到账
	//押金
	Mining_deposit      = uint64(100 * 10000 * 1e8) //见证人押金最少金额
	Mining_vote         = uint64(10 * 10000 * 1e8)  //社区节点投票押金最少金额
	Mining_light_min    = uint64(1000 * 1e8)        //轻节点押金最少金额
	CancelVote_Interval = uint64(100000)            //社区取消质押/轻节点取消投票间隔,默认10块高度

	Max_Community_Count = 1000 //全网最大社区数量,默认1000个
	Max_Light_Count     = 1000 //每个社区下的最大轻节点数量,默认1000个

	//其它配置
	EnableFreeGas      = true  //启用免gas费
	DisableCommunityTx = false //禁用社区交易
	DisableLightTx     = false //禁用轻节点交易
	Enable180DayLock   = false //是否启动180天冻结，线性释放

	//质押免gas费
	DepositFreeGasLimitHeight = uint64(3600 * 24 * 30) //质押免gas地址高度限制
	DepositFreeGasLimitCount  = uint64(1000 * 30)      //质押免gas地址次数限制
)

// 额外可选参数,有则按照配置,无则不修改
type ConfigExtra struct {
	WitnessBackupMax       int `json:"WitnessBackupMax,omitempty"`       //备用见证人排名靠前的最多数量，之后的人依然是选举中的候选见证人。31
	WitnessBackupRewardMax int `json:"WitnessBackupRewardMax,omitempty"` //有奖励的最大见证人数量
	WitnessAveRatio        int `json:"WitnessAveRatio,omitempty"`        //出块见证人+候选见证人平均分奖励
	//WitnessWeightRatio     int    //出块见证人加权分奖励
	MiningRewardInterval uint64 `json:"MiningRewardInterval,omitempty"` //主链发奖励间隔,默认600块高度(等于1及时到账),真实交易中是含有每个见证人的奖励的,但奖励会首先放入各自见证人的奖励池中.待时间间隔到了累计的奖励全部到账
	MiningDeposit        uint64 `json:"MiningDeposit,omitempty"`        //见证人押金最少金额
	MiningVote           uint64 `json:"MiningVote,omitempty"`           //社区节点投票押金最少金额
	MiningLightMin       uint64 `json:"MiningLightMin,omitempty"`       //轻节点押金最少金额
	CancelVoteInterval   uint64 `json:"CancelVoteInterval,omitempty"`   //社区取消质押/轻节点取消投票间隔,默认10块高度
	MaxCommunityCount    int    `json:"MaxCommunityCount,omitempty"`    //全网最大社区数量,默认1000个
	MaxLightCount        int    `json:"MaxLightCount,omitempty"`        //每个社区下的最大轻节点数量,默认1000个
}

/*
加载本地额外配置文件，并解析配置文件
*/
func ParseConfigExtra() {
	bs, err := loadConfigExtraLocal()
	if err != nil {
		//engine.Log.Info("没有读取到额外配置文件,使用缺省设置")
		return
	}
	err = parseConfigExtraJSON(bs)
	if err != nil {
		panic("解析额外配置文件错误：" + err.Error())
	}
}

/*
从本地路径中加载配置文件
*/
func loadConfigExtraLocal() ([]byte, error) {
	//fmt.Println("---------------Path_config_extra_Dir-----------------------")
	//fmt.Println(filepath.Join(Path_configDir, Path_config_extra))

	ok, err := utils.PathExists(filepath.Join(Path_configDir, Path_config_extra))
	if err != nil || !ok {
		return nil, errors.New("config extra not exist")
	}
	bs, err := os.ReadFile(filepath.Join(Path_configDir, Path_config_extra))
	if err != nil {
		return nil, err
	}
	return bs, nil
}

/*
解析json格式的配置项目
*/
func parseConfigExtraJSON(bs []byte) error {
	cfi := new(ConfigExtra)

	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(cfi)
	if err != nil {
		return err
	}

	Witness_backup_max = cfi.WitnessBackupMax
	Witness_backup_reward_max = cfi.WitnessBackupRewardMax
	if cfi.WitnessAveRatio < 0 || cfi.WitnessAveRatio > 100 {
		return errors.New("出块见证人+候选见证人平均分奖励比例参数错误")
	}
	Witness_Ave_Ratio = cfi.WitnessAveRatio
	//Witness_Weight_Ratio = 100 - cfi.WitnessAveRatio
	if cfi.MiningRewardInterval <= 0 {
		return errors.New("主链发奖励间隔参数错误")
	}
	Mining_Reward_Interval = cfi.MiningRewardInterval
	Mining_deposit = cfi.MiningDeposit * 1e8
	Mining_vote = cfi.MiningVote * 1e8
	Mining_light_min = cfi.MiningLightMin * 1e8
	CancelVote_Interval = cfi.CancelVoteInterval
	Max_Community_Count = cfi.MaxCommunityCount
	Max_Light_Count = cfi.MaxLightCount

	return nil
}
