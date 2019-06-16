package cli

import (
	"fmt"
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/types"

	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/cdp"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	"github.com/spf13/cobra"
)

// GetCmdModifyFtCdp cli command for creating and modifying FT cdps.
func GetCmdModifyFtCdp(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "modify-cdp-ft [ownerAddress] [collateralName] [collateralQuantity] [liquidityName]",
		Short: "create or modify a cdp",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			collateralQuantity, ok := sdk.NewIntFromString(args[2])
			if !ok {
				fmt.Printf("invalid collateral amount - %s \n", string(args[2]))
				return nil
			}

			// compose the collateral value
			collateral := types.Collateral{
				Token: cdp.BaseFT{
					TokenName: args[1],
				},
				Amount: collateralQuantity,
			}

			// compose the liquidity value
			liquidity := types.Liquidity{
				Coin: sdk.NewCoin(args[3], sdk.NewInt(0)),
			}

			msg := cdp.NewMsgCreateOrModifyCDP(cliCtx.GetFromAddress(), collateral, liquidity)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			cliCtx.PrintResponse = true
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdModifyNftCdp cli command for creating and modifying NFT cdps.
func GetCmdModifyNftCdp(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "modify-cdp-nft [ownerAddress] [collateralName] [collateralId] [collateralQuantity] [liquidityName]",
		Short: "create or modify a cdp",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			collateralAmount, ok := sdk.NewIntFromString(args[3])
			if !ok {
				fmt.Printf("invalid collateral amount - %s \n", string(args[3]))
				return nil
			}

			// build the collateral
			collateral := types.Collateral{
				Token: cdp.BaseNFT{
					Name: args[1],
					ID:   args[2],
				},
				Amount: collateralAmount,
			}

			// build the liquidity
			liquidity := types.Liquidity{
				Coin: sdk.NewCoin(args[4], sdk.NewInt(0)),
			}

			// create the message
			msg := cdp.NewMsgCreateOrModifyCDP(cliCtx.GetFromAddress(), collateral, liquidity)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			cliCtx.PrintResponse = true
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
