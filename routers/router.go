// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"github.com/astaxie/beego"
	"github.com/jualy007/GoTF/controllers"
)

func init() {
	ns := beego.NewNamespace("/v1",
		//beego.NSCond(func(ctx *beecontext.Context) bool {
		//	//if ctx.Input.Domain() == "api.beego.me" {
		//	//	return true
		//	//}
		//	//return false
		//	return true
		//}),
		//beego.NSBefore(func(ctx *beecontext.Context) {
		//	// The Authorization header should come in this format: Bearer <jwt>
		//	// The first thing we do is check that the JWT exists
		//	header := strings.Split(ctx.Input.Header("Authorization"), " ")
		//	if header[0] != "Bearer" {
		//		ctx.Abort(401, "Not authorized")
		//	}
		//}),
		//beego.NSBefore(func(ctx *beecontext.Context) {
		//	ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
		//	ctx.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS")
		//	ctx.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Origin,ContentType,Authorization,accept,accept-encoding, authorization, content-type")
		//	ctx.ResponseWriter.Header().Set("Access-Control-Max-Age", "1728000")
		//	ctx.ResponseWriter.Header().Set("Access-Control-Allow-Credentials", "true")
		//}),

		beego.NSRouter("*", &controllers.BaseController{}, "OPTIONS:Options"),

		beego.NSNamespace("/object",
			beego.NSInclude(
				&controllers.ObjectController{},
			),
		),
		beego.NSNamespace("/user",
			beego.NSInclude(
				&controllers.UserController{},
			),
		),
		beego.NSNamespace("/lightning",
			beego.NSInclude(
				&controllers.LightningController{},
			),
		),
		beego.NSNamespace("/debug/pprof",
			beego.NSInclude(
				&controllers.ProfController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
