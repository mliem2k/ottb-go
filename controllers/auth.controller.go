package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mliem2k/ottb-go/initializers"
	"github.com/mliem2k/ottb-go/models"
	"github.com/mliem2k/ottb-go/utils"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(DB *gorm.DB) AuthController {
	return AuthController{DB}
}

// SignUp User
func (ac *AuthController) SignUpUser(ctx *gin.Context) {
	// ac.DB.Logger.LogMode(logger.Info)
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	var payload *models.SignUpInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if payload.Password != payload.PasswordConfirm {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Passwords do not match"})
		return
	}

	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	now := time.Now()
	newUser := models.User{
		Name:      payload.Name,
		Username:  strings.ToLower(payload.Username),
		Email:     strings.ToLower(payload.Email),
		Password:  hashedPassword,
		Role:      "user",
		Verified:  false,
		Photo:     payload.Photo,
		Provider:  "local",
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := ac.DB.Create(&newUser)

	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "User with that email or username already exists"})
		return
	} else if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Something bad happened"})
		return
	}

	// Set up SMTP server configuration.
	smtpServer := config.SmtpServer
	smtpPort := config.SmtpPort
	username := config.SmtpUser
	password := config.SmtpPass

	from := config.SmtpFrom
	to := newUser.Email
	subject := "Verify your OTTB account"
	verificationLink := config.ServerOrigin + "/api/auth/verifyemail/" + newUser.ID.String()

	// HTML body content
	body := `
	<html>
	<head>
		<title>Verify Your OTTB Account</title>
	</head>
	<body>
		<p>Hello,</p>
		<p>Please click the following link to verify your OTTB account:</p>
		<p><a href="` + verificationLink + `">Verify Email</a></p>
		<p>If you didn't request this, please ignore this email.</p>
		<p>Thank you!</p>
	</body>
	</html>
	`

	// Set up the email message.
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	// Set up authentication information.
	d := gomail.NewDialer(smtpServer, smtpPort, username, password)

	// Send the email.
	if err := d.DialAndSend(msg); err != nil {
		log.Println("Failed to send email:", err)
		// If there's an error, delete the newly created user
		if deleteErr := ac.DB.Delete(&newUser).Error; deleteErr != nil {
			log.Println("Failed to delete user:", deleteErr)
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Failed to send email."})
		return
	}

	log.Println("Email sent successfully!")

	userResponse := &models.UserResponse{
		ID:        newUser.ID,
		Name:      newUser.Name,
		Username:  newUser.Username,
		Email:     newUser.Email,
		Photo:     newUser.Photo,
		Role:      newUser.Role,
		Provider:  newUser.Provider,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}
	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": gin.H{"user": userResponse}})
}

func (ac *AuthController) VerifyEmail(ctx *gin.Context) {
	userId := ctx.Param("userId")
	var user models.User
	result := ac.DB.First(&user, "id = ?", userId)
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid user ID"})
		return
	}
	user.Verified = true
	ac.DB.Save(&user)
	htmlResponse := `
	<html>
	<head>
		<title>OTTB Email Verification</title>
	</head>
	<body style="font-family: Arial, sans-serif; background-color: #f0f0f0; padding: 20px;">
		<div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 5px; padding: 20px; box-shadow: 0px 2px 5px 0px rgba(0,0,0,0.1);">
			<h1 style="color: #333333; text-align: center;">OTTB Email Verified Successfully</h1>
			<p style="color: #666666; text-align: center;">Your email has been successfully verified.</p>
		</div>
	</body>
	</html>
`

	ctx.Header("Content-Type", "text/html")
	ctx.String(http.StatusOK, htmlResponse)
}

func (ac *AuthController) SignInUser(ctx *gin.Context) {
	var payload *models.SignInInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var user models.User
	result := ac.DB.First(&user, "username = ?", strings.ToLower(payload.Username))
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid username or Password"})
		return
	}

	if err := utils.VerifyPassword(user.Password, payload.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid username or Password"})
		return
	}

	config, _ := initializers.LoadConfig(".")

	// Generate Tokens
	access_token, err := utils.CreateToken(config.AccessTokenExpiresIn, user.ID, config.AccessTokenPrivateKey)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	refresh_token, err := utils.CreateToken(config.RefreshTokenExpiresIn, user.ID, config.RefreshTokenPrivateKey)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.SetCookie("access_token", access_token, config.AccessTokenMaxAge*60, "/", config.ServerOrigin, false, true)
	ctx.SetCookie("refresh_token", refresh_token, config.RefreshTokenMaxAge*60, "/", config.ServerOrigin, false, true)
	ctx.SetCookie("logged_in", "true", config.AccessTokenMaxAge*60, "/", config.ServerOrigin, false, false)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "access_token": access_token})
}

// Refresh Access Token
func (ac *AuthController) RefreshAccessToken(ctx *gin.Context) {
	message := "could not refresh access token"

	cookie, err := ctx.Cookie("refresh_token")

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	config, _ := initializers.LoadConfig(".")

	sub, err := utils.ValidateToken(cookie, config.RefreshTokenPublicKey)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var user models.User
	result := ac.DB.First(&user, "id = ?", fmt.Sprint(sub))
	if result.Error != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": "the user belonging to this token no logger exists"})
		return
	}

	access_token, err := utils.CreateToken(config.AccessTokenExpiresIn, user.ID, config.AccessTokenPrivateKey)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.SetCookie("access_token", access_token, config.AccessTokenMaxAge*60, "/", config.ServerOrigin, false, true)
	ctx.SetCookie("logged_in", "true", config.AccessTokenMaxAge*60, "/", config.ServerOrigin, false, false)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "access_token": access_token})
}

func (ac *AuthController) LogoutUser(ctx *gin.Context) {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}
	ctx.SetCookie("access_token", "", -1, "/", config.ServerOrigin, false, true)
	ctx.SetCookie("refresh_token", "", -1, "/", config.ServerOrigin, false, true)
	ctx.SetCookie("logged_in", "", -1, "/", config.ServerOrigin, false, false)

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}
