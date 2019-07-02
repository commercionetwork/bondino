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
		keeper.AddAsset(ctx, asset.AssetName, asset.AssetCode, asset.Description)
	}

	for _, oracle := range genState.Oracles {
		keeper.AddOracle(ctx, oracle.OracleAddress)
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		[]Asset{
			{Type: _FT, AssetName: "btc", AssetCode: _CODEFT, Description: "a description"},
			{Type: _NFT, AssetName: "xrp", AssetCode: "01", Description: "the standard"},
			{Type: _FT, AssetName: "eth", AssetCode: _CODEFT, Description: "ethereum coin"},
			{Type: _FT, AssetName: "atm", AssetCode: _CODEFT, Description: "cosmos coin"},
			{Type: _FT, AssetName: "acme", AssetCode: _CODEFT, Description: "test coin"},
		},
		[]Oracle{
			{OracleAddress: "tmnt1gnw3y7cpf6yd30dkgs5stymu2qmhm73r3cjufr"},
			{OracleAddress: "tmnt1pqeqzuju3czmk3qjm7x58w37qexfy26g4gwsdw"},
		},
	}
}

// ValidateGenesis performs basic validation of genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	for _, asset := range data.Assets {
		if asset.Type != _FT && asset.Type != _NFT {
			return sdk.ErrInternal("Invalid asset type, must be FT or NFT")
		}
		if asset.Type == _FT && asset.AssetCode != _CODEFT {
			return sdk.ErrInternal("Invalid FT code, must be 0")
		}
		if len(asset.AssetName) == 0 {
			return sdk.ErrInternal("Asset name cant be empty")
		}
		if len(asset.Description) == 0 {
			return sdk.ErrInternal("Asset description cant be empty")
		}
	}

	for _, oracle := range data.Oracles {
		if len(oracle.OracleAddress) == 0 {
			return sdk.ErrInternal("Oracle address cant be empty")
		}
	}

	return nil
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	// TODO implement this
	return DefaultGenesisState()
}
