package pricefeed

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"strings"
)

// price Takes an [assetcode] and returns CurrentPrice for that asset
// pricefeed Takes an [assetcode] and returns the raw []PostedPrice for that asset
// assets Returns []Assets in the pricefeed system

const (
	// QueryCurrentPrice command for current price queries
	QueryCurrentPrice = "price"

	// QueryRawPrices command for raw price queries
	QueryRawPrices = "rawprices"

	// QueryAssets command for assets query
	QueryAssets = "assets"

	// QueryPendingPrices command for pending prices
	QueryPendingPrices = "pending-prices"
)

// implement fmt.Stringer
func (a Asset) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s, AssetName: %s, Description: %s`, a.AssetCode, a.AssetName, a.Description))
}

// implement fmt.Stringer
func (a PendingPriceAsset) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetName: %s, AssetCode: %s`, a.AssetName, a.AssetCode))
}

// QueryRawPricesResp response to a rawprice query
type QueryRawPricesResp []string

// implement fmt.Stringer
func (n QueryRawPricesResp) String() string {
	return strings.Join(n[:], "\n")
}

// QueryAssetsResp response to a assets query
type QueryAssetsResp []string

// implement fmt.Stringer
func (n QueryAssetsResp) String() string {
	return strings.Join(n[:], "\n")
}

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryCurrentPrice:
			return queryCurrentPrice(ctx, path[1:], req, keeper)
		//case QueryRawPrices:
		//	return queryRawPrices(ctx, path[1:], req, keeper)
		case QueryAssets:
			return queryAssets(ctx, req, keeper)
		case QueryPendingPrices:
			return queryPendingPrices(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown pricefeed query endpoint")
		}
	}

}

func queryCurrentPrice(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	assetName := path[0]
	assetCode := path[1]
	_, found := keeper.GetAsset(ctx, assetCode, assetName)
	if !found {
		return []byte{}, sdk.ErrUnknownRequest("asset not found")
	}
	currentPrice := keeper.GetCurrentPrice(ctx, assetCode, assetName)

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, currentPrice)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryRawPrices(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var priceList QueryRawPricesResp
	assetName := path[0]
	assetCode := path[1]
	_, found := keeper.GetAsset(ctx, assetCode, assetName)
	if !found {
		return []byte{}, sdk.ErrUnknownRequest("asset not found")
	}
	rawPrices := keeper.GetRawPrices(ctx, assetCode, assetName)
	for _, price := range rawPrices {
		priceList = append(priceList, price.String())
	}
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, priceList)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryAssets(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var assetList QueryAssetsResp
	assets := keeper.GetAssets(ctx)
	for _, asset := range assets {
		assetList = append(assetList, asset.String())
	}
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, assetList)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryPendingPrices(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var assetList QueryAssetsResp
	assets := keeper.GetPendingPriceAssets(ctx)
	for _, asset := range assets {
		assetList = append(assetList, asset.String())
	}
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, assetList)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}
