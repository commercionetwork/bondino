package cdp

type FT interface {
	GetName() string
}

type BaseFT struct {
	TokenName string
}

var _ FT = (*BaseFT)(nil)

func (bft BaseFT) GetName() string {
	return bft.TokenName
}

func (bft BaseFT) TokenType() string {
	return _FT
}
