package genesis

import (
	"strings"

	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

// _chainID get chainID
func (g *Genesis) _chainID() (chainID string) {
	return *g.sdk.Helper().StateHelper().GetEx("/genesis/chainid", &chainID).(*string)
}

// _setChainID set chainID
func (g *Genesis) _setChainID(chainID string) {
	key := "/genesis/chainid"
	g.sdk.Helper().StateHelper().Set(key, &chainID)
	g.appStateKeys = append(g.appStateKeys, key)
}

// _chkChainID check chainID exist
func (g *Genesis) _chkChainID() (ok bool) {
	return g.sdk.Helper().StateHelper().Check("/genesis/chainid")
}

// _setStrategys set reward strategy
func (g *Genesis) _setRewardStrategys(rewardStrategys []RewardStrategy) {
	key := "/rewardstrategys"
	g.sdk.Helper().StateHelper().Set(key, &rewardStrategys)
	g.appStateKeys = append(g.appStateKeys, key)
}

// validator
func (g *Genesis) _validator(nodeAddr string) ValidatorStore {
	return *g.sdk.Helper().StateHelper().GetEx("/validator/"+nodeAddr, &ValidatorStore{}).(*ValidatorStore)
}

func (g *Genesis) _setValidator(validator ValidatorStore) {
	key := "/validator/" + validator.NodeAddr
	g.sdk.Helper().StateHelper().Set(key, &validator)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _chkValidator(validatorAddr string) bool {
	return g.sdk.Helper().StateHelper().Check("/validator/" + validatorAddr)
}

func (g *Genesis) _setValidators(values []string) {
	key := "/validators/all/0"
	g.appStateKeys = append(g.appStateKeys, key)
	g.sdk.Helper().StateHelper().Set(key, &values)
}

// _setGenesisToken set genesis token
func (g *Genesis) _setGenesisToken(token std.Token) {
	key := "/genesis/token"
	g.sdk.Helper().StateHelper().Set(key, &token)
	g.appStateKeys = append(g.appStateKeys, key)
}

// _setAccount set account info
func (g *Genesis) _setAccountInfo(accountAddr, tokenAddr string, accountInfo std.AccountInfo) {
	key := "/account/ex/" + accountAddr + "/token/" + tokenAddr
	g.sdk.Helper().StateHelper().Set(key, &accountInfo)
	g.appStateKeys = append(g.appStateKeys, key)
	g.accountChildKeys = append(g.accountChildKeys, key)
}

func (g *Genesis) _setContract(contract std.Contract) {
	key := "/contract/" + contract.Address
	g.sdk.Helper().StateHelper().Set(key, &contract)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setToken(token std.Token) {
	key := "/token/" + token.Address
	g.sdk.Helper().StateHelper().Set(key, &token)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setTokenName(name string, addr types.Address) {
	key := "/token/name/" + strings.ToLower(name)
	g.sdk.Helper().StateHelper().Set(key, &addr)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setTokenSymbol(symbol string, addr types.Address) {
	key := "/token/symbol/" + strings.ToLower(symbol)
	g.sdk.Helper().StateHelper().Set(key, &addr)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setTokenBaseGasPrice(value int64) {
	key := "/token/basegasprice"
	g.sdk.Helper().StateHelper().Set(key, &value)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setOrganization(org std.Organization) {
	key := "/organization/" + org.OrgID
	g.sdk.Helper().StateHelper().Set(key, &org)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setOrgAuthDeployContract(orgID string, addr types.Address) {
	key := "/organization/" + orgID + "/auth"
	g.appStateKeys = append(g.appStateKeys, key)
	g.sdk.Helper().StateHelper().Set(key, &addr)
}

func (g *Genesis) _setContractMeta(meta std.ContractMeta) {
	key := "/contract/code/" + meta.ContractAddr
	g.sdk.Helper().StateHelper().Set("/contract/code/"+meta.ContractAddr, &meta)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setContractVersionInfo(info std.ContractVersionList, orgID string) {
	key := "/contract/" + orgID + "/" + info.Name
	g.sdk.Helper().StateHelper().Set("/contract/"+orgID+"/"+info.Name, &info)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setWorldAppState(state AppState) {
	key := "/world/appstate"
	g.sdk.Helper().StateHelper().Set(key, &state)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setAccountChildKeys(addr types.Address, value []string) {
	key := "/account/ex/" + addr
	g.sdk.Helper().StateHelper().Set(key, &value)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setGenesisContracts(contractAddrs []string) {
	key := "/genesis/contracts"
	g.sdk.Helper().StateHelper().Set("/genesis/contracts", &contractAddrs)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setAccountContracts(accountAddr types.Address, contractAddrs []types.Address) {
	key := "/account/ex/" + accountAddr + "/contracts"
	g.sdk.Helper().StateHelper().Set(key, &contractAddrs)
	g.accountChildKeys = append(g.accountChildKeys, key)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setGenesisContract(contract std.Contract) {
	key := "/genesis/sc/" + contract.Address
	g.sdk.Helper().StateHelper().Set(key, &contract)
	g.appStateKeys = append(g.appStateKeys, key)
}

func (g *Genesis) _setAllToken(value []types.Address) {
	g.appStateKeys = append(g.appStateKeys, std.KeyOfAllToken())
	g.sdk.Helper().StateHelper().Set(std.KeyOfAllToken(), &value)
}

func (g *Genesis) _setChainVersion(version int) {
	key := "/genesis/chainversion"
	g.appStateKeys = append(g.appStateKeys, key)
	g.sdk.Helper().StateHelper().Set(key, &version)
}

func (g *Genesis) _setGenesisOrgID(orgID string) {
	key := "/genesis/orgid"
	g.appStateKeys = append(g.appStateKeys, key)
	g.sdk.Helper().StateHelper().Set(key, &orgID)
}

func (g *Genesis) _setMineContract(m []std.MineContract) {
	key := std.KeyOfMineContracts()
	g.appStateKeys = append(g.appStateKeys, key)
	g.sdk.Helper().StateHelper().Set(key, &m)
}

func (g *Genesis) _effectHeightContractAddrs(height string) (contractWithHeight []std.ContractWithEffectHeight) {
	return *g.sdk.Helper().StateHelper().GetEx("/"+height, &contractWithHeight).(*[]std.ContractWithEffectHeight)
}

func (g *Genesis) _setEffectHeightContractAddrs(height string, contractWithHeight []std.ContractWithEffectHeight) {
	key := "/" + height
	g.appStateKeys = append(g.appStateKeys, key)
	g.sdk.Helper().StateHelper().Set(key, &contractWithHeight)
}

// ChainValidator
func (g *Genesis) _setChainValidator(chainID string, cvp map[string]ValidatorStore) {
	key := "/ibc/" + chainID
	g.appStateKeys = append(g.appStateKeys, key)
	g.sdk.Helper().StateHelper().Set(key, &cvp)
}

func (g *Genesis) _setGasPriceRatio(gasPriceRatio string) {
	g.appStateKeys = append(g.appStateKeys, std.KeyOfGasPriceRatio())
	g.sdk.Helper().StateHelper().Set(std.KeyOfGasPriceRatio(), gasPriceRatio)
}

// set open URL
func (g *Genesis) _setOpenURLs(chainID string, urls []string) {
	key := "/sidechain/" + chainID + "/openurls"
	g.appStateKeys = append(g.appStateKeys, key)
	g.sdk.Helper().StateHelper().Set(key, &urls)
}

func (g *Genesis) _setAllOrganization(value []string) {
	g.appStateKeys = append(g.appStateKeys, "/organization/all/0")
	g.sdk.Helper().StateHelper().Set("/organization/all/0", &value)
}

// bvm switch
func (g *Genesis) _BVMStatus() bool {
	var enable bool
	return *g.sdk.Helper().StateHelper().GetEx(g.keyOfBVMStatus(), &enable).(*bool)
}

func (g *Genesis) _setBVMStatus(enable bool) {
	g.appStateKeys = append(g.appStateKeys, g.keyOfBVMStatus())
	g.sdk.Helper().StateHelper().Set(g.keyOfBVMStatus(), &enable)
}

func (g *Genesis) keyOfBVMStatus() string {
	return "/bvm/status"
}
