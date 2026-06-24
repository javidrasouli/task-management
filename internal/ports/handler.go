package ports

import "github.com/gin-gonic/gin"

type Handler interface {
	GetTasks(c *gin.Context)
	GetTaskById(c *gin.Context)
	CreateTask(c *gin.Context)
	UpdateTask(c *gin.Context)
	DeleteTask(c *gin.Context)
	UpdateTaskStatus(c *gin.Context)
}
