package routes

import (
	"echo-app/internal/repositories"
	s "echo-app/internal/server"
	"echo-app/internal/server/handlers"
	"echo-app/internal/server/middleware"
	"echo-app/internal/services/post"
	"echo-app/internal/services/token"
	"echo-app/internal/services/user"
	"echo-app/internal/slogx"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func ConfigureRoutes(tracer slogx.TraceStarter, server *s.Server) {
	userRepository := repositories.NewUserRepository(server.DB)
	userService := user.NewService(userRepository)

	postRepository := repositories.NewPostRepository(server.DB)
	postService := post.NewService(postRepository)

	postHandler := handlers.NewPostHandlers(postService)
	authHandler := handlers.NewAuthHandler(userService, server)
	registerHandler := handlers.NewRegisterHandler(userService)

	server.Echo.Use(middleware.NewRequestLogger(tracer))

	server.Echo.GET("/swagger/*", echoSwagger.WrapHandler)

	server.Echo.POST("/login", authHandler.Login)
	server.Echo.POST("/register", registerHandler.Register)
	server.Echo.POST("/refresh", authHandler.RefreshToken)

	r := server.Echo.Group("", middleware.NewRequestDebugger())

	// Configure middleware with the custom claims type
	config := echojwt.Config{
		NewClaimsFunc: func(_ echo.Context) jwt.Claims {
			return new(token.JwtCustomClaims)
		},
		SigningKey: []byte(server.Config.Auth.AccessSecret),
	}
	r.Use(echojwt.WithConfig(config))

	r.GET("/posts", postHandler.GetPosts)
	r.POST("/posts", postHandler.CreatePost)
	r.DELETE("/posts/:id", postHandler.DeletePost)
	r.PUT("/posts/:id", postHandler.UpdatePost)
}
