package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/visionprice/proveedores-backend/src/core/errors"
	"github.com/visionprice/proveedores-backend/src/core/responses"
	"github.com/visionprice/proveedores-backend/src/feature/payments/application/payment_usecase"
	"github.com/visionprice/proveedores-backend/src/feature/payments/domain/entities"
)

// maxWebhookBody caps the size of a webhook body we will read (256 KB).
const maxWebhookBody = 256 << 10

// PaymentController handles checkout creation and gateway webhooks.
type PaymentController struct {
	useCase *payment_usecase.PaymentUseCase
}

// NewPaymentController creates a new PaymentController.
func NewPaymentController(useCase *payment_usecase.PaymentUseCase) *PaymentController {
	return &PaymentController{useCase: useCase}
}

// CreateCheckout godoc
// @Summary      Iniciar checkout de suscripción
// @Description  Crea una sesión de pago recurrente en la pasarela y retorna la URL de redirección
// @Tags         Payments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      entities.CheckoutRequest  true  "Plan y pasarela"
// @Success      200   {object}  responses.APIResponse{data=entities.CheckoutOutput}
// @Failure      400   {object}  responses.APIResponse
// @Failure      401   {object}  responses.APIResponse
// @Failure      501   {object}  responses.APIResponse
// @Failure      503   {object}  responses.APIResponse
// @Router       /api/v1/subscription/checkout [post]
func (ctrl *PaymentController) CreateCheckout(c *gin.Context) {
	providerID, exists := c.Get("provider_id")
	if !exists {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Proveedor no autenticado", nil)
		return
	}

	if !ctrl.useCase.Enabled() {
		responses.ErrorResponse(c, http.StatusServiceUnavailable, "Los pagos no están habilitados", nil)
		return
	}

	var req entities.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Datos de checkout inválidos", nil)
		return
	}

	out, err := ctrl.useCase.CreateCheckout(c.Request.Context(), providerID.(string), req.PlanCode, req.Gateway, req.PaymentToken)
	if err != nil {
		handlePaymentError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Checkout creado", out)
}

// ConektaWebhook handles POST /webhooks/conekta (public, verified by signature).
func (ctrl *PaymentController) ConektaWebhook(c *gin.Context) {
	ctrl.handleWebhook(c, "conekta")
}

// PayPalWebhook handles POST /webhooks/paypal (public, verified by signature).
func (ctrl *PaymentController) PayPalWebhook(c *gin.Context) {
	ctrl.handleWebhook(c, "paypal")
}

func (ctrl *PaymentController) handleWebhook(c *gin.Context, gateway string) {
	body, err := readLimited(c, maxWebhookBody)
	if err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Cuerpo del webhook inválido", nil)
		return
	}

	if err := ctrl.useCase.HandleWebhook(c.Request.Context(), gateway, body, c.Request.Header); err != nil {
		handlePaymentError(c, err)
		return
	}

	responses.SuccessResponse(c, http.StatusOK, "Webhook procesado", nil)
}

// readLimited reads up to limit bytes of the request body.
func readLimited(c *gin.Context, limit int64) ([]byte, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	return c.GetRawData()
}

func handlePaymentError(c *gin.Context, err error) {
	var domainErr *domainErrors.DomainError
	if errors.As(err, &domainErr) {
		switch {
		case errors.Is(domainErr.Base, domainErrors.ErrNotImplemented):
			responses.ErrorResponse(c, http.StatusNotImplemented, domainErr.Message, nil)
			return
		case errors.Is(domainErr.Base, domainErrors.ErrValidation):
			responses.ErrorResponse(c, http.StatusBadRequest, domainErr.Message, nil)
			return
		case errors.Is(domainErr.Base, domainErrors.ErrNotFound):
			responses.ErrorResponse(c, http.StatusNotFound, domainErr.Message, nil)
			return
		case errors.Is(domainErr.Base, domainErrors.ErrUnauthorized):
			responses.ErrorResponse(c, http.StatusUnauthorized, domainErr.Message, nil)
			return
		}
	}
	responses.ErrorResponse(c, http.StatusInternalServerError, "Error interno del servidor", nil)
}
