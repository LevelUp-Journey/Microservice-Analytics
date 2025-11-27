package controllers

import (
	"github.com/nanab/analytics-service/analytics/application/commandservices"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SyncController maneja las peticiones de sincronización
type SyncController struct {
	syncService *commandservices.SyncService
}

// NewSyncController crea una nueva instancia del controlador
func NewSyncController(syncService *commandservices.SyncService) *SyncController {
	return &SyncController{
		syncService: syncService,
	}
}

// RegisterRoutes registra las rutas del controlador
func (c *SyncController) RegisterRoutes(router *gin.RouterGroup) {
	sync := router.Group("/sync")
	{
		sync.POST("/events", c.SyncEvents)
	}
}

// SyncEvents sincroniza todos los eventos del tópico de Kafka
// @Summary Sincronizar eventos de Kafka
// @Description Obtiene todos los eventos del tópico execution.analytics de Kafka y los guarda en la base de datos
// @Tags Sync
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/sync/events [post]
func (c *SyncController) SyncEvents(ctx *gin.Context) {
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
		"message":      "Sync completed successfully",
		"synced_count": count,
		"status":       "success",
	})
}
