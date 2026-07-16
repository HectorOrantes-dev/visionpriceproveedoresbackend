package domain

import (
	"reflect"
	"testing"
)

func TestExpandCategorias(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{"vacio sin filtro", "", nil},
		{"solo espacios sin filtro", "   ", nil},
		{
			"pintura con acentos y mayusculas",
			"Pintura",
			[]string{"%pintura%", "%pintar%", "%color%", "%recubrimiento%", "%esmalte%", "%vinilica%", "%laca%", "%barniz%"},
		},
		{
			"varias categorias separadas por coma sin duplicar",
			"pintura, piso",
			[]string{
				"%pintura%", "%pintar%", "%color%", "%recubrimiento%", "%esmalte%", "%vinilica%", "%laca%", "%barniz%",
				"%piso%", "%loseta%", "%losa%", "%ceramica%", "%porcelanato%", "%porcelanico%",
			},
		},
		{"categoria desconocida usa el termino tal cual", "electricidad", []string{"%electricidad%"}},
		{
			"cruceta",
			"cruceta",
			[]string{"%cruceta%", "%crucetas%", "%separador%", "%espaciador%", "%nivelador%"},
		},
		{
			"pegazulejo",
			"pegazulejo",
			[]string{"%pegazulejo%", "%pega azulejo%", "%pegazulejos%", "%adhesivo%", "%cemento cola%", "%mortero%"},
		},
		{
			"boquillado",
			"boquillado",
			[]string{"%boquillado%", "%boquilla%", "%boquillas%", "%junta%", "%fragua%"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ExpandCategorias(c.raw)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("ExpandCategorias(%q) = %v, want %v", c.raw, got, c.want)
			}
		})
	}
}
