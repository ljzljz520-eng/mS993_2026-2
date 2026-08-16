package domain

import "github.com/shopspring/decimal"

type ProductType string

const (
	ProductCandle  ProductType = "candle"
	ProductOil     ProductType = "essential-oil"
	ProductStone   ProductType = "diffuser-stone"
	ProductGiftBox ProductType = "gift-box"
)

type Product struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Type     ProductType     `json:"type"`
	Price    decimal.Decimal `json:"price"`
	Stock    int             `json:"stock"`
	ImageURL string          `json:"imageUrl"`
}

type StockChange struct {
	ProductID string `json:"productId"`
	Delta     int    `json:"delta"`
	Reason    string `json:"reason"`
}

type OperationLog struct {
	Sequence  int    `json:"sequence"`
	Action    string `json:"action"`
	ProductID string `json:"productId,omitempty"`
	TaskID    string `json:"taskId,omitempty"`
	Detail    string `json:"detail"`
}
