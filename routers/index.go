package routers

import (
	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/suntory/middleware"
	"github.com/Gaku0607/suntory/model"
	"github.com/Gaku0607/suntory/process"
)

func Routers(c *augo.Collector) {

	c.Use(middleware.VerificationPath)

	{
		//mid-month-inventory
		c.Handler(absoluteServicePath(model.MID_MONTH_INVENTORY_METHOD), process.NewMonthInventoryTable().MonthInventory)

		//southern-mnagement-table
		c.Handler(absoluteServicePath(model.SOUTH_MANAGE_TABLE_METHOD), process.NewSouthMangeTable().ManageTable)
	}

}
