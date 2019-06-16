package cdp

import (
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/pricefeed"
	sdk "github.com/cosmos/cosmos-sdk/types"
<<<<<<< Updated upstream
=======
	"github.com/commercionetwork/cosmos-hackathom-2019/blockchain/x/pricefeed"
>>>>>>> Stashed changes
)

type bankKeeper interface {
	GetCoins(sdk.Context, sdk.AccAddress) sdk.Coins
	HasCoins(sdk.Context, sdk.AccAddress, sdk.Coins) bool
	AddCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Error)
	SubtractCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Error)
}

type pricefeedKeeper interface {
<<<<<<< Updated upstream
	GetCurrentPrice(sdk.Context, string) pricefeed.CurrentPrice
=======
	GetCurrentPrice(context sdk.Context, assetCode string, assetName string) pricefeed.CurrentPrice
>>>>>>> Stashed changes
	// These are used for testing TODO replace mockApp with keeper in tests to remove these
	AddAsset(context sdk.Context, assetCode string, assetString string)
	SetPrice(context sdk.Context, oracle sdk.AccAddress, assetCode string, assetName string, price sdk.Dec, expiry sdk.Int) (pricefeed.PostedPrice, sdk.Error)
	SetCurrentPrices(sdk.Context) sdk.Error
}
