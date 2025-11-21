package handler

import (
	"cruder/internal/controller"
	"cruder/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New(router *gin.Engine, userController *controller.UserController) *gin.Engine {
	// Apply JSON logger middleware to all routes
	router.Use(middleware.JSONLogger())

	v1 := router.Group("/api/v1")
	{
		userGroup := v1.Group("/users")
		{
			userGroup.GET("/", userController.GetAllUsers)
			userGroup.GET("/username/:username", userController.GetUserByUsername)
			userGroup.GET("/id/:id", userController.GetUserByID)

			userGroup.POST("/", userController.CreateUser)        // Task3
			userGroup.PATCH("/:uuid", userController.UpdateUser)  // Task3
			userGroup.DELETE("/:uuid", userController.DeleteUser) // Task3
		}
	}
	return router
}
