package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"

	"genesis-smcrunsvc/genesis/genesis"
	"github.com/bcbchain/bclib/socket"
	abci "github.com/bcbchain/bclib/tendermint/abci/types"
	tmcommon "github.com/bcbchain/bclib/tendermint/tmlibs/common"
	"github.com/bcbchain/bclib/tendermint/tmlibs/log"
	"github.com/bcbchain/bclib/types"
	bcbgls "github.com/bcbchain/sdk/common/gls"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/rlp"
	"github.com/bcbchain/sdk/sdk/std"
	sdkType "github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/sdkimpl"
	"github.com/bcbchain/sdk/sdkimpl/helper"
	"github.com/bcbchain/sdk/sdkimpl/llstate"
	"github.com/bcbchain/sdk/sdkimpl/object"
	"github.com/bcbchain/sdk/sdkimpl/sdkhelper"
	"github.com/spf13/cobra"
)

// 这个是genesis合约的运行服务，其他合约的需要自动生成，与这个不一样。这个经过单独处理。
var (
	logger          log.Loggerf
	flagRPCPort     int
	flagCallbackURL string

	p      *socket.ConnectionPool
	header *abci.Header
)

func pool() *socket.ConnectionPool {
	if p == nil {
		var err error
		p, err = socket.NewConnectionPool(flagCallbackURL, 4, logger)
		if err != nil {
			panic(err)
		}
	}

	return p
}

//adapter回调函数
func set(transID, txID int64, value map[string][]byte) {

	data := make(map[string]string)
	for k, v := range value {
		if v == nil {
			data[k] = string([]byte{})
		} else {
			data[k] = string(v)
		}
	}

	cli, err := pool().GetClient()
	if err != nil {
		panic(err)
	}
	defer pool().ReleaseClient(cli)

	result, err := cli.Call("set", map[string]interface{}{"transID": transID, "txID": txID, "data": data}, 10)
	if err != nil {
		panic(err)
	}

	if bRet, ok := result.(bool); !ok {
		msg := fmt.Sprintf("socket set error: result type is %s", reflect.TypeOf(result).String())
		panic(msg)
	} else if bRet == false {
		msg := "socket set error: return false"
		panic(msg)
	}
}

func get(transID, txID int64, key string) []byte {

	cli, err := pool().GetClient()
	if err != nil {
		panic(err)
	}
	defer pool().ReleaseClient(cli)

	result, err := cli.Call("get", map[string]interface{}{"transID": transID, "txID": txID, "key": key}, 10)
	if err != nil {
		panic(err)
	}

	return []byte(result.(string))
}

func build(transID int64, txID int64, contractMeta std.ContractMeta) std.BuildResult {

	cli, err := pool().GetClient()
	if err != nil {
		panic(err)
	}
	defer pool().ReleaseClient(cli)
	resBytes, _ := jsoniter.Marshal(contractMeta)
	result, err := cli.Call("build", map[string]interface{}{"transID": transID, "txID": txID, "contractMeta": string(resBytes)}, 180)
	if err != nil {
		panic(err)
	}

	var buildResult std.BuildResult
	err = jsoniter.Unmarshal([]byte(result.(string)), &buildResult)
	if err != nil {
		panic(err)
	}

	return buildResult
}

func getBlock(transID, height int64) std.Block {

	if height == 0 {
		block := std.Block{
			ChainID:         header.ChainID,
			BlockHash:       header.LastBlockID.Hash, //todo
			Height:          header.Height,
			Time:            header.Time,
			NumTxs:          header.NumTxs,
			DataHash:        header.DataHash,
			ProposerAddress: header.ProposerAddress,
			RewardAddress:   header.RewardAddress,
			RandomNumber:    header.RandomeOfBlock,
			Version:         header.Version,
			LastBlockHash:   header.LastBlockID.Hash,
			LastCommitHash:  header.LastCommitHash,
			LastAppHash:     header.LastAppHash,
			LastFee:         int64(header.LastFee),
		}

		return block
	}

	cli, err := pool().GetClient()
	if err != nil {
		panic(err)
	}
	defer pool().ReleaseClient(cli)
	result, err := cli.Call("block", map[string]interface{}{"height": height}, 10)
	if err != nil {
		panic(err)
	}

	logger.Infof("getBlock result type is %s", reflect.TypeOf(result).String())
	resBytes, _ := jsoniter.Marshal(result.(map[string]interface{}))
	var blockResult std.Block
	err = jsoniter.Unmarshal(resBytes, &blockResult)
	if err != nil {
		panic(err)
	}

	return blockResult
}

var Routes = map[string]socket.CallBackFunc{
	"Invoke":          Invoke,
	"McCommitTrans":   McCommitTrans,
	"McDirtyTrans":    McDirtyTrans,
	"McDirtyTransTx":  McDirtyTransTx,
	"McDirtyToken":    McDirtyToken,
	"McDirtyContract": McDirtyContract,
	"SetLogLevel":     SetLogLevel,
	"InitSoftForks":   InitSoftForks,
	"Health":          Health,
}

//RunRPC starts RPC service
func RunRPC(port int) error {
	logger = log.NewTMLogger(".", "smcsvc")
	logger.AllowLevel("info")
	logger.SetOutputAsync(true)
	logger.SetOutputToFile(true)
	logger.SetOutputToScreen(false)
	logger.SetOutputFileSize(20000000)

	sdkhelper.Init(transfer, build, set, get, getBlock, nil, &logger)

	svr, err := socket.NewServer("tcp://0.0.0.0:"+fmt.Sprintf("%d", port), Routes, 10, logger)
	if err != nil {
		tmcommon.Exit(err.Error())
	}

	// start server and wait forever
	err = svr.Start()
	if err != nil {
		tmcommon.Exit(err.Error())
	}

	return nil
}

//Invoke invoke function
func Invoke(req map[string]interface{}) (interface{}, error) {
	logger.Info("genesis contract Invoke")

	transID := int64(req["transID"].(float64))
	txID := int64(req["txID"].(float64))
	mCallParam := req["callParam"].(map[string]interface{})
	var callParam types.RPCInvokeCallParam
	jsonStr, _ := jsoniter.Marshal(mCallParam)
	err := jsoniter.Unmarshal(jsonStr, &callParam)
	if err != nil {
		panic(err)
	}

	logger.Debug("Genesis Invoke()", "transID", transID, "txID", txID)
	logger.Trace("Genesis Invoke()", "callParam", callParam)

	smc := NewSMC(transID, txID, callParam)

	contractStub := genesis.New(logger)
	var response types.Response
	bcbgls.Mgr.SetValues(bcbgls.Values{bcbgls.SDKKey: smc}, func() {
		response = contractStub.Invoke(smc)
	})
	//response := contractStub.Invoke(smc)
	smc.(*sdkimpl.SmartContract).Commit()

	resBytes, _ := jsoniter.Marshal(response)

	logger.Debugf("Genesis Invoke() return, result: %s", string(resBytes))
	return string(resBytes), nil
}

func NewSMC(transID, txID int64, callParam types.RPCInvokeCallParam) sdk.ISmartContract {
	sdkReceipts := make([]sdkType.KVPair, 0)
	for _, v := range callParam.Receipts {
		sdkReceipts = append(sdkReceipts, sdkType.KVPair{Key: v.Key, Value: v.Value})
	}

	items := make([]sdkType.HexBytes, 0)
	for _, item := range callParam.Message.Items {
		items = append(items, []byte(item))
	}

	var genesisInfo string
	err := rlp.DecodeBytes(items[0], &genesisInfo)
	if err != nil {
		panic(err)
	}
	req := genesis.RequestInitChain{}
	err = jsoniter.Unmarshal([]byte(genesisInfo), &req)
	if err != nil {
		panic(err)
	}

	smc := &sdkimpl.SmartContract{}
	bcbgls.Mgr.SetValues(bcbgls.Values{bcbgls.SDKKey: smc}, func() {
		llState := llstate.NewLowLevelSDB(smc, transID, txID)
		smc.SetLlState(llState)
		block := object.NewBlock(smc, req.ChainID, "", sdkType.Hash{}, sdkType.Hash{},
			0, 0, 0, "", "",
			sdkType.HexBytes{}, sdkType.Hash{}, sdkType.Hash{}, sdkType.Hash{}, 0)
		smc.SetBlock(block)
		helperObj := helper.NewHelper(smc)
		smc.SetHelper(helperObj)
		blh := helper.BlockChainHelper{}
		genesisOrgID := blh.CalcOrgID("genesis")
		contract := object.NewContractFromSTD(smc, &std.Contract{Address: callParam.Message.Contract, Name: "genesis", OrgID: genesisOrgID})
		msg := object.NewMessage(smc, contract, "", items, callParam.Sender, callParam.Payer,
			nil, nil)
		smc.SetMessage(msg)
	})
	return smc
}

//McCommitTrans commit transaction data of memory cache
func McCommitTrans(req map[string]interface{}) (interface{}, error) {
	transID := int64(req["transID"].(float64))
	sdkhelper.McCommit(transID)
	return true, nil
}

//McDirtyTrans dirty transaction data of memory cache
func McDirtyTrans(req map[string]interface{}) (interface{}, error) {
	transID := int64(req["transID"].(float64))
	sdkhelper.McDirtyTrans(transID)
	return true, nil
}

//McDirtyTransTx dirty tx data of transaction of memory cache
func McDirtyTransTx(req map[string]interface{}) (interface{}, error) {
	transID := int64(req["transID"].(float64))
	txID := int64(req["txID"].(float64))
	sdkhelper.McDirtyTransTx(transID, txID)
	return true, nil
}

//McDirtyToken dirty token data of memory cache
func McDirtyToken(req map[string]interface{}) (interface{}, error) {
	tokenAddr := req["tokenAddr"].(string)
	sdkhelper.McDirtyToken(tokenAddr)
	return true, nil
}

//McDirtyContract dirty contract data of memory cache
func McDirtyContract(req map[string]interface{}) (interface{}, error) {
	contractAddr := req["contractAddr"].(string)
	sdkhelper.McDirtyContract(contractAddr)
	return true, nil
}

//SetLogLevel sets log level
func SetLogLevel(req map[string]interface{}) (interface{}, error) {
	level := req["level"].(string)
	logger.AllowLevel(level)
	return true, nil
}

//InitSoftForks init docker soft forks
func InitSoftForks(req map[string]interface{}) (interface{}, error) {

	return true, nil
}

// Health return health message
func Health(req map[string]interface{}) (interface{}, error) {
	return "health", nil
}

//TransferFunc is used to transfer token for crossing contract invoking.
// nolint unhandled
func transfer(sdk sdk.ISmartContract, tokenAddr, to types.Address, value bn.Number, note string) ([]sdkType.KVPair, sdkType.Error) {
	logger.Debug("TransferFunc", "tokenAddress", tokenAddr, "to", to, "value", value)
	contract := sdk.Helper().ContractHelper().ContractOfToken(tokenAddr)
	logger.Info("Contract", "address", contract.Address(), "name", contract.Name(), "version", contract.Version())
	originMessage := sdk.Message()

	//todo 改成计算的方式
	mID := "af0228bc"
	newSdk := sdkhelper.OriginNewMessage(sdk, contract, mID, nil)

	//todo: 打包参数到message data
	// or 寻找一种方法生成InterfaceStub, 现在的设计会导致循环调用，不可使用。
	tobyte, _ := rlp.EncodeToBytes(to)
	valuebyte, _ := rlp.EncodeToBytes(value.Bytes())

	itemsbyte := make([]sdkType.HexBytes, 0)
	itemsbyte = append(itemsbyte, tobyte)
	itemsbyte = append(itemsbyte, valuebyte)

	newmsg := object.NewMessage(newSdk, newSdk.Message().Contract(), mID, itemsbyte,
		newSdk.Message().Sender().Address(), newSdk.Message().Payer().Address(), newSdk.Message().Origins(), nil)
	newSdk.(*sdkimpl.SmartContract).SetMessage(newmsg)
	contractStub := genesis.New(logger)
	response := contractStub.Invoke(newSdk)
	logger.Debug("Invoke response", "code", response.Code, "tags", response.Tags)
	if response.Code != sdkType.CodeOK {
		return nil, sdkType.Error{ErrorCode: response.Code, ErrorDesc: response.Log}
	}

	// read receipts from response and append to original sdk message
	recKV := make([]sdkType.KVPair, 0)
	for _, v := range response.Tags {
		recKV = append(recKV, sdkType.KVPair{Key: v.Key, Value: v.Value})
	}
	newSdk.(*sdkimpl.SmartContract).SetMessage(originMessage)
	return recKV, sdkType.Error{ErrorCode: sdkType.CodeOK}
}

//RootCmd cmd
var RootCmd = &cobra.Command{
	Use:   "smcrunsvc",
	Short: "grpc",
	Long:  "smcsvc rpc console",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunRPC(flagRPCPort)
	},
}

func main() {
	go func() {
		if e := http.ListenAndServe(":2019", nil); e != nil {
			fmt.Println("pprof cannot start!!!")
		}
	}()

	err := excute()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func excute() error {
	addFlags()
	addCommand()
	return RootCmd.Execute()
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start the smc_service",
	Long:  "start the smc_service",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunRPC(flagRPCPort)
	},
}

func addStartFlags() {
	startCmd.PersistentFlags().IntVarP(&flagRPCPort, "port", "p", 8080, "The port of the smc rpc service")
	startCmd.PersistentFlags().StringVarP(&flagCallbackURL, "callbackUrl", "c", "tcp://localhost:32333", "The url of the adapter callback")
}

func addFlags() {
	addStartFlags()
}

func addCommand() {
	RootCmd.AddCommand(startCmd)
}
