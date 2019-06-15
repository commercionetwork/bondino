package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/kava-labs/kava-devnet/blockchain/x/auction"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	auctionTxCmd := &cobra.Command{
		Use:   "auction",
		Short: "auction transactions subcommands",
	}

	auctionTxCmd.AddCommand(client.PostCommands(
		getCmdPlaceBid(cdc),
	)...)

	return auctionTxCmd
}

// getCmdPlaceBid cli command for creating and modifying cdps.
func getCmdPlaceBid(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "placebid [AuctionID] [Bidder] [Bid] [Lot]",
		Short: "place a bid on an auction",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			id, err := auction.NewIDFromString(args[0])
			if err != nil {
				fmt.Printf("invalid auction id - %s \n", string(args[0]))
				return err
			}

			bid, err := sdk.ParseCoin(args[2])
			if err != nil {
				fmt.Printf("invalid bid amount - %s \n", string(args[2]))
				return err
			}

			lot, err := sdk.ParseCoin(args[3])
			if err != nil {
				fmt.Printf("invalid lot - %s \n", string(args[3]))
				return err
			}
			msg := auction.NewMsgPlaceBid(id, cliCtx.GetFromAddress(), bid, lot)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			cliCtx.PrintResponse = true
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
