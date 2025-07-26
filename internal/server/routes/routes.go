package routes

import (
	"echo-app/internal/repositories"
	s "echo-app/internal/server"
	"echo-app/internal/server/handlers"
	"echo-app/internal/server/middleware"
	"echo-app/internal/services/post"
	"echo-app/internal/services/user"
	"echo-app/internal/slogx"
	"log/slog"

	echojwt "github.com/labstack/echo-jwt/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func ConfigureRoutes(tracer slogx.TraceStarter, server *s.Server) {
	userRepository := repositories.NewUserRepository(server.DB)
	userService := user.NewService(userRepository)

	postRepository := repositories.NewPostRepository(server.DB)
	postService := post.NewService(postRepository)

	postHandler := handlers.NewPostHandlers(postService)

	authHandler, err := handlers.NewAuthHandler(server, userService, userRepository, &server.Config.Auth)

	if err != nil {
		slog.Error("auth init error")
	}

	server.Echo.Use(middleware.NewRequestLogger(tracer))

	server.Echo.GET("/swagger/*", echoSwagger.WrapHandler)

	r := server.Echo.Group("", middleware.NewRequestDebugger())

	r.GET("/login", authHandler.InitiateLogin)
	r.GET("/callback", authHandler.HandleCallback)
	r.POST("/logout", authHandler.HandleLogout)

	// Protected routes with JWT middleware
	protected := r.Group("")
	protected.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:  []byte(server.Config.Auth.OIDCClientSecret),
		TokenLookup: "header:Authorization,cookie:access_token",
		ContextKey:  "user",
	}))

	r.GET("/posts", postHandler.GetPosts)
	r.POST("/posts", postHandler.CreatePost)
	r.DELETE("/posts/:id", postHandler.DeletePost)
	r.PUT("/posts/:id", postHandler.UpdatePost)
}
