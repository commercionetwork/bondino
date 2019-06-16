package cdp

import (
	"fmt"
	"github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/token"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Collateral struct {
	Token        token.Token `json:"token"`
	Qty          sdk.Int     `json:"qty"`
	InitialPrice sdk.Int     `json:"initial_price"`
}

func (c Collateral) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Collateral: 
  Token: 		 %s
  Quantity: 	 %s
  Initial price: %s`,
		c.Token,
		c.Qty,
		c.InitialPrice,
	))
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

// CDP is the state of a single Collateralized Debt Position.
type CDP struct {
	// ID       	[]byte								// Remove ids to make things easier
	Owner      sdk.AccAddress `json:"owner"`      // Account that authorizes changes to the CDP
	Collateral Collateral     `json:"collateral"` // Collateral given from the user into the system
	Liquidity  Liquidity      `json:"liquidity"`  // Liquidity given to the user
}

func (cdp CDP) IsUnderCollateralized(price sdk.Dec, liquidationRatio sdk.Dec) bool {
	collateralValue := sdk.NewDecFromInt(cdp.Collateral.Qty).Mul(price)
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

// byCollateralRatio is used to sort CDPs
type byCollateralRatio CDPs

func (cdps byCollateralRatio) Len() int      { return len(cdps) }
func (cdps byCollateralRatio) Swap(i, j int) { cdps[i], cdps[j] = cdps[j], cdps[i] }
func (cdps byCollateralRatio) Less(i, j int) bool {
	// Sort by "collateral ratio" ie collateralAmount/Debt
	// The comparison is: collat_i/debt_i < collat_j/debt_j
	// But to avoid division this can be rearranged to: collat_i*debt_j < collat_j*debt_i
	// Provided the values are positive, so check for positive values.
	if cdps[i].Collateral.Qty.IsNegative() ||
		cdps[i].Liquidity.Coin.Amount.IsNegative() ||
		cdps[j].Collateral.Qty.IsNegative() ||
		cdps[j].Liquidity.Coin.Amount.IsNegative() {
		panic("negative collateral and debt not supported in CDPs")
	}
	// TODO overflows could cause panics
	left := cdps[i].Collateral.Qty.Mul(cdps[j].Liquidity.Coin.Amount)
	right := cdps[j].Collateral.Qty.Mul(cdps[i].Liquidity.Coin.Amount)
	return left.LT(right)
}

// CollateralState stores global information tied to a particular collateral type.
type CollateralState struct {
	Denom     string  // Type of collateral
	TotalDebt sdk.Int // total debt collateralized by a this coin type
	//AccumulatedFees sdk.Int // Ignoring fees for now
}
