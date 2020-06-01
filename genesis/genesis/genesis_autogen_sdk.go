package genesis

import (
	"github.com/bcbchain/sdk/sdk"
)

//SetSdk This is a method of Genesis
func (g *Genesis) SetSdk(sdk sdk.ISmartContract) {
	g.sdk = sdk
}

//GetSdk This is a method of Genesis
func (g *Genesis) GetSdk() sdk.ISmartContract {
	return g.sdk
}
