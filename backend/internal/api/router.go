package api

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/task-manager/internal/websocket"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, hub *websocket.Hub, redisClient *redis.Client) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", healthCheck)

	// API routes
	api := router.Group("/api/v1")
	{
		// Project routes
		projectHandler := NewProjectHandler(db, hub)
		api.GET("/projects", projectHandler.GetProjects)
		api.GET("/projects/:id", projectHandler.GetProject)
		api.POST("/projects", projectHandler.CreateProject)
		api.PUT("/projects/:id", projectHandler.UpdateProject)
		api.DELETE("/projects/:id", projectHandler.DeleteProject)

		// Task routes
		taskHandler := NewTaskHandler(db, hub)
		api.GET("/tasks", taskHandler.GetTasks)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.POST("/tasks", taskHandler.CreateTask)
		api.PUT("/tasks/:id", taskHandler.UpdateTask)
		api.DELETE("/tasks/:id", taskHandler.DeleteTask)

		// Task dependency routes
		api.GET("/tasks/:id/dependencies", taskHandler.GetDependencies)
		api.POST("/tasks/:id/dependencies", taskHandler.AddDependency)
		api.DELETE("/tasks/:id/dependencies/:depends_on_id", taskHandler.RemoveDependency)

		// Comment routes
		commentHandler := NewCommentHandler(db, hub)
		api.GET("/tasks/:task_id/comments", commentHandler.GetComments)
		api.POST("/tasks/:task_id/comments", commentHandler.CreateComment)
		api.PUT("/comments/:comment_id", commentHandler.UpdateComment)
		api.DELETE("/comments/:comment_id", commentHandler.DeleteComment)
	}

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		handleWebSocket(hub, c)
	})

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
