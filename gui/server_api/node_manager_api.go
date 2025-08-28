package server_api

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
	gconfig "web3_gui/config"
	"web3_gui/gui/server_api/clents"
	pcli "web3_gui/gui/server_api/clents/jsonrpc2"
	"web3_gui/utils"
)

type YamlFile struct {
	PeerGroup map[string]string `json:"peer_group" yaml:"peer_group"`
	PeerList  []*Peer           `json:"peer_list" yaml:"peer_list"`
}

func NewYamlFile() *YamlFile {
	y := new(YamlFile)
	y.PeerGroup = make(map[string]string)
	y.PeerList = make([]*Peer, 0)
	return y
}

type Peer struct {
	ID             string         `json:"id" yaml:"-"`
	Group          string         `json:"group" yaml:"group"`
	Name           string         `json:"name"  yaml:"name"`
	Ip             string         `json:"ip"   yaml:"ip"`
	Port           int            `json:"port" yaml:"port"`
	Username       string         `json:"username" yaml:"username"`
	Password       string         `json:"password" yaml:"password"`
	Status         bool           `json:"status" yaml:"-"`          //是否开启
	HighestBlock   uint64         `json:"highest_block" yaml:"-"`   //所链接的节点的最高高度
	CurrentBlock   uint64         `json:"current_block" yaml:"-"`   // 块高度
	SnapshotHeight uint64         `json:"snapshot_height" yaml:"-"` //快照高度
	TotalBalance   float64        `json:"total_balance" yaml:"-"`   //总余额
	Types          map[int]uint   `json:"types" yaml:"-"`           //角色
	UsedTime       int64          `json:"used_time" yaml:"-"`       //耗时
	IsDel          bool           `json:"is_del" yaml:"-"`
	Error          string         `json:"error" yaml:"-"`
	UpdatedAt      int64          `json:"updated_at" yaml:"-"`
	DefaultAddress *AddressInfo   `json:"default_address" yaml:"-"`
	Addresses      []*AddressInfo `json:"addresses" yaml:"-"`
	usedTimeArr    []int64        //用于计算UsedTime的时间差统计
}

func (p *Peer) setUsedTime(starAt time.Time) {
	dur := time.Since(starAt).Milliseconds()
	if len(p.usedTimeArr) < 10 { //最多统计10次
		p.usedTimeArr = append(p.usedTimeArr, dur)
	} else {
		p.usedTimeArr = p.usedTimeArr[1:]
		p.usedTimeArr = append(p.usedTimeArr, dur)
	}

	sum := int64(0)
	for _, v := range p.usedTimeArr {
		sum += v
	}
	if len(p.usedTimeArr) > 0 {
		p.UsedTime = int64(float64(sum) / float64(len(p.usedTimeArr)))
	} else {
		p.UsedTime = 0
	}
}

type AddressInfo struct {
	Index         int     `json:"index"`
	Address       string  `json:"address"`
	Balance       float64 `json:"balance"`        //可用余额
	BalanceFrozen float64 `json:"balance_frozen"` //冻结余额
	BalanceLockup float64 `json:"balance_lockup"` //锁仓余额
	PeerType      int     `json:"peer_type"`
}

var PeerConfigFile = filepath.Join(gconfig.Path_configDir, "peerconfig.yaml")

func (this *SdkApi) Load() error {
	ok, err := utils.PathExists(PeerConfigFile)
	if err != nil {
		return err
	}
	if ok {
		bs, err := os.ReadFile(PeerConfigFile)
		if err != nil {
			return err
		}
		yf := new(YamlFile)
		yaml.Unmarshal(bs, yf)
		//utils.Log.Info().Int("组", len(yf.PeerGroup)).Send()
		for k, v := range yf.PeerList {
			yf.PeerList[k].ID = fmt.Sprintf("%s:%s", v.Group, v.Name)
			this.PeerList = append(this.PeerList, yf.PeerList[k])
			this.peerMap.Store(v.ID, v)
		}
		this.PeerGroup = yf.PeerGroup
	}
	//this.ctx = context.Background()
	go this.run()
	return nil
}

const UpdateDur = 200 * time.Millisecond //更新节点数据间隔时间

func (this *SdkApi) run() {
	doneChannel := make(map[string]chan int)
	for {
		select {
		//case <-this.Ctx.Done():
		//	return
		case <-time.After(UpdateDur):
			for _, peer := range this.getActivePeer() {
				if _, ok := doneChannel[peer.ID]; !ok {
					doneChannel[peer.ID] = make(chan int, 1)
					doneChannel[peer.ID] <- 0
				}
				select {
				case rc := <-doneChannel[peer.ID]:
					if rc > 0 {
						rc--
						doneChannel[peer.ID] <- rc
						continue
					}

					go func(peer *Peer) {
						err := update(peer)
						if err != nil {
							rc = 10
							doneChannel[peer.ID] <- rc
							return
						}
						doneChannel[peer.ID] <- rc
					}(peer)
				default:
					continue
				}

			}
		}
	}

}

func update(peer *Peer) error {
	cli := pcli.NewClient(fmt.Sprintf("%s:%d", peer.Ip, peer.Port), peer.Username, peer.Password)
	starAt := time.Now()
	listAccounts, err := cli.ListAccounts()
	if err != nil {
		peer.Error = err.Error()
		peer.setUsedTime(starAt)
		return err
	}

	if len(listAccounts) <= 0 {
		peer.Error = "缺省地址为空"
		peer.setUsedTime(starAt)
		return nil
	}
	addresses := []*AddressInfo{}
	totalBalance := uint64(0)
	types := make(map[int]uint, 0)
	for i, v := range listAccounts {
		balance := pcli.AmountDivCV(v.Value)
		addresses = append(addresses, &AddressInfo{
			Index:         i,
			Address:       v.AddrCoin,
			Balance:       balance,
			BalanceFrozen: pcli.AmountDivCV(v.ValueFrozen),
			BalanceLockup: pcli.AmountDivCV(v.ValueLockip),
			PeerType:      v.Type,
		})
		totalBalance += v.Value

		if v.Type == 1 || v.Type == 2 || v.Type == 3 {
			types[v.Type] = types[v.Type] + 1
		}
	}

	peer.Addresses = addresses
	peer.TotalBalance = pcli.AmountDivCV(totalBalance)
	peer.Types = types
	if len(addresses) > 0 {
		peer.DefaultAddress = addresses[0]
	}

	accountInfo, err := cli.GetInfo()
	if err != nil {
		peer.Error = err.Error()
		peer.setUsedTime(starAt)
		return err
	}

	endAt := time.Now().UnixMilli()
	peer.CurrentBlock = accountInfo.CurrentBlock
	peer.HighestBlock = accountInfo.HighestBlock
	peer.SnapshotHeight = accountInfo.SnapshotHeight
	peer.Error = ""
	peer.UpdatedAt = endAt
	peer.setUsedTime(starAt)
	return nil
}

func (this *SdkApi) AddPeer(p *Peer) (*Peer, error) {
	p.ID = fmt.Sprintf("%s:%s", p.Group, p.Name)
	//是否有重复
	_, ok := this.peerMap.Load(p.ID)
	if ok {
		return nil, errors.New("节点已存在请不要重复添加")
	}
	this.PeerList = append(this.PeerList, p)
	this.peerMap.Store(p.ID, p)
	this.PeerGroup[p.Group] = p.Group
	err := this.savePeerConfig()
	return p, err
}

func (this *SdkApi) EditPeer(p *Peer) (*Peer, error) {
	v, ok := this.peerMap.Load(p.ID)
	if !ok {
		return nil, errors.New("节点不存在！")
	}
	peer, ok := v.(*Peer)
	if !ok {
		return nil, errors.New("节点不存在！！")
	}

	id := fmt.Sprintf("%s:%s", p.Group, p.Name)
	if id != peer.ID {
		_, ok = this.peerMap.Load(id)
		if ok {
			return nil, errors.New("节点不能重复！")
		}
		this.peerMap.Store(id, peer)
		this.peerMap.Delete(p.ID)
	}
	peer.ID = id
	peer.Ip = p.Ip
	peer.Group = p.Group
	peer.Name = p.Name
	peer.Port = p.Port
	peer.Username = p.Username
	peer.Password = p.Password
	this.PeerGroup[p.Group] = p.Group

	err := this.savePeerConfig()
	return peer, err
}

func (this *SdkApi) DelPeer(id string) error {
	v, ok := this.peerMap.Load(id)
	if !ok {
		return errors.New("节点不存在！")
	}
	peer, ok := v.(*Peer)
	if !ok {
		return errors.New("节点不存在！")
	}
	peer.Status = false
	for k, p := range this.PeerList {
		if id == p.ID {
			this.PeerList = append(this.PeerList[:k], this.PeerList[k+1:]...)
			break
		}
	}
	this.savePeerConfig()
	time.Sleep(UpdateDur)
	this.peerMap.Delete(peer.ID)
	return nil
}

func (this *SdkApi) getActivePeer() (list []*Peer) {
	this.peerMap.Range(func(key, value any) bool {
		peer, ok := value.(*Peer)
		if ok {
			if peer.Status {
				list = append(list, peer)
			}
		}
		return true
	})
	return
}

func (this *SdkApi) GetPeerList() []*Peer {
	return this.PeerList
}

func (this *SdkApi) GetPeerGroup() map[string]string {
	return this.PeerGroup
}

func (this *SdkApi) SetStatus(id string, status bool) error {
	v, ok := this.peerMap.Load(id)
	if !ok {
		return errors.New("节点不存在！")
	}
	peer, ok := v.(*Peer)
	if !ok {
		return errors.New("节点不存在！")
	}
	peer.Status = status
	return nil
}

func (this *SdkApi) savePeerConfig() error {
	bs, err := yaml.Marshal(this)
	if err != nil {
		return err
	}
	return utils.SaveFile(PeerConfigFile, &bs)
}

func (this *SdkApi) AddGroup(group string) error {
	_, ok := this.PeerGroup[group]
	if ok {
		return errors.New("分组已存在！")
	}
	this.PeerGroup[group] = group
	return this.savePeerConfig()
}

func (this *SdkApi) EditGroup(group string, old string) error {
	_, ok := this.PeerGroup[old]
	if !ok {
		return errors.New("分组不存在！")
	}
	_, ok = this.PeerGroup[group]
	if ok {
		return errors.New("分组已存在！")
	}
	//修改数据
	this.PeerGroup[group] = group
	delete(this.PeerGroup, old)
	this.peerMap.Range(func(key, value any) bool {
		peer, ok := value.(*Peer)
		if ok {
			if peer.Group == old {
				peer.Group = group
				peer.ID = fmt.Sprintf("%s:%s", peer.Group, peer.Name)
			}
		}
		return true
	})
	return this.savePeerConfig()
}

func (this *SdkApi) DelGroup(group string) error {
	_, ok := this.PeerGroup[group]
	if !ok {
		return errors.New("分组不存在！")
	}
	//修改数据
	delete(this.PeerGroup, group)
	ids := make([]string, 0, 0)
	this.peerMap.Range(func(key, value any) bool {
		peer, ok := value.(*Peer)
		if ok {
			if peer.Group == group {
				peer.Status = false
				ids = append(ids, peer.ID)
				for k, p := range this.PeerList {
					if peer.ID == p.ID {
						this.PeerList = append(this.PeerList[:k], this.PeerList[k+1:]...)
						break
					}
				}
			}
		}
		return true
	})
	this.savePeerConfig()
	if len(ids) > 0 {
		time.Sleep(UpdateDur)
		for _, v := range ids {
			this.peerMap.Delete(v)
		}

	}
	return nil
}

func (this *SdkApi) getClentsAndPeer(id string) (c clents.PeerClient, p *Peer, e error) {
	v, ok := this.peerMap.Load(id)
	if !ok {
		return c, p, errors.New("节点不存在！")
	}
	p, ok = v.(*Peer)
	if !ok {
		return c, p, errors.New("节点不存在！！")
	}

	c = pcli.NewClient(fmt.Sprintf("%s:%d", p.Ip, p.Port), p.Username, p.Password)
	return
}

func (this *SdkApi) GetNewAddress(id, pwd string) error {
	cli, _, err := this.getClentsAndPeer(id)
	if err != nil {
		return err
	}
	err = cli.GetNewAddress(pwd)
	return err
}

func (this *SdkApi) SendToAddress(id, srcAddr string, toAddr string, pwd string, amount float64, gas float64) (*clents.RespSendToAddress, error) {
	cli, _, err := this.getClentsAndPeer(id)
	if err != nil {
		return nil, err
	}
	return cli.SendToAddress(srcAddr, toAddr, pwd, amount, gas)
}

func (this *SdkApi) Depositin(id string, amount float64, gas float64, rate float64, pwd string, payload string) error {
	cli, _, err := this.getClentsAndPeer(id)
	if err != nil {
		return err
	}
	return cli.Depositin(amount, gas, rate, pwd, payload)
}

func (this *SdkApi) Depositout(id string, witnessAddr string, amount float64, gas float64, pwd string) error {
	cli, _, err := this.getClentsAndPeer(id)
	if err != nil {
		return err
	}
	return cli.Depositout(witnessAddr, amount, gas, pwd)
}

func (this *SdkApi) VoteIn(id string, votetype uint16, address string, witness string, amount float64, gas float64, rate float64, pwd string, payload string) error {
	cli, _, err := this.getClentsAndPeer(id)
	if err != nil {
		return err
	}
	return cli.VoteIn(votetype, address, witness, amount, gas, rate, pwd, payload)
}

func (this *SdkApi) VoteOut(id string, votetype uint16, address string, amount float64, gas float64, rate float64, pwd string, payload string) error {
	cli, _, err := this.getClentsAndPeer(id)
	if err != nil {
		return err
	}
	return cli.VoteOut(votetype, address, amount, gas, rate, pwd, payload)
}

func (this *SdkApi) CommunityDistribute(id, srcaddress string, gas float64, pwd string) error {
	cli, _, err := this.getClentsAndPeer(id)
	if err != nil {
		return err
	}
	return cli.CommunityDistribute(srcaddress, gas, pwd)
}

func (this *SdkApi) NameIn(id, address string, amount float64, gas float64, pwd string, name string, netids string, addrcoins string, comment string) error {
	cli, _, err := this.getClentsAndPeer(id)
	if err != nil {
		return err
	}
	return cli.NameIn(address, amount, gas, pwd, name, netids, addrcoins, comment)
}

func (this *SdkApi) NameOut(id, address string, amount float64, gas float64, pwd string, name string, comment string) error {
	cli, _, err := this.getClentsAndPeer(id)
	if err != nil {
		return err
	}
	return cli.NameOut(address, amount, gas, pwd, name, comment)
}

func (this *SdkApi) GetNames(id string) ([]*clents.NameinfoVO, error) {
	cli, _, err := this.getClentsAndPeer(id)
	if err != nil {
		return nil, err
	}
	return cli.GetNames()
}
