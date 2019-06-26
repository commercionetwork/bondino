package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

type Collateral struct {
	Token        Token   `json:"token"`
	Amount       sdk.Int `json:"qty"`
	InitialPrice sdk.Int `json:"initial_price"`
}

func (c Collateral) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Collateral: 
  Token: 		 %s
  Quantity: 	 %s
  Initial price: %s`,
		c.Token,
		c.Amount,
		c.InitialPrice,
	))
}

//evaluate the collateral amount
func (c Collateral) CollateralValue() sdk.Int {
	return c.Amount.Mul(c.InitialPrice)
}

type Liquidity struct {
	Coin         sdk.Coin `json:"coin"`
	InitialPrice sdk.Int  `json:"initial_price"`
}

func (l Liquidity) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Liquidity:
  Coin:          %s
  Initial price: %s`,
		l.Coin,
		l.InitialPrice,
	))
}

// CurrentPrice struct that contains the metadata of a current price for a particular asset in the pricefeed module.
type CurrentPrice struct {
	AssetName string  `json:"asset_name"`
	AssetCode string  `json:"asset_code"`
	Price     sdk.Int `json:"price"`
	Expiry    sdk.Int `json:"expiry"` //this is the block height at which the price expires
}

// implement fmt.Stringer
func (cp CurrentPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s
Price: %s
Expiry: %s`, cp.AssetCode, cp.Price, cp.Expiry))
}

// CDP is the state of a single Collateralized Debt Position.
type CDP struct {
	// ID       	[]byte								// Remove ids to make things easier
	Owner      sdk.AccAddress `json:"owner"`      // Account that authorizes changes to the CDP
	Collateral Collateral     `json:"collateral"` // Collateral given from the user into the system
	Liquidity  Liquidity      `json:"liquidity"`  // Liquidity given to the user
}

func (cdp CDP) IsUnderCollateralized(price sdk.Int, liquidationRatio sdk.Dec) bool {
	collateralValue := sdk.NewDecFromInt(cdp.Collateral.Amount).MulInt(price)
	minCollateralValue := liquidationRatio.Mul(sdk.NewDecFromInt(cdp.Liquidity.Coin.Amount))
	return collateralValue.LT(minCollateralValue) // TODO LT or LTE?
}

func (cdp CDP) String() string {
	return strings.TrimSpace(fmt.Sprintf(`CDP:
  Owner:      %s
  Collateral: %s
  Liquidity:  %s`,
		cdp.Owner,
		cdp.Collateral,
		cdp.Liquidity,
	))
}

type CDPs []CDP

func (cdps CDPs) String() string {
	out := ""
	for _, cdp := range cdps {
		out += cdp.String() + "\n"
	}
	return out
}

// CollateralState stores global information tied to a particular collateral type.
type CollateralState struct {
	Denom     string  // Type of collateral
	TotalDebt sdk.Int // total debt collateralized by a this coin type
	//AccumulatedFees sdk.Int // Ignoring fees for now
}

type CdpModuleParams struct {
	GlobalDebtLimit  sdk.Int
	CollateralParams []CollateralParams
}

// Implement fmt.Stringer interface for cli querying
func (p CdpModuleParams) String() string {
	out := fmt.Sprintf(`Params:
	Global Debt Limit: %s
	Collateral Params:`,
		p.GlobalDebtLimit,
	)
	for _, cp := range p.CollateralParams {
		out += fmt.Sprintf(`
		%s
			Liquidation Ratio: %s
			Debt Limit:        %s`,
			cp.Denom,
			cp.LiquidationRatio,
			cp.DebtLimit,
		)
	}
	return out
}

// Helper methods to search the list of collateral params for a particular denom. Wouldn't be needed if amino supported maps.

func (p CdpModuleParams) GetCollateralParams(collateralDenom string) CollateralParams {
	// search for matching denom, return
	for _, cp := range p.CollateralParams {
		if cp.Denom == collateralDenom {
			return cp
		}
	}
	// panic if not found, to be safe
	panic("collateral params not found in module params")
}
func (p CdpModuleParams) IsCollateralPresent(collateralDenom string) bool {
	// search for matching denom, return
	for _, cp := range p.CollateralParams {
		if cp.Denom == collateralDenom {
			return true
		}
	}
	return false
}

type CollateralParams struct {
	Denom            string  // Coin name of collateral type
	LiquidationRatio sdk.Dec // The ratio (Collateral (priced in stable coin) / Debt) under which a CDP will be liquidated
	DebtLimit        sdk.Int // Maximum amount of debt allowed to be drawn from this collateral type
	//DebtFloor        sdk.Int // used to prevent dust
}
