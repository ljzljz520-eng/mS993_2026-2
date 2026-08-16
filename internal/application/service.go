package application

import (
	"fmt"

	"github.com/shopspring/decimal"

	"aroma-maintenance/internal/domain"
	"aroma-maintenance/internal/infrastructure/memory"
)

const fixtureTaskID = "task-stock-refresh-001"

type MaintenanceService struct {
	products *memory.ProductStore
	logs     *memory.LogStore
	tasks    *memory.TaskStore
	runner   *TaskRunner
}

func NewMaintenanceService() *MaintenanceService {
	products := memory.NewProductStore([]domain.Product{
		{ID: "candle-amber", Name: "琥珀木芯蜡烛", Type: domain.ProductCandle, Price: decimal.RequireFromString("129.00"), Stock: 18, ImageURL: "/assets/candle-amber.jpg"},
		{ID: "oil-lavender", Name: "薰衣草舒缓精油", Type: domain.ProductOil, Price: decimal.RequireFromString("88.00"), Stock: 32, ImageURL: "/assets/oil-lavender.jpg"},
		{ID: "stone-moon", Name: "月光扩香石", Type: domain.ProductStone, Price: decimal.RequireFromString("59.00"), Stock: 24, ImageURL: "/assets/stone-moon.jpg"},
		{ID: "gift-calm", Name: "静谧入眠礼盒", Type: domain.ProductGiftBox, Price: decimal.RequireFromString("239.00"), Stock: 9, ImageURL: "/assets/gift-calm.jpg"},
	})
	logs := memory.NewLogStore()
	tasks := memory.NewTaskStore([]domain.Task{{ID: fixtureTaskID, Kind: "stock-refresh", ProductID: "candle-amber", Delta: 1, Status: domain.TaskPending}})
	service := &MaintenanceService{products: products, logs: logs, tasks: tasks}
	service.runner = NewTaskRunner(tasks, func(task domain.Task, worker string) error {
		if _, ok := service.products.ApplyStockChange(task.ProductID, task.Delta); !ok {
			return fmt.Errorf("product %s not found", task.ProductID)
		}
		service.logs.Append(domain.OperationLog{Action: "task.stock-refresh", ProductID: task.ProductID, TaskID: task.ID, Detail: fmt.Sprintf("worker %s applied %+d stock", worker, task.Delta)})
		return nil
	})
	return service
}

func (s *MaintenanceService) ListProducts() []domain.Product {
	return s.products.ListProducts()
}

func (s *MaintenanceService) UpdatePrice(id, rawPrice string) (domain.Product, error) {
	price, err := decimal.NewFromString(rawPrice)
	if err != nil || price.IsNegative() {
		return domain.Product{}, fmt.Errorf("price must be a non-negative decimal")
	}
	product, ok := s.products.UpdatePrice(id, domain.Product{Price: price})
	if !ok {
		return domain.Product{}, fmt.Errorf("product %s not found", id)
	}
	s.logs.Append(domain.OperationLog{Action: "price.updated", ProductID: id, Detail: price.StringFixed(2)})
	return product, nil
}

func (s *MaintenanceService) UploadImage(id, imageURL string) (domain.Product, error) {
	if imageURL == "" {
		return domain.Product{}, fmt.Errorf("image URL is required")
	}
	product, ok := s.products.UpdateImage(id, imageURL)
	if !ok {
		return domain.Product{}, fmt.Errorf("product %s not found", id)
	}
	s.logs.Append(domain.OperationLog{Action: "image.updated", ProductID: id, Detail: imageURL})
	return product, nil
}

func (s *MaintenanceService) ChangeStock(id string, delta int, reason string) (domain.Product, error) {
	product, ok := s.products.ApplyStockChange(id, delta)
	if !ok {
		return domain.Product{}, fmt.Errorf("product %s not found", id)
	}
	s.logs.Append(domain.OperationLog{Action: "stock.updated", ProductID: id, Detail: fmt.Sprintf("%+d: %s", delta, reason)})
	return product, nil
}

func (s *MaintenanceService) Logs() []domain.OperationLog {
	return s.logs.List()
}

func (s *MaintenanceService) Tasks() []domain.Task {
	result := make([]domain.Task, 0, 1)
	if task, ok := s.tasks.Get(fixtureTaskID); ok {
		result = append(result, task)
	}
	return result
}

func (s *MaintenanceService) RunTask(taskID, worker string) domain.TaskRunResult {
	return s.runner.Run(taskID, worker)
}

func (s *MaintenanceService) SetBeforeCompleteHook(hook func(string)) {
	s.runner.BeforeComplete = hook
}

func FixtureTaskID() string {
	return fixtureTaskID
}
