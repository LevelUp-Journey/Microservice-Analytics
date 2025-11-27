package controllers

import (
	"analytics/analytics/application/commandservices"
	"analytics/analytics/application/queryservices"
	"analytics/analytics/domain/model/aggregates"
	"analytics/analytics/domain/model/valueobjects"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UserRegistrationAnalyticsController maneja las peticiones REST de analytics de registros de usuarios
type UserRegistrationAnalyticsController struct {
	queryService *queryservices.UserRegistrationAnalyticsQueryService
	syncService  *commandservices.UserRegistrationSyncService
}

// NewUserRegistrationAnalyticsController crea una nueva instancia del controlador
func NewUserRegistrationAnalyticsController(
	queryService *queryservices.UserRegistrationAnalyticsQueryService,
	syncService *commandservices.UserRegistrationSyncService,
) *UserRegistrationAnalyticsController {
	return &UserRegistrationAnalyticsController{
		queryService: queryService,
		syncService:  syncService,
	}
}

// RegisterRoutes registra las rutas del controlador
func (c *UserRegistrationAnalyticsController) RegisterRoutes(router *gin.RouterGroup) {
	userReg := router.Group("/user-registration-analytics")
	{
		userReg.GET("/user/:userId", c.GetByUserID)
		userReg.GET("/email/:email", c.GetByEmail)
		userReg.GET("/provider/:provider", c.GetByProvider)
		userReg.GET("/date-range", c.GetByDateRange)
		userReg.GET("/all", c.GetAll)

		// KPIs y estadísticas
		kpi := userReg.Group("/kpi")
		{
			kpi.GET("/total-users", c.GetTotalUsers)
			kpi.GET("/provider-stats", c.GetProviderStats)
			kpi.GET("/daily-registrations", c.GetDailyRegistrations)
			kpi.GET("/top-email-domains", c.GetTopEmailDomains)
		}

		// Sincronización
		userReg.POST("/sync", c.SyncFromKafka)
	}
}

// GetByUserID obtiene analytics por ID de usuario
// @Summary Obtener analytics por ID de usuario
// @Description Obtiene el análisis de registro de un usuario específico por su ID
// @Tags User Registration Analytics
// @Accept json
// @Produce json
// @Param userId path string true "ID del usuario (UUID)"
// @Success 200 {object} map[string]interface{} "Datos del registro del usuario"
// @Failure 400 {object} ErrorResponse "Solicitud inválida"
// @Failure 404 {object} ErrorResponse "Usuario no encontrado"
// @Router /api/v1/user-registration-analytics/user/{userId} [get]
func (c *UserRegistrationAnalyticsController) GetByUserID(ctx *gin.Context) {
	userIDStr := ctx.Param("userId")

	userID, err := valueobjects.NewUserID(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	userReg, err := c.queryService.GetByUserID(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if userReg == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "User registration not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	ctx.JSON(http.StatusOK, c.toResponse(userReg))
}

// GetByEmail obtiene analytics por email
// @Summary Obtener analytics por email
// @Description Obtiene el análisis de registro de un usuario por su email
// @Tags User Registration Analytics
// @Accept json
// @Produce json
// @Param email path string true "Email del usuario"
// @Success 200 {object} map[string]interface{} "Datos del registro del usuario"
// @Failure 400 {object} ErrorResponse "Solicitud inválida"
// @Failure 404 {object} ErrorResponse "Usuario no encontrado"
// @Router /api/v1/user-registration-analytics/email/{email} [get]
func (c *UserRegistrationAnalyticsController) GetByEmail(ctx *gin.Context) {
	emailStr := ctx.Param("email")

	email, err := valueobjects.NewEmail(emailStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_email",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	userReg, err := c.queryService.GetByEmail(ctx.Request.Context(), email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if userReg == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "User registration not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	ctx.JSON(http.StatusOK, c.toResponse(userReg))
}

// GetByProvider obtiene todos los usuarios registrados con un proveedor
// @Summary Obtener usuarios por proveedor
// @Description Obtiene todos los usuarios registrados con un proveedor específico (google, facebook, etc.)
// @Tags User Registration Analytics
// @Accept json
// @Produce json
// @Param provider path string true "Proveedor (google, facebook, github, twitter, local, apple, microsoft)"
// @Param limit query int false "Límite de resultados" default(50)
// @Param offset query int false "Offset para paginación" default(0)
// @Success 200 {object} map[string]interface{} "Lista de usuarios registrados"
// @Failure 400 {object} ErrorResponse "Solicitud inválida"
// @Router /api/v1/user-registration-analytics/provider/{provider} [get]
func (c *UserRegistrationAnalyticsController) GetByProvider(ctx *gin.Context) {
	providerStr := ctx.Param("provider")
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	provider, err := valueobjects.NewProvider(providerStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_provider",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	users, err := c.queryService.GetByProvider(ctx.Request.Context(), provider, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"total":  len(users),
		"limit":  limit,
		"offset": offset,
		"data":   c.toResponseList(users),
	})
}

// GetByDateRange obtiene registros en un rango de fechas
// @Summary Obtener registros por rango de fechas
// @Description Obtiene todos los registros de usuarios en un rango de fechas específico
// @Tags User Registration Analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Fecha de inicio (RFC3339, ej: 2024-01-01T00:00:00Z)"
// @Param end_date query string true "Fecha de fin (RFC3339, ej: 2024-12-31T23:59:59Z)"
// @Param limit query int false "Límite de resultados" default(50)
// @Param offset query int false "Offset para paginación" default(0)
// @Success 200 {object} map[string]interface{} "Lista de registros en el rango de fechas"
// @Failure 400 {object} ErrorResponse "Solicitud inválida"
// @Router /api/v1/user-registration-analytics/date-range [get]
func (c *UserRegistrationAnalyticsController) GetByDateRange(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_start_date",
			Message: "start_date must be in RFC3339 format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_end_date",
			Message: "end_date must be in RFC3339 format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	users, err := c.queryService.GetByDateRange(ctx.Request.Context(), startDate, endDate, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"start_date": startDate,
		"end_date":   endDate,
		"total":      len(users),
		"limit":      limit,
		"offset":     offset,
		"data":       c.toResponseList(users),
	})
}

// GetAll obtiene todos los registros con paginación
// @Summary Obtener todos los registros de usuarios
// @Description Obtiene todos los registros de usuarios con paginación
// @Tags User Registration Analytics
// @Accept json
// @Produce json
// @Param limit query int false "Límite de resultados" default(50)
// @Param offset query int false "Offset para paginación" default(0)
// @Success 200 {object} map[string]interface{} "Lista de todos los registros"
// @Failure 500 {object} ErrorResponse "Error interno del servidor"
// @Router /api/v1/user-registration-analytics/all [get]
func (c *UserRegistrationAnalyticsController) GetAll(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	users, err := c.queryService.GetAll(ctx.Request.Context(), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"total":  len(users),
		"limit":  limit,
		"offset": offset,
		"data":   c.toResponseList(users),
	})
}

// GetTotalUsers obtiene el total de usuarios registrados
// @Summary Obtener total de usuarios registrados
// @Description Obtiene el número total de usuarios registrados en el sistema
// @Tags User Registration Analytics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Total de usuarios"
// @Failure 500 {object} ErrorResponse "Error interno del servidor"
// @Router /api/v1/user-registration-analytics/kpi/total-users [get]
func (c *UserRegistrationAnalyticsController) GetTotalUsers(ctx *gin.Context) {
	total, err := c.queryService.GetTotalUsers(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"total_users": total,
	})
}

// GetProviderStats obtiene estadísticas por proveedor
// @Summary Obtener estadísticas por proveedor
// @Description Obtiene estadísticas de usuarios registrados agrupados por proveedor (google, facebook, etc.)
// @Tags User Registration Analytics
// @Accept json
// @Produce json
// @Success 200 {array} repositories.ProviderStats "Estadísticas por proveedor"
// @Failure 500 {object} ErrorResponse "Error interno del servidor"
// @Router /api/v1/user-registration-analytics/kpi/provider-stats [get]
func (c *UserRegistrationAnalyticsController) GetProviderStats(ctx *gin.Context) {
	stats, err := c.queryService.GetProviderStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// GetDailyRegistrations obtiene estadísticas diarias de registros
// @Summary Obtener estadísticas diarias de registros
// @Description Obtiene estadísticas de registros de usuarios agrupadas por día
// @Tags User Registration Analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Fecha de inicio (RFC3339, ej: 2024-01-01T00:00:00Z)"
// @Param end_date query string true "Fecha de fin (RFC3339, ej: 2024-12-31T23:59:59Z)"
// @Success 200 {array} repositories.DailyRegistrationStats "Estadísticas diarias"
// @Failure 400 {object} ErrorResponse "Solicitud inválida"
// @Failure 500 {object} ErrorResponse "Error interno del servidor"
// @Router /api/v1/user-registration-analytics/kpi/daily-registrations [get]
func (c *UserRegistrationAnalyticsController) GetDailyRegistrations(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_start_date",
			Message: "start_date must be in RFC3339 format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_end_date",
			Message: "end_date must be in RFC3339 format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	stats, err := c.queryService.GetDailyRegistrationStats(ctx.Request.Context(), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// GetTopEmailDomains obtiene los dominios de email más usados
// @Summary Obtener los dominios de email más usados
// @Description Obtiene los dominios de email más usados en los registros de usuarios
// @Tags User Registration Analytics
// @Accept json
// @Produce json
// @Param limit query int false "Límite de resultados" default(10)
// @Success 200 {array} repositories.EmailDomainStats "Top dominios de email"
// @Failure 500 {object} ErrorResponse "Error interno del servidor"
// @Router /api/v1/user-registration-analytics/kpi/top-email-domains [get]
func (c *UserRegistrationAnalyticsController) GetTopEmailDomains(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	stats, err := c.queryService.GetTopEmailDomains(ctx.Request.Context(), limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// SyncFromKafka sincroniza todos los eventos de Kafka
// @Summary Sincronizar eventos de registro de usuarios desde Kafka
// @Description Lee todos los mensajes disponibles del tópico iam.user.registered y los guarda en la base de datos
// @Tags User Registration Analytics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Resultado de la sincronización"
// @Failure 500 {object} ErrorResponse "Error interno del servidor"
// @Router /api/v1/user-registration-analytics/sync [post]
func (c *UserRegistrationAnalyticsController) SyncFromKafka(ctx *gin.Context) {
	count, err := c.syncService.SyncAllEvents(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "sync_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "Sync completed successfully",
		"events_synced": count,
	})
}

// Helper methods

func (c *UserRegistrationAnalyticsController) toResponse(userReg interface{}) gin.H {
	// Type assertion para acceder a los métodos del aggregate
	if ur, ok := userReg.(interface {
		ID() uint
		UserID() valueobjects.UserID
		ProfileID() valueobjects.ProfileID
		Username() string
		ProfileURL() *string
		RegisteredAt() time.Time
		CreatedAt() time.Time
		UpdatedAt() time.Time
	}); ok {
		return gin.H{
			"id":            ur.ID(),
			"user_id":       ur.UserID().Value(),
			"profile_id":    ur.ProfileID().Value(),
			"username":      ur.Username(),
			"profile_url":   ur.ProfileURL(),
			"registered_at": ur.RegisteredAt(),
			"created_at":    ur.CreatedAt(),
			"updated_at":    ur.UpdatedAt(),
		}
	}
	return gin.H{}
}

func (c *UserRegistrationAnalyticsController) toResponseList(users []*aggregates.UserRegistrationAnalytics) []gin.H {
	result := make([]gin.H, 0, len(users))

	for _, user := range users {
		result = append(result, gin.H{
			"id":            user.ID(),
			"user_id":       user.UserID().Value(),
			"profile_id":         user.ProfileID().Value(),
			"username":    user.Username(),
			"profile_url":   user.ProfileURL(),
			"registered_at": user.RegisteredAt(),
			"created_at":    user.CreatedAt(),
			"updated_at":    user.UpdatedAt(),
		})
	}

	return result
}
