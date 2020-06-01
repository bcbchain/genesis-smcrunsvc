package genesis

import (
	"genesis-smcrunsvc/genesis/stubcommon/common"
	stubType "genesis-smcrunsvc/genesis/stubcommon/types"
	"github.com/bcbchain/bclib/tendermint/tmlibs/log"
	bcType "github.com/bcbchain/bclib/types"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/rlp"
	"github.com/bcbchain/sdk/sdk/types"
)

//GenesisStub an object
type GenesisStub struct {
	logger log.Logger
}

var _ stubType.IContractStub = (*GenesisStub)(nil)

//New generate a stub
func New(logger log.Logger) stubType.IContractStub {
	return &GenesisStub{logger: logger}
}

//FuncRecover recover panic by Assert
func FuncRecover(response *bcType.Response) {
	if err := recover(); err != nil {
		if _, ok := err.(types.Error); ok {
			response.Code = err.(types.Error).ErrorCode
			response.Log = err.(types.Error).ErrorDesc
		} else {
			panic(err)
		}
	}
}

//Invoke invoke function
func (pbs *GenesisStub) Invoke(smc sdk.ISmartContract) (response bcType.Response) {
	defer FuncRecover(&response)

	var data string
	data = createGenesis(smc)
	response = common.CreateResponse(smc.Message(), data, 0, 0, 0)
	return
}

func createGenesis(smc sdk.ISmartContract) string {
	items := smc.Message().Items()
	sdk.Require(len(items) == 1, types.ErrStubDefined, "Invalid message data")

	var err error

	var genesisInfo string
	err = rlp.DecodeBytes(items[0], &genesisInfo)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	sdk.Require(len(genesisInfo) > 0,
		types.ErrInvalidParameter, "Invalid genesisInfo")
	req := RequestInitChain{}
	err = jsoniter.Unmarshal([]byte(genesisInfo), &req)
	sdk.RequireNotError(err, types.ErrInvalidParameter)

	contractObj := new(Genesis)
	contractObj.SetSdk(smc)
	rest0 := contractObj.CreateGenesis(req)
	resultList := make([]interface{}, 0)
	resultList = append(resultList, rest0)

	resBytes, _ := jsoniter.Marshal(resultList)
	return string(resBytes)
}
