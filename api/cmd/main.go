package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/koki-takada-1/go-rest-api/api/internal/handler"
	"github.com/koki-takada-1/go-rest-api/api/internal/models"
	"gorm.io/gorm"
)

const latest = "/v1"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	// Ginエンジンのインスタンスを作成
	r := gin.Default()
	// CORS for https://localhost:3005 and https://localhost:5100
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3001"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	r.Use(cors.New(config))

	db, err := models.SetupDB()
	if err != nil {
		panic("Failed to connect to database")
	}

	handler.SetupHandlers()

	ImportCSVToDatabase(db, "table-csv/parts.csv", "Parts")
	ImportCSVToDatabase(db, "table-csv/stock_frame.csv", "StockFrames")
	ImportCSVToDatabase(db, "table-csv/locations.csv", "Locations")
	ImportCSVToDatabase(db, "table-csv/part_locations.csv", "PartLocations")
	ImportCSVToDatabase(db, "table-csv/orders.csv", "Orders")

	v1 := r.Group(latest)

	// 認証が不要なルート
	v1.POST("/register", handler.RegisterUser)
	v1.POST("/login", handler.Login)
	v1.GET("/activateaccount", handler.ActivateAccount)

	auth := v1.Group("/")
	auth.Use(handler.Authenticate)
	// ルートURL ("/") に対するGETリクエストをハンドル
	auth.GET("/parts", handler.GetParts)
	auth.POST("/parts", handler.PostParts)
	// idに基づく部品情報詳細
	auth.GET("/parts/:id", handler.GetPartDetails)
	auth.DELETE("/parts/:id", handler.DeletePart)

	auth.GET("/orders", handler.GetOrders)
	auth.POST("/orders", handler.PostOrder)
	auth.DELETE("/orders/:id", handler.DeleteOrder)

	auth.GET("/stockframes", handler.GetStockFrame)
	auth.POST("/stockframes", handler.PostStockFrame)
	auth.DELETE("/stockframes/:id", handler.DeleteStockFrame)

	auth.GET("/locations", handler.GetLocation)
	auth.POST("/locations", handler.PostLocation)
	auth.DELETE("/locations/:id", handler.DeleteLocation)

	auth.POST("/partlocations", handler.PostPartLocation)
	auth.PATCH("/partlocations/part/:partId/location/:locationId", handler.PatchPartLocation)

	r.Run(":5100")
}

func ImportCSVToDatabase(db *gorm.DB, filePath string, tableName string) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	for _, record := range records {
		switch tableName {
		case "Parts":
			moq, _ := strconv.ParseUint(record[2], 10, 32)
			costUnitPrice, _ := strconv.ParseFloat(record[3], 64)
			contractUnitPrice, _ := strconv.ParseFloat(record[4], 64)
			part := models.Parts{
				Id:                record[0],
				Name:              record[1],
				Moq:               uint(moq),
				CostUnitPrice:     costUnitPrice,
				ContractUnitPrice: contractUnitPrice,
			}
			db.Create(&part)
		case "StockFrames":
			depot, _ := strconv.ParseBool(record[2])
			frame := models.StockFrames{
				Name:   record[0],
				Number: record[1],
				Depot:  depot,
			}
			db.Create(&frame)
		case "Locations":
			location := models.Locations{
				Id:             record[0],
				StockFrameName: record[1],
			}
			db.Create(&location)
		case "PartLocations":
			stock, _ := strconv.ParseUint(record[2], 10, 32)
			inTransit, _ := strconv.ParseUint(record[3], 10, 32)
			partLocation := models.PartLocations{
				PartId:     record[0],
				LocationId: record[1],
				Stock:      uint(stock),
				InTransit:  uint(inTransit),
			}
			db.Create(&partLocation)
		case "Orders":
			orderQuantity, _ := strconv.ParseUint(record[3], 10, 32)
			order := models.Orders{
				Id:            record[0],
				PartId:        record[1],
				Deadline:      record[2],
				OrderQuantity: uint(orderQuantity),
			}
			db.Create(&order)
		}
	}
}
