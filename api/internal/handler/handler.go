package handler

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/koki-takada-1/go-rest-api/api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB
)

func generateID() string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 7)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func Init() {
	var err error
	// または環境変数から取得
	// dbUser := os.Getenv("DB_USER")
	// dbPassword := os.Getenv("DB_PASSWORD")
	// dbName := os.Getenv("DB_DATABASE")
	// dbHost := os.Getenv("DB_HOST") // または環境変数から取得
	// dbPort := os.Getenv("DB_PORT") // または環境変数から取得
	dbUser := "postgres"
	dbPassword := "postgres"
	dbName := "postgres"
	dbHost := "db"
	dbPort := "5432"

	// DSNを構築
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo", dbHost, dbUser, dbPassword, dbName, dbPort)
	log.Println("DSNはこちら:", dsn)
	// GORMでデータベースに接続
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// データベースにテーブルを作成（マイグレーションの順序を修正）
	err = db.AutoMigrate(
		&models.StockFrames{},
		&models.Parts{},
		&models.Locations{},     // LocationsをPartsの後にマイグレート
		&models.PartLocations{}, // PartLocationsを最後にマイグレート
		&models.Orders{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}

// 　部品テーブルのレコード一覧get
func GetParts(c *gin.Context) {
	var parts []models.Parts
	db.Find(&parts)
	c.JSON(http.StatusOK, parts)
	// c.JSON(200, gin.H{"post":post})
}

// 部品レコードにレコードをpost
func PostParts(c *gin.Context) {

	var body models.Parts

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if result := db.Create(&body); result.Error != nil {
		c.JSON(http.StatusBadRequest, result.Error.Error())
		return
	}
	c.JSON(http.StatusOK, body)
}

func GetPartDetailsWithRelations(c *gin.Context) {
	partId := c.Param("id")
	var part models.Parts
	var orders []models.Orders
	var partLocations []models.PartLocations
	var locations []models.Locations
	var stockFrames []models.StockFrames

	// PartsテーブルからIdに該当するレコードを取得
	if err := db.Preload("PartLocations").Preload("Orders").First(&part, "id = ?", partId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Part not found"})
		return
	}

	// OrdersテーブルからPartIdに一致する全レコードを取得
	db.Where("part_id = ?", partId).Find(&orders)

	// PartLocationsテーブルからPartIdに一致する全レコードを取得
	db.Where("part_id = ?", partId).Find(&partLocations)

	// LocationsテーブルからPartLocationsのLocationIdに一致する全レコードを取得
	locationIds := make([]string, 0)
	for _, pl := range partLocations {
		locationIds = append(locationIds, pl.LocationId)
	}
	db.Where("id IN ?", locationIds).Find(&locations)

	// StockFrameテーブルからLocationsのStockFrameNameに一致する全レコードを取得
	stockFrameNames := make([]string, 0)
	for _, loc := range locations {
		stockFrameNames = append(stockFrameNames, loc.StockFrameName)
	}
	db.Where("name IN ?", stockFrameNames).Find(&stockFrames)

	c.JSON(http.StatusOK, gin.H{
		"Part":          part,
		"Orders":        orders,
		"PartLocations": partLocations,
		"Locations":     locations,
		"StockFrames":   stockFrames,
	})
}

func DeletePart(c *gin.Context) {
	partId := c.Param("id")

	if err := db.Where("Id = ?", partId).Delete(&models.Parts{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete part"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Part deleted successfully"})
}

func GetOrders(c *gin.Context) {
	var Orders []models.Orders
	db.Find(&Orders)
	c.JSON(http.StatusOK, Orders)
	// c.JSON(200, gin.H{"post":post})
}

func PostOrder(c *gin.Context) {
	var body struct {
		PartId        string `json:"PartId"`
		Deadline      string `json:"Deadline"`
		OrderQuantity uint   `json:"OrderQuantity"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newOrder := models.Orders{
		Id:            generateID(),
		PartId:        body.PartId,
		Deadline:      body.Deadline,
		OrderQuantity: body.OrderQuantity,
	}

	// IDの重複をチェック
	var existingOrder models.Orders
	if err := db.Where("id = ?", newOrder.Id).First(&existingOrder).Error; err == nil {
		// IDが既に存在する場合、新しいIDを生成
		newOrder.Id = generateID()
	}

	if err := db.Create(&newOrder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusOK, newOrder)
}

func DeleteOrder(c *gin.Context) {
	orderId := c.Param("id")

	if err := db.Where("id = ?", orderId).Delete(&models.Orders{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete part"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}

func GetStockFrame(c *gin.Context) {
	var StockFrame []models.StockFrames
	db.Find(&StockFrame)
	c.JSON(http.StatusOK, StockFrame)
	// c.JSON(200, gin.H{"post":post})
}

func PostStockFrame(c *gin.Context) {
	var body models.StockFrames

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if result := db.Create(&body); result.Error != nil {
		c.JSON(http.StatusBadRequest, result.Error.Error())
		return
	}

	c.JSON(http.StatusOK, body)
}

func DeleteStockFrame(c *gin.Context) {
	stockFrameId := c.Param("id")

	if err := db.Where("Name = ?", stockFrameId).Delete(&models.StockFrames{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete part"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "StockFrame deleted successfully"})
}

func GetLocation(c *gin.Context) {
	var locations []models.Locations
	db.Find(&locations)
	c.JSON(http.StatusOK, locations)
	// c.JSON(200, gin.H{"post":post})
}

func PostLocation(c *gin.Context) {
	var body models.Locations

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Create(&body)
	c.JSON(http.StatusOK, body)
}

func DeleteLocation(c *gin.Context) {
	locationId := c.Param("id")

	if err := db.Where("id = ?", locationId).Delete(&models.Locations{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete part"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Locations deleted successfully"})
}

func PostPartLocation(c *gin.Context) {
	var body models.PartLocations

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if result := db.Create(&body); result.Error != nil {
		c.JSON(http.StatusBadRequest, result.Error.Error())
		return
	}

	c.JSON(http.StatusOK, body)
}

func PatchPartLocation(c *gin.Context) {
	var body struct {
		Stock     uint `json:"Stock"`
		InTransit uint `json:"InTransit"`
	}

	partId := c.Param("partId")
	locationId := c.Param("locationId")

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 指定されたPartIdとLocationIdの組み合わせのPartLocationsレコードを検索し、StockとInTransitを更新
	result := db.Model(&models.PartLocations{}).Where("part_id = ? AND location_id = ?", partId, locationId).Updates(models.PartLocations{Stock: body.Stock, InTransit: body.InTransit})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update part location"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Part location not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Part location updated successfully"})
}
