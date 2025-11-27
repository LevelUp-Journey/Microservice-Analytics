package controllers

import (
	"analytics/analytics/application/queryservices"
	"analytics/analytics/domain/repositories"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsController maneja las peticiones REST de analytics
type AnalyticsController struct {
	queryService *queryservices.ExecutionAnalyticsQueryService
}

// NewAnalyticsController crea una nueva instancia del controlador
func NewAnalyticsController(queryService *queryservices.ExecutionAnalyticsQueryService) *AnalyticsController {
	return &AnalyticsController{
		queryService: queryService,
	}
}

// RegisterRoutes registra las rutas del controlador
func (c *AnalyticsController) RegisterRoutes(router *gin.RouterGroup) {
	analytics := router.Group("/analytics")
	{
		analytics.GET("/execution/:executionId", c.GetByExecutionID)
		analytics.GET("/student/:studentId", c.GetByStudentID)
		analytics.GET("/challenge/:challengeId", c.GetByChallengeID)
		analytics.GET("/date-range", c.GetByDateRange)

		kpi := analytics.Group("/kpi")
		{
			kpi.GET("/student/:studentId", c.GetStudentKPI)
			kpi.GET("/challenge/:challengeId", c.GetChallengeKPI)
			kpi.GET("/daily", c.GetDailyKPI)
			kpi.GET("/languages", c.GetLanguageKPI)
			kpi.GET("/top-failed-challenges", c.GetTopFailedChallenges)
		}
	}
}

// ErrorResponse representa una respuesta de error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// GetByExecutionID obtiene analytics por ID de ejecución - SIN DTOs, retorna aggregate directo
// @Summary Obtener analytics por ID de ejecución
// @Description Obtiene el análisis completo de una ejecución específica por su ID
// @Tags Analytics
// @Accept json
// @Produce json
// @Param executionId path string true "ID de la ejecución"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/analytics/execution/{executionId} [get]
func (c *AnalyticsController) GetByExecutionID(ctx *gin.Context) {
	executionID := ctx.Param("executionId")

	execution, err := c.queryService.GetByExecutionID(ctx.Request.Context(), executionID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if execution == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Execution analytics not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Transformación inline - NO DTO
	ctx.JSON(http.StatusOK, gin.H{
		"id":                execution.ID(),
		"execution_id":      execution.ExecutionID().Value(),
		"challenge_id":      execution.ChallengeID().Value(),
		"code_version_id":   execution.CodeVersionID(),
		"student_id":        execution.StudentID().Value(),
		"language":          execution.Language().Value(),
		"status":            execution.Status().Value(),
		"timestamp":         execution.Timestamp(),
		"execution_time_ms": execution.ExecutionTimeMs(),
		"exit_code":         execution.ExitCode(),
		"total_tests":       execution.TotalTests(),
		"passed_tests":      execution.PassedTests(),
		"failed_tests":      execution.FailedTests(),
		"success":           execution.Success(),
		"success_rate":      execution.CalculateSuccessRate(),
		"server_instance":   execution.ServerInstance(),
		"test_results": func() []gin.H {
			results := make([]gin.H, 0, len(execution.TestResults()))
			for _, tr := range execution.TestResults() {
				results = append(results, gin.H{
					"test_id":       tr.TestID().Value(),
					"test_name":     tr.TestName(),
					"passed":        tr.Passed(),
					"error_message": tr.ErrorMessage(),
				})
			}
			return results
		}(),
		"created_at": execution.CreatedAt(),
	})
}

// GetByStudentID obtiene todas las ejecuciones de un estudiante
// @Summary Obtener analytics por ID de estudiante
// @Description Obtiene todos los análisis de ejecuciones de un estudiante específico
// @Tags Analytics
// @Accept json
// @Produce json
// @Param studentId path string true "ID del estudiante"
// @Param page query int false "Número de página" default(1)
// @Param pageSize query int false "Tamaño de página" default(20)
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/analytics/student/{studentId} [get]
func (c *AnalyticsController) GetByStudentID(ctx *gin.Context) {
	studentID := ctx.Param("studentId")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	executions, err := c.queryService.GetByStudentID(ctx.Request.Context(), studentID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	data := make([]gin.H, 0, len(executions))
	for _, exec := range executions {
		data = append(data, gin.H{
			"id":                exec.ID(),
			"execution_id":      exec.ExecutionID().Value(),
			"challenge_id":      exec.ChallengeID().Value(),
			"student_id":        exec.StudentID().Value(),
			"language":          exec.Language().Value(),
			"status":            exec.Status().Value(),
			"timestamp":         exec.Timestamp(),
			"execution_time_ms": exec.ExecutionTimeMs(),
			"success":           exec.Success(),
			"success_rate":      exec.CalculateSuccessRate(),
			"passed_tests":      exec.PassedTests(),
			"total_tests":       exec.TotalTests(),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":      data,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetByChallengeID obtiene todas las ejecuciones de un challenge
// @Summary Obtener analytics por ID de challenge
// @Description Obtiene todos los análisis de ejecuciones de un challenge específico
// @Tags Analytics
// @Accept json
// @Produce json
// @Param challengeId path string true "ID del challenge"
// @Param page query int false "Número de página" default(1)
// @Param pageSize query int false "Tamaño de página" default(20)
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/analytics/challenge/{challengeId} [get]
func (c *AnalyticsController) GetByChallengeID(ctx *gin.Context) {
	challengeID := ctx.Param("challengeId")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	executions, err := c.queryService.GetByChallengeID(ctx.Request.Context(), challengeID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	data := make([]gin.H, 0, len(executions))
	for _, exec := range executions {
		data = append(data, gin.H{
			"id":                exec.ID(),
			"execution_id":      exec.ExecutionID().Value(),
			"challenge_id":      exec.ChallengeID().Value(),
			"student_id":        exec.StudentID().Value(),
			"language":          exec.Language().Value(),
			"status":            exec.Status().Value(),
			"timestamp":         exec.Timestamp(),
			"execution_time_ms": exec.ExecutionTimeMs(),
			"success":           exec.Success(),
			"success_rate":      exec.CalculateSuccessRate(),
			"passed_tests":      exec.PassedTests(),
			"total_tests":       exec.TotalTests(),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":      data,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetByDateRange obtiene ejecuciones en un rango de fechas
// @Summary Obtener analytics por rango de fechas
// @Description Obtiene todos los análisis de ejecuciones en un rango de fechas específico
// @Tags Analytics
// @Accept json
// @Produce json
// @Param startDate query string true "Fecha de inicio (RFC3339)"
// @Param endDate query string true "Fecha de fin (RFC3339)"
// @Param page query int false "Número de página" default(1)
// @Param pageSize query int false "Tamaño de página" default(20)
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/analytics/date-range [get]
func (c *AnalyticsController) GetByDateRange(ctx *gin.Context) {
	startDateStr := ctx.Query("startDate")
	endDateStr := ctx.Query("endDate")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid start date format. Use RFC3339",
			Code:    http.StatusBadRequest,
		})
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_date",
			Message: "Invalid end date format. Use RFC3339",
			Code:    http.StatusBadRequest,
		})
		return
	}

	executions, err := c.queryService.GetByDateRange(ctx.Request.Context(), startDate, endDate, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	data := make([]gin.H, 0, len(executions))
	for _, exec := range executions {
		data = append(data, gin.H{
			"id":                exec.ID(),
			"execution_id":      exec.ExecutionID().Value(),
			"challenge_id":      exec.ChallengeID().Value(),
			"student_id":        exec.StudentID().Value(),
			"language":          exec.Language().Value(),
			"timestamp":         exec.Timestamp(),
			"execution_time_ms": exec.ExecutionTimeMs(),
			"success":           exec.Success(),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":      data,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetStudentKPI obtiene KPIs de un estudiante
// @Summary Obtener KPIs de un estudiante
// @Description Obtiene las métricas clave de rendimiento de un estudiante específico
// @Tags KPI
// @Accept json
// @Produce json
// @Param studentId path string true "ID del estudiante"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/analytics/kpi/student/{studentId} [get]
func (c *AnalyticsController) GetStudentKPI(ctx *gin.Context) {
	studentID := ctx.Param("studentId")

	count, err := c.queryService.GetStudentExecutionCount(ctx.Request.Context(), studentID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	successRate, err := c.queryService.GetStudentSuccessRate(ctx.Request.Context(), studentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"student_id":       studentID,
		"total_executions": count,
		"success_rate":     successRate,
	})
}

// GetChallengeKPI obtiene KPIs de un challenge
// @Summary Obtener KPIs de un challenge
// @Description Obtiene las métricas clave de rendimiento de un challenge específico
// @Tags KPI
// @Accept json
// @Produce json
// @Param challengeId path string true "ID del challenge"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/analytics/kpi/challenge/{challengeId} [get]
func (c *AnalyticsController) GetChallengeKPI(ctx *gin.Context) {
	challengeID := ctx.Param("challengeId")

	count, err := c.queryService.GetChallengeExecutionCount(ctx.Request.Context(), challengeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	successRate, err := c.queryService.GetChallengeSuccessRate(ctx.Request.Context(), challengeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	avgTime, err := c.queryService.GetChallengeAverageExecutionTime(ctx.Request.Context(), challengeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"challenge_id":          challengeID,
		"total_executions":      count,
		"success_rate":          successRate,
		"avg_execution_time_ms": avgTime,
	})
}

// GetDailyKPI obtiene estadísticas diarias
// @Summary Obtener KPIs diarios
// @Description Obtiene las métricas agregadas por día
// @Tags KPI
// @Accept json
// @Produce json
// @Param limit query int false "Límite de resultados" default(30)
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/analytics/kpi/daily [get]
func (c *AnalyticsController) GetDailyKPI(ctx *gin.Context) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	if startDateStr := ctx.Query("startDate"); startDateStr != "" {
		var err error
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid start date format. Use RFC3339",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}

	if endDateStr := ctx.Query("endDate"); endDateStr != "" {
		var err error
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid end date format. Use RFC3339",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}

	stats, err := c.queryService.GetDailyStats(ctx.Request.Context(), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, transformDailyStats(stats))
}

// GetLanguageKPI obtiene estadísticas por lenguaje
// @Summary Obtener KPIs por lenguaje de programación
// @Description Obtiene las métricas agregadas por lenguaje de programación
// @Tags KPI
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/analytics/kpi/languages [get]
func (c *AnalyticsController) GetLanguageKPI(ctx *gin.Context) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := ctx.Query("startDate"); startDateStr != "" {
		var err error
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid start date format. Use RFC3339",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}

	if endDateStr := ctx.Query("endDate"); endDateStr != "" {
		var err error
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid end date format. Use RFC3339",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}

	stats, err := c.queryService.GetLanguageStats(ctx.Request.Context(), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	responses := make([]gin.H, 0, len(stats))
	for _, stat := range stats {
		responses = append(responses, gin.H{
			"language":         stat.Language,
			"total_executions": stat.TotalExecutions,
			"success_rate":     stat.SuccessRate,
		})
	}

	ctx.JSON(http.StatusOK, responses)
}

// GetTopFailedChallenges obtiene los challenges con más fallos
// @Summary Obtener top challenges con más fallos
// @Description Obtiene los challenges que tienen mayor cantidad de ejecuciones fallidas
// @Tags KPI
// @Accept json
// @Produce json
// @Param limit query int false "Límite de resultados" default(10)
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/analytics/kpi/top-failed-challenges [get]
func (c *AnalyticsController) GetTopFailedChallenges(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	stats, err := c.queryService.GetTopFailedChallenges(ctx.Request.Context(), limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	responses := make([]gin.H, 0, len(stats))
	for _, stat := range stats {
		responses = append(responses, gin.H{
			"challenge_id":          stat.ChallengeID,
			"total_executions":      stat.TotalExecutions,
			"success_rate":          stat.SuccessRate,
			"avg_execution_time_ms": stat.AvgExecTime,
		})
	}

	ctx.JSON(http.StatusOK, responses)
}

// Función auxiliar inline para transformar DailyStats - NO mapper class
func transformDailyStats(stats []repositories.DailyStats) []gin.H {
	responses := make([]gin.H, 0, len(stats))
	for _, stat := range stats {
		successRate := float64(0)
		if stat.TotalExecutions > 0 {
			successRate = (float64(stat.SuccessfulExecs) / float64(stat.TotalExecutions)) * 100.0
		}

		responses = append(responses, gin.H{
			"date":                  stat.Date.Format("2006-01-02"),
			"total_executions":      stat.TotalExecutions,
			"successful_executions": stat.SuccessfulExecs,
			"failed_executions":     stat.FailedExecs,
			"success_rate":          successRate,
			"avg_execution_time_ms": stat.AvgExecTime,
		})
	}
	return responses
}
