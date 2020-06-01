package genesis

import (
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

// RequestInitChain genesis info
type RequestInitChain struct {
	Validators []Validator  `json:"validators"`
	ChainID    string       `json:"chain_id"`
	AppState   InitAppState `json:"app_state"`
}

// ValidatorStore store validator info
type ValidatorStore struct {
	PubKey     types.PubKey `json:"nodepubkey,omitempty"` //节点公钥
	Power      int64        `json:"power,omitempty"`      //节点记账权重
	RewardAddr string       `json:"rewardaddr,omitempty"` //节点接收奖励的地址
	Name       string       `json:"name,omitempty"`       //节点名称
	NodeAddr   string       `json:"nodeaddr,omitempty"`   //节点地址
}

// Validator validator info
type Validator struct {
	PubKey     []byte `json:"pub_key,omitempty"`
	Power      int64  `json:"power,omitempty"`
	RewardAddr string `json:"reward_addr,omitempty"`
	Name       string `json:"name,omitempty"`
	NodeAddr   string `json:"node_addr,omitempty"`
}

type OrgBind struct {
	OrgName string        `json:"orgName"`
	Owner   types.Address `json:"owner"`
}

type MainChainInfo struct {
	OpenUrls   []string                  `json:"openUrls"`
	Validators map[string]ValidatorStore `json:"validators"`
}

// InitAppState initChain app state info
type InitAppState struct {
	Organization   string        `json:"organization"`
	GasPriceRatio  string        `json:"gas_price_ratio"`
	Token          std.Token     `json:"token"`
	RewardStrategy []Reward      `json:"rewardStrategy"`
	Contracts      []Contract    `json:"contracts"`
	OrgBind        OrgBind       `json:"orgBind"`
	MainChain      MainChainInfo `json:"mainChain"`
}

// Reward reward info
type Reward struct {
	Name          string `json:"name"`          // 被奖励者名称
	RewardPercent string `json:"rewardPercent"` // 奖励比例
	Address       string `json:"address"`       // 被奖励者地址
}

// Contract contract info
type Contract struct {
	Name       string         `json:"name,omitempty"`
	Version    string         `json:"version,omitempty"`
	Owner      string         `json:"owner,omitempty"`
	CodeByte   types.HexBytes `json:"codeByte,omitempty"`
	CodeHash   string         `json:"codeHash,omitempty"`
	CodeDevSig Signature      `json:"codeDevSig,omitempty"`
	CodeOrgSig Signature      `json:"codeOrgSig,omitempty"`
}

// Signature sig for contract code
type Signature struct {
	PubKey    string `json:"pubkey"`
	Signature string `json:"signature"`
}

// RewardStrategy stored reward list and effectHeight
type RewardStrategy struct {
	Strategy     []Reward `json:"rewardStrategy,omitempty"` //奖励策略
	EffectHeight uint64   `json:"effectHeight,omitempty"`   //生效高度
}

// ResponseInitChain return info for genesis
type ResponseInitChain struct {
	Code        uint32 `json:"code,omitempty"`
	Log         string `json:"log,omitempty"`
	GenAppState []byte `json:"gen_app_state,omitempty"`
}

// AppState world app state
type AppState struct {
	BlockHeight  int64          `json:"block_height,omitempty"`  //最后一个确认的区块高度
	AppHash      types.HexBytes `json:"app_hash,omitempty"`      //最后一个确认区块的AppHash
	ChainVersion int64          `json:"chain_version,omitempty"` //当前链版本
}

// SideChain Own Info
type SideChainOrg struct {
	OrgName string        `json:"orgName"`
	Owner   types.Address `json:"owner"`
}

const (
	// MaxNameLen name max length
	MaxNameLen = 40
	// MaxSymbolLen token symbol can be up to 20 characters
	MaxSymbolLen = 20
)
