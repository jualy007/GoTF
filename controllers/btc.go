package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type IndexController struct {
	beego.Controller
}

type UserDb struct {
	Id         int    `json:"id" orm:"column(id)"`
	Account    string `json:"account" orm:"column(account)"`
	Password   string `json:"password" orm:"column(password)"`
	CreateTime int    `json:"create_time" orm:"column(create_time)"`
	UpdateTime int    `json:"update_time" orm:"column(update_time)"`
	Status     int    `json:"status"  orm:"column(status)"`
	//json标签意义是定义此结构体解析为json或序列化输出json时value字段对应的key值,如不想此字段被解析可将标签设为`json:"-"`
	//orm的column标签意义是将orm查询结果解析到此结构体时每个结构体字段对应的数据表字段名
}

func (c *IndexController) Test() {
	var user []UserDb
	o := orm.NewOrm()
	_, err := o.Raw("SELECT * FROM user").QueryRows(&user) //将查询结果解析到结构体中,对应方式参考结构体中标签说明
	if err == nil {

	}
}
