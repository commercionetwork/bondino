package cdp

import (
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgCreateOrModifyCDP creates, adds/removes collateral/stable coin from a cdp
// TODO Make this more user friendly - maybe split into four functions.
type MsgCreateOrModifyCDP struct {
	Sender     sdk.AccAddress
	Collateral types.Collateral
	Liquidity  types.Liquidity
}

// NewMsgPlaceBid returns a new MsgPlaceBid.
func NewMsgCreateOrModifyCDP(sender sdk.AccAddress, collateral types.Collateral, liquidity types.Liquidity) MsgCreateOrModifyCDP {
	return MsgCreateOrModifyCDP{
		Sender:     sender,
		Collateral: collateral,
		Liquidity:  liquidity,
	}
}

// Route return the message type used for routing the message.
func (msg MsgCreateOrModifyCDP) Route() string { return "cdp" }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgCreateOrModifyCDP) Type() string { return "create_modify_cdp" } // TODO snake case?

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgCreateOrModifyCDP) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInternal("invalid (empty) sender address")
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgCreateOrModifyCDP) GetSignBytes() []byte {
	bz := moduleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgCreateOrModifyCDP) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgTransferCDP changes the ownership of a cdp
type MsgTransferCDP struct {
	// TODO
}
