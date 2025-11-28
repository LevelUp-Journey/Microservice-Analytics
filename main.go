package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nanab/analytics-service/analytics/application/commandservices"
	"github.com/nanab/analytics-service/analytics/application/queryservices"
	"github.com/nanab/analytics-service/analytics/infrastructure/config"
	"github.com/nanab/analytics-service/analytics/infrastructure/messaging/kafka"
	"github.com/nanab/analytics-service/analytics/infrastructure/persistence/postgres/repositories"
	"github.com/nanab/analytics-service/analytics/interfaces/rest/controllers"
	_ "github.com/nanab/analytics-service/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Analytics Microservice API
// @version 1.0
// @description Microservicio de análisis de ejecuciones de código con arquitectura DDD
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 100.102.208.55:8291
// @BasePath /
// @schemes http https
func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Conectar a la base de datos
	db, err := config.NewDatabase(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Crear repositorios
	executionRepository := repositories.NewPostgresExecutionAnalyticsRepository(db)
	userRegistrationRepository := repositories.NewPostgresUserRegistrationAnalyticsRepository(db)

	// Crear servicios de ejecución de código
	executionCommandService := commandservices.NewExecutionAnalyticsCommandService(executionRepository)
	executionQueryService := queryservices.NewExecutionAnalyticsQueryService(executionRepository)
	executionSyncService := commandservices.NewSyncService(
		cfg.Kafka.BootstrapServers,
		cfg.Kafka.Topic,
		executionRepository,
	)

	// Crear servicios de registro de usuarios
	userRegistrationCommandService := commandservices.NewUserRegistrationAnalyticsCommandService(userRegistrationRepository)
	userRegistrationQueryService := queryservices.NewUserRegistrationAnalyticsQueryService(userRegistrationRepository)
	userRegistrationSyncService := commandservices.NewUserRegistrationSyncService(
		cfg.Kafka.BootstrapServers,
		cfg.KafkaUserRegistration.Topic,
		userRegistrationRepository,
	)

	log.Println("Services initialized successfully")

	// Configurar Gin
	router := gin.Default()

	// Middleware CORS - Configuración permisiva para desarrollo
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Root endpoint - redirecciona a Swagger
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": cfg.ServiceDiscovery.ServiceName,
			"version": "1.0.0",
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Registrar rutas de API
	apiV1 := router.Group("/api/v1")

	// Controladores de ejecución de código
	analyticsController := controllers.NewAnalyticsController(executionQueryService)
	analyticsController.RegisterRoutes(apiV1)

	syncController := controllers.NewSyncController(executionSyncService)
	syncController.RegisterRoutes(apiV1)

	// Controladores de registro de usuarios
	userRegistrationController := controllers.NewUserRegistrationAnalyticsController(
		userRegistrationQueryService,
		userRegistrationSyncService,
	)
	userRegistrationController.RegisterRoutes(apiV1)

	// Iniciar servidor HTTP en goroutine
	srv := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: router,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", cfg.GetServerAddress())
		log.Printf("Swagger UI available at: http://%s/swagger/index.html", cfg.GetServerAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Registrar con Eureka si está habilitado
	var eurekaClient *config.EurekaClient
	if cfg.ServiceDiscovery.Enabled {
		// Usar EUREKA_INSTANCE_IP si está configurado, sino hostname, sino detectar automáticamente
		instanceIP := cfg.ServiceDiscovery.InstanceIP
		if instanceIP == "" && cfg.Server.Hostname != "" {
			instanceIP = cfg.Server.Hostname
		}
		// Si instanceIP sigue vacío o es 0.0.0.0, NewEurekaClient detectará automáticamente

		eurekaClient = config.NewEurekaClient(
			cfg.ServiceDiscovery.URL,
			cfg.ServiceDiscovery.ServiceName,
			instanceIP,
			cfg.Server.Port,
		)

		if err := eurekaClient.Register(); err != nil {
			log.Printf("Warning: Failed to register with Eureka: %v", err)
		} else {
			eurekaClient.StartHeartbeat()
		}
	}

	// Configurar consumidor de Kafka para ejecuciones de código con Azure Event Hub
	executionConsumerConfig := &kafka.ConsumerConfig{
		Brokers:          cfg.Kafka.BootstrapServers,
		GroupID:          cfg.Kafka.GroupID,
		Topic:            cfg.Kafka.Topic,
		SecurityProtocol: cfg.Kafka.SecurityProtocol,
		SaslMechanism:    cfg.Kafka.SaslMechanism,
		SaslUsername:     cfg.Kafka.SaslUsername,
		SaslPassword:     cfg.Kafka.SaslPassword,
		RequestTimeoutMs: cfg.Kafka.RequestTimeoutMs,
		SessionTimeoutMs: cfg.Kafka.SessionTimeoutMs,
		EnableAutoCommit: cfg.Kafka.EnableAutoCommit,
	}

	log.Println("Creating execution analytics Kafka consumer...")
	executionConsumer, err := kafka.NewConsumerWithConfig(executionConsumerConfig, executionCommandService)
	if err != nil {
		log.Fatalf("Failed to create execution Kafka consumer: %v", err)
	}

	// Configurar consumidor de Kafka para registros de usuarios con Azure Event Hub
	userRegistrationConsumerConfig := &kafka.ConsumerConfig{
		Brokers:          cfg.Kafka.BootstrapServers,
		GroupID:          cfg.KafkaUserRegistration.GroupID,
		Topic:            cfg.KafkaUserRegistration.Topic,
		SecurityProtocol: cfg.Kafka.SecurityProtocol,
		SaslMechanism:    cfg.Kafka.SaslMechanism,
		SaslUsername:     cfg.Kafka.SaslUsername,
		SaslPassword:     cfg.Kafka.SaslPassword,
		RequestTimeoutMs: cfg.Kafka.RequestTimeoutMs,
		SessionTimeoutMs: cfg.Kafka.SessionTimeoutMs,
		EnableAutoCommit: cfg.Kafka.EnableAutoCommit,
	}

	log.Println("Creating user registration Kafka consumer...")
	userRegistrationConsumer, err := kafka.NewUserRegistrationConsumerWithConfig(userRegistrationConsumerConfig, userRegistrationCommandService)
	if err != nil {
		log.Fatalf("Failed to create user registration Kafka consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar consumidor de ejecuciones
	go func() {
		log.Printf("Starting Kafka consumer for topic: %s", cfg.Kafka.Topic)
		if err := executionConsumer.Start(ctx); err != nil {
			log.Printf("Execution Kafka consumer error: %v", err)
		}
	}()

	// Iniciar consumidor de registros de usuarios
	go func() {
		log.Printf("Starting Kafka consumer for user registrations on topic: %s", cfg.KafkaUserRegistration.Topic)
		if err := userRegistrationConsumer.Start(ctx); err != nil {
			log.Printf("User registration Kafka consumer error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Deregistrar de Eureka
	if eurekaClient != nil {
		if err := eurekaClient.Deregister(); err != nil {
			log.Printf("Error deregistering from Eureka: %v", err)
		}
	}

	// Detener consumidores de Kafka
	cancel()
	if err := executionConsumer.Close(); err != nil {
		log.Printf("Error closing execution Kafka consumer: %v", err)
	}
	if err := userRegistrationConsumer.Close(); err != nil {
		log.Printf("Error closing user registration Kafka consumer: %v", err)
	}

	// Detener servidor HTTP
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
