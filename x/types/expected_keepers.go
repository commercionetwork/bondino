package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

type CdpKeeper interface {
	ModifyCDP(ctx sdk.Context, owner sdk.AccAddress, collateral Collateral, liquidity Liquidity) sdk.Error
	PartialSeizeCDP(ctx sdk.Context, owner sdk.AccAddress, collateral Collateral, collateralToSeize sdk.Int, debtToSeize sdk.Int) sdk.Error
	ReduceGlobalDebt(ctx sdk.Context, amount sdk.Int) sdk.Error
	GetStableDenom() string
	GetGovDenom() string
	GetParams(ctx sdk.Context) CdpModuleParams
	GetCDPs(ctx sdk.Context, collateralDenom string, price sdk.Dec) (CDPs, sdk.Error)
	GetCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string, nftID string) (CDP, bool)
	GetGlobalDebt(ctx sdk.Context) sdk.Int
	GetCollateralState(ctx sdk.Context, collateralDenom string) (CollateralState, bool)
	GetLiquidatorAccountAddress() sdk.AccAddress
	AddCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) (sdk.Coins, sdk.Error)
	SubtractCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) (sdk.Coins, sdk.Error)
	GetCoins(ctx sdk.Context, address sdk.AccAddress) sdk.Coins
	HasCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) bool
}

type PricefeedKeeper interface {
	GetCurrentPrice(context sdk.Context, assetCode string, assetName string) CurrentPrice
	// These are used for testing TODO replace mockApp with keeper in tests to remove these
	AddAsset(context sdk.Context, assetCode string, assetString string)
	SetPrice(context sdk.Context, oracle sdk.AccAddress, assetCode string, assetName string, price sdk.Dec, expiry sdk.Int) (PostedPrice, sdk.Error)
	SetCurrentPrices(sdk.Context) sdk.Error
}

// PostedPrice struct represented a price for an asset posted by a specific oracle
type PostedPrice struct {
	AssetName     string  `json:"asset_name"`
	AssetCode     string  `json:"asset_code"`
	OracleAddress string  `json:"oracle_address"`
	Price         sdk.Dec `json:"price"`
	Expiry        sdk.Int `json:"expiry"`
}

// implement fmt.Stringer
func (pp PostedPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s
OracleAddress: %s
Price: %s
Expiry: %s`, pp.AssetCode, pp.OracleAddress, pp.Price, pp.Expiry))
}
