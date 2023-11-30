package managing

type ReadUserResponse struct {
	Id           string `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	EmailAddress string `json:"emailAddress"`
}
