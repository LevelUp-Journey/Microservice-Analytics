package controllers

import (
	"github.com/nanab/analytics-service/analytics/application/commandservices"
	"github.com/nanab/analytics-service/analytics/application/queryservices"
	"github.com/nanab/analytics-service/analytics/domain/model/aggregates"
	"github.com/nanab/analytics-service/analytics/domain/model/valueobjects"
	"net/http"

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

		// KPIs y estadísticas
		kpi := userReg.Group("/kpi")
		{
			kpi.GET("/total-users", c.GetTotalUsers)
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

// SyncFromKafka sincroniza todos los eventos de Kafka
// @Summary Sincronizar eventos de registro de usuarios desde Kafka
// @Description Lee todos los mensajes disponibles del tópico community.registration y los guarda en la base de datos
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

func (c *UserRegistrationAnalyticsController) toResponse(userReg *aggregates.UserRegistrationAnalytics) gin.H {
	return gin.H{
		"id":            userReg.ID(),
		"user_id":       userReg.UserID().Value(),
		"profile_id":    userReg.ProfileID().Value(),
		"username":      userReg.Username(),
		"profile_url":   userReg.ProfileURL(),
		"registered_at": userReg.RegisteredAt(),
		"created_at":    userReg.CreatedAt(),
		"updated_at":    userReg.UpdatedAt(),
	}
}
