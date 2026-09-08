package localseed

import (
	"context"
	"testing"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/repository/memory"
	"github.com/adell/cloudops-release-intelligence/internal/service"
)

func TestCatalogSeesSeededGraph(t *testing.T) {
	store := memory.New()
	if err := Load(context.Background(), store, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	detail, err := service.NewCatalog(store).GetService(context.Background(), "checkout-api")
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.DependsOn) < 3 {
		t.Fatalf("checkout should depend on multiple services, got %d", len(detail.DependsOn))
	}
}
