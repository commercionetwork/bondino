<<<<<<< Updated upstream
module github.com/commercionetwork/cosmos-hackatom-2019/blockchain
=======
module github.com/commercionetwork/cosmos-hackathom-2019/blockchain
>>>>>>> Stashed changes

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.28.2-0.20190521100210-dd89c329516e
	github.com/cosmos/gaia v0.0.1-0.20190524130037-594c2adbe776
	github.com/gorilla/mux v1.7.2
	github.com/kava-labs/kava-devnet/blockchain v0.0.0-20190610193355-a54e810789f6
	github.com/rakyll/statik v0.1.6
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.31.5
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
