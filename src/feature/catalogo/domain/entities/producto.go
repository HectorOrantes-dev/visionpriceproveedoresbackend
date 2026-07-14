package entities

// Proveedor is the supplier data embedded in each product, as the gateway expects.
type Proveedor struct {
	ProveedorID string   `json:"proveedor_id"`
	Nombre      string   `json:"nombre"`
	DistanciaKm float64  `json:"distancia_km"`
	Lat         *float64 `json:"lat"`
	Lng         *float64 `json:"lng"`
}

// Producto is the catalog product shape returned to the gateway.
// JSON field names are fixed by the gateway contract; do not rename them.
type Producto struct {
	ProductoID       string    `json:"producto_id"`
	Nombre           string    `json:"nombre"`
	Categoria        string    `json:"categoria"`
	Unidad           string    `json:"unidad"`
	PrecioUnitario   float64   `json:"precio_unitario"`
	RendimientoM2    float64   `json:"rendimiento_m2"`
	PiezaLargoM      float64   `json:"pieza_largo_m"`
	PiezaAnchoM      float64   `json:"pieza_ancho_m"`
	PiezasPorPaquete int       `json:"piezas_por_paquete"`
	ImageURL         string    `json:"image_url"`
	Proveedor        Proveedor `json:"proveedor"`
}
