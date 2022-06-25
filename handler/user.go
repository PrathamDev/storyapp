package handler

import (
	"io"
	"os"
	"path/filepath"

	"github.com/PrathamDev/storyapp/database"
	"github.com/PrathamDev/storyapp/model"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CreateUser(c *fiber.Ctx) error {
	type NewUser struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	db := database.DB
	user := new(model.User)

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": nil})
	}

	hash, err := hashPassword(user.Password)

	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Couldn't hash password", "data": nil})
	}

	user.Password = hash
	if err := db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Couldn't create user", "data": nil})
	}

	newUser := NewUser{
		Username: user.Username,
		Email:    user.Email,
	}

	return c.JSON(fiber.Map{"status": "success", "message": "User created", "data": newUser})

}

func GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	db := database.DB

	var user model.User
	db.Find(&user, id)
	if user.IsEmpty() {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "No user found with ID", "data": nil})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "User found", "data": user})
}

func UploadProfilePicture(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	fileHeader, err := c.FormFile("file")
	println(fileHeader.Size)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Expected a file as file", "data": nil})
	}
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error opening the file", "data": nil})
	}
	defer file.Close()

	os.MkdirAll("files/"+claims["username"].(string), os.ModePerm)
	createdFile, err := os.Create(filepath.Join("files/"+claims["username"].(string), filepath.Base(fileHeader.Filename)))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error(), "data": nil})

	}
	defer createdFile.Close()

	if _, err := io.Copy(createdFile, file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error saving the file in server", "data": nil})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Uploaded profile picture successfully", "data": file})
}
