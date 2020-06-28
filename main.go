package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/jualy007/GoTF/models"
	"os"
	"path"

	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/config/yaml"
	"github.com/astaxie/beego/logs"
	lconfig "github.com/jualy007/GoTF/config"
	_ "github.com/jualy007/GoTF/routers"
	"github.com/urfave/cli/v2"
)

//go:generate sh -c "echo 'package routers; import \"github.com/astaxie/beego\"; func init() {beego.BConfig.RunMode = beego.DEV}' > routers/0.go"
//go:generate sh -c "echo 'package routers; import \"os\"; func init() {os.Exit(0)}' > routers/z.go"
//go:generate go run $GOFILE
//go:generate sh -c "rm routers/0.go routers/z.go"
func main() {
	app := cli.NewApp()

	app.Name = "GoTF"
	app.Version = "1.0.0"
	app.Usage = "GoTF Options"
	app.Description = "GO Test Framework"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "configdir",
			Usage:    "Configuration Files Directory",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "logdir",
			Usage:    "Log Directory",
			Required: true,
		},
	}

	app.Before = func(cctx *cli.Context) error {
		fmt.Println("Before Start......")
		return nil
	}

	app.Action = func(cctx *cli.Context) error {
		// Load Customer Beego Configurations
		bconf := path.Join(cctx.String("configdir"), "beego.yaml")
		err := beego.LoadAppConfig("yaml", bconf)

		if err != nil {
			fmt.Printf("Load Config Error!!! %s\n", err)
		}

		// Load Application Configurations
		aconf := path.Join(cctx.String("configdir"), "application.yaml")
		ctx := context.WithValue(context.Background(), "configfile", aconf)

		lconfig.GetCfg(ctx)

		//Init logs
		config := make(map[string]interface{})
		logfile := fmt.Sprintf("%v.log", beego.AppConfig.String("AppName"))
		logpath := path.Join(cctx.String("logdir"), logfile)
		config["filename"] = logpath
		config["level"] = logs.LevelDebug
		config["maxlines"] = 1000000
		config["maxfiles"] = 7
		config["maxdays"] = 30
		configJson, err := json.Marshal(config)
		if err != nil {
			fmt.Println("ERROR: Log config format is not correct!!!")
		}
		log := logs.NewLogger()
		log.SetLogger(logs.AdapterFile, string(configJson))

		//Init DB
		maxIdle := 30
		maxConn := 30
		err = orm.RegisterDataBase("default", "mysql", "root:root@tcp(127.0.0.1:3306)/test?charset=utf8", maxIdle, maxConn)

		if err != nil {
			fmt.Println("ERROR: Connect Database Error : ", err)
		}

		// 注册定义的model
		orm.RegisterModel(new(models.User))

		// 自动建表
		orm.RunSyncdb("default", false, true)

		beego.InsertFilter("/*", beego.BeforeExec, controllers.BeforeExecFilter, true, false)
		beego.InsertFilter("/*", beego.BeforeStatic, controllers.BeforeStaticFilter, true, false)
		beego.InsertFilter("/*", beego.AfterExec, controllers.AfterExecFilter, true, false)
		beego.InsertFilter("/*", beego.FinishRouter, controllers.FinishRouterFilter, true, false)

		if beego.BConfig.RunMode == "dev" {
			beego.BConfig.WebConfig.DirectoryIndex = true
			beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
		}
		beego.Run()
		return nil
	}

	app.After = func(cctx *cli.Context) error {
		fmt.Println("Start End......")
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Run Error!!! %s\n", err)
	}
}
