package router

import (
	_ "task-management/docs"
	"task-management/internal/delivery/handlers"
	"task-management/internal/delivery/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func Setup(h *handlers.TaskHandler, log *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(log), middleware.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	tasks := v1.Group("/tasks")
	{
		tasks.GET("", h.GetTasks)
		tasks.GET("/:id", h.GetTaskById)
		tasks.POST("", h.CreateTask)
		tasks.PUT("/:id", h.UpdateTask)
		tasks.DELETE("/:id", h.DeleteTask)
		tasks.PATCH("/:id/status", h.UpdateTaskStatus)
	}

	return r
}
