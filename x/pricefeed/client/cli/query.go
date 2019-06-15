package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/pricefeed"
	"github.com/spf13/cobra"
)


// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	pricefeedQueryCmd := &cobra.Command{
		Use:   "pricefeed",
		Short: "Querying commands for the pricefeed module",
	}

	pricefeedQueryCmd.AddCommand(client.GetCommands(
		getCmdCurrentPrice(queryRoute, cdc),
		getCmdRawPrices(queryRoute, cdc),
		getCmdAssets(queryRoute, cdc),
	)...)

	return pricefeedQueryCmd
}

// getCmdCurrentPrice queries the current price of an asset
func getCmdCurrentPrice(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "price [assetCode]",
		Short: "get the current price of an asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			assetCode := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/price/%s", queryRoute, assetCode), nil)
			if err != nil {
				fmt.Printf("error when querying current price - %s", err)
				fmt.Printf("could not get current price for - %s \n", string(assetCode))
				return nil
			}
			var out pricefeed.CurrentPrice
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}



// getCmdRawPrices queries the current price of an asset
func getCmdRawPrices(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "rawprices [assetCode]",
		Short: "get the raw oracle prices for an asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			assetCode := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/rawprices/%s", queryRoute, assetCode), nil)
			if err != nil {
				fmt.Printf("could not get raw prices for - %s \n", string(assetCode))
				return nil
			}
			var out pricefeed.QueryRawPricesResp
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// getCmdAssets queries list of assets in the pricefeed
func getCmdAssets(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "assets",
		Short: "get the assets in the pricefeed",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/assets", queryRoute), nil)
			if err != nil {
				fmt.Printf("could not get assets")
				return nil
			}
			var out pricefeed.QueryAssetsResp
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
