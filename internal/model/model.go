package model

type CatalogItem struct {
	RecordID     int64  `json:"recordId"`
	ItemName     string `json:"itemName"`
	Category     string `json:"category"`
	PriceCents   uint64 `json:"-"`
	PriceText    string `json:"priceText"`
	StockStatus  string `json:"stockStatus,omitempty"`
	Merchant     string `json:"merchant"`
	UpdatedAt    string `json:"updatedAt"`
	DisplayOrder int    `json:"-"`
}

type CatalogVersion struct {
	DatasetVersion string `json:"datasetVersion"`
	RecordCount    int    `json:"recordCount"`
	BlockCount     int    `json:"blockCount"`
	DomainSize     uint64 `json:"domainSize"`
}

type PayloadRecord struct {
	RecordID int64
	Blocks   []uint64
}
