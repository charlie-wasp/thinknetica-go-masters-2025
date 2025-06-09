package models

type Beer struct {
	ID      int    `json:"id"`
	Name    string `json:"name,omitempty"`
	Style   string `json:"style,omitempty"`
	Brewery string `json:"brewery,omitempty"`
}

type User struct {
	ID       int    `json:"id"`
	Fullname string `json:"fullname"`
}

type Review struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Rating  *int   `json:"rating"`
	UserID  int    `json:"user_id"`
	BeerID  int    `json:"beer_id"`
}
