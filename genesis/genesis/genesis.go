package genesis

import (
	"bytes"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"

	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/crypto/sha3"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
)

//Genesis is the genesis contract
//@:contract:genesis
//@:version:2.1
//@:organization:orgJgaGConUyK81zibntUBjQ33PKctpk1K1G
//@:author:5e8339cb1a5cce65602fd4f57e115905348f7e83bcbe38dd77694dbe1f8903c9
type Genesis struct {
	sdk sdk.ISmartContract

	chainID             string
	rewardStrategy      *[]RewardStrategy
	validator           *Validator
	token               *std.Token
	organization        *std.Organization
	contract            *std.Contract
	contractMeta        *std.ContractMeta
	contractVersionList *std.ContractVersionList
	worldAppState       *AppState

	appStateKeys     []string // calc genesis world app state hash
	accountChildKeys []string // account all child keys
}

//InitChain Constructor of this Genesis
//@:constructor
func (g *Genesis) InitChain() {

}

//CreateGenesis create genesis of the blockchain world
//@:public:method:gas[500]
// nolint gocyclo
func (g *Genesis) CreateGenesis(req RequestInitChain) (response ResponseInitChain) {
	genesisState := req.AppState
	helper := g.sdk.Helper().BlockChainHelper()

	sdk.Require(len(req.ChainID) > 0,
		types.ErrInvalidParameter, "Invalid ChainID.")
	g._setChainID(req.ChainID)

	// check and save reward strategy
	g.checkRewardStrategy(genesisState.RewardStrategy)
	initStrategy := RewardStrategy{genesisState.RewardStrategy, 1}
	rewardStrategy := make([]RewardStrategy, 0)
	rewardStrategy = append(rewardStrategy, initStrategy)
	g._setRewardStrategys(rewardStrategy)

	// check and save validators info
	if len(req.Validators) >= 4 {
		g.checkValidators(req.Validators)
	}

	var allNodeAddrs []string
	validatorStores := make([]ValidatorStore, len(req.Validators))
	for i, validator := range req.Validators {
		validator.NodeAddr = helper.CalcAccountFromPubKey(validator.PubKey)
		vs := ValidatorStore{
			PubKey:     validator.PubKey,
			Power:      validator.Power,
			RewardAddr: validator.RewardAddr,
			Name:       validator.Name,
			NodeAddr:   validator.NodeAddr,
		}
		validatorStores[i] = vs
		g._setValidator(vs)
		allNodeAddrs = append(allNodeAddrs, vs.NodeAddr)
	}
	g._setValidators(allNodeAddrs)

	// check genesis token
	token := genesisState.Token
	g.checkGenesisToken(token)

	// check and save genesis organization
	g.checkGenesisOrg(genesisState.Organization)
	var org std.Organization
	var allOrg []string
	org = std.Organization{
		OrgID:    helper.CalcOrgID(genesisState.Organization),
		Name:     genesisState.Organization,
		OrgOwner: token.Owner,
	}
	g._setOrganization(org)
	g._setGenesisOrgID(org.OrgID)
	g._setOrgAuthDeployContract(org.OrgID, org.OrgOwner)
	allOrg = append(allOrg, org.OrgID)

	if len(req.AppState.OrgBind.OrgName) > 0 {
		orgBind := std.Organization{
			OrgID:    helper.CalcOrgID(req.AppState.OrgBind.OrgName),
			Name:     req.AppState.OrgBind.OrgName,
			OrgOwner: req.AppState.OrgBind.Owner,
		}

		g._setOrganization(orgBind)
		g._setOrgAuthDeployContract(orgBind.OrgID, orgBind.OrgOwner)
		allOrg = append(allOrg, orgBind.OrgID)
	}
	g._setAllOrganization(allOrg)
	g.sdk.Helper().StateHelper().Flush()

	var contractAddrList []string
	var orgCodeHashBytes []byte
	var orgSignersStr []string
	var orgSigners []string
	var genTokenAddress string
	for _, contract := range genesisState.Contracts {
		// check genesis contact info
		g.checkGenesisContract(contract)

		// 如果本币地址存在，token-basic 合约地址就用本币地址，如果没有本币地址，就重新计算合约地址，本币地址即合约地址。
		contractAddr := helper.CalcContractAddress(
			contract.Name,
			contract.Version,
			org.OrgID)

		var tokenAddress types.Address
		if contract.Name == "token-basic" {
			if req.AppState.Token.Address != "" {
				tokenAddress = req.AppState.Token.Address
				contractAddr = req.AppState.Token.Address
			} else {
				tokenAddress = contractAddr
			}

			genTokenAddress = tokenAddress
		}

		contractAddrList = append(contractAddrList, contractAddr)

		codeDevSigBytes, _ := jsoniter.Marshal(contract.CodeDevSig)
		codeDevSigBytes, _ = jsoniter.Marshal(string(codeDevSigBytes))
		codeOrgSigBytes, _ := jsoniter.Marshal(contract.CodeOrgSig)
		codeOrgSigBytes, _ = jsoniter.Marshal(string(codeOrgSigBytes))
		sdk.Require(len(contract.CodeByte) > 0,
			types.ErrInvalidParameter, "Invalid contract code.")
		contractCodeHash, errOfDecode := hex.DecodeString(contract.CodeHash)
		sdk.Require(errOfDecode == nil,
			types.ErrInvalidParameter, "Invalid contract codeHash.")

		// build genesis contract
		conCodeHash, _ := hex.DecodeString(contract.CodeHash)
		buildResult := g.sdk.Helper().BuildHelper().Build(std.ContractMeta{
			Name:         contract.Name,
			ContractAddr: contractAddr,
			OrgID:        org.OrgID,
			Version:      contract.Version,
			EffectHeight: 1,
			LoseHeight:   0,
			CodeData:     contract.CodeByte,
			CodeHash:     conCodeHash,
			CodeDevSig:   codeDevSigBytes,
			CodeOrgSig:   codeOrgSigBytes,
		})
		sdk.Require(buildResult.Code == types.CodeOK, buildResult.Code, "build contract failed!"+buildResult.Error)

		if len(orgSigners) != 0 {
			ifExit := false
			for _, orgSigner := range orgSigners {
				if orgSigner == strings.ToLower(contract.CodeOrgSig.PubKey) {
					ifExit = true
				}
			}

			if !ifExit {
				orgSigners = append(orgSigners, strings.ToLower(contract.CodeOrgSig.PubKey))
			}

		} else {
			orgSigners = append(orgSigners, strings.ToLower(contract.CodeOrgSig.PubKey))
		}

		orgSignersStr = append(orgSignersStr, contract.CodeOrgSig.PubKey)
		orgCodeHashBytes = buildResult.OrgCodeHash
		org.ContractAddrList = contractAddrList
		org.OrgCodeHash = orgCodeHashBytes
		g._setOrganization(org)

		var contractOwner types.Address
		if contract.Owner != "" {
			contractOwner = contract.Owner
		} else {
			contractOwner = token.Owner
		}
		// save contract
		newContract := std.Contract{
			Address:      contractAddr,
			Account:      helper.CalcAccountFromName(contract.Name, org.OrgID),
			Owner:        contractOwner,
			Name:         contract.Name,
			Version:      contract.Version,
			CodeHash:     contractCodeHash,
			EffectHeight: 1,
			LoseHeight:   0,
			KeyPrefix:    "",
			Methods:      buildResult.Methods,
			Interfaces:   buildResult.Interfaces,
			Mine:         buildResult.Mine,
			IBCs:         buildResult.IBCs,
			Token:        tokenAddress,
			OrgID:        org.OrgID,
			ChainVersion: 2,
		}

		g._setContract(newContract)
		g._setGenesisContract(newContract)

		// save contract metadata
		g._setContractMeta(std.ContractMeta{
			Name:         contract.Name,
			ContractAddr: contractAddr,
			OrgID:        org.OrgID,
			Version:      newContract.Version,
			EffectHeight: newContract.EffectHeight,
			LoseHeight:   newContract.LoseHeight,
			CodeData:     contract.CodeByte,
			CodeHash:     contractCodeHash,
			CodeDevSig:   codeDevSigBytes,
			CodeOrgSig:   codeOrgSigBytes,
		})

		// save contract version and effect height list
		g._setContractVersionInfo(std.ContractVersionList{
			Name:             contract.Name,
			ContractAddrList: []string{contractAddr},
			EffectHeights:    []int64{newContract.EffectHeight},
		}, org.OrgID)

		newConWithHeight := std.ContractWithEffectHeight{
			ContractAddr: contractAddr,
			IsUpgrade:    false,
		}
		conWithHeight := g._effectHeightContractAddrs("1")
		conWithHeight = append(conWithHeight, newConWithHeight)
		g._setEffectHeightContractAddrs("1", conWithHeight)

		g.sdk.Helper().StateHelper().Flush()
	}
	sdk.Require(genTokenAddress != "",
		types.ErrInvalidParameter, "Must include genesis token contract \"token-basic\".")

	for _, orgSigner := range orgSigners {
		signerPubKey, err := hex.DecodeString(orgSigner)
		sdk.Require(err == nil,
			types.ErrInvalidParameter, "Invalid orgSinger's pubKey.")
		org.Signers = append(org.Signers, types.PubKey(signerPubKey))
	}

	g._setOrganization(org)

	// set genesis contract address list
	g._setGenesisContracts(contractAddrList)

	// set contract owner's contract address list
	g._setAccountContracts(token.Owner, contractAddrList)

	// set genesis token
	token.Address = genTokenAddress
	g._setGenesisToken(token)
	g._setToken(token)
	g._setTokenName(token.Name, token.Address)
	g._setTokenSymbol(token.Symbol, token.Address)
	g._setTokenBaseGasPrice(token.GasPrice)

	// this is genesis, so no query required
	var allTokenAddr []types.Address = nil
	allTokenAddr = append(allTokenAddr, genTokenAddress)
	g._setAllToken(allTokenAddr)

	// set genesis token owner balance
	balance := std.AccountInfo{
		Address: token.Address,
		Balance: token.TotalSupply,
	}
	g._setAccountInfo(token.Owner, token.Address, balance)
	// set account all child keys
	g._setAccountChildKeys(token.Owner, g.accountChildKeys)

	// set chain version
	g._setChainVersion(2)

	// set gasPriceRatio
	if req.AppState.GasPriceRatio != "" {
		g._setGasPriceRatio(req.AppState.GasPriceRatio)
	}

	// set validators by chainID --sideChain
	sideChainValidators := make(map[string]ValidatorStore)
	for _, node := range validatorStores {
		sideChainValidators[node.NodeAddr] = node
	}
	g._setChainValidator(req.ChainID, sideChainValidators)

	// set validators by chainID --mainChain
	mainChainID := g.sdk.Helper().BlockChainHelper().GetMainChainID()
	if len(req.AppState.MainChain.Validators) > 0 {
		mainChainValidators := req.AppState.MainChain.Validators
		g._setChainValidator(mainChainID, mainChainValidators)

		// set mainChain openURLs
		g._setOpenURLs(mainChainID, req.AppState.MainChain.OpenUrls)
	}

	// BVM 默认关闭
	g._setBVMStatus(false)

	// 这部分要在最后面，确保所有 key 都参与计算 appHash
	genAppState := g.calcWorldAppStateHash()
	g._setWorldAppState(genAppState)

	response.Code = types.CodeOK
	genAppStateBytes, err := jsoniter.Marshal(&genAppState)
	sdk.Require(err == nil,
		types.ErrInvalidParameter, "Marshal app state failed.")
	response.GenAppState = genAppStateBytes

	return
}

func (g *Genesis) checkRewardStrategy(rwdStrategy []Reward) {

	var percent float64
	var haveNameOfValidators bool

	for _, st := range rwdStrategy {
		// check name length
		sdk.Require(len(st.Name) > 0 && len(st.Name) <= MaxNameLen,
			types.ErrInvalidParameter, "Invalid name in strategy")

		// check percent format
		if strings.Contains(st.RewardPercent, ".") {
			index := strings.IndexByte(st.RewardPercent, '.')
			sub := []byte(st.RewardPercent)[index+1:]
			sdk.Require(len(sub) == 2,
				types.ErrInvalidParameter, "Invalid reward percent")
		}

		nodePer, err := strconv.ParseFloat(st.RewardPercent, 64)
		sdk.RequireNotError(err, types.ErrInvalidParameter)

		percent = percent + nodePer
		sdk.Require(nodePer > 0.00 && percent <= 100.00,
			types.ErrInvalidParameter, "Invalid reward percent")

		// Check Address, check name with "validators", for validators, we don't care about its address
		if st.Name == "validators" {
			haveNameOfValidators = true
		} else {
			sdk.RequireAddress(st.Address)
		}
	}

	sdk.Require(percent == 100.00,
		types.ErrInvalidParameter, "Invalid reward percent")
	sdk.Require(haveNameOfValidators,
		types.ErrInvalidParameter, "Lose name of validators")
}

func (g *Genesis) checkValidators(validators []Validator) {

	totalPower := bn.N(0)
	maxPower := bn.N(0)
	for _, validator := range validators {

		power := bn.N(validator.Power)
		totalPower = totalPower.Add(power)

		if power.IsGreaterThan(maxPower) {
			maxPower = power
		}
	}

	// If the maxPower is equal to or over 1/3 totalPower, return smc.Error to avoid this happens
	sdk.Require(maxPower.IsLessThan(totalPower.DivI(3)),
		types.ErrInvalidParameter, "Invalid power, max power is greater than or equal to 1/3 total power.")
}

func (g *Genesis) checkGenesisToken(token std.Token) {
	sdk.Require(len(token.Owner) != 0,
		types.ErrInvalidParameter, "Genesis token must has owner.")
	sdk.Require(len(token.Name) > 0 && len(token.Name) < MaxNameLen && len(token.Symbol) > 0 && len(token.Symbol) <= MaxSymbolLen,
		types.ErrInvalidParameter, "Invalid name or symbol.")
	sdk.Require(token.GasPrice != 0,
		types.ErrInvalidParameter, "Invalid gasPrice")

	//If chainID is mainChain ID
	if !g.sdk.Helper().BlockChainHelper().IsSideChain() {
		sdk.Require(token.TotalSupply.IsPositive(),
			types.ErrInvalidParameter, "Invalid totalSupply")
	}
}

func (g *Genesis) checkGenesisOrg(name string) {
	sdk.Require(name == "genesis",
		types.ErrInvalidParameter, "Invalid genesis org name.")
}

func (g *Genesis) checkGenesisContract(contract Contract) {
	sdk.Require(contract.Name != "" && contract.Version != "",
		types.ErrInvalidParameter, "Contract name can not empty.")
	sdk.Require(len(contract.CodeByte) > 0 && contract.CodeHash != "",
		types.ErrInvalidParameter, "Invalid contract code or codeHash.")
	sdk.Require(contract.CodeDevSig.PubKey != "" && contract.CodeDevSig.Signature != "",
		types.ErrInvalidParameter, "Code dev sig can not empty.")

	devPubKey, err := hex.DecodeString(contract.CodeDevSig.PubKey)
	sdk.Require(err == nil && len(devPubKey) == 32,
		types.ErrInvalidParameter, "Invalid codeDevSig.")

	sdk.Require(contract.CodeOrgSig.PubKey != "" && contract.CodeOrgSig.Signature != "",
		types.ErrInvalidParameter, "Code org sig can not empty.")

	devOrgKey, err := hex.DecodeString(contract.CodeOrgSig.PubKey)
	sdk.Require(err == nil && len(devOrgKey) == 32,
		types.ErrInvalidParameter, "Invalid codeDevSig.")
}

// nolint gocyclo
func (g *Genesis) calcWorldAppStateHash() AppState {

	sort.Strings(g.appStateKeys)
	var buf bytes.Buffer
	for _, k := range g.appStateKeys {
		obj := g.sdk.Helper().StateHelper().Get(k, new(interface{}))
		v, _ := jsoniter.Marshal(obj)
		buf.Write([]byte(k))
		buf.Write(v)
	}
	appStateHash := sha3.Sum256(buf.Bytes())
	genAppState := AppState{
		BlockHeight:  0,
		AppHash:      appStateHash,
		ChainVersion: 2,
	}

	return genAppState
}
