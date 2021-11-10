package process

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/suntory/model"
	"github.com/Gaku0607/suntory/store"
	"github.com/Gaku0607/suntory/tool"
	"github.com/xuri/excelize/v2"
)

var InventoryData [][]string

type MonthInventoryTable struct {
	mmi            model.MidMonthInventory
	middate        string //月中的基準日期
	dateCol        int    //商品日期
	stopDateTcol   int    //出貨停止日的目標欄位
	remaindaysTcol int    //離出貨停止日的相對天數
}

func NewMonthInventoryTable() *MonthInventoryTable {
	m := &MonthInventoryTable{}
	m.mmi = model.Environment.MMI
	m.stopDateTcol = excelgo.TwentysixToTen(m.mmi.StopDateTcol)
	m.remaindaysTcol = excelgo.TwentysixToTen(m.mmi.RemainDaysTcol)
	return m
}

func (mi *MonthInventoryTable) MonthInventory(c *augo.Context) {
	sourc, _ := c.Get("sourc")

	s := sourc.(*excelgo.Sourc)

	headers, rows, err := mi.monthinventory(s)
	if err != nil {
		c.AbortWithError(err)
		return
	}

	//排序
	excelgo.Sort(rows, mi.remaindaysTcol-1, excelgo.ReverseOrder)

	t := time.Now()

	name := fmt.Sprintf(mi.mmi.FileNameFormat, t.Format("2006"), t.Format("01"))

	path := excelgo.CheckFileName(filepath.Join(model.Result_Path, name))

	if err := store.Store.ExportMidMonthInventory(s, headers, rows, path, mi.remaindaysTcol); err != nil {
		c.AbortWithError(err)
	}
}

func (mi *MonthInventoryTable) monthinventory(s *excelgo.Sourc) ([]interface{}, [][]interface{}, error) {

	var newitem map[string][]string = make(map[string][]string)
	var newdateitem map[string][]string = make(map[string][]string)

	//進行篩選後格式化
	rows, err := s.Transform(s.FilterAll(s.Rows))
	if err != nil {
		return nil, nil, err
	}

	nl, nf, err := mi.nameList()
	if err != nil {
		return nil, nil, err
	}

	vl, df, err := mi.validityPeriodList()
	if err != nil {
		return nil, nil, err
	}

	itemcol := s.GetCol(mi.mmi.ItemCodeSpan)
	mi.dateCol = s.GetCol(mi.mmi.DateSpan).Col

	//匯出用的Headers
	var headers []interface{} = mi.mmi.ExportHeaders

	//添加日期Headers
	if len(rows) > 0 && len(rows[0]) > 0 {
		mi.middate = rows[0][0].(string)
		headers = append(headers, mi.middate)
	}

	//新增欄位
	s.Cols = append(s.Cols, mi.mmi.NewCols...)
	if err := s.ResetCol(append(s.Headers, mi.mmi.NewHeaders...)); err != nil {
		return nil, nil, err
	}

	for i, row := range rows {

		itemcode := row[itemcol.Col].(string)

		//當商品碼長度小於五並為數字時變更itemcode
		if len(itemcode) < 5 && tool.IsNumeric(itemcode) {

			zerocount := 5 - len(itemcode)
			newcode := ""

			for i := 0; i < zerocount; i++ {
				newcode += "0"
			}

			rows[i][itemcol.Col] = newcode + itemcode
			itemcode = rows[i][itemcol.Col].(string)
		}

		name, exist := nl[itemcode]
		if !exist {

			if InventoryData == nil {
				file, err := excelgo.OpenFile(mi.mmi.InventoryPath)
				if err != nil {
					return nil, nil, err
				}

				s := excelgo.NewSourc(mi.mmi.InventorySheet)
				if err := s.Init(file); err != nil {
					return nil, nil, err
				}

				InventoryData = s.Rows
			}

			for _, row := range InventoryData {
				if itemcode == row[2] { //InventoryData index 2 為商品code 3 為中文名稱
					name = row[3]
					newitem[itemcode] = []string{name} //新增至庫存
					break
				}
			}
			if name == "" {
				return nil, nil, fmt.Errorf("Check the item without code: %s", itemcode)
			}
		}

		safedays, exist := vl[itemcode]

		if !exist {
			safedays = model.NilCell
			newdateitem[itemcode] = []string{name, safedays}
		}

		rows[i], err = mi.format(s, append(row, name, safedays), safedays, row[mi.dateCol].(string))
		if err != nil {
			return nil, nil, err
		}
	}

	//更新中文名稱表
	if len(newitem) != 0 {
		if err := store.Store.UpdateMasterTable(nf, len(nl)+1, newitem); err != nil {
			return nil, nil, err
		}
	}

	//更新安全庫存表
	if len(newdateitem) != 0 {
		if err := store.Store.UpdateMasterTable(df, len(vl)+1, newdateitem); err != nil {
			return nil, nil, err
		}
	}

	InventoryData = nil

	return headers, rows, err
}

//中文名對照表
func (mi *MonthInventoryTable) nameList() (map[string]string, *excelize.File, error) {
	f, err := excelize.OpenFile(mi.mmi.NameListPath)
	if err != nil {
		return nil, nil, err
	}
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, nil, err
	}

	namelist := make(map[string]string, 0)
	for i, row := range rows {
		if len(row) < 2 {
			return nil, nil, fmt.Errorf("Error comes from the NameList %d line", i+1)
		}
		namelist[row[0]] = row[1]
	}
	return namelist, f, err
}

//安全庫存表
func (mi *MonthInventoryTable) validityPeriodList() (map[string]string, *excelize.File, error) {
	f, err := excelize.OpenFile(mi.mmi.ValidityPeriodListPath)
	if err != nil {
		return nil, nil, err
	}
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, nil, err
	}

	validitylist := make(map[string]string, 0)
	for i, row := range rows {
		if len(row) < 3 {
			return nil, nil, fmt.Errorf("Error comes from the ValidityPeriodList %d line", i+1)
		}
		validitylist[row[0]] = row[2]
	}

	return validitylist, f, err
}

func (mi *MonthInventoryTable) format(s *excelgo.Sourc, oldrow []interface{}, safedays, expirydate string) ([]interface{}, error) {

	if safedays == model.NilCell {
		safedays = "0" //當安全日查無時設為0方便計算
	}

	days, err := strconv.Atoi(safedays) //安全停止日
	if err != nil {
		return nil, err
	}

	row := make([]interface{}, len(mi.mmi.ExportHeaders)+1)
	fn := s.IteratorByTCol()

	for {
		tcol, exist := fn()
		if !exist {
			break
		}

		row[tcol.TCol-1] = tcol.Format(oldrow[tcol.ParentCol.Col])
	}

	t, _ := time.Parse("20060102", expirydate)
	m, _ := time.ParseDuration("-24h")

	stopday := t.Add(m * time.Duration(days)) //有效期限-安全停止日 判斷商品出貨停止日期
	row[mi.stopDateTcol-1] = stopday.Format("2006/01/02")

	t, _ = time.Parse("2006-01-02", mi.middate)
	row[mi.remaindaysTcol-1] = int(stopday.Sub(t).Hours() / 24) //以當月份月中為基準 判斷離停止日期還有幾天

	return row, nil
}
