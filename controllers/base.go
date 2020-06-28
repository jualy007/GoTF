package controllers

import "github.com/astaxie/beego"

type BaseController struct {
	beego.Controller
}

// @Title Support CORS
// @Description Support CORS
// @Success 200 {string} "OK"
// @router / [options]
func (this *BaseController) Options() {
	this.Data["json"] = map[string]interface{}{"status": 200, "message": "ok", "data": ""}
	this.ServeJSON()
}
