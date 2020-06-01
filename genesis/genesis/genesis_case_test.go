package genesis

import (
	"fmt"
	"github.com/bcbchain/sdk/sdk/bn"
	"github.com/bcbchain/sdk/sdk/jsoniter"
	"github.com/bcbchain/sdk/sdk/std"
	"github.com/bcbchain/sdk/sdk/types"
	"github.com/bcbchain/sdk/utest"
	"testing"

	"gopkg.in/check.v1"
)

//Test This is a function
func Test(t *testing.T) { check.TestingT(t) }

//MySuite This is a struct
type MySuite struct{}

var _ = check.Suite(&MySuite{})

//TestGenesis_InitChain This is a method of MySuite
func (mysuit *MySuite) TestGenesis_InitChain(c *check.C) {
	utest.Init(orgID)
	contractOwner := utest.DeployContract(c, contractName, orgID, contractMethods, contractInterfaces)
	test := NewTestObject(contractOwner)

	req := RequestInitChain{
		Validators: []Validator{
			{
				PubKey:     []byte("abcdabcdabcdabcdabcdabcdabcdabc2"),
				Power:      50,
				RewardAddr: "reward_address1",
				Name:       "validators1",
				NodeAddr:   "node-address1",
			},
			{
				PubKey:     []byte("abcdabcdabcdabcdabcdabcdabcdabc3"),
				Power:      50,
				RewardAddr: "reward_address2",
				Name:       "validators2",
				NodeAddr:   "node-address2",
			},
			{
				PubKey:     []byte("abcdabcdabcdabcdabcdabcdabcdabc4"),
				Power:      50,
				RewardAddr: "reward_address3",
				Name:       "validators3",
				NodeAddr:   "node-address3",
			},
			{
				PubKey:     []byte("abcdabcdabcdabcdabcdabcdabcdabc5"),
				Power:      50,
				RewardAddr: "reward_address4",
				Name:       "validators4",
				NodeAddr:   "node-address4",
			},
		},
		ChainID: "testChainID",
		AppState: InitAppState{
			Organization: "genesis",
			Token: std.Token{
				Address:          "token_address",
				Owner:            "owner_address",
				Name:             "token_name",
				Symbol:           "ton",
				TotalSupply:      bn.N(1000),
				AddSupplyEnabled: false,
				BurnEnabled:      false,
				GasPrice:         100,
			},
			RewardStrategy: []Reward{
				{
					Name:          "validators",
					RewardPercent: "100.00",
					Address:       "rewarder_addr"},
			},
			Contracts: []Contract{
				{
					Name:     "token-basic",
					Version:  "v1.0",
					CodeHash: "f1edf8f50848b8fa121a24e2a3a83cc5c8cbf85d6ce23a3a8413f46a717beda1",
					CodeDevSig: Signature{
						PubKey:    "f1edf8f50848b8fa121a24e2a3a83cc5c8cbf85d6ce23a3a8413f46a717beda1",
						Signature: "2",
					},
					CodeOrgSig: Signature{
						PubKey:    "f1edf8f50848b8fa121a24e2a3a83cc5c8cbf85d6ce23a3a8413f46a717beda1",
						Signature: "1",
					},
				},
			},
		},
	}
	reqBytes, _ := jsoniter.Marshal(req)
	fmt.Println(string(reqBytes))

	testCases := []struct {
		initInfo string
		err      types.Error
	}{
		{string(reqBytes), types.Error{ErrorCode: types.CodeOK}},
	}

	for _, v := range testCases {
		result, err := test.run().CreateGenesis(v.initInfo)
		fmt.Println(err)
		utest.Assert(err.ErrorCode == v.err.ErrorCode)
		if v.err.ErrorCode == types.CodeOK {
			fmt.Println(result)
		} else {
			fmt.Println(err.ErrorDesc)
		}
	}

}
