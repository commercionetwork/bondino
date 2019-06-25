package cdp

import (
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

// generic sealed codec to be used throughout module
var moduleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	moduleCdc = cdc.Seal()
}

// RegisterCodec registers concrete types on the codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateOrModifyCDP{}, "cdp/MsgCreateOrModifyCDP", nil)
	cdc.RegisterConcrete(MsgTransferCDP{}, "cdp/MsgTransferCDP", nil)
	cdc.RegisterInterface((*types.Token)(nil), nil)
	cdc.RegisterConcrete(BaseFT{}, "cdp/BaseFT", nil)
	cdc.RegisterConcrete(BaseNFT{}, "cdp/BaseNFT", nil)
}
