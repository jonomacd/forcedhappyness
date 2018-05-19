package domain

type ErrorSetter interface {
	SetErrorMessage(err string)
}

type BasePage struct {
	HasSession  bool
	SessionUser User

	Next     string
	Previous string

	Error      string
	ErrorToast string
}

func (bp *BasePage) SetErrorMessage(err string) {
	bp.Error = err
}

func NewErrorBase(err string) *BasePage {
	return &BasePage{
		Error: err,
	}
}
