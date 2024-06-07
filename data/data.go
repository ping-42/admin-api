package data

type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var Items = []Item{
	{ID: 1, Name: "Item 1", Price: 10.0},
	{ID: 2, Name: "Item 2", Price: 20.0},
	{ID: 3, Name: "Item 3", Price: 30.0},
}
