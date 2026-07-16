package product_usecase

import (
	"fmt"
	"math"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	catalogoDomain "github.com/visionprice/proveedores-backend/src/feature/catalogo/domain"
	"github.com/visionprice/proveedores-backend/src/feature/products/domain/entities"
)

// rendimientoTolerancia permite un margen de redondeo entre rendimiento_m2 y
// piezas_por_paquete × pieza_largo_m × pieza_ancho_m antes de rechazar el dato
// como inconsistente.
const rendimientoTolerancia = 0.02 // 2%

// applyCategoryRules valida (y completa cuando falta) los campos de cálculo
// según el rol del material — ver RolMaterial. Esto evita que el motor de
// kits/crucetas del backend principal reciba productos sin los datos que
// necesita para calcular piezas, cajas o sacos, o con rendimiento_m2
// inconsistente con la geometría de la pieza.
//
// Se aplica sobre el estado final del producto (tras mezclar los cambios), no
// solo sobre los campos que llegaron en el request, para que tanto crear como
// editar un producto queden con datos coherentes.
func applyCategoryRules(p *entities.Product) error {
	switch catalogoDomain.RolMaterial(p.Category) {

	case catalogoDomain.RolPrincipal:
		if p.PiezaLargoM <= 0 || p.PiezaAnchoM <= 0 {
			return domainErrors.NewDomainError(domainErrors.ErrValidation,
				"Para materiales de piso/azulejo/zoclo, pieza_largo_m y pieza_ancho_m son obligatorios y deben ser mayores a 0 (se usan para calcular piezas y crucetas)")
		}
		if p.PiezasPorPaquete <= 0 {
			return domainErrors.NewDomainError(domainErrors.ErrValidation,
				"Para materiales de piso/azulejo/zoclo, piezas_por_paquete es obligatorio y debe ser mayor a 0 (agrupa piezas en cajas)")
		}

		esperado := float64(p.PiezasPorPaquete) * p.PiezaLargoM * p.PiezaAnchoM
		if p.RendimientoM2 == 0 {
			// No vino: se calcula a partir de la geometría de la pieza.
			p.RendimientoM2 = esperado
		} else if math.Abs(p.RendimientoM2-esperado) > esperado*rendimientoTolerancia {
			return domainErrors.NewDomainError(domainErrors.ErrValidation,
				fmt.Sprintf(
					"rendimiento_m2 (%.4f) no coincide con piezas_por_paquete × pieza_largo_m × pieza_ancho_m (%.4f). Corrígelo o quítalo para que se calcule automáticamente.",
					p.RendimientoM2, esperado))
		}

	case catalogoDomain.RolAdhesivo:
		if p.RendimientoM2 <= 0 {
			return domainErrors.NewDomainError(domainErrors.ErrValidation,
				"Para adhesivo (pegazulejo), rendimiento_m2 (m² que cubre 1 saco) es obligatorio y debe ser mayor a 0")
		}

	case catalogoDomain.RolBoquilla:
		if p.RendimientoM2 <= 0 {
			return domainErrors.NewDomainError(domainErrors.ErrValidation,
				"Para boquilla/emboquillado, rendimiento_m2 (m² que cubre 1 saco) es obligatorio y debe ser mayor a 0")
		}

	case catalogoDomain.RolCruceta:
		if p.PiezasPorPaquete <= 0 {
			return domainErrors.NewDomainError(domainErrors.ErrValidation,
				"Para crucetas, piezas_por_paquete (unidades por bolsa/paquete) es obligatorio y debe ser mayor a 0")
		}
	}

	return nil
}
