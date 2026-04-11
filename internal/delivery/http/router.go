package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/emadhejazian/subscription_service/internal/delivery/http/handler"
	"github.com/emadhejazian/subscription_service/internal/delivery/http/middleware"
	pgRepo "github.com/emadhejazian/subscription_service/internal/repository/postgres"
	"github.com/emadhejazian/subscription_service/internal/usecase"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	// Repositories
	productRepo := pgRepo.NewProductRepo(db)
	subscriptionRepo := pgRepo.NewSubscriptionRepo(db)
	voucherRepo := pgRepo.NewVoucherRepo(db)

	// Usecases
	productUC := usecase.NewProductUsecase(productRepo)
	voucherUC := usecase.NewVoucherUsecase(voucherRepo)
	subscriptionUC := usecase.NewSubscriptionUsecase(subscriptionRepo, productRepo, voucherRepo)

	// Handlers
	productHandler := handler.NewProductHandler(productUC)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionUC)
	voucherHandler := handler.NewVoucherHandler(voucherUC, productUC)

	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.GET("", productHandler.GetAll)
			products.GET("/:id", productHandler.GetByID)
		}

		vouchers := v1.Group("/vouchers")
		{
			vouchers.POST("/validate", voucherHandler.Validate)
		}

		subscriptions := v1.Group("/subscriptions")
		subscriptions.Use(middleware.Auth())
		{
			subscriptions.POST("", subscriptionHandler.Buy)
			subscriptions.GET("/me", subscriptionHandler.GetMySubscription)
			subscriptions.GET("/:id", subscriptionHandler.GetByID)
			subscriptions.POST("/:id/pause", subscriptionHandler.Pause)
			subscriptions.POST("/:id/unpause", subscriptionHandler.Unpause)
			subscriptions.POST("/:id/cancel", subscriptionHandler.Cancel)
		}
	}

	return r
}
