package dto

type Privileges struct {
	IsAuthorized bool     `json:"isAuthorized"`
	UserId       string   `json:"userId"`
	Pages        Pages    `json:"pages"`
	Entities     Entities `json:"entities"`
}

type Pages struct {
	Main         []string `json:"main"`
	Login        []string `json:"login"`
	Calculations []string `json:"calculations"`
	Requests     []string `json:"requests"`
}

type Entities struct {
	News         []string `json:"news"`
	Promo        []string `json:"promo"`
	Documents    []string `json:"documents"`
	Calculations []string `json:"calculations"`
	Reports      []string `json:"reports"`
	Requests     []string `json:"requests"`
}
