package models

type Parts struct {
	Id                string  `gorm:"primarykey;size:10"`
	Name              string  `gorm:"size:255"`
	Moq               uint    `gorm:"type:smallint"`
	CostUnitPrice     float64 `gorm:"type:numeric"`
	ContractUnitPrice float64 `gorm:"type:numeric"`
}

type Orders struct {
	Id            string `gorm:"primarykey;size:7"`
	PartId        string `gorm:"size:10"`
	Deadline      string `gorm:"type:date"`
	OrderQuantity uint   `gorm:"type:smallint"`
	Parts         Parts  `gorm:"foreignKey:PartId;references:Id"`
}

type StockFrames struct {
	Name   string `gorm:"primarykey;size:50"` // 在庫枠名
	Number string `gorm:"size:2"`             // 在庫枠No.
	Depot  bool   `gorm:"type:boolean"`
}

type Locations struct {
	Id             string      `gorm:"primarykey;size:6"` // ロケーションId
	StockFrameName string      `gorm:"size:50"`
	StockFrames    StockFrames `gorm:"foreignKey:StockFrameName;references:Name"` // 在庫枠名
}

type PartLocations struct {
	PartId     string `gorm:"primaryKey;"`
	LocationId string `gorm:"primaryKey;"`
	Stock      uint   `gorm:"type:smallint"`
	InTransit  uint   `gorm:"type:smallint"`
	// Parts      Parts     `gorm:"foreignKey:Id;references:PartId"`
	// Locations  Locations `gorm:"foreignKey:Id;references:LocationId"`
	Parts     Parts     `gorm:"foreignKey:PartId;references:Id"`
	Locations Locations `gorm:"foreignKey:LocationId;references:Id"`
}
