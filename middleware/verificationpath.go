package middleware

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/suntory/model"
)

func VerificationPath(c *augo.Context) {
	switch c.Request.Method() {

	case model.MID_MONTH_INVENTORY_METHOD:
		xlsxpath, csvpath, err := checkFileCount(c.Request, 1, 1)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		model.Environment.MMI.InventoryPath = xlsxpath[0]
		s, err := getSourc(csvpath[0], c.Request.Method())
		if err != nil {
			c.AbortWithError(err)
			return
		}

		c.Set("sourc", s)

	case model.SOUTH_MANAGE_TABLE_METHOD:
		xlsxpaths, _, err := checkFileCount(c.Request, 0, 2)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		sfs, err := getSources(xlsxpaths, c.Request.Method())
		if err != nil {
			c.AbortWithError(err)
			return
		}

		//Manage表
		c.Set("managesourc", sfs[0])
		//Inventory表
		c.Set("Inventorysourc", sfs[1])
	}
}

//確認是否為指定的檔名
func checkFileCount(req *augo.Request, csvcount, xlsxcount int) ([]string, []string, error) {

	var (
		cvstotal  int
		xlsxtotal int
		xlsxpaths []string
		csvpath   []string
	)

	for i := 0; i < len(req.Files); i++ {

		switch filepath.Ext(req.Files[i]) {
		case ".xlsx":
			xlsxpaths = append(xlsxpaths, req.Files[i])
			xlsxtotal++
		case ".csv":
			csvpath = append(csvpath, req.Files[i])
			cvstotal++
		case ".DS_Store":
			continue
		default:
			return nil, nil, fmt.Errorf("%s Incorrect file format", req.Files[i])
		}
	}

	if csvcount != cvstotal && csvcount != -1 {
		return nil, nil, errors.New("Too many .csv files")
	}
	if xlsxcount != xlsxtotal && xlsxcount != -1 {
		return nil, nil, errors.New("Too many .xlsx files")
	}

	return xlsxpaths, csvpath, nil
}

func getSources(paths []string, method string) ([]*excelgo.Sourc, error) {
	switch method {
	case model.MID_MONTH_INVENTORY_METHOD:
		s, err := getSourc(paths[0], method)
		if err != nil {
			return nil, err
		}
		return []*excelgo.Sourc{s}, nil

	case model.SOUTH_MANAGE_TABLE_METHOD:
		var managesourc, inventorysourc excelgo.Sourc
		var s *excelgo.Sourc
		for _, path := range paths {
			//判斷有無管理表
			if strings.Contains(filepath.Base(path), model.Environment.SMC.ManageFileNameForamt) {
				managesourc = model.Environment.SMC.ManageSourc.Sourc
				s = &managesourc
			} else {
				inventorysourc = model.Environment.SMC.InventorySourc.Sourc
				s = &inventorysourc
			}

			f, err := excelgo.OpenFile(path)
			if err != nil {
				return nil, err
			}
			if err := s.Init(f); err != nil {
				return nil, err
			}
		}

		if &managesourc == nil || &inventorysourc == nil {
			return nil, errors.New("Input file type is wrong")
		}

		return []*excelgo.Sourc{&managesourc, &inventorysourc}, nil

	default:
		return nil, fmt.Errorf("input method:=%s, is not format", method)
	}
}

func getSourc(path, method string) (*excelgo.Sourc, error) {

	switch method {
	case model.MID_MONTH_INVENTORY_METHOD:
		s := model.Environment.MMI.MidMonthCsvSourc.Sourc
		f, err := excelgo.OpenFile(path)
		if err != nil {
			return nil, err
		}
		return &s, s.Init(f)
	default:
		return nil, fmt.Errorf("input method:=%s, is not format", method)
	}
}
