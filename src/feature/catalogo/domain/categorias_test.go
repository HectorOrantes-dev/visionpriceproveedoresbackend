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
			[]string{`\ypintura\y`, `\ypintar\y`, `\ycolor\y`, `\yrecubrimiento\y`, `\yesmalte\y`, `\yvinilica\y`, `\ylaca\y`, `\ybarniz\y`},
		},
		{
			"varias categorias separadas por coma sin duplicar",
			"pintura, piso",
			[]string{
				`\ypintura\y`, `\ypintar\y`, `\ycolor\y`, `\yrecubrimiento\y`, `\yesmalte\y`, `\yvinilica\y`, `\ylaca\y`, `\ybarniz\y`,
				`\ypiso\y`, `\yloseta\y`, `\ylosa\y`, `\yceramica\y`, `\yporcelanato\y`, `\yporcelanico\y`,
			},
		},
		{"categoria desconocida usa el termino tal cual", "electricidad", []string{`\yelectricidad\y`}},
		{
			"cruceta",
			"cruceta",
			[]string{`\ycruceta\y`, `\ycrucetas\y`, `\yseparador\y`, `\yespaciador\y`, `\ynivelador\y`},
		},
		{
			"pegazulejo",
			"pegazulejo",
			[]string{`\ypegazulejo\y`, `\ypega azulejo\y`, `\ypegazulejos\y`, `\yadhesivo\y`, `\ycemento cola\y`, `\ymortero\y`},
		},
		{
			"emboquillado",
			"emboquillado",
			[]string{`\yemboquillado\y`, `\yboquillado\y`, `\yboquilla\y`, `\yboquillas\y`, `\yjunta\y`, `\yfragua\y`},
		},
		{
			"azulejo ya no debe matchear dentro de pegazulejo (limite de palabra)",
			"azulejo",
			[]string{`\yazulejo\y`, `\ymosaico\y`, `\yrevestimiento\y`, `\ytalavera\y`},
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
