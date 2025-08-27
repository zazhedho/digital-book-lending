package request

type AddBook struct {
	Title    string `json:"title" binding:"required"`
	Author   string `json:"author" binding:"required"`
	ISBN     string `json:"isbn" binding:"required"`
	Category string `json:"category" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gte=0"`
}

type UpdateBook struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	ISBN     string `json:"isbn"`
	Category string `json:"category"`
	Quantity int    `json:"quantity" binding:"omitempty,gte=0"`
}
