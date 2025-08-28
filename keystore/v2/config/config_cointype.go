package config

// coin
var (
	BTC            = "BTC"
	LTC            = "LTC"
	OKEXCHAIN      = "OKEXCHAIN"
	OKEXCHAIN_TEST = "OKEXCHAIN_TEST"
	BSV            = "BSV"
	DASH           = "DASH"
	DOT            = "DOT"
	AVAXC          = "AVAX"
	ETH            = "ETH"
	ETC            = "ETC"
	BSC            = "BNB"
	TRX            = "TRX"
	ADA            = "ADA"
	XRP            = "XRP"
	SOL            = "SOL"
	POLYGON        = "MATIC"
	FIL            = "FIL"
	BCH            = "BCH"
	DOGE           = "DOGE"
	ZCASH          = "ZCASH"
	NEO            = "NEO"
	DGB            = "DGB"
	ONT            = "ONT"
	IOST           = "IOST"
	UGAS           = "UGAS"
	UGASBMUPC      = "UGASBMUPC"
	UGASBMNRL      = "UGASBMNRL"
	UGASBMTPR      = "UGASBMTPR"
	WAN            = "WAN"
	IPC            = "IPC"
	CKB            = "CKB"
	XMR            = "XMR"

	SHIB = "SHIB"
	USDT = "USDT"
	USDC = "USDC"
	UNI  = "UNI"
	DAI  = "DAI"

	//CoinName_custom = "TEST"
)

const (
	Zero      uint32 = 0
	ZeroQuote uint32 = 0x80000000

	PurposeBIP44 uint32 = 0x8000002C // 44' BIP44
	PurposeBIP49 uint32 = 0x80000031 // 49' BIP49
	PurposeBIP84 uint32 = 0x80000054 // 84' BIP84

	//Apostrophe uint32 = 0x80000000

)

var (
	COINADDR_version = VERSION_v2 //地址版本
)

const (
	BTCToken uint32 = 0x10000000
	ETHToken uint32 = 0x20000000

	// 钱包各个币种的coinType编号，满足bip44协议
	// https://github.com/satoshilabs/slips/blob/master/slip-0044.md#registered-coin-types
	COINTYPE_BTC        = ZeroQuote + 0
	COINTYPE_BTCTestnet = ZeroQuote + 1
	COINTYPE_LTC        = ZeroQuote + 2
	COINTYPE_DOGE       = ZeroQuote + 3
	COINTYPE_DASH       = ZeroQuote + 5
	COINTYPE_GL         = ZeroQuote + 49
	COINTYPE_ETH        = ZeroQuote + 60
	COINTYPE_BSC        = COINTYPE_ETH
	COINTYPE_POLYGON    = COINTYPE_ETH
	COINTYPE_AVAXC      = COINTYPE_ETH
	COINTYPE_ETC        = ZeroQuote + 61
	COINTYPE_ATOM       = ZeroQuote + 118
	COINTYPE_XMR        = ZeroQuote + 128
	COINTYPE_XRP        = ZeroQuote + 144
	COINTYPE_BCH        = ZeroQuote + 145
	COINTYPE_EOS        = ZeroQuote + 194
	COINTYPE_TRX        = ZeroQuote + 195
	COINTYPE_CKB        = ZeroQuote + 309
	COINTYPE_DOT        = ZeroQuote + 354
	COINTYPE_KSM        = ZeroQuote + 434
	COINTYPE_FIL        = ZeroQuote + 461
	COINTYPE_SOL        = ZeroQuote + 501
	COINTYPE_BNB        = ZeroQuote + 714
	COINTYPE_NEO        = ZeroQuote + 888
	COINTYPE_ONT        = ZeroQuote + 1024
	COINTYPE_XTZ        = ZeroQuote + 1729
	COINTYPE_ADA        = ZeroQuote + 1815
	COINTYPE_QTUM       = ZeroQuote + 2301

	COINTYPE_USDT = BTCToken + 1 // btc token
	COINTYPE_IOST = ETHToken + 1 // eth token
	COINTYPE_USDC = ETHToken + 2 // eth token
)
