package full_node

import (
	"github.com/craftto/go-tron/pkg/client"
	"github.com/craftto/go-tron/pkg/trc20"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"math/big"
	"time"
	usdtconfig "web3_gui/cross_chain/usdt/config"
	"web3_gui/utils"
)

var rpcClient *client.GrpcClient

/*
连接全节点
*/
func ConnFullNode() error {
	var err error
	for _, one := range usdtconfig.NODE_rpc_addr_trx {
		err = connFullNodeOne(one)
		if err != nil {
			continue
		}
	}
	return err
}

func connFullNodeOne(grpcUrl string) error {
	c, err := client.NewGrpcClient(grpcUrl,
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Minute,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return err
	}
	rpcClient = c
	return nil
}

/*
获取节点区块高度
*/
func GetNodeBlockHeight() (int64, utils.ERROR) {
	nowBlock, err := rpcClient.GetNowBlock()
	if err != nil {
		return 0, utils.NewErrorBus(usdtconfig.ERROR_CODE_conn_node_fail_trx, err.Error())
	}
	return nowBlock.BlockHeader.RawData.Number, utils.NewErrorSuccess()
}

/*
获取usdt余额
*/
func GetUSDTBalance(address string) (*big.Int, error) {
	t, err := trc20.NewTrc20(rpcClient, usdtconfig.TrxUsdtAddr)
	if err != nil {
		utils.Log.Error().Msgf("GetTrc20Amount NewTrc20 addr:%s err:%v", address, err)
		return nil, err
	}
	amount, err := t.GetBalance(address)
	if err != nil {
		return nil, err
	}
	//fmt.Println("地址余额", decimal.NewFromBigInt(amount, 0).Div(decimal.New(1, 6)).String())
	return amount, err
}

/*
获取Trx余额
*/
func GetTrxBalance(address string) (int64, error) {
	acc, err := rpcClient.GetAccount(address)
	if err != nil {
		return 0, err
	}
	return acc.Balance, nil
}
