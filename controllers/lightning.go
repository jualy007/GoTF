package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/jualy007/GoTF/blockchain/lighting"
	"github.com/jualy007/GoTF/config"
	"github.com/jualy007/GoTF/models"
)

type LightningController struct {
	beego.Controller
}

func (this *LightningController) Prepare() {
	//Update Avaliable Lnd Node
	models.Healthcheck()

	if models.Livenode.Size() == 0 {
		this.Data["status"] = 500
		this.Data["data"] = nil
		this.Data["msg"] = "No avaliable Lnd nodes"
		this.ServeJSON()
		this.StopRun()
	}
}

// @Title GetInfo
// @Description Get Lnd Info
// @Param	LndName		path 	string	true		"The Lnd Name which we want get full info"
// @Success 200 {object}
// @Failure 403 : Lnd not exist for The Lnd Name
// @router /:lndname [get]
func (this *LightningController) Info() {
	lndname := this.Ctx.Input.Param(":lndname")

	if models.Livenode.Contains(lndname) {
		lndinfo := config.Cfg.Lnds[lndname]
		adapter, err := lighting.NewAdapter(lndinfo)

		if err != nil {
			this.Data["status"] = 403
			this.Data["msg"] = err.Error()
			this.Data["data"] = nil
		} else {
			info, err := adapter.GetInfo()

			if err != nil {
				this.Data["status"] = 403
				this.Data["msg"] = err.Error()
				this.Data["data"] = nil
			} else {
				this.Data["status"] = 200
				this.Data["msg"] = ""
				this.Data["json"] = &info
			}
		}
	} else {
		this.Data["status"] = 403
		this.Data["msg"] = fmt.Sprintf("No avaliable lnd node %v", lndname)
		this.Data["data"] = nil
	}

	this.ServeJSON()
}

// @Title QueryRoutes
// @Description Query Routes Info
// @Param	PayRequests		path 	string	true		"The LND Payment Requests"
// @Success 200 {object}
// @Failure 403 : Decode Failed
// @router /routes/:pay_req [get]
func (this *LightningController) QueryRoutes() {
	pay_req := this.Ctx.Input.Param(":pay_req")

	lndinfo := config.Cfg.Lnds["atoken-lightning-01"]
	adapter, err := lighting.NewAdapter(lndinfo)

	if err != nil {
		this.Data["status"] = 403
		this.Data["msg"] = err.Error()
		this.Data["data"] = nil
	} else {
		info := adapter.QueryRoute(pay_req)
		this.Data["status"] = 200
		this.Data["msg"] = ""
		this.Data["json"] = &info
	}

	this.ServeJSON()
}

// @Title SendPayment
// @Description SendPayment
// @Param	uid		path 	string	true		"The uid you want to update"
// @Param	body	body 	models.User	true		"body for user content"
// @Success 200 {object} models.User
// @Failure 403 :uid is not int
// @router /:uid [put]
func (this *LightningController) SP() {

	adapter, err := lighting.NewAdapter(lndinfo)
}
