package domain

import "testing"

func TestRolMaterial(t *testing.T) {
	cases := []struct {
		categoria string
		want      Rol
	}{
		{"", ""},
		{"Piso", RolPrincipal},
		{"Loseta 60x60", RolPrincipal},
		{"Azulejo", RolPrincipal},
		{"Zoclo", RolPrincipal},
		{"Cruceta", RolCruceta},
		{"Crucetas 2mm", RolCruceta},
		{"Pegazulejo", RolAdhesivo},
		{"Pega azulejo gris", RolAdhesivo},
		{"Emboquillado", RolBoquilla},
		{"Boquilla blanca", RolBoquilla},
		{"Pintura", ""},
		{"Electricidad", ""},
	}

	for _, c := range cases {
		t.Run(c.categoria, func(t *testing.T) {
			got := RolMaterial(c.categoria)
			if got != c.want {
				t.Errorf("RolMaterial(%q) = %q, want %q", c.categoria, got, c.want)
			}
		})
	}
}
