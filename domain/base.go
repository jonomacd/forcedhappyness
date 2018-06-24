package domain

import "time"

type ErrorSetter interface {
	SetErrorMessage(err string)
}

var (
	backgrounds = []string{
		"bg1.jpg",
		"bg2.jpg",
		"bg3.jpg",
		"bg4.jpg",
		"bg5.jpg",
		"bg6.jpg",
		"bg7.jpg",
		"bg8.jpg",
		"bg9.jpg",
	}
)

type BasePage struct {
	HasSession  bool
	SessionUser User

	Next     string
	Previous string

	Error       string
	ErrorToast  string
	QueryParams map[string]string
}

func (bp *BasePage) SetErrorMessage(err string) {
	bp.Error = err
}

func (bp BasePage) BackgroundImage() string {
	hour := time.Now().Hour()
	return backgrounds[hour%len(backgrounds)]
}

func NewErrorBase(err string) *BasePage {
	return &BasePage{
		Error: err,
	}
}
