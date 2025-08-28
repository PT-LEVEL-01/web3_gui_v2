package keystore

import (
	"testing"
	"web3_gui/utils"
)

func TestMnemonic(t *testing.T) {
	example_CreateMnemonic()
}

func example_CreateMnemonic() {
	mnemonic, err := CreateMnemonic(MnemonicSize128)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}
	utils.Log.Info().Str("生成助记词", mnemonic).Send()
}
