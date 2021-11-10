package model

import "github.com/Gaku0607/excelgo"

//**************************************
//***************EnvPaths***************
//**************************************

var EnvironmentDir string //配置文件資料夾

var Services_Dir string //服務資料夾

var Result_Path string //完成時返回地址

var EnvPath = "/Users/gaku/Documents/GitHub/sumtory_auto/main/.env"

//**************************************
//***************EnvSources*************
//**************************************

var Environment SystemParms

type SystemParms struct {
	SMC SouthManagementCompute `json:"south_management_compute"` //南部庫存管理表
	MMI MidMonthInventory      `json:"mid_month_inventory"`      // 月中庫存表
}

//***********Mid-MonthInventory**********

type MidMonthInventory struct {
	MidMonthCsvSourc       `json:"csv_sourc"`
	NewCols                []*excelgo.Col `json:"new_cols"`
	NewHeaders             []string       `json:"new_headers"`
	ExportHeaders          []interface{}  `json:"export_headers"`
	ValidityPeriodListPath string         `json:"validity_period_list_path"`
	NameListPath           string         `json:"name_list_path"`
	InventoryPath          string         `json:"-"`
	InventorySheet         string         `json:"Inventory_sheet"`
	FileNameFormat         string         `json:"file_name_format"`
	StopDateTcol           string         `json:"stop_date_tcol"`
	RemainDaysTcol         string         `json:"remain_days_tcol"`
}

type MidMonthCsvSourc struct {
	excelgo.Sourc `json:"sourc"`
	ItemCodeSpan  string `json:"itme_code_span"`
	TotalSpan     string `json:"total_span"`
	DateSpan      string `json:"date_span"`
}

//*************SouthManage*************

type SouthManagementCompute struct {
	ManageSourc          `json:"mange_sourc"`
	InventorySourc       `json:"inventory_sourc"`
	TotalPCSTCol         string `json:"total_pcs_tcol"`
	DifferenceTCol       string `json:"difference_tcol"`
	FileNameFormat       string `json:"file_name_format"`        //輸出後的檔名格式
	ManageFileNameForamt string `json:"manage_file_name_format"` //輸入時判斷管理表的依據
}

//管理
type ManageSourc struct {
	excelgo.Sourc
	CodeSpan            string `json:"code_span"`
	PSCSpan             string `json:"psc_span"`
	ItemComparisonTable []*ItemRatio
}

//庫存
type InventorySourc struct {
	excelgo.Sourc
	CodeSpan      string `json:"code_span"`
	InventorySpan string `json:"inventory_span"`
}

//庫存之間的倍率轉換表
type ItemRatio struct {
	Main string //itemcode
	Subs []*Sub
}

type Sub struct {
	Sub        string //itemcode
	Proportion int    // 倍率
	Isdivision bool   // 是否為除法 否為乘法
}
