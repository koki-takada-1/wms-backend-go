package handler

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/koki-takada-1/go-rest-api/api/internal/models"

	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func SetupHandlers() {
	db = models.GetDB() // データベース接続をグローバル変数に設定
}

func generateID() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 7)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

// 　部品テーブルのレコード一覧get
func GetParts(c *gin.Context) {
	var parts []models.Parts
	if err := db.Find(&parts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, parts)
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

func GetPartDetails(c *gin.Context) {
	partId := c.Param("id") // URLからPartのIDを取得
	var part models.Parts
	var partLocations []models.PartLocations
	var orders []models.Orders
	var locations []models.Locations

	// PartsテーブルからIdに該当するレコードを取得
	if err := db.First(&part, "id = ?", partId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Part not found"})
		return
	}

	// PartLocationsテーブルからPartIdに一致する全レコードを取得
	if err := db.Where("part_id = ?", partId).Find(&partLocations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve part locations"})
		return
	}

	// OrdersテーブルからPartIdに一致する全レコードを取得
	if err := db.Where("part_id = ?", partId).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}

	// PartLocationsから取得したLocationIdを使用してLocationsテーブルからレコードを取得
	locationIds := make([]string, len(partLocations))
	for i, pl := range partLocations {
		locationIds[i] = pl.LocationId
	}
	if err := db.Where("id IN ?", locationIds).Find(&locations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve locations"})
		return
	}

	// 成功した場合、Part、PartLocations、Orders、およびLocationsのデータをJSON形式で返す
	c.JSON(http.StatusOK, gin.H{
		"Part":          part,
		"PartLocations": partLocations,
		"Orders":        orders,
		"Locations":     locations,
	})
}
