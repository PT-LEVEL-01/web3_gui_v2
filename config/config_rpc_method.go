package config

const (
	RPC_Method_Sharebox_GetFileOrder = "Sharebox_GetFileOrder"

	RPC_Method_chain_Info                   = "Info"
	RPC_Method_chain_BlockHeight            = "BlockHeight"
	RPC_Method_chain_Balance                = "Balance"
	RPC_Method_chain_AddressList            = "AddressList"
	RPC_Method_chain_AddressCreate          = "AddressCreate"
	RPC_Method_chain_AddressNonce           = "AddressNonce"
	RPC_Method_chain_AddressNonceMore       = "AddressNonceMore"
	RPC_Method_chain_AddressValidate        = "AddressValidate"
	RPC_Method_chain_AddressInfo            = "AddressInfo"
	RPC_Method_chain_AddressBalance         = "AddressBalance"
	RPC_Method_chain_AddressBalanceMore     = "AddressBalanceMore"
	RPC_Method_chain_AddressAllBalanceRange = "AddressAllBalanceRange"
	RPC_Method_chain_AddressVoteInfo        = "AddressVoteInfo"   //
	RPC_Method_chain_TransactionRecord      = "TransactionRecord" //

	RPC_Method_chain_MnemonicImport      = "MnemonicImport"
	RPC_Method_chain_MnemonicExport      = "MnemonicExport"
	RPC_Method_chain_MnemonicImportEncry = "MnemonicImportEncry"
	RPC_Method_chain_MnemonicExportEncry = "MnemonicExportEncry"

	RPC_Method_chain_SendToAddress       = "SendToAddress"
	RPC_Method_chain_SendToAddressMore   = "SendToAddressMore"
	RPC_Method_chain_PayOrder            = "PayOrder"
	RPC_Method_chain_TxStatusOnChain     = "TxStatusOnChain"
	RPC_Method_chain_TxStatusOnChainMore = "TxStatusOnChainMore"
	RPC_Method_chain_TxProto64ByHash16   = "TxProto64ByHash16"
	RPC_Method_chain_TxJsonByHash16      = "TxJsonByHash16"

	RPC_Method_chain_BlockProto64ByHash16       = "BlockProto64ByHash16"
	RPC_Method_chain_BlockJsonByHash16          = "BlockJsonByHash16"
	RPC_Method_chain_BlockProto64ByHeight       = "BlockProto64ByHeight"
	RPC_Method_chain_BlockJsonByHeight          = "BlockJsonByHeight"
	RPC_Method_chain_BlocksProto64ByHeightRange = "BlocksProto64ByHeightRange"
	RPC_Method_chain_BlocksJsonByHeightRange    = "BlocksJsonByHeightRange"

	RPC_Method_chain_ChainDepositAll     = "ChainDepositAll"     //
	RPC_Method_chain_ChainDepositNodeNum = "ChainDepositNodeNum" //

	RPC_Method_chain_WitnessDepositIn     = "WitnessDepositIn"
	RPC_Method_chain_WitnessDepositOut    = "WitnessDepositOut"
	RPC_Method_chain_WitnessFeatureList   = "WitnessFeatureList"   //
	RPC_Method_chain_WitnessFeatureInfo   = "WitnessFeatureInfo"   //
	RPC_Method_chain_WitnessCandidateList = "WitnessCandidateList" //
	RPC_Method_chain_WitnessCandidateInfo = "WitnessCandidateInfo" //
	RPC_Method_chain_WitnessZoneDeposit   = "WitnessZoneDeposit"   //
	RPC_Method_chain_WitnessList          = "WitnessList"          //
	RPC_Method_chain_WitnessInfo          = "WitnessInfo"          //

	RPC_Method_chain_CommunityList             = "CommunityList"             //
	RPC_Method_chain_CommunityInfo             = "CommunityInfo"             //
	RPC_Method_chain_CommunityDepositIn        = "CommunityDepositIn"        //
	RPC_Method_chain_CommunityDepositOut       = "CommunityDepositOut"       //
	RPC_Method_chain_CommunityShowRewardPool   = "CommunityShowRewardPool"   //
	RPC_Method_chain_CommunityDistributeReward = "CommunityDistributeReward" //

	RPC_Method_chain_LightList             = "LightList"             //
	RPC_Method_chain_LightInfo             = "LightInfo"             //
	RPC_Method_chain_LightDepositIn        = "LightDepositIn"        //
	RPC_Method_chain_LightDepositOut       = "LightDepositOut"       //
	RPC_Method_chain_LightVoteIn           = "LightVoteIn"           //
	RPC_Method_chain_LightVoteOut          = "LightVoteOut"          //
	RPC_Method_chain_LightDistributeReward = "LightDistributeReward" //

	RPC_Method_chain_TokenPublish           = "TokenPublish"
	RPC_Method_chain_TokenBalance           = "TokenBalance"
	RPC_Method_chain_TokenSendToAddress     = "TokenSendToAddress"
	RPC_Method_chain_TokenSendToAddressMore = "TokenSendToAddressMore"
	RPC_Method_chain_TokenInfo              = "TokenInfo"
	RPC_Method_chain_TokenList              = "TokenList"

	RPC_Method_chain_ContractCreate        = "ContractCreate"        //
	RPC_Method_chain_ContractPushTxProto64 = "ContractPushTxProto64" //
	RPC_Method_chain_ContractInfo          = "ContractInfo"          //
	RPC_Method_chain_ContractCall          = "ContractCall"          //
	RPC_Method_chain_ContractCallStack     = "ContractCallStack"     //
	RPC_Method_chain_ContractEvent         = "ContractEvent"         //

	RPC_Method_chain_ERC20Create = "ERC20Create" //
	RPC_Method_chain_ERC20Info   = "ERC20Info"   //

	RPC_Method_chain_GetTxGas = "GetTxGas" //

	RPC_Method_chain_OfflineTxSendToAddress       = "OfflineTxSendToAddress"
	RPC_Method_chain_OfflineTxCreateContract      = "OfflineTxCreateContract"
	RPC_Method_chain_OfflineTxCommunityDepositIn  = "OfflineTxCommunityDepositIn"
	RPC_Method_chain_OfflineTxCommunityDepositOut = "OfflineTxCommunityDepositOut"
	RPC_Method_chain_OfflineTxLightDepositIn      = "OfflineTxLightDepositIn"
	RPC_Method_chain_OfflineTxLightDepositOut     = "OfflineTxLightDepositOut"
	RPC_Method_chain_OfflineTxLightVoteIn         = "OfflineTxLightVoteIn"
	RPC_Method_chain_OfflineTxLightVoteOut        = "OfflineTxLightVoteOut"

	RPC_Method_chain_PushTxProto64 = "PushTxProto64"

	RPC_Method_chain_TestReceiveCoin = "TestReceiveCoin" //
)
