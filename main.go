package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"

	_ "b2time/docs" // load generated docs

	jwtware "github.com/gofiber/jwt/v2"
	"github.com/gofiber/swagger"
	"github.com/golang-jwt/jwt/v4"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var books []Book

func checkMiddleware(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	if claims["role"] != "admin" {
		return fiber.ErrUnauthorized
	}
	return c.Next()
}

// @title						Book API
// @description				This is a sample server for a book API.
// @version					1.0
// @host						localhost:8080
// @BasePath					/
// @schemes					http
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("load . env error")
	}
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/swagger/*", swagger.HandlerDefault) //default

	books = append(books, Book{ID: 1, Title: "mikelopster", Author: "mike"})
	books = append(books, Book{ID: 2, Title: "MM", Author: "mike"})

	app.Post("/login", Login)

	//JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}))
	app.Use(checkMiddleware)

	app.Get("/books", getBooks)
	app.Get("/books/:id", getBook)
	app.Post("/books", creactBook)
	app.Put("/books/:id", updateBook)
	app.Delete("/books/:id", deleteBook)

	app.Post("/upload", uploadFile)

	app.Get("/test-html", testHTML)

	app.Get("/config", getEnv)

	app.Listen(":8080")

}

func uploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	err = c.SaveFile(file, "./uploads/"+file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.SendString("File upload complete")
}
func testHTML(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "Hello, World!",
		"Name":  "Mike",
	})
}
func getEnv(c *fiber.Ctx) error {
	secret := os.Getenv("SECRET")
	if secret == "" {
		secret = "defaulsecret"
	}
	return c.JSON(fiber.Map{
		"SECRET": secret,
	})
}

type User = struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var memberUser = User{
	Email:    "user@example.com",
	Password: "password123",
}

func Login(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if user.Email != memberUser.Email || user.Password != memberUser.Password {
		return fiber.ErrUnauthorized
	}
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["role"] = "admin"
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(fiber.Map{
		"message": "Login success",
		"token":   t,
	})
}
