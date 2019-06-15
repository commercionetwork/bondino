module github.com/cosmos/cosmos-sdk

require (
	github.com/bartekn/go-bip39 v0.0.0-20171116152956-a05967ea095d
	github.com/bgentry/speakeasy v0.1.0
	github.com/btcsuite/btcd v0.0.0-20190427004231-96897255fd17
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/cosmos/ledger-cosmos-go v0.10.3
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.1
	github.com/gorilla/mux v1.7.2
	github.com/kava-labs/kava-devnet/blockchain v0.0.0-20190610193355-a54e810789f6
	github.com/mattn/go-isatty v0.0.7
	github.com/pelletier/go-toml v1.4.0
	github.com/pkg/errors v0.8.1
	github.com/rakyll/statik v0.1.6
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/iavl v0.12.2
	github.com/tendermint/tendermint v0.31.5
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f
	google.golang.org/genproto v0.0.0-20180831171423-11092d34479b // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
