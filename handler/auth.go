package handler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PrathamDev/storyapp/config"
	"github.com/PrathamDev/storyapp/database"
	"github.com/PrathamDev/storyapp/model"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func getUserByEmail(e string) (*model.User, error) {
	db := database.DB
	var user model.User

	if err := db.Where(&model.User{Email: e}).Find(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func getUserById(id int) (*model.User, error) {
	db := database.DB
	var user model.User

	db.Find(&user, id)

	if user.Password == "" {
		return nil, errors.New("invalid refresh token")
	}

	return &user, nil
}

func getUserByUsername(u string) (*model.User, error) {
	db := database.DB
	var user model.User

	if err := db.Where(&model.User{Username: u}).Find(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func checkPasswordHash(password string, hashString string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashString), []byte(password))
	return err == nil
}

func generateTokenPair(user model.User) (fiber.Map, error) {
	accessTokenClaims := jwt.MapClaims{
		"username": user.Username,
		"user_id":  user.ID,
		"role":     user.Role,
		"sub":      user.ID,
		"exp":      time.Now().Add(time.Minute * 60).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	signedAccessToken, err := accessToken.SignedString([]byte(config.FetchKey("SECRET")))
	if err != nil {
		return nil, err
	}

	refreshTokenClaims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 720).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	signedRefreshToken, err := refreshToken.SignedString([]byte(config.FetchKey("SECRET")))
	if err != nil {
		return nil, err
	}

	return fiber.Map{
		"access_token":  signedAccessToken,
		"refresh_token": signedRefreshToken,
	}, nil
}

func Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}

	var input LoginInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error on login request", "data": err})
	}

	identity := input.Identity
	pass := input.Password

	var foundUser *model.User
	var errorMsg string
	var err error
	if strings.Contains(identity, "@") {
		foundUser, err = getUserByEmail(identity)
		if err != nil {
			errorMsg = "No user found with email"
		}
	} else {
		foundUser, err = getUserByUsername(identity)
		if err != nil {
			errorMsg = "No user found with username"
		}
	}

	if errorMsg != "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": errorMsg, "data": nil})
	}

	if !checkPasswordHash(pass, foundUser.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Invalid password", "data": nil})
	}

	tokens, err := generateTokenPair(*foundUser)

	if err != nil {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": err.Error(), "data": nil})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Success login", "data": tokens})
}

func SignupWithEmailAndPassword(c *fiber.Ctx) error {
	type NewUser struct {
		Username string    `json:"username"`
		Email    string    `json:"email"`
		PhotoURL string    `json:"photourl"`
		Tokens   fiber.Map `json:"tokens"`
	}
	db := database.DB
	user := new(model.User)

	if err := c.BodyParser(&user); err != nil || user.Email == "" || user.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Email and Password is required", "data": nil})
	}

	foundUserWithEmail, err := getUserByEmail(user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error fetching details of the user with the given email", "data": nil})
	}
	if foundUserWithEmail.IsEmpty() {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "User with the given email already exist", "data": foundUserWithEmail})
	}

	hash, err := hashPassword(user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error while hashing the password", "data": nil})
	}
	user.Password = hash
	if err := db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error while creating the user", "data": nil})
	}

	tokens, err := generateTokenPair(*user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Error while generating the tokens", "data": nil})
	}
	var newUser NewUser
	newUser.Email = user.Email
	newUser.PhotoURL = user.PhotoURL
	newUser.Username = user.Username
	newUser.Tokens = tokens

	return c.JSON(fiber.Map{"status": "success", "message": "User created successfully", "data": newUser})
}

func Token(c *fiber.Ctx) error {
	type TokenRegBody struct {
		RefreshToken string `json:"refresh_token"`
	}

	var tokenReq TokenRegBody

	if err := c.BodyParser(&tokenReq); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Expected refresh_token in the body", "data": err})
	}

	token, err := jwt.Parse(tokenReq.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(config.FetchKey("SECRET")), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int(claims["sub"].(float64))

		user, err := getUserById(userID)
		if err != nil {
			return err
		}
		tokens, err := generateTokenPair(*user)

		if err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).JSON(tokens)
	}
	return err
}
