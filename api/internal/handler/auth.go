package handler

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/koki-takada-1/go-rest-api/api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

var privateKey *ecdsa.PrivateKey
var publicKey *ecdsa.PublicKey

func init() {
	generateSecretKey()
}

func generateSecretKey() {
	var err error
	privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate ECDSA key: %v", err)
	}
	publicKey = &privateKey.PublicKey
}

func generateConfirmToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func createToken(userID string) (string, error) {
	if privateKey == nil {
		return "", fmt.Errorf("privateKey is not initialized")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	return token, err
}

func RegisterUser(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// ConfirmTokenの生成
	confirmToken := generateConfirmToken()

	user := models.User{
		Email:        input.Email,
		PassWordHash: string(hashedPassword),
		ConfirmToken: confirmToken, // ConfirmTokenをユーザーレコードに追加
	}

	// ユーザーをデータベースに保存
	if result := db.Create(&user); result.Error != nil {
		c.JSON(http.StatusBadRequest, result.Error.Error())
		return
	}

	// アクティベーションメールを送信
	sendActivationEmail(c, user.Email, confirmToken)

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful, please check your email to activate your account"})
}

func sendActivationEmail(c *gin.Context, email, token string) {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_FROM"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "部品問い合わせシステムのアカウント登録")
	m.SetBody("text/html", "こちらのリンクから登録を完了させてください: <a href=\"http://localhost:5100/v1/activateaccount?token="+token+"\">Activate</a>")
	smtpPort := 587
	d := gomail.NewDialer(os.Getenv("SMTP_HOST"), smtpPort, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASSWORD"))

	if err := d.DialAndSend(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send activation email"})
		return
	}
}

func ActivateAccount(c *gin.Context) {
	tokenString := c.Query("token")

	// ConfirmTokenを使用してユーザーを検索
	var user models.User
	if err := db.Where("confirm_token = ?", tokenString).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// トークンが有効であれば、ユーザーのアカウントをアクティベート
	now := time.Now()      // 現在時刻を取得
	user.VerifiedAt = &now // ポインタを代入
	user.ConfirmToken = "" // トークンをクリア
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user verification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account activated successfully"})
}

func Login(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Where("email = ?", credentials.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PassWordHash), []byte(credentials.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenString, err := createToken(user.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func Authenticate(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	email := c.GetHeader("Email") // ユーザー識別のためのEmailヘッダーを想定

	if email == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Email header is missing"})
		return
	}

	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials"})
		return
	}

	// ユーザーが登録を完了しているか確認
	if user.VerifiedAt == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Account not activated"})
		return
	}

	token, err := verifyToken(tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		c.Set("userID", claims["userId"])
		c.Next() // 認証が成功し、かつアカウントがアクティブであれば次のハンドラーに処理を移す
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
	}
}
