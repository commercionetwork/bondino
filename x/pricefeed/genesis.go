package pricefeed

import sdk "github.com/cosmos/cosmos-sdk/types"

// GenesisState state at gensis
type GenesisState struct {
	Assets  []Asset
	Oracles []Oracle
}

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, genState GenesisState) {
	for _, asset := range genState.Assets {
		keeper.AddAsset(ctx, asset.AssetCode, asset.Description)
	}

	for _, oracle := range genState.Oracles {
		keeper.AddOracle(ctx, oracle.OracleAddress)
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		[]Asset{
			{Type: "ft", AssetName: "btc", Description: "a description"},
			{Type: "nft", AssetName: "xrp", Description: "the standard"},
		},
		[]Oracle{}}
}

// ValidateGenesis performs basic validation of genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	// TODO
	return nil
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	// TODO implement this
	return DefaultGenesisState()
}
