package cli

import (
	"fmt"
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/pool"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

func GetCmdGetFunds(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get-funds [address]",
		Short: "get the funds for a specified account",
		Long:  "Get the current funds value for the given account address.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			ownerAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(pool.QueryFundsParams{
				Owner: ownerAddress,
			})
			if err != nil {
				return err
			}

			res, err := cliCtx.QueryWithData(
				fmt.Sprintf("custom/%s/%s", queryRoute, pool.QueryReadFunds),
				bz,
			)
			if err != nil {
				return err
			}
			var funds sdk.Coin
			cdc.MustUnmarshalJSON(res, &funds)
			return cliCtx.PrintOutput(funds)
		},
	}
}

func GetCmdGetAllFunds(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "funds",
		Short: "get the total pool funds",
		Long:  "Get the total pool funds",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, pool.QueryTotalFunds), nil)
			if err != nil {
				return err
			}

			var funds sdk.Coins
			cdc.MustUnmarshalJSON(res, &funds)
			return cliCtx.PrintOutput(funds)
		},
	}
}
