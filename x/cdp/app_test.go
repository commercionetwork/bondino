package cdp

import (
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/types"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestApp_CreateModifyDeleteCDP(t *testing.T) {
	// Setup
	mapp, keeper := setUpMockAppWithoutGenesis()
	genAccs, addrs, _, privKeys := mock.CreateGenAccounts(1, cs(c("xrp", 100)))
	testAddr := addrs[0]
	testPrivKey := privKeys[0]
	mock.SetGenesis(mapp, genAccs)
	// setup pricefeed, TODO can this be shortened a bit?
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	keeper.pricefeed.AddAsset(ctx, "xrp", "xrp test")
	_, err := keeper.pricefeed.SetPrice(ctx, sdk.AccAddress{}, "", "xrp", i(5), i(6))
	if err != nil {
		panic(err)
	}
	err = keeper.pricefeed.SetCurrentPrices(ctx)
	if err != nil {
		panic(err)
	}
	mapp.EndBlock(abci.RequestEndBlock{})
	mapp.Commit()

	//creating collateral
	var collateralTest = types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
		Amount: i(10), InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(20)}}}
	var collateralTest2 = types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
		Amount: i(4000), InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "pdw", Amount: i(2000)}}}
	var collateralTest3 = types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
		Amount: i(-50), InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "btc", Amount: i(-10)}}}

	// Create CDP
	msgs := []sdk.Msg{NewMsgCreateOrModifyCDP(testAddr, collateralTest.Collateral, collateralTest.Liquidity)}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, abci.Header{Height: mapp.LastBlockHeight() + 1}, msgs, []uint64{0}, []uint64{0}, true, true, testPrivKey)

	mock.CheckBalance(t, mapp, testAddr, cs(c("xrp", 5), c("xrp", 90)))

	// Modify CDP
	msgs = []sdk.Msg{NewMsgCreateOrModifyCDP(testAddr, collateralTest2.Collateral, collateralTest2.Liquidity)}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, abci.Header{Height: mapp.LastBlockHeight() + 1}, msgs, []uint64{0}, []uint64{1}, true, true, testPrivKey)

	mock.CheckBalance(t, mapp, testAddr, cs(c("pdw", 10), c("pdw", 50)))

	// Delete CDP
	msgs = []sdk.Msg{NewMsgCreateOrModifyCDP(testAddr, collateralTest3.Collateral, collateralTest3.Liquidity)}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, abci.Header{Height: mapp.LastBlockHeight() + 1}, msgs, []uint64{0}, []uint64{2}, true, true, testPrivKey)

	mock.CheckBalance(t, mapp, testAddr, cs(c("xrp", 100)))
}
