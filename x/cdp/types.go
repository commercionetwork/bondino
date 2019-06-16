package cdp

import "github.com/commercionetwork/cosmos-hackatom-2019/blockchain/x/types"

// byCollateralRatio is used to sort CDPs
type byCollateralRatio types.CDPs

func (cdps byCollateralRatio) Len() int      { return len(cdps) }
func (cdps byCollateralRatio) Swap(i, j int) { cdps[i], cdps[j] = cdps[j], cdps[i] }
func (cdps byCollateralRatio) Less(i, j int) bool {
	// Sort by "collateral ratio" ie collateralAmount/Debt
	// The comparison is: collat_i/debt_i < collat_j/debt_j
	// But to avoid division this can be rearranged to: collat_i*debt_j < collat_j*debt_i
	// Provided the values are positive, so check for positive values.
	if cdps[i].Collateral.Amount.IsNegative() ||
		cdps[i].Liquidity.Coin.Amount.IsNegative() ||
		cdps[j].Collateral.Amount.IsNegative() ||
		cdps[j].Liquidity.Coin.Amount.IsNegative() {
		panic("negative collateral and debt not supported in CDPs")
	}
	// TODO overflows could cause panics
	left := cdps[i].Collateral.Amount.Mul(cdps[j].Liquidity.Coin.Amount)
	right := cdps[j].Collateral.Amount.Mul(cdps[i].Liquidity.Coin.Amount)
	return left.LT(right)
}
