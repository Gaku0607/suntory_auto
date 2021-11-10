package store

import (
	"fmt"
	"strconv"

	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/suntory/model"
	"github.com/xuri/excelize/v2"
)

var Store *ExcelStore

const (
	NEW_SHEET = "Sheet1"
)

type ExcelStore struct {
}

//匯出月中表
func (e *ExcelStore) ExportMidMonthInventory(s *excelgo.Sourc, headers []interface{}, rows [][]interface{}, path string, remaindaytcol int) error {
	f := excelize.NewFile()
	sheetid := f.NewSheet(NEW_SHEET)

	orginstyle, err := f.NewStyle(&excelgo.Origin_Style)
	if err != nil {
		return err
	}

	//負數style
	negativestyle, err := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{excelgo.Bottom, excelgo.Left, excelgo.Right, excelgo.Top},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#f6b4af"}},
		Alignment: excelgo.Alignment_Center,
		Font:      excelgo.Font,
	})

	if err != nil {
		return err
	}

	//低於150style
	riskstyle, err := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{excelgo.Bottom, excelgo.Left, excelgo.Right, excelgo.Top},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#ffeb7f"}},
		Alignment: excelgo.Alignment_Center,
		Font:      excelgo.Font,
	})

	if err != nil {
		return err
	}

	index := 1
	//開始位置
	start := e.getaddr("A", index)
	//結束位置
	header := excelgo.ConvertToLetter(len(headers)) //row末端標頭
	end := e.getaddr(header, index)

	//設置欄位頭
	e.addrow(f, sheetid, start, headers)

	if err := f.SetCellStyle(NEW_SHEET, start, end, orginstyle); err != nil {
		return err
	}

	for _, row := range rows {
		index++
		day := row[remaindaytcol-1].(int)
		style := orginstyle
		start = e.getaddr("A", index)
		end = e.getaddr(header, index)
		if day < 0 {
			style = negativestyle
		} else if day < 150 {
			style = riskstyle
		}

		if err := f.SetCellStyle(NEW_SHEET, start, end, style); err != nil {
			return err
		}

		e.addrow(f, sheetid, start, row)
	}

	index += 1
	//設置Totalformula
	if err := e.addCell(f, f.GetSheetName(sheetid), e.getaddr("A", index), "總共"); err != nil {
		return err
	}

	totalcol := "G"
	if err := e.addCellFormula(f, f.GetSheetName(sheetid), e.getaddr(totalcol, index), fmt.Sprintf("SUM(%s2:%s%d)", totalcol, totalcol, index-1)); err != nil {
		return err
	}

	if err := f.SetCellStyle(NEW_SHEET, e.getaddr("A", index), e.getaddr(totalcol, index), orginstyle); err != nil {
		return err
	}

	return f.SaveAs(path)
}

//更新MasterFile表
func (e *ExcelStore) UpdateMasterTable(f *excelize.File, start int, newitems map[string][]string) error {

	var idx int

	for code, items := range newitems {

		row := make([]interface{}, len(items)+1)
		row[0] = code

		for i, item := range items {
			row[i+1] = item
		}

		e.addrow(f, 0, "A"+strconv.Itoa(start+idx), row)
		idx++
	}
	return f.Save()
}

//南部庫存管理表
func (e *ExcelStore) ExportManageFile(manageysourc, inventorysourc *excelgo.Sourc, result, difference []interface{}, path string) error {

	totaltcol := model.Environment.SMC.TotalPCSTCol
	differenctcol := model.Environment.SMC.DifferenceTCol

	file, err := excelize.OpenFile(manageysourc.Path)
	if err != nil {
		return err
	}

	//負數style
	negativenumStyle, err := file.NewStyle(&excelize.Style{
		Border:    excelgo.Origin_Style.Border,
		Alignment: excelgo.Alignment_Right,
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#f9c9c6"}},
		Font:      &excelize.Font{Family: "Meiryo UI", Color: "#c32316"},
	})
	if err != nil {
		return err
	}

	originstyle, err := file.NewStyle(&excelize.Style{
		Border:    excelgo.Origin_Style.Border,
		Alignment: excelgo.Alignment_Right,
		Fill:      excelgo.Fill,
		Font:      &excelize.Font{Family: "Meiryo UI"},
	})

	if err != nil {
		return err
	}
	for i, val := range result {
		//開始高度
		index := i + 2
		style := originstyle
		//非結構體累型實際路紀錄欄位
		if _, ok := val.(excelgo.NilRow); !ok {

			if err := e.addCell(file, manageysourc.SheetName, e.getaddr(totaltcol, index), val); err != nil {
				return err
			}
			//差額小於0
			if difference[i].(int) < 0 {
				style = negativenumStyle
			}

			if err := file.SetCellStyle(manageysourc.SheetName, e.getaddr(differenctcol, index), e.getaddr(differenctcol, index), style); err != nil {
				return err
			}

			if err := e.addCell(file, manageysourc.SheetName, e.getaddr(differenctcol, index), difference[i]); err != nil {
				return err
			}

			for _, f := range manageysourc.Formulas {
				e.addCellFormula(file, f.TSheet, e.getaddr(f.TColStr, index), fmt.Sprintf(f.FormulaStr, index, index, index))
			}
		}
	}
	return file.SaveAs(path)
}

//以Tcol的方式匯入欄位
func (e *ExcelStore) addCell(f *excelize.File, sheet, addr string, val interface{}) error {
	return f.SetCellValue(sheet, addr, val)
}

//對內容進行全部匯出
func (e *ExcelStore) Export(path string, data [][]interface{}) error {
	nf := excelize.NewFile()
	Id := nf.NewSheet(NEW_SHEET)
	if err := e.exportExcel(nf, Id, data); err != nil {
		return err
	}
	return nf.SaveAs(path)
}

//已行的方式寫入檔案
func (e *ExcelStore) addrow(f *excelize.File, sheetid int, addr string, row []interface{}) {
	f.SetSheetRow(f.GetSheetName(sheetid), addr, &row)
}

//已Excel的格式匯出
func (e *ExcelStore) exportExcel(nf *excelize.File, sheetId int, rows [][]interface{}) error {
	for i, row := range rows {
		e.addrow(nf, sheetId, "A"+strconv.Itoa(i+1), row)
	}
	return nil
}

//設置公式
func (e *ExcelStore) addCellFormula(file *excelize.File, sheet, addr, formula string) error {
	return file.SetCellFormula(sheet, addr, formula)
}

//修改Sheet的定義名稱
func (e *ExcelStore) modityDefinedName(file *excelize.File, name *excelize.DefinedName) error {

	if err := file.DeleteDefinedName(&excelize.DefinedName{Scope: name.Scope, Name: name.Name}); err != nil {
		return err
	}
	return file.SetDefinedName(name)
}

func (e *ExcelStore) insertRows(count, Position int, file *excelize.File, sheet string) error {

	for i := 0; i < count; i++ {
		if err := file.DuplicateRow(sheet, Position); err != nil {
			return err
		}
	}

	return nil
}

func (e *ExcelStore) CreateSheet(file *excelize.File, name string) int {
	return e.createSheet(file, name, name, 0)
}

func (e *ExcelStore) ChangeSheetSort(file *excelize.File, sheetId, to int) {
}

func (e *ExcelStore) createSheet(file *excelize.File, name string, base string, id int) int {
	sheetid := 0
	if sheetid = file.NewSheet(name); sheetid == 0 {
		return e.createSheet(file, fmt.Sprintf("%s(%d)", base, id), base, id+1)
	}
	return sheetid
}

func (e *ExcelStore) getaddr(header string, index int) string {
	return header + strconv.Itoa(index)
}
