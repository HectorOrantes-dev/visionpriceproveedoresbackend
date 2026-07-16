package product_usecase

import (
	"testing"

	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

func TestApplyCategoryRules_Principal(t *testing.T) {
	t.Run("falta geometria de pieza", func(t *testing.T) {
		p := &entities.Product{Category: "piso", PiezasPorPaquete: 5}
		if err := applyCategoryRules(p); err == nil {
			t.Fatal("esperaba error por falta de pieza_largo_m/pieza_ancho_m")
		}
	})

	t.Run("falta piezas_por_paquete", func(t *testing.T) {
		p := &entities.Product{Category: "azulejo", PiezaLargoM: 0.25, PiezaAnchoM: 0.4}
		if err := applyCategoryRules(p); err == nil {
			t.Fatal("esperaba error por falta de piezas_por_paquete")
		}
	})

	t.Run("rendimiento_m2 vacio se autocalcula", func(t *testing.T) {
		p := &entities.Product{Category: "piso", PiezaLargoM: 0.25, PiezaAnchoM: 0.4, PiezasPorPaquete: 5}
		if err := applyCategoryRules(p); err != nil {
			t.Fatalf("no esperaba error, got %v", err)
		}
		want := 0.5
		if p.RendimientoM2 != want {
			t.Errorf("RendimientoM2 = %v, want %v", p.RendimientoM2, want)
		}
	})

	t.Run("rendimiento_m2 consistente se acepta", func(t *testing.T) {
		p := &entities.Product{Category: "piso", PiezaLargoM: 0.25, PiezaAnchoM: 0.4, PiezasPorPaquete: 5, RendimientoM2: 0.5}
		if err := applyCategoryRules(p); err != nil {
			t.Fatalf("no esperaba error, got %v", err)
		}
	})

	t.Run("rendimiento_m2 inconsistente se rechaza", func(t *testing.T) {
		p := &entities.Product{Category: "piso", PiezaLargoM: 0.25, PiezaAnchoM: 0.4, PiezasPorPaquete: 5, RendimientoM2: 3.0}
		if err := applyCategoryRules(p); err == nil {
			t.Fatal("esperaba error por rendimiento_m2 inconsistente")
		}
	})
}

func TestApplyCategoryRules_Adhesivo(t *testing.T) {
	t.Run("sin rendimiento_m2 se rechaza", func(t *testing.T) {
		p := &entities.Product{Category: "pegazulejo"}
		if err := applyCategoryRules(p); err == nil {
			t.Fatal("esperaba error por falta de rendimiento_m2")
		}
	})

	t.Run("con rendimiento_m2 se acepta sin geometria", func(t *testing.T) {
		p := &entities.Product{Category: "pegazulejo", RendimientoM2: 4}
		if err := applyCategoryRules(p); err != nil {
			t.Fatalf("no esperaba error, got %v", err)
		}
	})
}

func TestApplyCategoryRules_Boquilla(t *testing.T) {
	t.Run("sin rendimiento_m2 se rechaza", func(t *testing.T) {
		p := &entities.Product{Category: "emboquillado"}
		if err := applyCategoryRules(p); err == nil {
			t.Fatal("esperaba error por falta de rendimiento_m2")
		}
	})

	t.Run("con rendimiento_m2 se acepta", func(t *testing.T) {
		p := &entities.Product{Category: "boquilla blanca", RendimientoM2: 6}
		if err := applyCategoryRules(p); err != nil {
			t.Fatalf("no esperaba error, got %v", err)
		}
	})
}

func TestApplyCategoryRules_Cruceta(t *testing.T) {
	t.Run("sin piezas_por_paquete se rechaza", func(t *testing.T) {
		p := &entities.Product{Category: "cruceta"}
		if err := applyCategoryRules(p); err == nil {
			t.Fatal("esperaba error por falta de piezas_por_paquete")
		}
	})

	t.Run("con piezas_por_paquete se acepta sin geometria ni rendimiento", func(t *testing.T) {
		p := &entities.Product{Category: "crucetas 2mm", PiezasPorPaquete: 200}
		if err := applyCategoryRules(p); err != nil {
			t.Fatalf("no esperaba error, got %v", err)
		}
	})
}

func TestApplyCategoryRules_CategoriaSinRol(t *testing.T) {
	p := &entities.Product{Category: "pintura"}
	if err := applyCategoryRules(p); err != nil {
		t.Fatalf("categorias sin rol no deben validarse, got %v", err)
	}
}
