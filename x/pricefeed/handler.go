package pricefeed

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler handles all pricefeed type messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgPostPrice:
			{
				// TODO: After posting the price, send the user the specific amount from the pool if he was waiting for it
				// TODO: Update any CDP regarding that token to change the debt value accordingly
				return HandleMsgPostPrice(ctx, k, msg)
			}
		default:
			errMsg := fmt.Sprintf("unrecognized pricefeed message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// price feed questions:
// do proposers need to post the round in the message? If not, how do we determine the round?

// HandleMsgPostPrice handles prices posted by oracles
func HandleMsgPostPrice(ctx sdk.Context, k Keeper, msg MsgPostPrice) sdk.Result {

	// TODO cleanup message validation and errors
	err := k.ValidatePostPrice(ctx, msg)
	if err != nil {
		return err.Result()
	}

	_, err = k.SetPrice(ctx, msg.From, msg.AssetName, msg.AssetCode, msg.Price, msg.Expiry)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}

// EndBlocker updates the current pricefeed
func EndBlocker(ctx sdk.Context, k Keeper) sdk.Tags {
	// TODO val_state_change.go is relevant if we want to rotate the oracle set

	// Running in the end blocker ensures that prices will update at most once per block,
	// which seems preferable to having state storage values change in response to multiple transactions
	// which occur during a block
	//TODO use an iterator and update the prices for all assets in the store
	err := k.SetCurrentPrices(ctx)
	if err != nil {
		panic(err)
	}

	err = k.cdpKeeper.ModifyCDPType(ctx, "")
	if err != nil {
		panic(err)
	}

	return sdk.Tags{}
}
