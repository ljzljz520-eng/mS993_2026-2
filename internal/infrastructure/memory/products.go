package memory

import (
	"sync"

	"aroma-maintenance/internal/domain"
)

type ProductStore struct {
	mu       sync.RWMutex
	products map[string]domain.Product
	order    []string
}

func NewProductStore(fixture []domain.Product) *ProductStore {
	products := make(map[string]domain.Product, len(fixture))
	order := make([]string, 0, len(fixture))
	for _, product := range fixture {
		products[product.ID] = product
		order = append(order, product.ID)
	}
	return &ProductStore{products: products, order: order}
}

func (s *ProductStore) ListProducts() []domain.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.Product, 0, len(s.order))
	for _, id := range s.order {
		result = append(result, s.products[id])
	}
	return result
}

func (s *ProductStore) GetProduct(id string) (domain.Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	product, ok := s.products[id]
	return product, ok
}

func (s *ProductStore) UpdatePrice(id string, product domain.Product) (domain.Product, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.products[id]
	if !ok {
		return domain.Product{}, false
	}
	current.Price = product.Price
	s.products[id] = current
	return current, true
}

func (s *ProductStore) UpdateImage(id, imageURL string) (domain.Product, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	product, ok := s.products[id]
	if !ok {
		return domain.Product{}, false
	}
	product.ImageURL = imageURL
	s.products[id] = product
	return product, true
}

func (s *ProductStore) ApplyStockChange(id string, delta int) (domain.Product, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	product, ok := s.products[id]
	if !ok {
		return domain.Product{}, false
	}
	product.Stock += delta
	s.products[id] = product
	return product, true
}
