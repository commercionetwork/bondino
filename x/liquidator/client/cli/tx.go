package cli

import (
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/cdp"
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/types"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/spf13/cobra"

	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/liquidator"
)

func GetCmd_SeizeAndStartCollateralAuction(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seize [cdp-owner] [collateralName] [collateralID] [collateralQuantity]",
		Short: "seize funds from a CDP and send to auction",
		Long: `Seize a fixed amount of collateral and liquidity from a CDP then start an auction with the collateral.
The amount of collateral seized is given by the 'AuctionSize' module parameter or, if there isn't enough collateral in the CDP, all the CDP's collateral is seized.
Debt is seized in proportion to the collateral seized so that the CDP stays at the same collateral to debt ratio.
A 'forward-reverse' auction is started selling the seized collateral for some stable coin, with a maximum bid of stable coin set to equal the debt seized.
As this is a forward-reverse auction type, if the max stable coin is bid then bidding continues by bidding down the amount of collateral taken by the bidder. At the end, extra collateral is returned to the original CDP owner.`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Setup
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// Validate inputs
			sender := cliCtx.GetFromAddress()
			cdpOwner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			name := args[1]
			// TODO validate denom?
			id := args[2]
			amount, _ := sdk.NewIntFromString(args[4])

			//composing collateral
			var collateral types.Collateral
			if id != "" {
				collateral = types.Collateral{
					Token: cdp.BaseNFT{
						Name: name,
						ID:   id,
					},
					Amount: amount,
				}
			} else {
				collateral = types.Collateral{
					Token: cdp.BaseFT{
						TokenName: name,
					},
					Amount: amount,
				}
			}

			// Prepare and send message
			msgs := []sdk.Msg{liquidator.MsgSeizeAndStartCollateralAuction{
				Sender:     sender,
				CdpOwner:   cdpOwner,
				Collateral: collateral,
			}}
			// TODO print out results like auction ID?
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, msgs)
		},
	}
	return cmd
}

func GetCmd_StartDebtAuction(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "start a debt auction, minting gov coin to cover debt",
		Long:  "Start a reverse auction, selling off minted gov coin to raise a fixed amount of stable coin.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Setup
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			sender := cliCtx.GetFromAddress()

			// Prepare and send message
			msgs := []sdk.Msg{liquidator.MsgStartDebtAuction{
				Sender: sender,
			}}
			// TODO print out results like auction ID?
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, msgs)
		},
	}
	return cmd
}
