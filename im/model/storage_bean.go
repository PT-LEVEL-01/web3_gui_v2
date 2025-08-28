package model

import (
	"bytes"
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"sync"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/keystore/v2/base58"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
云存储服务器信息
*/
type StorageServerInfo struct {
	Addr          nodeStore.AddressNet //
	Nickname      string               //
	IsOpen        bool                 //是否打开
	Directory     []string             //提供空间目录
	Selling       uint64               //正在售卖的容量 单位：1G
	SellingLock   uint64               //订单锁定容量 单位：1G
	Sold          uint64               //已经卖出 单位：1G
	PriceUnit     uint64               //单价 单位：1G/1天
	UserFreelimit uint64               //用户空闲空间限制,用户只能购买这么多剩余空间，当空间不够时才能继续购买 单位：1G
	UserCanTotal  uint64               //每个用户可以购买的空间总量 单位：1G
	UseTimeMax    uint64               //每个订单租用时间最大值 单位：天
	RenewalTime   uint64               //续费时间，订单到期前多少天可以开始续费。等于0时，不能再续费了 单位：天
	Count         uint64               //在线次数
}

func CreateStorageServerInfo() *StorageServerInfo {
	return &StorageServerInfo{
		Nickname:    utils.BuildName(),
		RenewalTime: config.STORAGE_server_RenewalTime,
		UseTimeMax:  config.STORAGE_server_UseTimeMax,
	}
}

/*
序列化
*/
func (this *StorageServerInfo) Proto() (*[]byte, error) {
	//isOpen := uint32(0)
	//if this.IsOpen {
	//	isOpen = 2
	//} else {
	//	isOpen = 1
	//}
	bhp := go_protos.StorageServerInfo{
		Addr:          this.Addr.GetAddr(),
		Nickname:      []byte(this.Nickname),
		IsOpen:        this.IsOpen,
		Directory:     this.Directory,
		Selling:       this.Selling,
		SellingLock:   this.SellingLock,
		Sold:          this.Sold,
		PriceUnit:     this.PriceUnit,
		UserFreelimit: this.UserFreelimit,
		UserCanTotal:  this.UserCanTotal,
		UseTimeMax:    this.UseTimeMax,  //每个订单租用时间最大值 单位：天
		RenewalTime:   this.RenewalTime, //续费时间，订单到期前多少天可以开始续费。等于0时，不能再续费了 单位：天
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

func (this *StorageServerInfo) ConverVO() *StorageServerInfoVO {
	sslVO := StorageServerInfoVO{
		Addr:          this.Addr.B58String(),
		Nickname:      this.Nickname,
		IsOpen:        this.IsOpen,        //是否打开
		Directory:     this.Directory,     //提供空间目录
		Selling:       this.Selling,       //正在售卖的容量 单位：1G
		SellingLock:   this.SellingLock,   //订单锁定容量 单位：1G
		Sold:          this.Sold,          //已经卖出 单位：1G
		PriceUnit:     this.PriceUnit,     //单价 单位：1G/1天
		UserFreelimit: this.UserFreelimit, //用户空闲空间限制,用户只能购买这么多剩余空间，当空间不够时才能继续购买 单位：1G
		UserCanTotal:  this.UserCanTotal,  //每个用户可以购买的空间总量 单位：1G
		UseTimeMax:    this.UseTimeMax,    //每个订单租用时间最大值 单位：天
		RenewalTime:   this.RenewalTime,   //续费时间，订单到期前多少天可以开始续费。等于0时，不能再续费了 单位：天
		Count:         this.Count,         //在线次数
	}
	return &sslVO
}

func ParseStorageServerInfo(bs []byte) (*StorageServerInfo, error) {
	isInfo := new(go_protos.StorageServerInfo)
	err := proto.Unmarshal(bs, isInfo)
	if err != nil {
		return nil, err
	}
	info := StorageServerInfo{
		Addr:          *nodeStore.NewAddressNet(isInfo.Addr), //
		Nickname:      string(isInfo.Nickname),               //
		IsOpen:        isInfo.IsOpen,                         //
		Directory:     isInfo.Directory,                      //提供空间目录
		Selling:       isInfo.Selling,                        //正在售卖的容量 单位：1G
		SellingLock:   isInfo.SellingLock,                    //订单锁定容量 单位：1G
		Sold:          isInfo.Sold,                           //已经卖出 单位：1G
		PriceUnit:     isInfo.PriceUnit,                      //单价 单位：1G/1天
		UserFreelimit: isInfo.UserFreelimit,                  //用户空闲空间限制,用户只能购买这么多剩余空间，当空间不够时才能继续购买 单位：1G
		UserCanTotal:  isInfo.UserCanTotal,                   //每个用户可以购买的空间总量 单位：1G
		UseTimeMax:    isInfo.UseTimeMax,                     //每个订单租用时间最大值 单位：天
		RenewalTime:   isInfo.RenewalTime,                    //续费时间，订单到期前多少天可以开始续费。等于0时，不能再续费了 单位：天
	}
	//if isInfo.IsOpen == 2 {
	//	info.IsOpen = true
	//} else {
	//	info.IsOpen = false
	//}
	return &info, nil
}

type StorageServerInfoVO struct {
	Addr              string   //
	Nickname          string   //
	IsOpen            bool     //是否打开
	Directory         []string //提供空间目录
	DirectoryFreeSize []uint64 //每个目录的剩余空间
	Selling           uint64   //正在售卖的容量 单位：1G
	SellingLock       uint64   //订单锁定容量 单位：1G
	Sold              uint64   //已经卖出 单位：1G
	PriceUnit         uint64   //单价 单位：1G/1天
	UserFreelimit     uint64   //用户空闲空间限制,用户只能购买这么多剩余空间，当空间不够时才能继续购买 单位：1G
	UserCanTotal      uint64   //每个用户可以购买的空间总量 单位：1G
	UseTimeMax        uint64   //每个订单租用时间最大值 单位：天
	RenewalTime       uint64   //续费时间，订单到期前多少天可以开始续费。等于0时，不能再续费了 单位：天
	Count             uint64   //在线次数
}

//type StorageServerList struct {
//	Addr          nodeStore.AddressNet
//	Nickname      string
//	Count         uint64
//	Open          bool
//	Selling       uint64 //正在售卖的容量 单位：1G
//	Sold          uint64 //已经卖出 单位：1G
//	PriceUnit     uint64 //单价 单位：1G
//	UserFreelimit uint64 //用户空闲空间限制,用户只能购买这么多剩余空间，当空间不够时才能继续购买 单位：1G
//	UserCanTotal  uint64 //每个用户可以购买的空间总量 单位：1G
//}

//type StorageServerListVO struct {
//	Addr          string
//	Nickname      string
//	Count         uint64
//	Open          bool
//	Selling       uint64
//	Sold          uint64
//	PriceUnit     uint64
//	UserFreelimit uint64 //用户空闲空间限制,用户只能购买这么多剩余空间，当空间不够时才能继续购买 单位：1G
//	UserCanTotal  uint64 //每个用户可以购买的空间总量 单位：1G
//	SpacesTotal   uint64 //购买的容量
//	SpacesUse     uint64 //已经使用的容量
//}

/*
订单
*/
type OrderForm struct {
	GoodsId            []byte                   //商品id
	Number             []byte                   //订单编号，全局自增长ID
	PreNumber          []byte                   //续费订单引用之前的订单编号
	UserAddr           nodeStore.AddressNet     //消费者地址
	ServerAddr         nodeStore.AddressNet     //服务器地址
	ServerAddrCoin     coin_address.AddressCoin //服务器收款地址
	SpaceTotal         uint64                   //购买空间数量 单位：1G
	UseTime            uint64                   //空间使用时间 单位：1天
	TotalPrice         uint64                   //订单总金额
	ChainTx            []byte                   //区块链上的交易
	TxHash             []byte                   //已经上链的交易hash
	CreateTime         int64                    //订单创建时间
	TimeOut            int64                    //订单过期时间
	PayLockBlockHeight uint64                   //未支付的订单是支付限制区块高度，已支付的订单是服务超期高度
	LockHeightOnChain  uint64                   //已支付订单的锁定上链高度，超过这个高度还未上链的，就不能上链了
}

func (this *OrderForm) Proto() (*[]byte, error) {
	bhp := go_protos.StorageOrderForm{
		Number:             this.Number,
		UserAddr:           this.UserAddr.GetAddr(),
		ServerAddr:         this.ServerAddr.GetAddr(),
		ServerAddrCoin:     this.ServerAddrCoin,
		SpaceTotal:         this.SpaceTotal,
		UseTime:            this.UseTime,
		TotalPrice:         this.TotalPrice,
		ChainTx:            this.ChainTx,
		TxHash:             this.TxHash,
		CreateTime:         this.CreateTime,
		TimeOut:            this.TimeOut,
		GoodsId:            this.GoodsId,
		PayLockBlockHeight: this.PayLockBlockHeight,
		LockHeightOnChain:  this.LockHeightOnChain,
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

func (this *OrderForm) ConverVO() *OrderFormVO {
	return &OrderFormVO{
		Number:             hex.EncodeToString(this.Number),  //订单编号，全局自增长ID
		UserAddr:           this.UserAddr.B58String(),        //消费者地址
		ServerAddr:         this.ServerAddr.B58String(),      //服务器地址
		ServerAddrCoin:     this.ServerAddrCoin.B58String(),  //服务器收款地址
		SpaceTotal:         this.SpaceTotal,                  //购买空间数量 单位：1G
		UseTime:            this.UseTime,                     //空间使用时间 单位：1天
		TotalPrice:         this.TotalPrice,                  //订单总金额
		ChainTx:            hex.EncodeToString(this.ChainTx), //区块链上的交易
		TxHash:             hex.EncodeToString(this.TxHash),  //已经上链的交易hash
		CreateTime:         this.CreateTime,                  //订单创建时间
		TimeOut:            this.TimeOut,                     //订单过期时间
		GoodsId:            hex.EncodeToString(this.GoodsId), //
		PayLockBlockHeight: this.PayLockBlockHeight,          //支付限制区块高度
		LockHeightOnChain:  this.LockHeightOnChain,           //
	}
}

/*
解析表单
*/
func ParseOrderForm(bs []byte) (*OrderForm, error) {
	if bs == nil {
		return nil, nil
	}
	sof := new(go_protos.StorageOrderForm)
	err := proto.Unmarshal(bs, sof)
	if err != nil {
		return nil, err
	}
	of := OrderForm{
		Number:             sof.Number,                                   //订单编号，全局自增长ID
		UserAddr:           *nodeStore.NewAddressNet(sof.UserAddr),       //消费者地址
		ServerAddr:         *nodeStore.NewAddressNet(sof.ServerAddr),     //服务器地址
		ServerAddrCoin:     coin_address.AddressCoin(sof.ServerAddrCoin), //
		SpaceTotal:         sof.SpaceTotal,                               //购买空间数量 单位：1G
		UseTime:            sof.UseTime,                                  //空间使用时间 单位：1天
		TotalPrice:         sof.TotalPrice,                               //订单总金额
		ChainTx:            sof.ChainTx,                                  //区块链上的交易
		TxHash:             sof.TxHash,                                   //已经上链的交易hash
		CreateTime:         sof.CreateTime,                               //订单创建时间
		TimeOut:            sof.TimeOut,                                  //订单过期时间
		GoodsId:            sof.GoodsId,                                  //
		PayLockBlockHeight: sof.PayLockBlockHeight,                       //
		LockHeightOnChain:  sof.LockHeightOnChain,                        //
	}
	return &of, nil
}

/*
订单
*/
type OrderFormVO struct {
	Number             string //订单编号，全局自增长ID
	UserAddr           string //消费者地址
	ServerAddr         string //服务器地址
	ServerAddrCoin     string //服务器收款地址
	SpaceTotal         uint64 //购买空间数量 单位：1G
	UseTime            uint64 //空间使用时间 单位：1天
	TotalPrice         uint64 //订单总金额
	ChainTx            string //区块链上的交易
	TxHash             string //已经上链的交易hash
	CreateTime         int64  //订单创建时间
	TimeOut            int64  //订单过期时间
	GoodsId            string //商品id
	PayLockBlockHeight uint64 //支付限制区块高度
	LockHeightOnChain  uint64 //已支付待上链的锁定区块高度
	Status             int    //状态。1=未支付;2=已支付，但未上链;3=已支付，已上链;4=过期;
}

/*
目录索引
*/
type DirectoryIndex struct {
	ID        []byte               //文件夹唯一ID采用全局自增长ID
	ParentID  []byte               //父文件夹ID
	Name      string               //文件夹名称
	Dirs      []*DirectoryIndex    //文件夹中包含文件夹
	Files     []*FileIndex         //文件列表
	DirsID    [][]byte             //包含文件夹的ID
	FilesID   [][]byte             //包含文件的ID
	UAddr     nodeStore.AddressNet //所属用户
	ParentDir *DirectoryIndex      //
}

func (this *DirectoryIndex) Proto() (*[]byte, error) {
	dirs := make([][]byte, 0, len(this.DirsID))
	for _, one := range this.DirsID {
		dirs = append(dirs, one)
	}
	files := make([][]byte, 0, len(this.FilesID))
	for _, one := range this.FilesID {
		files = append(files, one)
	}
	bhp := go_protos.StorageDirectoryIndex{
		ID:       this.ID,       //文件夹唯一ID采用全局自增长ID
		ParentID: this.ParentID, //
		Name:     this.Name,     //文件夹名称
		//Dirs    []DirectoryIndex //文件夹中包含文件夹
		//Files   []FileIndex      //文件列表
		Dirs:     dirs,                 //包含文件夹的ID
		Files:    files,                //包含文件的ID
		UserAddr: this.UAddr.GetAddr(), //
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

/*
解析目录索引
*/
func ParseDirectoryIndex(bs []byte) (*DirectoryIndex, error) {
	if bs == nil {
		return nil, nil
	}
	sdi := new(go_protos.StorageDirectoryIndex)
	err := proto.Unmarshal(bs, sdi)
	if err != nil {
		return nil, err
	}
	of := DirectoryIndex{
		ID:       sdi.ID,       //文件夹唯一ID采用全局自增长ID
		Name:     sdi.Name,     //文件夹名称
		ParentID: sdi.ParentID, //
		//Dirs    []DirectoryIndex //文件夹中包含文件夹
		//Files   []FileIndex      //文件列表
		DirsID:  sdi.Dirs,                               //包含文件夹的ID
		FilesID: sdi.Files,                              //包含文件的ID
		UAddr:   *nodeStore.NewAddressNet(sdi.UserAddr), //
	}
	return &of, nil
}

/*
将包含的子目录和文件也序列化
*/
func (this *DirectoryIndex) ProtoDirAndFile() (*[]byte, error) {
	dirs := make([][]byte, 0, len(this.Dirs)+1)
	bs, err := this.Proto()
	if err != nil {
		return nil, err
	}
	//排在第一位的是最顶层目录
	dirs = append(dirs, *bs)
	for _, one := range this.Dirs {
		bsOne, err := one.Proto()
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, *bsOne)
	}
	files := make([][]byte, 0, len(this.Files))
	for _, one := range this.Files {
		bsOne, err := one.Proto()
		if err != nil {
			return nil, err
		}
		files = append(files, *bsOne)
	}
	bhp := go_protos.Bytes{
		List:  dirs,
		List2: files,
	}
	*bs, err = bhp.Marshal()
	return bs, err
}

/*
解析目录索引，包含其中的子目录和文件
*/
func ParseDirectoryIndexMore(bs []byte) (*DirectoryIndex, error) {
	if bs == nil {
		return nil, nil
	}
	bss := new(go_protos.Bytes)
	err := proto.Unmarshal(bs, bss)
	if err != nil {
		return nil, err
	}
	dirs := make([]*DirectoryIndex, 0, len(bss.List))
	for _, one := range bss.List {
		dIndex, err := ParseDirectoryIndex(one)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, dIndex)
	}
	files := make([]*FileIndex, 0, len(bss.List2))
	for _, one := range bss.List2 {
		fIndex, err := ParseFileIndex(one)
		if err != nil {
			return nil, err
		}
		files = append(files, fIndex)
	}
	d := dirs[0]
	d.Dirs = dirs[1:]
	d.Files = files
	return d, nil
}

/*
目录索引
*/
type DirectoryIndexVO struct {
	ID       string             //文件夹唯一ID采用全局自增长ID
	ParentID string             //父文件夹ID
	Name     string             //文件夹名称
	Dirs     []DirectoryIndexVO //文件夹中包含文件夹
	Files    []FileIndexVO      //文件列表
	DirsID   []string           //包含文件夹的ID
	FilesID  []string           //包含文件的ID
}

func (this *DirectoryIndex) ConverVO() *DirectoryIndexVO {
	sslVO := DirectoryIndexVO{
		ID:       string(base58.Encode(this.ID)),              //文件夹唯一ID采用全局自增长ID
		ParentID: string(base58.Encode(this.ParentID)),        //
		Name:     this.Name,                                   //文件夹名称
		Dirs:     make([]DirectoryIndexVO, 0, len(this.Dirs)), //文件夹中包含文件夹
		Files:    make([]FileIndexVO, 0, len(this.Files)),     //文件列表
		DirsID:   make([]string, 0, len(this.DirsID)),         //包含文件夹的ID
		FilesID:  make([]string, 0, len(this.FilesID)),        //包含文件的ID
	}
	for _, one := range this.DirsID {
		sslVO.DirsID = append(sslVO.DirsID, string(base58.Encode(one)))
	}
	for _, one := range this.FilesID {
		sslVO.FilesID = append(sslVO.FilesID, string(base58.Encode(one)))
	}
	for _, one := range this.Dirs {
		sslVO.Dirs = append(sslVO.Dirs, *one.ConverVO())
	}
	for _, one := range this.Files {
		sslVO.Files = append(sslVO.Files, *one.ConverVO())
	}
	return &sslVO
}

/*
文件索引
*/
type FileIndex struct {
	ID               []byte        //
	Hash             []byte        //文件加密后的hash值
	DirID            [][]byte      //所属文件夹ID
	UserAddr         [][]byte      //所属用户地址
	Pwds             [][]byte      //所属用户密码
	Version          uint16        //版本号
	Name             []string      //多用户的不同文件名称
	FileSize         uint64        //文件总大小
	ChunkCount       uint32        //分片总量
	ChunkOneSize     uint64        //每一个分片大小
	Chunks           [][]byte      //每一个分片ID
	SupplierIDs      [][]byte      //每个分片文件上传者ID
	PullIDs          [][]byte      //每个分片文件下载者ID
	ChunkOffsetIndex []uint64      //每一个分片已经传输的大小，用于显示已经上传/下载的大小
	PermissionType   []uint8       //权限类型 0=仅自己可访问;1=仅自己授权者可访问;2=所有人可访问;
	EncryptionType   uint32        //加密类型
	Time             []int64       //创建时间
	Status           uint32        //内存中的状态，不用实例化
	AbsPath          string        //文件的路径
	Error            string        //报错信息
	Lock             *sync.RWMutex //
}

func (this *FileIndex) Proto() (*[]byte, error) {
	pts := make([]uint32, 0, len(this.PermissionType))
	for _, one := range this.PermissionType {
		pts = append(pts, uint32(one))
	}
	bhp := go_protos.StorageFileIndex{
		ID:             this.ID,              //
		Hash:           this.Hash,            //文件加密后的hash值
		DirID:          this.DirID,           //
		UserAddr:       this.UserAddr,        //用户地址
		Pwds:           this.Pwds,            //
		Version:        uint32(this.Version), //版本号
		Name:           this.Name,            //文件名称
		FileSize:       this.FileSize,        //
		ChunkCount:     this.ChunkCount,      //分片总量
		ChunkOneSize:   this.ChunkOneSize,    //每一个分片大小
		Chunks:         this.Chunks,          //每一个分片ID
		PullIDs:        this.PullIDs,         //
		PermissionType: pts,                  //权限类型 0=仅自己可访问;1=仅自己授权者可访问;2=所有人可访问;
		Time:           this.Time,            //
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

/*
检查用户是否在其中
*/
func (this *FileIndex) CheckHaveUser(uAddr nodeStore.AddressNet) bool {
	for _, one := range this.UserAddr {
		if bytes.Equal(one, uAddr.GetAddr()) {
			return true
		}
	}
	return false
}

/*
减少一个用户所有者
*/
func (this *FileIndex) SubUser(uAddr nodeStore.AddressNet) bool {
	for i, one := range this.UserAddr {
		if bytes.Equal(one, uAddr.GetAddr()) {
			temp := this.UserAddr[:i]
			temp = append(temp, this.UserAddr[i+1:]...)
			this.UserAddr = temp

			temp = this.DirID[:i]
			temp = append(temp, this.DirID[i+1:]...)
			this.DirID = temp

			temp = this.Pwds[:i]
			temp = append(temp, this.Pwds[i+1:]...)
			this.Pwds = temp

			names := this.Name[:i]
			names = append(names, this.Name[i+1:]...)
			this.Name = names

			ptypes := this.PermissionType[:i]
			ptypes = append(ptypes, this.PermissionType[i+1:]...)
			this.PermissionType = ptypes

			times := this.Time[:i]
			times = append(times, this.Time[i+1:]...)
			this.Time = times
			return true
		}
	}
	return false
}

/*
筛选出指定用户
*/
func (this *FileIndex) FilterUser(uAddr nodeStore.AddressNet) bool {
	for i, one := range this.UserAddr {
		if bytes.Equal(one, uAddr.GetAddr()) {
			this.UserAddr = [][]byte{this.UserAddr[i]}
			this.DirID = [][]byte{this.DirID[i]}
			this.Pwds = [][]byte{this.Pwds[i]}
			this.Name = []string{this.Name[i]}
			this.PermissionType = []uint8{this.PermissionType[i]}
			this.Time = []int64{this.Time[i]}
			return true
		}
	}
	return false
}

/*
解析文件索引
*/
func ParseFileIndex(bs []byte) (*FileIndex, error) {
	if bs == nil {
		return nil, nil
	}
	sdi := new(go_protos.StorageFileIndex)
	err := proto.Unmarshal(bs, sdi)
	if err != nil {
		return nil, err
	}
	fi, err := ConverFileIndex(sdi)
	return fi, err
}

func ConverFileIndex(sfi *go_protos.StorageFileIndex) (*FileIndex, error) {
	fi := FileIndex{
		ID:               sfi.ID,                                    //
		Hash:             sfi.Hash,                                  //文件加密后的hash值
		DirID:            sfi.DirID,                                 //
		UserAddr:         sfi.UserAddr,                              //用户地址
		Pwds:             sfi.Pwds,                                  //
		Version:          uint16(sfi.Version),                       //版本号
		Name:             sfi.Name,                                  //文件名称
		FileSize:         sfi.FileSize,                              //
		ChunkCount:       sfi.ChunkCount,                            //分片总量
		ChunkOneSize:     sfi.ChunkOneSize,                          //每一个分片大小
		Chunks:           sfi.Chunks,                                //每一个分片ID
		PullIDs:          sfi.PullIDs,                               //
		ChunkOffsetIndex: make([]uint64, len(sfi.Chunks)),           //
		PermissionType:   make([]uint8, 0, len(sfi.PermissionType)), //权限类型 0=仅自己可访问;1=仅自己授权者可访问;2=所有人可访问;
		Time:             sfi.Time,                                  //
		Lock:             new(sync.RWMutex),                         //
	}
	for _, one := range sfi.PermissionType {
		fi.PermissionType = append(fi.PermissionType, uint8(one))
	}
	return &fi, nil
}

/*
文件索引
*/
type FileIndexVO struct {
	ID             string   //文件加密后的hash值
	Hash           string   //
	DirID          string   //所属文件夹ID
	UserAddr       string   //用户地址
	Version        uint16   //版本号
	Name           string   //文件名称
	FileSize       uint64   //文件总大小
	ChunkCount     uint32   //分片总量
	ChunkOneSize   uint64   //每一个分片大小
	Chunks         []string //每一个分片ID
	PermissionType uint8    //权限类型 0=仅自己可访问;1=仅自己授权者可访问;2=所有人可访问;
	EncryptionType uint32   //加密类型
	Time           int64    //创建时间
	Status         uint32   //内存中的状态，不用实例化
	AbsPath        string   //文件的路径
	Error          string   //报错信息
}

func (this *FileIndex) ConverVO() *FileIndexVO {
	sslVO := FileIndexVO{
		ID:             string(base58.Encode(this.ID)),       //文件加密后的hash值
		Hash:           string(base58.Encode(this.Hash)),     //
		DirID:          string(base58.Encode(this.DirID[0])), //所属文件夹ID
		Version:        this.Version,                         //版本号
		Name:           this.Name[0],                         //文件名称
		FileSize:       this.FileSize,                        //文件总大小
		ChunkCount:     this.ChunkCount,                      //分片总量
		ChunkOneSize:   this.ChunkOneSize,                    //每一个分片大小
		Chunks:         make([]string, 0, len(this.Chunks)),  //每一个分片ID
		PermissionType: this.PermissionType[0],               //权限类型 0=仅自己可访问;1=仅自己授权者可访问;2=所有人可访问;
		EncryptionType: this.EncryptionType,                  //加密类型
		Time:           this.Time[0],                         //创建时间
		Status:         this.Status,                          //内存中的状态，不用实例化
		AbsPath:        this.AbsPath,                         //文件的路径
		Error:          this.Error,                           //报错信息
	}
	for _, one := range this.Chunks {
		sslVO.Chunks = append(sslVO.Chunks, hex.EncodeToString(one))
	}
	return &sslVO
}

func (this *FileIndex) Conver() *go_protos.StorageFileIndex {
	sfi := go_protos.StorageFileIndex{
		ID:             this.ID,
		Hash:           this.Hash,
		DirID:          this.DirID,
		UserAddr:       this.UserAddr,
		Pwds:           this.Pwds,
		Version:        uint32(this.Version),
		Name:           this.Name,
		FileSize:       this.FileSize,
		ChunkCount:     this.ChunkCount,
		ChunkOneSize:   this.ChunkOneSize,
		Chunks:         this.Chunks,
		PullIDs:        this.PullIDs,
		PermissionType: make([]uint32, 0, len(this.PermissionType)),
		Time:           this.Time,
	}
	for _, one := range this.PermissionType {
		sfi.PermissionType = append(sfi.PermissionType, uint32(one))
	}
	return &sfi
}

/*
下载任务
*/
//type DownloadTaskVO struct {
//	DbId       string       //
//	ServerAddr string       //
//	FileIndex  *FileIndexVO //
//	LocalPath  string       //保存到本地的路径
//	Status     int          //状态
//}
