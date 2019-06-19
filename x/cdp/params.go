package cdp

import (
	"github.com/commercionetwork/cosmos-hackatom-2019/x/types"

	"github.com/cosmos/cosmos-sdk/x/params"
)

/*
How this uses the sdk params module:
 - Put all the params for this module in one struct `CDPModuleParams`
 - Store this in the keeper's paramSubspace under one key
 - Provide a function to load the param struct all at once `keeper.GetParams(ctx)`
It's possible to set individual key value pairs within a paramSubspace, but reading and setting them is awkward (an empty variable needs to be created, then Get writes the value into it)
This approach will be awkward if we ever need to write individual parameters (because they're stored all together). If this happens do as the sdk modules do - store parameters separately with custom get/set func for each.
*/

var moduleParamsKey = []byte("CdpModuleParams")

func createParamsKeyTable() params.KeyTable {
	return params.NewKeyTable(
		moduleParamsKey, types.CdpModuleParams{},
	)
}
