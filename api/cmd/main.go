package main

import (
	"github.com/gin-gonic/gin"
	"github.com/koki-takada-1/go-rest-api/api/internal/handler"
)

const latest = "/v1"

func main() {
	// Ginエンジンのインスタンスを作成
	r := gin.Default()
	v1 := r.Group(latest)

	handler.Init()
	// ルートURL ("/") に対するGETリクエストをハンドル
	v1.GET("/parts", handler.GetParts)
	v1.POST("/parts", handler.PostParts)
	// idに基づく部品情報詳細
	v1.GET("/parts/:id", handler.GetPartDetailsWithRelations)
	v1.DELETE("/parts/:id", handler.DeletePart)

	v1.GET("/orders", handler.GetOrders)
	v1.POST("/orders", handler.PostOrder)
	v1.DELETE("/orders/:id", handler.DeleteOrder)

	v1.GET("/stockframes", handler.GetStockFrame)
	v1.POST("/stockframes", handler.PostStockFrame)
	v1.DELETE("/stockframes/:id", handler.DeleteStockFrame)

	v1.GET("/locations", handler.GetLocation)
	v1.POST("/locations", handler.PostLocation)
	v1.DELETE("/locations/:id", handler.DeleteLocation)

	v1.POST("/partlocations", handler.PostPartLocation)
	v1.PATCH("/partlocations/part/:partId/location/:locationId", handler.PatchPartLocation)

	r.Run(":5100")
}
