package main

import (
	"Uranus/api"
	"Uranus/app"
	"Uranus/model"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
	"time"
)

// Claims struct for JWT
type Claims struct {
	UserId int `json:"user_id"`
	jwt.RegisteredClaims
}

var jwtKey = []byte("82jhdksl#")

func JWTMiddleware(c *gin.Context) {
	fmt.Println("..... A")
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		return
	}

	fmt.Println("..... B")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		return
	}

	tokenStr := parts[1]
	claims := &Claims{} // Use custom claims

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Print the full payload including your custom claim
	fmt.Printf("JWT payload: %+v\n", claims)
	fmt.Println("UserId from token:", claims.UserId)

	// Store the claims or just the UserId in the context
	c.Set("claims", claims)
	c.Set("userId", claims.UserId)

	c.Next()
}

func login(c *gin.Context) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	fmt.Println("email address:", creds.Email)
	fmt.Println("password:", creds.Password)
	hashedPassword, err := app.EncryptPassword(creds.Password)
	fmt.Println("hashedPassword:", hashedPassword)

	user, err := model.GetUser(app.GApp, c, creds.Email)
	if err != nil {
		http.Error(c.Writer, "No user", http.StatusBadRequest)
		return
	}
	user.Print()

	err = app.ComparePasswords(user.PasswordHash, creds.Password)
	if err != nil {
		fmt.Println("Passwords do NOT match!")
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserId: user.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	fmt.Println("token:", tokenStr)
	c.JSON(http.StatusOK, gin.H{"token": tokenStr})
}

func protectedRouteTest(c *gin.Context) {
	// Optional: get user info from context if JWTMiddleware stored it
	/*userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "login failed"})
		return
	}
	*/

	fmt.Printf("You are logged in :-)")

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func main() {

	config_file_name := "config.json"
	if len(os.Args) > 1 {
		config_file_name = os.Args[1] // First argument
	}
	fmt.Println("config_file:", config_file_name)

	err := app.GApp.LoadConfig(config_file_name)
	if err != nil {
		panic(err)
	}

	err = app.GApp.PrepareSql()
	if err != nil {
		panic(err)
	}

	app.GApp.InitDB()
	defer app.GApp.DbPool.Close()

	// Create a Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New() // Use `Default()` for built-in logging and recovery
	// Add middleware explicitly
	// router.Use(gin.Logger())
	// router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8000"}, // Your frontend origin
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Register routes
	apiRoute := router.Group("/api")
	{
		apiRoute.POST("/login", login)
		apiRoute.GET("/query", api.QueryHandler)
		apiRoute.GET("/protected", JWTMiddleware, protectedRouteTest)

		apiRoute.POST("/event", JWTMiddleware, api.CreateEventHandler)
	}

	// Start the server (Gin handles everything)
	fmt.Println("Server running")
	if err := router.Run(":9090"); err != nil {
		fmt.Println("Server error:", err)
	}
}
