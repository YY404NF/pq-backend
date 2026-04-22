package payload

import (
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/YY404NF/pq-backend/internal/model"
)

const (
	ItemNameBytes    = 48
	CategoryBytes    = 24
	StockStatusBytes = 16
	MerchantBytes    = 32
	UpdatedAtBytes   = 24
	TotalBytes       = 160
	BlockCount       = TotalBytes / 8
)

func EncodeCatalogItem(item model.CatalogItem) []uint64 {
	buf := make([]byte, TotalBytes)
	offset := 0

	writeUint64(buf[offset:], uint64(item.RecordID))
	offset += 8
	writeString(buf[offset:offset+ItemNameBytes], item.ItemName)
	offset += ItemNameBytes
	writeString(buf[offset:offset+CategoryBytes], item.Category)
	offset += CategoryBytes
	writeUint64(buf[offset:], item.PriceCents)
	offset += 8
	writeString(buf[offset:offset+StockStatusBytes], item.StockStatus)
	offset += StockStatusBytes
	writeString(buf[offset:offset+MerchantBytes], item.Merchant)
	offset += MerchantBytes
	writeString(buf[offset:offset+UpdatedAtBytes], item.UpdatedAt)

	blocks := make([]uint64, BlockCount)
	for i := range blocks {
		blocks[i] = binary.LittleEndian.Uint64(buf[i*8 : (i+1)*8])
	}
	return blocks
}

func DecodeCatalogItem(blocks []uint64) model.CatalogItem {
	buf := make([]byte, TotalBytes)
	for i := range blocks {
		binary.LittleEndian.PutUint64(buf[i*8:(i+1)*8], blocks[i])
	}

	offset := 0
	recordID := int64(binary.LittleEndian.Uint64(buf[offset:]))
	offset += 8

	item := model.CatalogItem{
		RecordID: recordID,
		ItemName: readString(buf[offset : offset+ItemNameBytes]),
	}
	offset += ItemNameBytes
	item.Category = readString(buf[offset : offset+CategoryBytes])
	offset += CategoryBytes
	item.PriceCents = binary.LittleEndian.Uint64(buf[offset:])
	item.PriceText = FormatPrice(item.PriceCents)
	offset += 8
	item.StockStatus = readString(buf[offset : offset+StockStatusBytes])
	offset += StockStatusBytes
	item.Merchant = readString(buf[offset : offset+MerchantBytes])
	offset += MerchantBytes
	item.UpdatedAt = readString(buf[offset : offset+UpdatedAtBytes])
	return item
}

func FormatPrice(cents uint64) string {
	return fmt.Sprintf("¥%d.%02d", cents/100, cents%100)
}

func writeUint64(dst []byte, value uint64) {
	binary.LittleEndian.PutUint64(dst[:8], value)
}

func writeString(dst []byte, value string) {
	copy(dst, truncateUTF8(value, len(dst)))
}

func readString(src []byte) string {
	return strings.TrimRight(string(src), "\x00")
}

func truncateUTF8(value string, limit int) []byte {
	if len(value) <= limit {
		return []byte(value)
	}
	result := make([]byte, 0, limit)
	for _, r := range value {
		size := utf8.RuneLen(r)
		if size < 0 || len(result)+size > limit {
			break
		}
		result = utf8.AppendRune(result, r)
	}
	return result
}
