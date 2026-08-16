package application

import (
	"sync"
	"testing"
)

func TestMaintenanceDeskFixtureAndChanges(t *testing.T) {
	service := NewMaintenanceService()
	products := service.ListProducts()
	if len(products) != 4 {
		t.Fatalf("got %d products", len(products))
	}
	if _, err := service.UpdatePrice("oil-lavender", "99.50"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UploadImage("oil-lavender", "/uploads/lavender.png"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ChangeStock("oil-lavender", -2, "sample sale"); err != nil {
		t.Fatal(err)
	}
	product, _ := service.products.GetProduct("oil-lavender")
	if product.Price.StringFixed(2) != "99.50" || product.Stock != 30 || product.ImageURL != "/uploads/lavender.png" {
		t.Fatalf("unexpected product: %+v", product)
	}
	if len(service.Logs()) != 3 {
		t.Fatalf("got %d logs", len(service.Logs()))
	}
}

func TestTwoWorkersProduceOneTaskOutcome(t *testing.T) {
	service := NewMaintenanceService()
	var claimed sync.WaitGroup
	claimed.Add(2)
	release := make(chan struct{})
	service.SetBeforeCompleteHook(func(string) {
		claimed.Done()
		<-release
	})
	var wait sync.WaitGroup
	results := make(chan string, 2)
	for _, worker := range []string{"worker-a", "worker-b"} {
		wait.Add(1)
		go func(worker string) {
			defer wait.Done()
			results <- string(service.RunTask(FixtureTaskID(), worker).Status)
		}(worker)
	}
	claimed.Wait()
	close(release)
	wait.Wait()
	close(results)
	statuses := make(map[string]int)
	for status := range results {
		statuses[status]++
	}
	product, _ := service.products.GetProduct("candle-amber")
	if product.Stock != 19 {
		t.Fatalf("stock changed %d times", product.Stock-18)
	}
	if len(service.Logs()) != 1 {
		t.Fatalf("got %d task logs", len(service.Logs()))
	}
	if statuses[string("executed")] != 1 || statuses[string("already-processed")] != 1 {
		t.Fatalf("unexpected statuses: %#v", statuses)
	}
}
