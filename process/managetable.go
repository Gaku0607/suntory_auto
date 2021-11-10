package process

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/suntory/model"
	"github.com/Gaku0607/suntory/store"
)

type SouthMangeTable struct {
	smc model.SouthManagementCompute
}

func NewSouthMangeTable() *SouthMangeTable {
	s := &SouthMangeTable{}
	s.smc = model.Environment.SMC
	return s
}

func (mt *SouthMangeTable) ManageTable(c *augo.Context) {
	managesourc, _ := c.Get("managesourc")
	inventorysourc, _ := c.Get("Inventorysourc")

	ms := managesourc.(*excelgo.Sourc)
	is := inventorysourc.(*excelgo.Sourc)

	result, difference, err := mt.calculation(ms, is)
	if err != nil {
		c.AbortWithError(err)
		return
	}

	//檔名為時間戳取下合併
	time := filepath.Base(is.Path)

	path := filepath.Join(model.Result_Path, fmt.Sprintf(mt.smc.FileNameFormat, time))
	path = excelgo.CheckFileName(path)

	if err := store.Store.ExportManageFile(ms, is, result, difference, path); err != nil {
		c.AbortWithError(err)
	}
}

func (mt *SouthMangeTable) calculation(managesourc, inventorysourc *excelgo.Sourc) ([]interface{}, []interface{}, error) {

	//管理表商品碼欄
	managecodecol := managesourc.GetCol(mt.smc.ManageSourc.CodeSpan)
	//管理閥值欄
	mangepsccol := managesourc.GetCol(mt.smc.ManageSourc.PSCSpan)
	//庫存表商品碼欄
	inventorycodecol := inventorysourc.GetCol(mt.smc.InventorySourc.CodeSpan)
	//庫存欄位
	inventorycol := inventorysourc.GetCol(mt.smc.InventorySourc.InventorySpan)

	inventoryrows, err := inventorysourc.Transform(inventorysourc.FilterAll(inventorysourc.Rows))
	if err != nil {
		return nil, nil, err
	}

	managerows, err := managesourc.Transform(managesourc.Rows)
	if err != nil {
		return nil, nil, err
	}

	var result []interface{}
	var difference []interface{}

	var Nil excelgo.NilRow

	for _, mrow := range managerows {

		var count int  //數量
		var exist bool //品項庫存是否存在

		//該行為空白或為指定類型時跳過
		if len(mrow) == 0 || mrow[managecodecol.Col].(string) == "品番" {
			result = append(result, Nil)
			difference = append(difference, Nil)
			continue
		}

		code := mrow[managecodecol.Col].(string)

		//查詢商品的倍率表
		count, exist = mt.selectItemMagnification(inventorycol.Col, inventorycodecol.Col, inventoryrows, code)

		//字尾包含Ｐ的Code 去Ｐ之後查出來的值需要在除以3
		if !exist {
			if strings.HasSuffix(code, "P") {
				val := mt.find(inventorycol.Col, inventorycodecol.Col, inventoryrows, code[:len(code)-1])
				count = val / 3
			}

			val := mt.find(inventorycol.Col, inventorycodecol.Col, inventoryrows, code)
			count += val
		}

		//計算差額
		pscstr := mrow[mangepsccol.Col].(string)
		psc, err := strconv.Atoi(pscstr)
		if err != nil {
			return nil, nil, err
		}
		difference = append(difference, count-psc)

		//判斷該商品是庫存為0還是無品項
		if count == 0 && !mt.isItemExist(inventorycol.Col, inventorycodecol.Col, inventoryrows, code) {
			result = append(result, model.NilCell)
			continue
		}

		result = append(result, count)
	}

	return result, difference, nil
}

func (mt *SouthMangeTable) selectItemMagnification(inventorycol, inventorycodecol int, inventoryrows [][]interface{}, code string) (int, bool) {
	var val int
	//查詢商品的倍率表
	for _, tg := range mt.smc.ManageSourc.ItemComparisonTable {
		if code == tg.Main { //當有匹配對象
			for _, sb := range tg.Subs {
				count := mt.find(inventorycol, inventorycodecol, inventoryrows, sb.Sub)
				if sb.Isdivision {
					val += count / sb.Proportion
				} else {
					val += count * sb.Proportion
				}
			}
			return val, true
		}
	}
	return 0, false
}

func (mt *SouthMangeTable) find(inventorycol, inventorycodecol int, inventoryrows [][]interface{}, code string) int {
	var val int
	for _, irow := range inventoryrows {
		if irow[inventorycodecol].(string) == code {
			val += irow[inventorycol].(int)
		}
	}
	return val
}

func (mt *SouthMangeTable) isItemExist(inventorycol, inventorycodecol int, inventoryrows [][]interface{}, code string) bool {
	for _, tg := range mt.smc.ItemComparisonTable {
		if code == tg.Main {
			for _, sb := range tg.Subs {
				if exist := mt.isExist(inventorycol, inventorycodecol, inventoryrows, sb.Sub); exist {
					return exist
				}
			}
			return false
		}
	}
	return mt.isExist(inventorycol, inventorycodecol, inventoryrows, code)
}

func (mt *SouthMangeTable) isExist(inventorycol, inventorycodecol int, inventoryrows [][]interface{}, code string) bool {
	for _, irow := range inventoryrows {
		if irow[inventorycodecol].(string) == code {
			return true
		}
	}
	return false
}
