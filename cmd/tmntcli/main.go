package main

import (
	"net/http"
	"os"
	"path"

	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/app"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	at "github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	crisisclient "github.com/cosmos/cosmos-sdk/x/crisis/client"
	distcmd "github.com/cosmos/cosmos-sdk/x/distribution"
	distClient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrcli "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	dist "github.com/cosmos/cosmos-sdk/x/distribution/client/rest"
	gv "github.com/cosmos/cosmos-sdk/x/gov"
	govClient "github.com/cosmos/cosmos-sdk/x/gov/client"
	gov "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintclient "github.com/cosmos/cosmos-sdk/x/mint/client"
	mintrest "github.com/cosmos/cosmos-sdk/x/mint/client/rest"
	paramcli "github.com/cosmos/cosmos-sdk/x/params/client/cli"
	paramsrest "github.com/cosmos/cosmos-sdk/x/params/client/rest"
	sl "github.com/cosmos/cosmos-sdk/x/slashing"
	slashingclient "github.com/cosmos/cosmos-sdk/x/slashing/client"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	st "github.com/cosmos/cosmos-sdk/x/staking"
	stakingclient "github.com/cosmos/cosmos-sdk/x/staking/client"
	staking "github.com/cosmos/cosmos-sdk/x/staking/client/rest"

	auctionclient "github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/auction/client"
	auctionrest "github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/auction/client/rest"
	cdpclient "github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/cdp/client"
	liquidatorclient "github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/liquidator/client"
	poolclient "github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/pool/client"
	priceclient "github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/pricefeed/client"
	pricerest "github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/pricefeed/client/rest"

	_ "github.com/cosmos/gaia/cmd/gaiacli/statik"
)

func main() {
	// Configure cobra to sort commands
	cobra.EnableCommandSorting = false

	// Instantiate the codec for the command line application
	cdc := app.MakeCodec()

	// configure the address prefixes and seal config
	app.SetAddressPrefixes()

	mc := []sdk.ModuleClient{
		govClient.NewModuleClient(gv.StoreKey, cdc, paramcli.GetCmdSubmitProposal(cdc), distrcli.GetCmdSubmitProposal(cdc)),
		distClient.NewModuleClient(distcmd.StoreKey, cdc),
		stakingclient.NewModuleClient(st.StoreKey, cdc),
		mintclient.NewModuleClient(mint.StoreKey, cdc),
		slashingclient.NewModuleClient(sl.StoreKey, cdc),
		crisisclient.NewModuleClient(sl.StoreKey, cdc),
		priceclient.NewModuleClient("pricefeed", cdc),
		cdpclient.NewModuleClient("cdp", cdc),
		auctionclient.NewModuleClient("auction", cdc),
		liquidatorclient.NewModuleClient("liquidator", cdc),
		poolclient.NewModuleClient("pool", cdc),
	}

	rootCmd := &cobra.Command{
		Use:   "tmntcli",
		Short: "TMNT command line interface",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		client.ConfigCmd(app.DefaultCLIHome),
		queryCmd(cdc, mc),
		txCmd(cdc, mc),
		client.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		client.LineBreak,
		keys.Commands(),
		client.LineBreak,
		version.Cmd,
		client.NewCompletionCmd(rootCmd, true),
	)

	executor := cli.PrepareMainCmd(rootCmd, "tmnt", app.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func queryCmd(cdc *amino.Codec, mc []sdk.ModuleClient) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
		authcmd.GetAccountCmd(at.StoreKey, cdc),
	)

	for _, m := range mc {
		mQueryCmd := m.GetQueryCmd()
		if mQueryCmd != nil {
			queryCmd.AddCommand(mQueryCmd)
		}
	}

	return queryCmd
}

func txCmd(cdc *amino.Codec, mc []sdk.ModuleClient) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		authcmd.GetSignCommand(cdc),
		authcmd.GetMultiSignCommand(cdc),
		tx.GetBroadcastCommand(cdc),
		tx.GetEncodeCommand(cdc),
		client.LineBreak,
	)

	for _, m := range mc {
		txCmd.AddCommand(m.GetTxCmd())
	}

	return txCmd
}

func registerRoutes(rs *lcd.RestServer) {
	registerSwaggerUI(rs)
	rpc.RegisterRoutes(rs.CliCtx, rs.Mux)
	tx.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	auth.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, at.StoreKey)
	bank.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	dist.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, distcmd.StoreKey)
	staking.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	slashing.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	gov.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, paramsrest.ProposalRESTHandler(rs.CliCtx, rs.Cdc), dist.ProposalRESTHandler(rs.CliCtx, rs.Cdc))
	mintrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	pricerest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, "pricefeed")
	auctionrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	//liquidatorrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
}

func registerSwaggerUI(rs *lcd.RestServer) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}
	staticServer := http.FileServer(statikFS)
	rs.Mux.PathPrefix("/swagger-ui/").Handler(http.StripPrefix("/swagger-ui/", staticServer))
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
