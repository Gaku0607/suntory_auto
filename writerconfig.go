package suntory

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/suntory/model"
	"github.com/joho/godotenv"
)

func WriterConfig() error {

	if err := godotenv.Load(model.EnvPath); err != nil {
		return err
	}
	if err := loadPath(); err != nil {
		return err
	}

	s := &model.SystemParms{}

	if err := mid_month_inventory(&s.MMI); err != nil {
		return err
	}

	if err := manage_goods_parms(&s.SMC); err != nil {
		return err
	}

	return nil
}

func writerconfig(filename string, data []byte) error {
	return ioutil.WriteFile(filepath.Join(model.EnvironmentDir, filename), data, 0777)
}

func mid_month_inventory(sm *model.MidMonthInventory) error {

	const (
		originsheetname = "Sheet1"
	)

	tcolfn := func(sheet, header string) *excelgo.TargetCol {
		return excelgo.NewTCol(model.MID_MONTH_INVENTORY_METHOD, sheet, header)
	}

	//MidMonthCsvSourc
	mms := &model.MidMonthCsvSourc{}
	{
		mms.ItemCodeSpan = "Product code"
		codecol := excelgo.NewCol(mms.ItemCodeSpan)
		codecol.TCol = []*excelgo.TargetCol{tcolfn(originsheetname, "A")}

		palletNocol := excelgo.NewCol("Pallet no")
		palletNocol.TCol = []*excelgo.TargetCol{tcolfn(originsheetname, "B")}

		pickingkyecol := excelgo.NewCol("Picking kye")
		pickingkyecol.TCol = []*excelgo.TargetCol{tcolfn(originsheetname, "D")}

		mms.DateSpan = "Cust ord Ref"
		custordref := excelgo.NewCol(mms.DateSpan)
		custordref.Filter = excelgo.Filter{IsTarget: false, Target: []string{"00000000", "0"}}
		custordref.TCol = []*excelgo.TargetCol{tcolfn(originsheetname, "E"), tcolfn(originsheetname, "F")}

		mms.TotalSpan = "Qty."
		totalcol := excelgo.NewCol(mms.TotalSpan)
		totalcol.Numice = excelgo.Numice{IsNumice: true}
		totalcol.TCol = []*excelgo.TargetCol{tcolfn(originsheetname, "G")}

		categorycol := excelgo.NewCol("Category-1")
		categorycol.Filter = excelgo.Filter{IsTarget: true, Target: []string{"NM"}}
		mms.Sourc = *excelgo.NewSourc(
			originsheetname,
			codecol,
			palletNocol,
			pickingkyecol,
			custordref,
			totalcol,
			categorycol,
		)
		mms.Sourc.Formulas = excelgo.Formulas{
			excelgo.NewFormula(model.MID_MONTH_INVENTORY_METHOD, `=TEXT(E%d,"0000""/""00""/""00")`, originsheetname, "F"),
			excelgo.NewFormula(model.MID_MONTH_INVENTORY_METHOD, `=F%d-H%d`, originsheetname, "I"),
			excelgo.NewFormula(model.MID_MONTH_INVENTORY_METHOD, "I%d-$J$1", originsheetname, "J"),
		}
	}

	sm.MidMonthCsvSourc = *mms
	sm.RemainDaysTcol = "J"
	sm.StopDateTcol = "I"
	sm.InventorySheet = "庫存彙總表"
	sm.ValidityPeriodListPath = "/Users/gaku/suntory_test/config/商品安全日.xlsx"
	sm.NameListPath = "/Users/gaku/suntory_test/config/商品中文名.xlsx"
	sm.NewHeaders = []string{"中文名稱", "安全庫存日"}

	namecol := excelgo.NewCol("中文名稱")
	namecol.TCol = []*excelgo.TargetCol{tcolfn(originsheetname, "C")}

	datecol := excelgo.NewCol("安全庫存日")
	datecol.TCol = []*excelgo.TargetCol{tcolfn(originsheetname, "H")}

	sm.NewCols = []*excelgo.Col{namecol, datecol}
	sm.ExportHeaders = []interface{}{"商品CODE", "Pallet no", "中文名稱", "Picking kye", "Cust ord Ref", "賞味期限", "加總-Qty.", "安全庫存日", "出荷停止日"}
	sm.FileNameFormat = "%s %s 月中庫存.xlsx"
	data, err := json.Marshal(sm)
	if err != nil {
		return err
	}
	return writerconfig(model.MID_MONTH_INVENTORY_BASE, data)
}

func manage_goods_parms(smc *model.SouthManagementCompute) error {

	const (
		manage_sheet    = "台南基準在庫値(商品+販促物)"
		inventory_sheet = "庫存彙總表"
	)

	smc.TotalPCSTCol = "G"
	smc.DifferenceTCol = "H"

	smc.ManageFileNameForamt = "管理表"
	smc.FileNameFormat = "南部在庫管理-%s"

	//inventorySourc
	{
		smc.InventorySourc.InventorySpan = "結存"
		inventory := excelgo.NewCol(smc.InventorySourc.InventorySpan)
		inventory.Numice = excelgo.Numice{IsNumice: true}

		smc.InventorySourc.CodeSpan = "商品型號"
		code := excelgo.NewCol(smc.InventorySourc.CodeSpan)

		book := excelgo.NewCol("帳面")
		book.Filter = excelgo.Filter{IsTarget: true, Target: []string{"P100/S", "P104/S"}}
		sourc := excelgo.NewSourc(
			inventory_sheet,
			book,
			inventory,
			code,
		)
		smc.InventorySourc.Sourc = *sourc
	}

	//ManageSourc
	{

		smc.ManageSourc.CodeSpan = "品番"
		codecol := excelgo.NewCol(smc.ManageSourc.CodeSpan)
		codecol.Impurity = excelgo.Impurity{IsSplit: true, Contains: []string{"'"}}

		smc.ManageSourc.PSCSpan = "閾値(pcs)"
		psccol := excelgo.NewCol(smc.ManageSourc.PSCSpan)

		sourc := excelgo.NewSourc(
			manage_sheet,
			codecol,
			psccol,
		)

		sourc.Formulas = []*excelgo.Formula{excelgo.NewFormula(model.MID_MONTH_INVENTORY_BASE, "=ROUNDUP((G%d-F%d)/D%d,0)", manage_sheet, "I")}

		smc.ManageSourc.Sourc = *sourc

		item1 := &model.ItemRatio{
			Main: "VAS006",
			Subs: []*model.Sub{
				{
					Sub:        "52106",
					Isdivision: true,
					Proportion: 6,
				},
			},
		}
		smc.ManageSourc.ItemComparisonTable = []*model.ItemRatio{item1}

	}

	data, err := json.Marshal(&smc)
	if err != nil {
		return err
	}

	return writerconfig(model.SOUTH_MANAGE_TABLE_BASE, data)
}
