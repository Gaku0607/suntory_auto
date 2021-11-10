package main

import (
	"time"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/suntory"
	"github.com/Gaku0607/suntory/formatter"
	"github.com/Gaku0607/suntory/routers"
)

func main() {
	// sumtory.WriterConfig()

	//發生致命錯誤時打印並結束程序
	go suntory.RecoveryPrint()

	//讀取所有配置
	if err := suntory.LoadEnvironment(); err != nil {
		suntory.ErrChan <- err
	}

	//建立服務資料夾
	if err := routers.MakeServiceRouters(); err != nil {
		suntory.ErrChan <- err
	}

	//設置各個服務匯出時的格式
	formatter.SetExportFormat()

	//設置環境
	augo.SetSystemVersion(augo.MacOS)
	//設置服務Title
	augo.SetLogTitle("SUNTORY")

	c := augo.DefautCollector(
		augo.ResultLogKey(func(c *augo.Context) augo.LogKey {
			return map[string]interface{}{}
		}),
	)
	//註冊所有服務路由
	routers.Routers(c)

	engine := augo.NewEngine(
		augo.MaxThread(2),
		augo.ScanIntval(time.Second*2),
		augo.SetCollector(c),
	)

	engine.Run()

	engine.Wait()

	if err := engine.Close(); err != nil {
		panic(err.Error())
	}
}
