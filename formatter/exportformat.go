package formatter

import (
	"time"

	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/suntory/model"
)

func SetExportFormat() {

	excelgo.FormatCategory.SetFormatCategory(
		model.MID_MONTH_INVENTORY_METHOD,
		"Sheet1",
		"F",
		func(i interface{}) interface{} {
			t := i.(string)
			time, _ := time.Parse("20060102", t)
			return time.Format("2006/01/02")
		},
	)
}
