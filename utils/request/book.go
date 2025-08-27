package request

type AddBook struct {
	Title    string `json:"title" binding:"required"`
	Author   string `json:"author" binding:"required"`
	ISBN     string `json:"isbn" binding:"required"`
	Category string `json:"category" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gte=0"`
}
