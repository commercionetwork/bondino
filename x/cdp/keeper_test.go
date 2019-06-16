package cdp

import (
	"fmt"
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/types"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// How could one reduce the number of params in the test cases. Create a table driven test for each of the 4 add/withdraw collateral/debt?

var _, addrs = mock.GeneratePrivKeyAddressPairs(1)
var ownerAddr = addrs[0]
var collateralTest = types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
	Amount: i(100), InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(1)}}}
var collateralTest2 = types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
	Amount: i(100), InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "pdw", Amount: i(1)}}}
var collateralTest3 = types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
	Amount: i(100), InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "btc", Amount: i(1)}}}

func TestKeeper_ModifyCDP(t *testing.T) {

	type state struct { // TODO this allows invalid state to be set up, should it?
		CDP             types.CDP
		OwnerCoins      sdk.Coins
		GlobalDebt      sdk.Int
		CollateralState types.CollateralState
	}
	type args struct {
		owner              sdk.AccAddress
		collateralName     string
		changeInCollateral sdk.Int
		changeInDebt       sdk.Int
	}

	tests := []struct {
		name       string
		priorState state
		price      string
		// also missing CDPModuleParams
		args          args
		expectPass    bool
		expectedState state
	}{
		{
			"addCollateralAndDecreaseDebt",
			state{types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
				Amount: i(100), InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(1)}}},
				cs(c("xrp", 10),
					c("gov", 2)), i(2), types.CollateralState{Denom: "xrp", TotalDebt: i(2)}},
			"10.345", args{ownerAddr, "xrp", i(10), i(-1)},
			true, state{types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT}, Amount: i(100),
				InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(2)}}},
				cs( /*  0xrp  */ c("gov", 2)), i(1), types.CollateralState{Denom: "xrp", TotalDebt: i(2)}},
		},
		{
			"removeTooMuchCollateral",
			state{types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
				Amount: i(1000), InitialPrice: i(200)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(1)}}},
				cs(c("xrp", 10),
					c("gov", 10)), i(200), types.CollateralState{Denom: "xrp", TotalDebt: i(2)}},
			"10.345", args{ownerAddr, "xrp", i(-601), i(0)},
			false, state{types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT}, Amount: i(1000),
				InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(2)}}},
				cs( /*  0xrp  */ c("gov", 10)), i(200), types.CollateralState{Denom: "xrp", TotalDebt: i(200)}},
		},
		{
			"withdrawTooMuchStableCoin",
			state{types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
				Amount: i(1000), InitialPrice: i(200)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(1)}}},
				cs(c("xrp", 10),
					c("gov", 10)), i(200), types.CollateralState{Denom: "xrp", TotalDebt: i(200)}},
			"1.00", args{ownerAddr, "xrp", i(0), i(301)},
			false, state{types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT}, Amount: i(1000),
				InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(200)}}},
				cs( /*  0xrp  */ c("gov", 10)), i(200), types.CollateralState{Denom: "xrp", TotalDebt: i(200)}},
		},
		{
			"createCDPAndWithdrawStable",
			state{types.CDP{}, cs(c("xrp", 10),
				c("gov", 10)), i(0), types.CollateralState{Denom: "xrp", TotalDebt: i(2)}},
			"1.00", args{ownerAddr, "xrp", i(5), i(2)},
			false, state{types.CDP{}, cs(c("xrp", 10),
				c("gov", 10)), i(0), types.CollateralState{Denom: "xrp", TotalDebt: i(2)}},
		},
		{
			"emptyCDP",
			state{types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
				Amount: i(1000), InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(200)}}},
				cs(c("xrp", 10), c("gov", 201)),
				i(200), types.CollateralState{Denom: "xrp", TotalDebt: i(200)}},
			"1.00", args{ownerAddr, "xrp", i(0), i(301)},
			true, state{types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
				Amount: i(1000), InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(200)}}},
				cs(c("xrp", 10), c("gov", 201)),
				i(200), types.CollateralState{Denom: "xrp", TotalDebt: i(200)}},
		},
		{
			"invalidCollateralType",
			state{types.CDP{}, cs(c("shitcoin", 500000)), i(0),
				types.CollateralState{}}, "0.000001",
			args{ownerAddr, "shitcoin", i(500000), i(1)},
			false, state{types.CDP{}, cs(c("shitcoin", 500000)), i(0),
				types.CollateralState{}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup keeper
			mapp, keeper := setUpMockAppWithoutGenesis()
			// initialize cdp owner account with coins
			genAcc := auth.BaseAccount{
				Address: ownerAddr,
				Coins:   tc.priorState.OwnerCoins,
			}
			mock.SetGenesis(mapp, []auth.Account{&genAcc})
			// create a new context
			header := abci.Header{Height: mapp.LastBlockHeight() + 1}
			mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
			ctx := mapp.BaseApp.NewContext(false, header)
			// setup store state
			keeper.pricefeed.AddAsset(ctx, "xrp", "xrp test")
			keeper.pricefeed.SetPrice(ctx, sdk.AccAddress{}, "NFT1", "nft", i(100), i(100))
			keeper.pricefeed.SetCurrentPrices(ctx)
			if tc.priorState.CDP.Collateral.Token.GetName() != "" { // check if the prior CDP should be created or not (see if an empty one was specified)
				keeper.setCDP(ctx, tc.priorState.CDP)
			}
			keeper.setGlobalDebt(ctx, tc.priorState.GlobalDebt)
			if tc.priorState.CollateralState.Denom != "" {
				keeper.setCollateralState(ctx, tc.priorState.CollateralState)
			}

			// call func under test
			err := keeper.ModifyCDP(ctx, tc.args.owner, collateralTest.Collateral, collateralTest.Liquidity)
			mapp.EndBlock(abci.RequestEndBlock{})
			mapp.Commit()

			// check for err
			if tc.expectPass {
				require.NoError(t, err, fmt.Sprint(err))
			} else {
				require.Error(t, err)
			}
			// get new state for verification
			actualCDP, found := keeper.GetCDP(ctx, tc.args.owner, collateralTest.Collateral.Token.GetName(), "")
			actualGDebt := keeper.GetGlobalDebt(ctx)
			actualCstate, _ := keeper.GetCollateralState(ctx, collateralTest.Collateral.Token.GetName())
			// check state
			require.Equal(t, tc.expectedState.CDP, actualCDP)
			if tc.expectedState.CDP.Collateral.Token.GetName() == "" { // if the expected CDP is blank, then expect the CDP to have been deleted (hence not found)
				require.False(t, found)
			} else {
				require.True(t, found)
			}
			require.Equal(t, tc.expectedState.GlobalDebt, actualGDebt)
			require.Equal(t, tc.expectedState.CollateralState, actualCstate)
			// check owner balance
			mock.CheckBalance(t, mapp, ownerAddr, tc.expectedState.OwnerCoins)
		})
	}
}

// TODO change to table driven test to test more test cases
func TestKeeper_PartialSeizeCDP(t *testing.T) {
	// Setup
	collateralTest := types.CDP{Owner: ownerAddr, Collateral: types.Collateral{Token: BaseFT{TokenName: _FT},
		Amount: i(100), InitialPrice: i(1)}, Liquidity: types.Liquidity{Coin: sdk.Coin{Denom: "xrp", Amount: i(1)}}}

	const collateral = "xrp"
	mapp, keeper := setUpMockAppWithoutGenesis()
	genAccs, addrs, _, _ := mock.CreateGenAccounts(1, cs(c(collateral, 100)))
	testAddr := addrs[0]
	mock.SetGenesis(mapp, genAccs)
	// setup pricefeed
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	keeper.pricefeed.AddAsset(ctx, collateral, "test description")
	keeper.pricefeed.SetPrice(ctx, sdk.AccAddress{}, "", collateralTest.Collateral.Token.GetName(), i(10), i(10))
	keeper.pricefeed.SetCurrentPrices(ctx)
	// Create CDP
	err := keeper.ModifyCDP(ctx, collateralTest.Owner, collateralTest.Collateral, collateralTest.Liquidity)
	require.NoError(t, err)
	// Reduce price
	keeper.pricefeed.SetPrice(ctx, sdk.AccAddress{}, "", collateralTest.Collateral.Token.GetName(), i(10), i(10))
	keeper.pricefeed.SetCurrentPrices(ctx)

	// Seize entire CDP
	err = keeper.PartialSeizeCDP(ctx, testAddr, collateralTest.Collateral, i(10), i(5))

	// Check
	require.NoError(t, err)
	_, found := keeper.GetCDP(ctx, ownerAddr, collateralTest.Collateral.Token.GetName(), "")
	require.False(t, found)
	collateralState, found := keeper.GetCollateralState(ctx, collateral)
	require.True(t, found)
	require.Equal(t, sdk.ZeroInt(), collateralState.TotalDebt)
}

func TestKeeper_GetCDPs(t *testing.T) {
	// setup keeper
	mapp, keeper := setUpMockAppWithoutGenesis()
	mock.SetGenesis(mapp, []auth.Account(nil))
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	// setup CDPs
	_, addrs := mock.GeneratePrivKeyAddressPairs(2)
	cdps := types.CDPs{
		collateralTest,
		collateralTest2,
		collateralTest3,
	}
	for _, cdp := range cdps {
		keeper.setCDP(ctx, cdp)
	}

	// Check nil params returns all CDPs
	returnedCdps, err := keeper.GetCDPs(ctx, "", sdk.Int{})
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs{
			collateralTest,
			collateralTest2,
			collateralTest3,
		},
		returnedCdps,
	)
	// Check correct CDPs filtered by collateral and sorted
	var num, _ = sdk.NewIntFromString("0.00000001")
	returnedCdps, err = keeper.GetCDPs(ctx, "xrp", num)
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs{
			collateralTest,
			collateralTest2,
			collateralTest3,
		},
		returnedCdps,
	)
	returnedCdps, err = keeper.GetCDPs(ctx, "xrp", sdk.Int{})
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs{
			collateralTest,
			collateralTest2,
			collateralTest3,
		},
		returnedCdps,
	)
	var num2, _ = sdk.NewIntFromString("0.9")
	returnedCdps, err = keeper.GetCDPs(ctx, "xrp", num2)
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs{
			collateralTest,
			collateralTest2,
			collateralTest3,
		},
		returnedCdps,
	)
	// Check high price returns no CDPs
	var num3, _ = sdk.NewIntFromString("999999999.99")
	returnedCdps, err = keeper.GetCDPs(ctx, "xrp", num3)
	require.NoError(t, err)
	require.Equal(t,
		types.CDPs(nil),
		returnedCdps,
	)
	// Check unauthorized collateral denom returns error
	var num4, _ = sdk.NewIntFromString("999999999.99")
	_, err = keeper.GetCDPs(ctx, "a non existent coin", num4)
	require.Error(t, err)
	// Check price without collateral returns error
	_, err = keeper.GetCDPs(ctx, "", d("0.34023"))
	require.Error(t, err)
	// Check deleting a CDP removes it
	keeper.deleteCDP(ctx, cdps[0])
	returnedCdps, err = keeper.GetCDPs(ctx, "", sdk.Dec{})
	require.NoError(t, err)
	require.Equal(t,
		CDPs{
			{addrs[0], "btc", i(10), i(20)},
			{addrs[1], "xrp", i(4000), i(2000)}},
		returnedCdps,
	)
}
func TestKeeper_GetSetDeleteCDP(t *testing.T) {
	// setup keeper, create CDP
	mapp, keeper := setUpMockAppWithoutGenesis()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	_, addrs := mock.GeneratePrivKeyAddressPairs(1)
	cdp := CDP{addrs[0], "xrp", i(412), i(56)}

	// write and read from store
	keeper.setCDP(ctx, cdp)
	readCDP, found := keeper.GetCDP(ctx, cdp.Owner, cdp.CollateralDenom)

	// check before and after match
	require.True(t, found)
	require.Equal(t, cdp, readCDP)

	// delete auction
	keeper.deleteCDP(ctx, cdp)

	// check auction does not exist
	_, found = keeper.GetCDP(ctx, cdp.Owner, cdp.CollateralDenom)
	require.False(t, found)
}
func TestKeeper_GetSetGDebt(t *testing.T) {
	// setup keeper, create GDebt
	mapp, keeper := setUpMockAppWithoutGenesis()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	gDebt := i(4120000)

	// write and read from store
	keeper.setGlobalDebt(ctx, gDebt)
	readGDebt := keeper.GetGlobalDebt(ctx)

	// check before and after match
	require.Equal(t, gDebt, readGDebt)
}

func TestKeeper_GetSetCollateralState(t *testing.T) {
	// setup keeper, create CState
	mapp, keeper := setUpMockAppWithoutGenesis()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	collateralState := CollateralState{"xrp", i(15400)}

	// write and read from store
	keeper.setCollateralState(ctx, collateralState)
	readCState, found := keeper.GetCollateralState(ctx, collateralState.Denom)

	// check before and after match
	require.Equal(t, collateralState, readCState)
	require.True(t, found)
}
