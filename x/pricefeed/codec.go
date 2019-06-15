package pricefeed

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgPostPrice{}, "pricefeed/MsgPostPrice", nil)
}

// generic sealed codec to be used throughout module
var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}
