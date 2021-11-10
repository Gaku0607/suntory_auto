package routers

import (
	"os"
	"path/filepath"

	"github.com/Gaku0607/suntory/model"
	"github.com/Gaku0607/suntory/tool"
)

func MakeServiceRouters() error {

	//mid-month-inventory
	if err := MakeServiceRouter(model.MID_MONTH_INVENTORY_METHOD); err != nil {
		return err
	}

	//south-manage-table
	if err := MakeServiceRouter(model.SOUTH_MANAGE_TABLE_METHOD); err != nil {
		return err
	}

	//返回檔案地址
	if exist := tool.IsExist(model.Result_Path); !exist {
		if err := os.MkdirAll(model.Result_Path, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func MakeServiceRouter(method string) error {
	path := absoluteServicePath(method)
	if exist := tool.IsExist(path); !exist {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func absoluteServicePath(method string) string {
	return filepath.Join(model.Services_Dir, method)
}
