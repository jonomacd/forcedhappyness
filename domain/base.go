package domain

type BasePage struct {
	HasSession  bool
	SessionUser User

	Next     string
	Previous string
}
