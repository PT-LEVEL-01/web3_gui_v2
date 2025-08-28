package chain_plus

import (
	"math/big"
	"strings"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common"
)

func PackParamsERC20(byteCode, abiStr string, initialSupply *big.Int, tokenName string, decimalUnits uint8, tokenSymbol string) (string, error) {
	abiUtil, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return "", err
	}
	payload, err := abiUtil.Pack("", initialSupply, tokenName, decimalUnits, tokenSymbol)
	if err != nil {
		return "", err
	}
	input := byteCode + common.Bytes2Hex(payload)
	return input, nil
}
