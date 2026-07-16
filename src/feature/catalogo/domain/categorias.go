package domain

import "strings"

// Lógica de negocio: identificar qué proveedores tienen el material del item.
//
// La app manda una o varias "categorías canónicas" (ej. "pintura", "piso")
// derivadas del item que se está cotizando (ej. "cambio de pintura" → "pintura").
// Aquí las expandimos a los términos con los que un proveedor pudo haber
// nombrado su categoría en su catálogo ("Pintura", "colores", "recubrimiento"…),
// para que el filtro alcance a todos los proveedores del material sin exigir que
// escriban exactamente la misma palabra.
//
// El resultado son patrones para comparar con ILIKE ('%pintura%'), de modo que
// una categoría del proveedor como "Pintura vinílica" o "Colores para muro"
// también coincida.

// sinonimos mapea cada categoría canónica a las palabras clave equivalentes.
// La clave y los valores van en minúsculas y sin acentos (ver normalizar).
var sinonimos = map[string][]string{
	"pintura":           {"pintura", "pintar", "color", "recubrimiento", "esmalte", "vinilica", "laca", "barniz"},
	"impermeabilizante": {"impermeabilizante", "impermeabilizar", "impermeabilizacion", "sellador"},
	"piso":              {"piso", "loseta", "losa", "ceramica", "porcelanato", "porcelanico"},
	"azulejo":           {"azulejo", "mosaico", "revestimiento", "talavera"},
	"zoclo":             {"zoclo", "zocalo", "rodapie"},
	"cruceta":           {"cruceta", "crucetas", "separador", "espaciador", "nivelador"},
	"pegazulejo":        {"pegazulejo", "pega azulejo", "pegazulejos", "adhesivo", "cemento cola", "mortero"},
	"emboquillado":      {"emboquillado", "boquillado", "boquilla", "boquillas", "junta", "fragua"},
}

// ExpandCategorias toma el valor crudo del parámetro `categoria` (una o varias
// categorías separadas por coma) y devuelve los patrones ILIKE con los que se
// filtra el catálogo. Devuelve nil (sin filtro) cuando la entrada está vacía,
// para conservar el comportamiento "sin categoría = todos los productos".
func ExpandCategorias(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	vistos := map[string]struct{}{}
	patrones := []string{}
	add := func(termino string) {
		p := "%" + termino + "%"
		if _, ok := vistos[p]; ok {
			return
		}
		vistos[p] = struct{}{}
		patrones = append(patrones, p)
	}

	for _, parte := range strings.Split(raw, ",") {
		cat := normalizar(parte)
		if cat == "" {
			continue
		}
		if syns, ok := sinonimos[cat]; ok {
			for _, s := range syns {
				add(s)
			}
		} else {
			// Categoría no catalogada: filtramos por el término tal cual llegó.
			add(cat)
		}
	}

	if len(patrones) == 0 {
		return nil
	}
	return patrones
}

// normalizar deja el término en minúsculas, sin espacios sobrantes ni acentos,
// para que "Pintura", "pintura" y "PINTURÁ" resuelvan a la misma clave.
func normalizar(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	reemplazos := map[rune]rune{
		'á': 'a', 'é': 'e', 'í': 'i', 'ó': 'o', 'ú': 'u', 'ü': 'u', 'ñ': 'n',
	}
	var b strings.Builder
	for _, r := range s {
		if rep, ok := reemplazos[r]; ok {
			b.WriteRune(rep)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
