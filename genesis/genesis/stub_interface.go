package genesis

import (
	bcType "github.com/bcbchain/bclib/types"

	"genesis-smcrunsvc/genesis/stubcommon/common"
	stubType "genesis-smcrunsvc/genesis/stubcommon/types"
	tmcommon "github.com/bcbchain/bclib/tendermint/tmlibs/common"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/types"
)

//InterfaceGenesisStub interface stub
type InterfaceGenesisStub struct {
	smc sdk.ISmartContract
}

var _ stubType.IContractIntfcStub = (*InterfaceGenesisStub)(nil)

//NewInterStub new interface stub
func NewInterStub(smc sdk.ISmartContract) stubType.IContractIntfcStub {
	return &InterfaceGenesisStub{smc: smc}
}

//GetSdk get sdk
func (inter *InterfaceGenesisStub) GetSdk() sdk.ISmartContract {
	return inter.smc
}

//SetSdk set sdk
func (inter *InterfaceGenesisStub) SetSdk(smc sdk.ISmartContract) {
	inter.smc = smc
}

//Invoke invoke function
func (inter *InterfaceGenesisStub) Invoke(methodID string, p interface{}) (response bcType.Response) {
	defer FuncRecover(&response)

	// 生成手续费收据
	fee, gasUsed, feeReceipt, err := common.FeeAndReceipt(inter.smc, true)
	if err.ErrorCode != types.CodeOK {
		response = common.CreateResponse(inter.smc.Message(), "", fee, gasUsed, inter.smc.Tx().GasLimit())
		return
	}
	response.Fee = fee
	response.GasUsed = gasUsed
	response.Tags = append(response.Tags, tmcommon.KVPair{Key: feeReceipt.Key, Value: feeReceipt.Value})

	var data string
	switch methodID {
	}
	response = common.CreateResponse(inter.smc.Message(), data, fee, gasUsed, inter.smc.Tx().GasLimit())
	return
}
