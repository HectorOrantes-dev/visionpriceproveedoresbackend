package domain

import "strings"

// Rol clasifica la categoría de un producto según cómo el motor de kits/crucetas
// del backend principal calcula cantidades para ese material.
type Rol string

const (
	// RolPrincipal agrupa piso/azulejo/zoclo: se vende por caja y el cálculo de
	// piezas/crucetas depende de la geometría de la pieza.
	RolPrincipal Rol = "principal"
	// RolAdhesivo es pegazulejo: se vende por saco, el cálculo solo usa rendimiento_m2.
	RolAdhesivo Rol = "adhesivo"
	// RolCruceta es separadores: se vende por paquete/bolsa, el cálculo solo usa piezas_por_paquete.
	RolCruceta Rol = "cruceta"
	// RolBoquilla es emboquillado: se vende por saco, el cálculo solo usa rendimiento_m2.
	RolBoquilla Rol = "boquilla"
)

// rolesOrdenados mapea cada rol a sus categorías canónicas (ver sinonimos en
// categorias.go). El orden importa: se revisan primero las categorías
// compuestas más específicas (cruceta, pegazulejo, emboquillado) porque
// algunos de sus sinónimos contienen como subcadena palabras de la categoría
// "azulejo" (p. ej. "pegazulejo" contiene "azulejo"); si "principal" se
// revisara primero, "Pegazulejo" clasificaría mal como piso/azulejo.
var rolesOrdenados = []struct {
	rol     Rol
	canones []string
}{
	{RolCruceta, []string{"cruceta"}},
	{RolAdhesivo, []string{"pegazulejo"}},
	{RolBoquilla, []string{"emboquillado"}},
	{RolPrincipal, []string{"piso", "azulejo", "zoclo"}},
}

// RolMaterial clasifica la categoría cruda de un producto (tal como la escribe
// el proveedor, ej. "Loseta 60x60" o "Pega azulejo gris") en un rol de
// cálculo, reutilizando los mismos sinónimos que ExpandCategorias usa para
// la búsqueda. Devuelve "" si la categoría no corresponde a ningún rol con
// reglas especiales (ej. pintura, impermeabilizante, u otra categoría libre).
func RolMaterial(categoria string) Rol {
	cat := normalizar(categoria)
	if cat == "" {
		return ""
	}

	// Coincidencia exacta con la categoría canónica.
	for _, grupo := range rolesOrdenados {
		for _, canon := range grupo.canones {
			if cat == canon {
				return grupo.rol
			}
		}
	}

	// Coincidencia por sinónimo contenido en el texto libre, en el orden de
	// especificidad definido en rolesOrdenados.
	for _, grupo := range rolesOrdenados {
		for _, canon := range grupo.canones {
			for _, syn := range sinonimos[canon] {
				if strings.Contains(cat, syn) {
					return grupo.rol
				}
			}
		}
	}

	return ""
}
