package product_usecase

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// topProductsLimit caps how many materials the "top" list returns.
const topProductsLimit = 5

// GetMetricsSummary builds the provider's catalog/inventory metrics from the
// products table (there is no sales data in the system).
func (uc *ProductUseCase) GetMetricsSummary(ctx context.Context, providerID string) (*entities.MetricsSummary, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	inventoryValue, unitsInStock, averagePrice, totalMaterials, err := uc.repo.MetricsAggregate(ctx, pid)
	if err != nil {
		slog.Error("metrics: failed to aggregate", "error", err, "provider_id", providerID)
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al calcular métricas")
	}

	distribution, err := uc.repo.CategoryDistribution(ctx, pid)
	if err != nil {
		slog.Error("metrics: failed category distribution", "error", err, "provider_id", providerID)
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al calcular métricas")
	}
	// Share = percentage of total inventory value.
	for i := range distribution {
		distribution[i].Share = percentage(distribution[i].Value, inventoryValue)
	}

	monthly, err := uc.repo.MonthlyNewMaterials(ctx, pid)
	if err != nil {
		slog.Error("metrics: failed monthly materials", "error", err, "provider_id", providerID)
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al calcular métricas")
	}

	return &entities.MetricsSummary{
		InventoryValue: inventoryValue,
		UnitsInStock:   unitsInStock,
		AveragePrice:   averagePrice,
		TotalMaterials: totalMaterials,
		MonthlyNew:     monthly,
		Distribution:   distribution,
	}, nil
}

// GetTopProducts returns the provider's top materials by inventory value.
func (uc *ProductUseCase) GetTopProducts(ctx context.Context, providerID string) ([]entities.TopProduct, error) {
	pid, err := uuid.Parse(providerID)
	if err != nil {
		return nil, domainErrors.NewDomainError(domainErrors.ErrValidation, "ID de proveedor inválido")
	}

	inventoryValue, _, _, _, err := uc.repo.MetricsAggregate(ctx, pid)
	if err != nil {
		slog.Error("metrics: failed aggregate for top", "error", err, "provider_id", providerID)
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener top de materiales")
	}

	top, err := uc.repo.TopByInventoryValue(ctx, pid, topProductsLimit)
	if err != nil {
		slog.Error("metrics: failed top products", "error", err, "provider_id", providerID)
		return nil, domainErrors.NewDomainError(domainErrors.ErrInternal, "Error al obtener top de materiales")
	}
	for i := range top {
		top[i].Share = percentage(top[i].Amount, inventoryValue)
	}

	return top, nil
}

// percentage returns value/total*100, guarding against division by zero.
func percentage(value, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return value / total * 100
}
