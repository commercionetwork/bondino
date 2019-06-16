package pricefeed

import (
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/types"
	"sort"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO refactor constants to app.go
const (
	// ModuleKey is the name of the module
	ModuleName = "pricefeed"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// Parameter store default namestore
	DefaultParamspace = ModuleName

	// Store prefix for the raw pricefeed of an asset
	RawPriceFeedPrefix = StoreKey + ":raw:"

	// Store prefix for the required prices
	RequiredPricesPrefix = StoreKey + ":requiredPrices:"

	// Store prefix for the current price of an asset
	CurrentPricePrefix = StoreKey + ":currentprice:"

	// Store Prefix for the assets in the pricefeed system
	AssetPrefix = StoreKey + ":assets"

	// OraclePrefix store prefix for the oracle accounts
	OraclePrefix = StoreKey + ":oracles"

	// EstimableAssetPrefix store prefix for the estimable assets
	EstimableAssetPrefix = StoreKey + ":estimableassets"
)

// Keeper struct for pricefeed module
type Keeper struct {
	priceStoreKey          sdk.StoreKey
	pricesRequestsStoreKey sdk.StoreKey
	cdc                    *codec.Codec
	codespace              sdk.CodespaceType
	cdpKeeper              types.CdpKeeper
}

// NewKeeper returns a new keeper for the pricefeed modle
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, codespace sdk.CodespaceType, cdpKeeper types.CdpKeeper) Keeper {
	return Keeper{
		priceStoreKey: storeKey,
		cdc:           cdc,
		codespace:     codespace,
		cdpKeeper:     cdpKeeper,
	}
}

func (k Keeper) combineAssetInfo(assetCode string, assetName string) string {
	return assetCode + "++" + assetName
}

func (k Keeper) getAssetCodeAndName(code string) (string, string) {
	result := strings.Split(code, "++")
	return result[0], result[1]
}

// AddOracle adds an Oracle to the store
func (k Keeper) AddOracle(ctx sdk.Context, address string) {

	oracles := k.GetOracles(ctx)
	oracles = append(oracles, Oracle{OracleAddress: address})
	store := ctx.KVStore(k.priceStoreKey)
	store.Set(
		[]byte(OraclePrefix), k.cdc.MustMarshalBinaryBare(oracles),
	)
}

// AddAsset adds an asset to the store
func (k Keeper) AddAsset(ctx sdk.Context, assetCode string, desc string) {
	assets := k.GetAssets(ctx)
	assets = append(assets, Asset{AssetCode: assetCode, Description: desc})
	store := ctx.KVStore(k.priceStoreKey)
	store.Set([]byte(AssetPrefix), k.cdc.MustMarshalBinaryBare(assets))
}

// SetPrice updates the posted price for a specific oracle
func (k Keeper) SetPrice(ctx sdk.Context, oracle sdk.AccAddress, assetName string, assetCode string, price sdk.Int, expiry sdk.Int) (types.PostedPrice, sdk.Error) {
	// If the expiry is less than or equal to the current blockheight, we consider the price valid
	if expiry.GTE(sdk.NewInt(ctx.BlockHeight())) {
		store := ctx.KVStore(k.priceStoreKey)
		prices := k.GetRawPrices(ctx, assetCode, assetName)
		var index int
		found := false
		for i := range prices {
			if prices[i].OracleAddress == oracle.String() {
				index = i
				found = true
				break
			}
		}
		// set the price for that particular oracle
		if found {
			prices[index] = types.PostedPrice{AssetCode: assetCode, OracleAddress: oracle.String(), Price: price, Expiry: expiry}
		} else {
			prices = append(prices, types.PostedPrice{
				AssetName:     assetName,
				AssetCode:     assetCode,
				OracleAddress: oracle.String(),
				Price:         price,
				Expiry:        expiry,
			})
			index = len(prices) - 1
		}

		store.Set([]byte(RawPriceFeedPrefix+k.combineAssetInfo(assetCode, assetName)), k.cdc.MustMarshalBinaryBare(prices))
		err := k.cdpKeeper.ModifyCDPType(ctx, assetName, assetCode)
		if err != nil {
			return types.PostedPrice{}, err
		}

		return prices[index], nil
	}

	return types.PostedPrice{}, ErrExpired(k.codespace)

}

// SetCurrentPrices updates the price of an asset to the median of all valid oracle inputs
func (k Keeper) SetCurrentPrices(ctx sdk.Context) sdk.Error {
	assets := k.GetAssets(ctx)
	for _, v := range assets {
		assetCode := v.AssetCode
		assetName := v.AssetName
		prices := k.GetRawPrices(ctx, assetCode, assetName)
		var notExpiredPrices []types.CurrentPrice
		// filter out expired prices
		for _, v := range prices {
			if v.Expiry.GTE(sdk.NewInt(ctx.BlockHeight())) {
				notExpiredPrices = append(notExpiredPrices, types.CurrentPrice{
					AssetCode: v.AssetCode,
					AssetName: v.AssetName,
					Price:     v.Price,
					Expiry:    v.Expiry,
				})
			}
		}
		l := len(notExpiredPrices)
		var medianPrice sdk.Int
		var expiry sdk.Int
		// TODO make threshold for acceptance (ie. require 51% of oracles to have posted valid prices
		if l == 0 {
			// Error if there are no valid prices in the raw pricefeed
			// return ErrNoValidPrice(k.codespace)
			medianPrice = sdk.NewInt(0)
			expiry = sdk.NewInt(0)
		} else if l == 1 {
			// Return immediately if there's only one price
			medianPrice = notExpiredPrices[0].Price
			expiry = notExpiredPrices[0].Expiry
		} else {
			// sort the prices
			sort.Slice(notExpiredPrices, func(i, j int) bool {
				return notExpiredPrices[i].Price.LT(notExpiredPrices[j].Price)
			})
			// If there's an even number of prices
			if l%2 == 0 {
				// TODO make sure this is safe.
				// Since it's a price and not a blance, division with precision loss is OK.
				price1 := notExpiredPrices[l/2-1].Price
				price2 := notExpiredPrices[l/2].Price
				sum := price1.Add(price2)
				divsor := sdk.NewInt(2)
				medianPrice = sum.Quo(divsor)
				// TODO Check if safe, makes sense
				// Takes the average of the two expiries rounded down to the nearest Int.
				expiry = notExpiredPrices[l/2-1].Expiry.Add(notExpiredPrices[l/2].Expiry).Quo(sdk.NewInt(2))
			} else {
				// integer division, so we'll get an integer back, rounded down
				medianPrice = notExpiredPrices[l/2].Price
				expiry = notExpiredPrices[l/2].Expiry
			}
		}

		store := ctx.KVStore(k.priceStoreKey)
		currentPrice := types.CurrentPrice{
			AssetCode: assetCode,
			Price:     medianPrice,
			Expiry:    expiry,
		}
		store.Set(
			[]byte(CurrentPricePrefix+k.combineAssetInfo(assetCode, assetName)), k.cdc.MustMarshalBinaryBare(currentPrice),
		)
	}

	return nil
}

// GetPendingPriceAssets returns the list of all those assets which prices are still pending
func (k Keeper) GetPendingPriceAssets(ctx sdk.Context) []PendingPriceAsset {
	store := ctx.KVStore(k.priceStoreKey)
	bz := store.Get([]byte(RequiredPricesPrefix))

	var estimableAssets []PendingPriceAsset
	k.cdc.MustUnmarshalBinaryBare(bz, &estimableAssets)

	return estimableAssets
}

// GetOracles returns the oracles in the pricefeed store
func (k Keeper) GetOracles(ctx sdk.Context) []Oracle {
	store := ctx.KVStore(k.priceStoreKey)
	bz := store.Get([]byte(OraclePrefix))
	var oracles []Oracle
	k.cdc.MustUnmarshalBinaryBare(bz, &oracles)
	return oracles
}

// GetAssets returns the assets in the pricefeed store
func (k Keeper) GetAssets(ctx sdk.Context) []Asset {
	store := ctx.KVStore(k.priceStoreKey)
	bz := store.Get([]byte(AssetPrefix))
	var assets []Asset
	k.cdc.MustUnmarshalBinaryBare(bz, &assets)
	return assets
}

// GetAsset returns the asset if it is in the pricefeed system
func (k Keeper) GetAsset(ctx sdk.Context, assetCode string, assetName string) (Asset, bool) {
	assets := k.GetAssets(ctx)

	for i := range assets {
		if assets[i].AssetCode == assetCode && assets[i].AssetName == assetName {
			return assets[i], true
		}
	}
	return Asset{}, false

}

// GetOracle returns the oracle address as a string if it is in the pricefeed store
func (k Keeper) GetOracle(ctx sdk.Context, oracle string) (Oracle, bool) {
	oracles := k.GetOracles(ctx)

	for i := range oracles {
		if oracles[i].OracleAddress == oracle {
			return oracles[i], true
		}
	}
	return Oracle{}, false

}

// Deve essere estratto l'oracolo preposto a valutare quel tipo di NFT
// poi deve essere registrato un msg con l'indicazione del tipo di NFT,
// il suo ID, l'oracolo preposto e il fatto che sia o meno stato valutato
// Questo elemento del keystore serve per restitutire un messaggio agli oracoli perché valutino l'NFT
func (k Keeper) AskForPrice(ctx sdk.Context, assetCode string, assetName string) {

	// recover the existing prices, if any
	store := ctx.KVStore(k.pricesRequestsStoreKey)
	existing := store.Get([]byte(RequiredPricesPrefix))

	var requiredPrices []PendingPriceAsset
	if existing != nil {
		k.cdc.MustUnmarshalBinaryBare(existing, &requiredPrices)
	}

	// update the required prices
	requiredPrices = append(requiredPrices, PendingPriceAsset{AssetName: assetName, AssetCode: assetCode})

	// TODO: this should probably take into consideration the fact that the price may be have been asked before.
	// In this case it should be better to save the block height at which it has been retrieved the last time, and later
	// decide whenever it is better to require it again or not
	store.Set([]byte(RequiredPricesPrefix), k.cdc.MustMarshalBinaryBare(requiredPrices))
}

// GetCurrentPrice fetches the current median price of all oracles for a specific asset
func (k Keeper) GetCurrentPrice(ctx sdk.Context, assetCode string, assetName string) types.CurrentPrice {

	store := ctx.KVStore(k.priceStoreKey)
	var storedPriceKey string

	bz := store.Get([]byte(storedPriceKey))

	var price types.CurrentPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &price)

	return price
}

// GetRawPrices fetches the set of all prices posted by oracles for an asset
func (k Keeper) GetRawPrices(ctx sdk.Context, assetCode string, assetName string) []types.PostedPrice {
	store := ctx.KVStore(k.priceStoreKey)
	bz := store.Get([]byte(RawPriceFeedPrefix + assetCode))
	var prices []types.PostedPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &prices)
	return prices
}

// ValidatePostPrice makes sure the person posting the price is an oracle
func (k Keeper) ValidatePostPrice(ctx sdk.Context, msg MsgPostPrice) sdk.Error {

	_, assetFound := k.GetAsset(ctx, msg.AssetCode, msg.AssetName)
	if !assetFound {
		return ErrInvalidAsset(k.codespace)
	}
	_, oracleFound := k.GetOracle(ctx, msg.From.String())
	if !oracleFound {
		return ErrInvalidOracle(k.codespace)
	}

	return nil
}
