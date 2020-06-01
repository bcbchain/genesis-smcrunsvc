package common

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	types2 "github.com/bcbchain/bclib/types"
	"github.com/bcbchain/sdk/sdk"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/crypto/sha3"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/sdkimpl"
	"github.com/bcbchain/sdk/sdkimpl/object"
)

// TODO 这个stub 只是为了genesis合约，其他合约不适用！！！
//CreateResponse create response data
//TODO: Fee 和 gasUsed 只是本合约调用消耗的gas， 跨合约调用消耗的无法获取到，
//       建议此处不赋值，而是由bcchain根据gaslimit-gasused计算出总的gasUsed.
func CreateResponse(message sdk.IMessage, data string, fee, gasUsed, gasLimit int64) (response types2.Response) {
	response.Code = types.CodeOK
	response.Data = data
	response.GasLimit = gasLimit
	return
}

//FeeAndReceipt pay fee for the calling and emit fee receipt
func FeeAndReceipt(smc sdk.ISmartContract, bMethod bool) (fee, gasUsed int64, receipt types.KVPair, err types.Error) {

	err.ErrorCode = types.CodeOK
	//Get gas price
	var gasprice int64
	if smc.Message().Contract().Token() == "" {
		gasprice = smc.Helper().TokenHelper().BaseGasPrice()
	} else {
		gasprice = smc.Helper().TokenHelper().Token().GasPrice()
	}
	//calculate fee
	var methods []std.Method
	if bMethod {
		methods = smc.Message().Contract().Methods()
	} else {
		methods = smc.Message().Contract().Interfaces()
	}
	var gas int64
	for _, m := range methods {
		if m.MethodID == smc.Message().MethodID() {
			gas = m.Gas
			break
		}
	}
	gasAbs := int64(math.Abs(float64(gas))) //abs number

	gasLeft := smc.Tx().GasLeft()
	if gasLeft < gasAbs {
		gasUsed = gasLeft
		err.ErrorCode = types.ErrGasNotEnough
	} else {
		gasUsed = gasAbs
	}
	fee = gasprice * gasUsed

	payer := smc.Message().Payer()
	token := smc.Helper().GenesisHelper().Token().Address()
	balance := payer.BalanceOfToken(token)
	if balance.IsLessThanI(fee) {
		fee = balance.V.Int64()
		balance = bn.N(0)
		err.ErrorCode = types.ErrFeeNotEnough
	} else {
		balance = balance.SubI(fee)
	}
	payer.(*object.Account).SetBalanceOfToken(token, balance)

	//Set gasLeft to tx
	gasLeft = gasLeft - gasUsed
	smc.Tx().(*object.Tx).SetGasLeft(gasLeft)
	//emit receipt
	feeReceipt := std.Fee{
		Token: smc.Helper().GenesisHelper().Token().Address(),
		From:  payer.Address(),
		Value: fee,
	}
	receipt = emitFeeReceipt(smc, feeReceipt)

	return
}

func CalcKey(name, version string) string {
	if strings.HasPrefix(name, "token-templet-") {
		name = "token-issue"
	}
	return name + "_" + strings.Replace(version, ".", "_", -1)
}

func emitFeeReceipt(smc sdk.ISmartContract, receipt std.Fee) types.KVPair {
	bz, err := jsoniter.Marshal(receipt)
	if err != nil {
		sdkimpl.Logger.Fatalf("[sdk]Cannot marshal receipt data=%v", receipt)
		sdkimpl.Logger.Flush()
		panic(err)
	}

	rcpt := std.Receipt{
		Name:         receiptName(receipt),
		ContractAddr: smc.Message().Contract().Address(),
		Bytes:        bz,
		Hash:         nil,
	}
	rcpt.Hash = sha3.Sum256([]byte(rcpt.Name), []byte(rcpt.ContractAddr), bz)
	resBytes, _ := jsoniter.Marshal(rcpt) // nolint unhandled

	result := types.KVPair{
		Key:   []byte(fmt.Sprintf("/%d/%s", len(smc.Message().(*object.Message).OutputReceipts()), rcpt.Name)),
		Value: resBytes,
	}

	return result
}

func receiptName(receipt interface{}) string {
	typeOfInterface := reflect.TypeOf(receipt).String()

	if strings.HasPrefix(typeOfInterface, "std.") {
		prefixLen := len("std.")
		return "std::" + strings.ToLower(typeOfInterface[prefixLen:prefixLen+1]) + typeOfInterface[prefixLen+1:]
	}

	return typeOfInterface
}
