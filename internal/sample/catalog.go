package sample

import (
	"fmt"

	"github.com/YY404NF/pq-backend/internal/model"
	"github.com/YY404NF/pq-backend/internal/payload"
)

func CatalogItems() []model.CatalogItem {
	categories := []string{"基础商品", "研究设备", "隐私终端", "演示样例"}
	stockStatuses := []string{"有货", "低库存", "限量", "预售"}

	items := make([]model.CatalogItem, 0, 32)
	for i := 0; i < 32; i++ {
		name := fmt.Sprintf("商品%d", i+1)
		category := categories[i%len(categories)]
		merchant := fmt.Sprintf("商家%d", i+1)
		status := stockStatuses[(i/4)%len(stockStatuses)]
		priceCents := uint64(15900 + i*875)
		updatedAt := fmt.Sprintf("2026-03-%02d 10:%02d", (i%28)+1, (i*7)%60)

		items = append(items, model.CatalogItem{
			RecordID:     int64(i),
			ItemName:     name,
			Category:     category,
			PriceCents:   priceCents,
			PriceText:    payload.FormatPrice(priceCents),
			StockStatus:  status,
			Merchant:     merchant,
			UpdatedAt:    updatedAt,
			DisplayOrder: i,
		})
	}
	return items
}
