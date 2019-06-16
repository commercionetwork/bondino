package cdp

import (
	"bytes"
	"fmt"
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/types"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// GovDenom asset code of the governance coin
const GovDenom = "tmnt"

// Keeper cdp Keeper
type Keeper struct {
	storeKey       sdk.StoreKey
	pricefeed      types.PricefeedKeeper
	bank           bankKeeper
	paramsSubspace params.Subspace
	cdc            *codec.Codec
}

// NewKeeper creates a new keeper
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, subspace params.Subspace, pricefeed types.PricefeedKeeper, bank bankKeeper) Keeper {
	subspace = subspace.WithKeyTable(createParamsKeyTable())
	return Keeper{
		storeKey:       storeKey,
		pricefeed:      pricefeed,
		bank:           bank,
		paramsSubspace: subspace,
		cdc:            cdc,
	}
}

func (k Keeper) getAssetCodeAndName(token types.Token) (string, string) {

	assetName := token.GetName()

	var assetCode string
	switch token := token.(type) {
	case BaseFT:
		assetCode = ""
		break
	case NFT:
		assetCode = token.GetID()
		break
	}

	return assetCode, assetName
}

// ModifyCDP creates, changes, or deletes a CDP
// TODO can/should this function be split up?
func (k Keeper) ModifyCDP(ctx sdk.Context, owner sdk.AccAddress, collateral types.Collateral, liquidity types.Liquidity) sdk.Error {

	// Phase 1: Get state, make changes in memory and check if they're ok.

	collateralName := collateral.Token.GetName()

	// Check collateral type ok
	p := k.GetParams(ctx)
	if !p.IsCollateralPresent(collateralName) { // maybe abstract this logic into GetCDP
		return sdk.ErrInternal("collateral type not enabled to create CDPs")
	}

	// Check the owner has enough collateral and stable coins

	// adding collateral to CDP
	if collateral.Amount.IsPositive() {
		ok := k.bank.HasCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin(collateralName, collateral.Amount)))
		if !ok {
			return sdk.ErrInsufficientCoins("not enough collateral in sender's account")
		}
	}
	// reducing liquidity, by adding stable coin to CDP
	if liquidity.Coin.Amount.IsNegative() {
		ok := k.bank.HasCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin(collateral.Token.GetName(), liquidity.Coin.Amount.Neg())))
		if !ok {
			return sdk.ErrInsufficientCoins("not enough stable coin in sender's account")
		}
	}

	// Change collateral and debt recorded in CDP

	// Get CDP (or create if not exists)
	var cdp types.CDP
	var found bool

	switch collToken := (collateral.Token).(type) {
	case BaseFT:
		cdp, found = k.GetCDP(ctx, owner, collateralName, "")
	case BaseNFT:
		cdp, found = k.GetCDP(ctx, owner, collateralName, collToken.ID)
	}

	if !found {
		cdp = types.CDP{Owner: owner, Collateral: collateral, Liquidity: liquidity}
	}
	// Add/Subtract collateral and debt
	if cdp.Collateral.Amount.IsNegative() {
		return sdk.ErrInternal(" can't withdraw more collateral than exists in CDP")
	}
	if cdp.Liquidity.Coin.Amount.IsNegative() {
		return sdk.ErrInternal("can't pay back more debt than exists in CDP")
	}

	assetCode, assetName := k.getAssetCodeAndName(cdp.Collateral.Token)

	collateralCurrentPrice := k.pricefeed.GetCurrentPrice(ctx, assetCode, assetName)

	// if the price is zero, then ask for the price of the token
	if collateralCurrentPrice.Price.IsZero() {
		k.pricefeed.AskForPrice(ctx, assetCode, assetName)
	}

	isUnderCollateralized := cdp.IsUnderCollateralized(
		collateralCurrentPrice.Price,
		p.GetCollateralParams(cdp.Collateral.Token.GetName()).LiquidationRatio,
	)

	if isUnderCollateralized {
		return sdk.ErrInternal("Change to CDP would put it below liquidation ratio")
	}
	// TODO check for dust

	// Add/Subtract from global debt limit
	gDebt := k.GetGlobalDebt(ctx)
	gDebt = gDebt.Add(liquidity.Coin.Amount)
	if gDebt.IsNegative() {
		return sdk.ErrInternal("global debt can't be negative") // This should never happen if debt per CDP can't be negative
	}
	if gDebt.GT(p.GlobalDebtLimit) {
		return sdk.ErrInternal("change to CDP would put the system over the global debt limit")
	}

	// Add/Subtract from collateral debt limit
	collateralState, found := k.GetCollateralState(ctx, cdp.Collateral.Token.GetName())
	if !found {
		collateralState = types.CollateralState{Denom: cdp.Collateral.Token.GetName(), TotalDebt: sdk.ZeroInt()} // Already checked that this denom is authorized, so ok to create new CollateralState
	}
	collateralState.TotalDebt = collateralState.TotalDebt.Add(liquidity.Coin.Amount)
	if collateralState.TotalDebt.IsNegative() {
		return sdk.ErrInternal("total debt for this collateral type can't be negative") // This should never happen if debt per CDP can't be negative
	}
	if collateralState.TotalDebt.GT(p.GetCollateralParams(cdp.Collateral.Token.GetName()).DebtLimit) {
		return sdk.ErrInternal("change to CDP would put the system over the debt limit for this collateral type")
	}

	// Phase 2: Update all the state

	// change owner's coins (increase or decrease)
	var err sdk.Error
	if collateral.Amount.IsNegative() {
		_, err = k.bank.AddCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin(collateral.Token.GetName(), collateral.Amount.Neg())))
	} else {
		_, err = k.bank.SubtractCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin(collateral.Token.GetName(), collateral.Amount)))
	}
	if err != nil {
		panic(err) // this shouldn't happen because coin balance was checked earlier
	}
	if liquidity.Coin.Amount.IsNegative() {
		_, err = k.bank.SubtractCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin(collateral.Token.GetName(), liquidity.Coin.Amount.Neg())))
	} else {
		_, err = k.bank.AddCoins(ctx, owner, sdk.NewCoins(sdk.NewCoin(collateral.Token.GetName(), liquidity.Coin.Amount)))
	}
	if err != nil {
		panic(err) // this shouldn't happen because coin balance was checked earlier
	}

	//TODO Here calculate liquidityValue
	// Set CDP
	liquidityCurrentPrice := k.pricefeed.GetCurrentPrice(ctx, "", liquidity.Coin.Denom)
	if liquidityCurrentPrice.Price.IsZero() {
		return sdk.ErrInvalidCoins("Liquidity price cant be equal to zero")
	}

	// get the collateral value = price * quantity
	collateralValue := collateralCurrentPrice.Price.Mul(collateral.Amount)

	// get the liquidity amount = collateral-value / liquidity price
	cdp.Liquidity.Coin.Amount = collateralValue.Quo(liquidityCurrentPrice.Price)

	if cdp.Collateral.Amount.IsZero() && cdp.Liquidity.Coin.Amount.IsZero() { // TODO maybe abstract this logic into setCDP
		k.deleteCDP(ctx, cdp)
	} else {
		k.setCDP(ctx, cdp)
	}
	// set total debts
	k.setGlobalDebt(ctx, gDebt)
	k.setCollateralState(ctx, collateralState)

	return nil
}

func (k Keeper) ModifyCDPType(ctx sdk.Context, assetName string, assetCode string) sdk.Error {

	// Get all cdps with assetName
	cdps, _ := k.GetCDPs(ctx, assetName, sdk.NewInt(0))
	for _, cdp := range cdps {

		switch token := cdp.Collateral.Token.(type) {
		case BaseFT:
			{
				// this shouldn't verify but we've included it for completeness
				// ideally no CDP can exist with a collateral being a FT with price zero
				// this would throw an error inside the ModifyCDP method
				if token.GetName() == assetName {
					err := k.ModifyCDP(ctx, cdp.Owner, cdp.Collateral, cdp.Liquidity)
					if err != nil {
						return err
					}
				}
				break
			}

		case BaseNFT:
			{
				// get the token based on the asset name and the asset code and update the CDP
				// this will trigger the funds being moved into the user wallet from the pool
				if token.Name == assetName && token.ID == assetCode {
					err := k.ModifyCDP(ctx, cdp.Owner, cdp.Collateral, cdp.Liquidity)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// TODO
// // TransferCDP allows people to transfer ownership of their CDPs to others
// func (k Keeper) TransferCDP(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, collateralDenom string) sdk.Error {
// 	return nil
// }

// PartialSeizeCDP removes collateral and debt from a CDP and decrements global debt counters. It does not move collateral to another account so is unsafe.
// TODO should this be made safer by moving collateral to liquidatorModuleAccount ? If so how should debt be moved?
func (k Keeper) PartialSeizeCDP(ctx sdk.Context, owner sdk.AccAddress, collateral types.Collateral, collateralToSeize sdk.Int, debtToSeize sdk.Int) sdk.Error {
	// Get CDP

	var cdp types.CDP
	var found bool

	switch collToken := (collateral.Token).(type) {
	case BaseFT:
		cdp, found = k.GetCDP(ctx, owner, collateral.Token.GetName(), "")
	case BaseNFT:
		cdp, found = k.GetCDP(ctx, owner, collateral.Token.GetName(), collToken.ID)
	}
	if !found {
		return sdk.ErrInternal("could not find CDP")
	}

	assetCode, assetName := k.getAssetCodeAndName(cdp.Collateral.Token)

	// Check if CDP is undercollateralized
	p := k.GetParams(ctx)
	isUnderCollateralized := cdp.IsUnderCollateralized(
		k.pricefeed.GetCurrentPrice(ctx, assetCode, assetName).Price,
		p.GetCollateralParams(cdp.Collateral.Token.GetName()).LiquidationRatio,
	)
	if !isUnderCollateralized {
		return sdk.ErrInternal("CDP is not currently under the liquidation ratio")
	}

	// Remove Collateral
	if collateralToSeize.IsNegative() {
		return sdk.ErrInternal("cannot seize negative collateral")
	}
	//cdp.CollateralAmount = cdp.CollateralAmount.Sub(collateralToSeize)
	if cdp.Collateral.Amount.IsNegative() {
		return sdk.ErrInternal("can't seize more collateral than exists in CDP")
	}

	// Remove Debt
	if debtToSeize.IsNegative() {
		return sdk.ErrInternal("cannot seize negative debt")
	}
	//cdp.Debt = cdp.Debt.Sub(debtToSeize)
	if cdp.Liquidity.Coin.Amount.IsNegative() {
		return sdk.ErrInternal("can't seize more debt than exists in CDP")
	}

	// Update debt per collateral type
	collateralState, found := k.GetCollateralState(ctx, cdp.Collateral.Token.GetName())
	if !found {
		return sdk.ErrInternal("could not find collateral state")
	}
	collateralState.TotalDebt = collateralState.TotalDebt.Sub(debtToSeize)
	if collateralState.TotalDebt.IsNegative() {
		return sdk.ErrInternal("Total debt per collateral type is negative.") // This should not happen given the checks on the CDP.
	}

	// Note: Global debt is not decremented here. It's only decremented when debt and stable coin are annihilated (aka heal)
	// TODO update global seized debt? this is what maker does (named vice in Vat.grab) but it's not used anywhere

	// Store updated state

	if cdp.Collateral.Amount.IsZero() && cdp.Liquidity.Coin.Amount.IsZero() { // TODO maybe abstract this logic into setCDP
		k.deleteCDP(ctx, cdp)
	} else {
		k.setCDP(ctx, cdp)
	}
	k.setCollateralState(ctx, collateralState)
	return nil
}

// ReduceGlobalDebt decreases the stored global debt counter. It is used by the liquidator when it annihilates debt and stable coin.
// TODO Can the interface between cdp and liquidator modules be improved so that this function doesn't exist?
func (k Keeper) ReduceGlobalDebt(ctx sdk.Context, amount sdk.Int) sdk.Error {
	if amount.IsNegative() {
		return sdk.ErrInternal("reduction in global debt must be a positive amount")
	}
	newGDebt := k.GetGlobalDebt(ctx).Sub(amount)
	if newGDebt.IsNegative() {
		return sdk.ErrInternal("cannot reduce global debt by amount specified")
	}
	k.setGlobalDebt(ctx, newGDebt)
	return nil
}

// deprecated - use collateral.Token.GetName() instead
func (k Keeper) GetStableDenom() string {
	return ""
}
func (k Keeper) GetGovDenom() string {
	return GovDenom
}

// ---------- Module Parameters ----------

func (k Keeper) GetParams(ctx sdk.Context) types.CdpModuleParams {
	var p types.CdpModuleParams
	k.paramsSubspace.Get(ctx, moduleParamsKey, &p)
	return p
}

// This is only needed to be able to setup the store from the genesis file. The keeper should not change any of the params itself.
func (k Keeper) setParams(ctx sdk.Context, cdpModuleParams types.CdpModuleParams) {
	k.paramsSubspace.Set(ctx, moduleParamsKey, &cdpModuleParams)
}

// ---------- Store Wrappers ----------

func (k Keeper) getCDPKeyPrefix(collateralDenom string) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("cdp"),
			[]byte(collateralDenom),
		},
		nil, // no separator
	)
}
func (k Keeper) getCDPKey(owner sdk.AccAddress, collateralDenom string) []byte {
	return bytes.Join(
		[][]byte{
			k.getCDPKeyPrefix(collateralDenom),
			[]byte(owner.String()),
		},
		nil, // no separator
	)
}
func (k Keeper) GetCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string, nftID string) (types.CDP, bool) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get CDP
	bz := store.Get(k.getCDPKey(owner, collateralDenom))
	// unmarshal
	if bz == nil {
		return types.CDP{}, false
	}
	var cdp types.CDP
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &cdp)
	return cdp, true
}
func (k Keeper) setCDP(ctx sdk.Context, cdp types.CDP) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// marshal and set
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(cdp)
	store.Set(k.getCDPKey(cdp.Owner, cdp.Collateral.Token.GetName()), bz)
}
func (k Keeper) deleteCDP(ctx sdk.Context, cdp types.CDP) { // TODO should this id the cdp by passing in owner,collateralDenom pair?
	// get store
	store := ctx.KVStore(k.storeKey)
	// delete key
	store.Delete(k.getCDPKey(cdp.Owner, cdp.Collateral.Token.GetName()))
}

// GetCDPs returns all CDPs, optionally filtered by collateral type and liquidation price.
// `price` filters for CDPs that will be below the liquidation ratio when the collateral is at that specified price.
func (k Keeper) GetCDPs(ctx sdk.Context, collateralDenom string, price sdk.Int) (types.CDPs, sdk.Error) {
	// Validate inputs
	parameters := k.GetParams(ctx)
	if len(collateralDenom) != 0 && !parameters.IsCollateralPresent(collateralDenom) {
		return nil, sdk.ErrInternal("collateral denom not authorized")
	}
	if len(collateralDenom) == 0 && !price.IsNegative() {
		return nil, sdk.ErrInternal("cannot specify price without collateral denom")
	}

	// Get an iterator over CDPs
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, k.getCDPKeyPrefix(collateralDenom)) // could be all CDPs is collateralDenom is ""

	// Decode CDPs into slice
	var cdps types.CDPs
	for ; iter.Valid(); iter.Next() {
		var cdp types.CDP
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &cdp)
		cdps = append(cdps, cdp)
	}

	// Sort by collateral ratio (collateral/debt)
	sort.Sort(byCollateralRatio(cdps)) // TODO this doesn't make much sense across different collateral types

	// Filter for CDPs that would be under-collateralized at the specified price
	// If price is nil or -ve, skip the filtering as it would return all CDPs anyway
	if !price.IsNegative() {
		var filteredCDPs types.CDPs
		for _, cdp := range cdps {
			if cdp.IsUnderCollateralized(price, parameters.GetCollateralParams(collateralDenom).LiquidationRatio) {
				filteredCDPs = append(filteredCDPs, cdp)
			} else {
				break // break early because list is sorted
			}
		}
		cdps = filteredCDPs
	}

	return cdps, nil
}

var globalDebtKey = []byte("globalDebt")

func (k Keeper) GetGlobalDebt(ctx sdk.Context) sdk.Int {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get bytes
	bz := store.Get(globalDebtKey)
	// unmarshal
	if bz == nil {
		panic("global debt not found")
	}
	var globalDebt sdk.Int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &globalDebt)
	return globalDebt
}
func (k Keeper) setGlobalDebt(ctx sdk.Context, globalDebt sdk.Int) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// marshal and set
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(globalDebt)
	store.Set(globalDebtKey, bz)
}

func (k Keeper) getCollateralStateKey(collateralDenom string) []byte {
	return []byte(collateralDenom)
}
func (k Keeper) GetCollateralState(ctx sdk.Context, collateralDenom string) (types.CollateralState, bool) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get bytes
	bz := store.Get(k.getCollateralStateKey(collateralDenom))
	// unmarshal
	if bz == nil {
		return types.CollateralState{}, false
	}
	var collateralState types.CollateralState
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &collateralState)
	return collateralState, true
}
func (k Keeper) setCollateralState(ctx sdk.Context, collateralstate types.CollateralState) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// marshal and set
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(collateralstate)
	store.Set(k.getCollateralStateKey(collateralstate.Denom), bz)
}

// ---------- Weird Bank Stuff ----------
// This only exists because module accounts aren't really a thing yet.
// Also because we need module accounts that allow for burning/minting.

// These functions make the CDP module act as a bank keeper, ie it fulfills the bank.Keeper interface.
// It intercepts calls to send coins to/from the liquidator module account, otherwise passing the calls onto the normal bank keeper.

// Not sure if module accounts are good, but they make the auction module more general:
// - startAuction would just "mints" coins, relying on calling function to decrement them somewhere
// - closeAuction would have to call something specific for the receiver module to accept coins (like liquidationKeeper.AddStableCoins)

// The auction and liquidator modules can probably just use SendCoins to keep things safe (instead of AddCoins and SubtractCoins).
// So they should define their own interfaces which this module should fulfill, rather than this fulfilling the entire bank.Keeper interface.

// bank.Keeper interfaces:
// type SendKeeper interface {
// 	type ViewKeeper interface {
// 		GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
// 		HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool
// 		Codespace() sdk.CodespaceType
// 	}
// 	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
// 	GetSendEnabled(ctx sdk.Context) bool
// 	SetSendEnabled(ctx sdk.Context, enabled bool)
// }
// type Keeper interface {
// 	SendKeeper
// 	SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error
// 	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
// 	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
// 	InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) (sdk.Tags, sdk.Error)
// 	DelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
// 	UndelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)

var LiquidatorAccountAddress = sdk.AccAddress([]byte("whatever"))
var liquidatorAccountKey = []byte("liquidatorAccount")

func (k Keeper) GetLiquidatorAccountAddress() sdk.AccAddress {
	return LiquidatorAccountAddress
}

type LiquidatorModuleAccount struct {
	Coins sdk.Coins // keeps track of seized collateral, surplus usdx, and mints/burns gov coins
}

func (k Keeper) AddCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) (sdk.Coins, sdk.Error) {
	// intercept module account
	if address.Equals(LiquidatorAccountAddress) {
		if !amount.IsValid() {
			return nil, sdk.ErrInvalidCoins(amount.String())
		}
		// remove gov token from list
		filteredCoins := stripGovCoin(amount)
		// add coins to module account
		lma := k.getLiquidatorModuleAccount(ctx)
		updatedCoins := lma.Coins.Add(filteredCoins)
		if updatedCoins.IsAnyNegative() {
			return amount, sdk.ErrInsufficientCoins(fmt.Sprintf("insufficient account funds; %s < %s", lma.Coins, amount))
		}
		lma.Coins = updatedCoins
		k.setLiquidatorModuleAccount(ctx, lma)
		return updatedCoins, nil
	} else {
		return k.bank.AddCoins(ctx, address, amount)
	}
}

// TODO abstract stuff better
func (k Keeper) SubtractCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) (sdk.Coins, sdk.Error) {
	// intercept module account
	if address.Equals(LiquidatorAccountAddress) {
		if !amount.IsValid() {
			return nil, sdk.ErrInvalidCoins(amount.String())
		}
		// remove gov token from list
		filteredCoins := stripGovCoin(amount)
		// subtract coins from module account
		lma := k.getLiquidatorModuleAccount(ctx)
		updatedCoins, isNegative := lma.Coins.SafeSub(filteredCoins)
		if isNegative {
			return amount, sdk.ErrInsufficientCoins(fmt.Sprintf("insufficient account funds; %s < %s", lma.Coins, amount))
		}
		lma.Coins = updatedCoins
		k.setLiquidatorModuleAccount(ctx, lma)
		return updatedCoins, nil
	} else {
		return k.bank.SubtractCoins(ctx, address, amount)
	}
}

// TODO Should this return anything for the gov coin balance? Currently returns nothing.
func (k Keeper) GetCoins(ctx sdk.Context, address sdk.AccAddress) sdk.Coins {
	if address.Equals(LiquidatorAccountAddress) {
		return k.getLiquidatorModuleAccount(ctx).Coins
	} else {
		return k.bank.GetCoins(ctx, address)
	}
}

// TODO test this with unsorted coins
func (k Keeper) HasCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) bool {
	if address.Equals(LiquidatorAccountAddress) {
		return true
	} else {
		return k.getLiquidatorModuleAccount(ctx).Coins.IsAllGTE(stripGovCoin(amount))
	}
}

func (k Keeper) getLiquidatorModuleAccount(ctx sdk.Context) LiquidatorModuleAccount {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get bytes
	bz := store.Get(liquidatorAccountKey)
	if bz == nil {
		return LiquidatorModuleAccount{} // TODO is it safe to do this, or better to initialize the account explicitly
	}
	// unmarshal
	var lma LiquidatorModuleAccount
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &lma)
	return lma
}
func (k Keeper) setLiquidatorModuleAccount(ctx sdk.Context, lma LiquidatorModuleAccount) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(lma)
	store.Set(liquidatorAccountKey, bz)
}
func stripGovCoin(coins sdk.Coins) sdk.Coins {
	filteredCoins := sdk.NewCoins()
	for _, c := range coins {
		if c.Denom != GovDenom {
			filteredCoins = append(filteredCoins, c)
		}
	}
	return filteredCoins
}
