package client

// interface implemented by FT and NFT
type Token interface {
	TokenType() string
	GetName() string
}
