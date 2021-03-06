package cdp

import (
	"fmt"
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QueryGetCdps   = "cdps"
	QueryGetParams = "params"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryGetCdps:
			return queryGetCdps(ctx, req, keeper)
		case QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown cdp query endpoint")
		}
	}
}

type QueryCdpsParams struct {
	CollateralName        string         // get CDPs collateral name
	NftID                 string         //  get CDPs collateral ID
	Owner                 sdk.AccAddress //   get CDPs belonging to this owner
	UnderCollateralizedAt sdk.Int        //    get CDPs that will be below the liquidation ratio when the collateral is at this price.
}

// queryGetCdps fetches CDPs, optionally filtering by any of the query params (in QueryCdpsParams).
// While CDPs do not have an ID, this method can be used to get one CDP by specifying the collateral and owner.
func queryGetCdps(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// Decode request
	var requestParams QueryCdpsParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	var cdps types.CDPs
	var cdp types.CDP
	var found bool

	if len(requestParams.Owner) != 0 {
		if len(requestParams.CollateralName) != 0 {
			if len(requestParams.NftID) != 0 { //ID specified, NFT
				cdp, found = keeper.GetCDP(ctx, requestParams.Owner, requestParams.CollateralName, requestParams.NftID)
			} else { // empty ID field, FT
				cdp, found = keeper.GetCDP(ctx, requestParams.Owner, requestParams.CollateralName, "")
			}
			if !found {
				cdp = types.CDP{Owner: requestParams.Owner, Collateral: types.Collateral{}, Liquidity: types.Liquidity{}}
			}
			cdps = types.CDPs{cdp}
		} else {
			// owner, but no collateral specified - get all CDPs for one address
			return nil, sdk.ErrInternal("getting all CDPs belonging to one owner not implemented")
		}
	} else {
		// owner not specified -- get all CDPs or all CDPs of one collateral type, optionally filtered by price
		var errSdk sdk.Error // := doesn't work here
		cdps, errSdk = keeper.GetCDPs(ctx, requestParams.CollateralName, requestParams.UnderCollateralizedAt)
		if errSdk != nil {
			return nil, errSdk
		}

	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, cdps)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// queryGetParams fetches the cdp module parameters
// TODO does this need to exist? Can you use cliCtx.QueryStore instead?
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
