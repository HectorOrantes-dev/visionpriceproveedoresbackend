package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/visionprice/proveedores-backend/src/feature/catalogo/application/catalogo_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/catalogo/domain/entities"
)

// defaultRadioKm is used when radio_km is missing or not positive.
const defaultRadioKm = 10.0

// CatalogoController serves catalog queries to the gateway. Responses are raw
// JSON arrays (the gateway accepts an array directly), not the API envelope.
type CatalogoController struct {
	useCase *catalogo_usecase.CatalogoUseCase
}

// NewCatalogoController creates a new CatalogoController.
func NewCatalogoController(useCase *catalogo_usecase.CatalogoUseCase) *CatalogoController {
	return &CatalogoController{useCase: useCase}
}

// ProductosCercanos handles GET /productos/cercanos?lat=&lng=&radio_km=&categoria=
func (ctrl *CatalogoController) ProductosCercanos(c *gin.Context) {
	lat, errLat := strconv.ParseFloat(c.Query("lat"), 64)
	lng, errLng := strconv.ParseFloat(c.Query("lng"), 64)
	if errLat != nil || errLng != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros 'lat' y 'lng' son obligatorios y numéricos"})
		return
	}

	radioKm, err := strconv.ParseFloat(c.Query("radio_km"), 64)
	if err != nil || radioKm <= 0 {
		radioKm = defaultRadioKm
	}

	categoria := strings.TrimSpace(c.Query("categoria"))

	productos, err := ctrl.useCase.ProductosCercanos(c.Request.Context(), lat, lng, radioKm, categoria)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar productos cercanos"})
		return
	}
	c.JSON(http.StatusOK, productos)
}

// ProductosPorIDs handles GET /productos?ids=12,30
func (ctrl *CatalogoController) ProductosPorIDs(c *gin.Context) {
	idsParam := strings.TrimSpace(c.Query("ids"))
	if idsParam == "" {
		c.JSON(http.StatusOK, []entities.Producto{})
		return
	}

	parts := strings.Split(idsParam, ",")
	ids := make([]string, 0, len(parts))
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		ids = append(ids, s)
	}

	productos, err := ctrl.useCase.ProductosPorIDs(c.Request.Context(), ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar productos"})
		return
	}
	c.JSON(http.StatusOK, productos)
}
