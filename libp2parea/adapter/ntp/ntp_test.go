package ntp

import (
	"testing"
	"time"
	"web3_gui/utils"
)

func TestNtp(t *testing.T) {
	time.Sleep(time.Second * 5)

	t1 := time.Now()
	ntpt := GetNtpTime()
	t2 := time.Now()
	utils.Log.Info().Msgf(ntpt.String())
	utils.Log.Info().Msgf(t1.String())
	utils.Log.Info().Msgf(t2.String())
	utils.Log.Info().Msgf(GetNtpOffsetTime().String())

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 30)
		t := GetNtpTime()
		utils.Log.Info().Msgf(t.String())
		utils.Log.Info().Msgf(GetNtpOffsetTime().String())
		utils.Log.Info().Msgf("----------------")
	}
}

func TestNtpSync(t *testing.T) {
	time.Sleep(time.Second * 5)

	t1 := time.Now()
	ntpt := GetNtpTime()
	t2 := time.Now()
	utils.Log.Info().Msgf(ntpt.String())
	utils.Log.Info().Msgf(t1.String())
	utils.Log.Info().Msgf(t2.String())
	utils.Log.Info().Msgf(GetNtpOffsetTime().String())

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 30)
		// TriggerNtpSync()
		TriggerNtpSyncWaitRes()
		t := GetNtpTime()
		utils.Log.Info().Msgf(t.String())
		utils.Log.Info().Msgf(GetNtpOffsetTime().String())
		utils.Log.Info().Msgf("----------------")
	}
}

func TestNtpSystemTimeChange(t *testing.T) {
	// 测试修改系统时间, ntp的同步机制是否正常
	time.Sleep(time.Second * 5)

	utils.Log.Info().Msgf("Ntp 差值:%s", GetNtpOffsetTime().String())

	for {
		time.Sleep(time.Second * 5)
		utils.Log.Info().Msgf("----------------")
		utils.Log.Info().Msgf("Ntp 时间:%s", GetNtpTime().String())
		utils.Log.Info().Msgf("Ntp 差值:%s", GetNtpOffsetTime().String())
		utils.Log.Info().Msgf("----------------")
	}
}
