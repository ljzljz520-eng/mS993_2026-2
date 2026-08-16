package domain

type ProductRepository interface {
	ListProducts() []Product
	GetProduct(id string) (Product, bool)
	UpdatePrice(id string, price Product) (Product, bool)
	UpdateImage(id, imageURL string) (Product, bool)
	ApplyStockChange(id string, delta int) (Product, bool)
}

type LogRepository interface {
	Append(log OperationLog)
	List() []OperationLog
}

type TaskRepository interface {
	Get(id string) (Task, bool)
	Complete(id string) bool
}
