package main

import (
	"log"
	"quotepro-backend/config"
	"quotepro-backend/handlers"
	"quotepro-backend/middleware"
	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db := models.InitDB(cfg.DBPath)
	models.AutoMigrate(db)

	r := gin.Default()
	r.Use(middleware.CORS())

	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register(db))
			auth.POST("/login", handlers.Login(db, cfg.JWTSecret))
			auth.GET("/profile", middleware.Auth(cfg.JWTSecret), handlers.Profile(db))
		}

		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			products := protected.Group("/products")
			{
				products.GET("", handlers.ListProducts(db))
				products.GET("/:id", handlers.GetProduct(db))
				products.POST("", handlers.CreateProduct(db))
				products.PUT("/:id", handlers.UpdateProduct(db))
				products.DELETE("/:id", handlers.DeleteProduct(db))
				products.POST("/import", handlers.ImportProducts(db))
			}

			quotes := protected.Group("/quotes")
			{
				quotes.POST("/parse", handlers.ParseRequirement(cfg))
				quotes.POST("/generate", handlers.GenerateQuote(cfg))
				quotes.POST("", handlers.CreateQuote(db))
				quotes.GET("", handlers.ListQuotes(db))
				quotes.GET("/:id", handlers.GetQuote(db))
				quotes.PUT("/:id", handlers.UpdateQuote(db))
				quotes.DELETE("/:id", handlers.DeleteQuote(db))
				quotes.POST("/:id/duplicate", handlers.DuplicateQuote(db))
			}

			templates := protected.Group("/templates")
			{
				templates.GET("", handlers.ListTemplates(db))
				templates.GET("/:id", handlers.GetTemplate(db))
				templates.POST("", handlers.CreateTemplate(db))
				templates.PUT("/:id", handlers.UpdateTemplate(db))
				templates.DELETE("/:id", handlers.DeleteTemplate(db))
			}

			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", handlers.DashboardStats(db))
				dashboard.GET("/recent", handlers.RecentQuotes(db))
			}

			api.POST("/upload", middleware.Auth(cfg.JWTSecret), handlers.Upload())
		}
	}

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
