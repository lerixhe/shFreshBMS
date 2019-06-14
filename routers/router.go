package routers

import (
	"shFreshBMS/controllers"

	"github.com/astaxie/beego/context"

	"github.com/astaxie/beego"
)

func init() {
	//设置路由过滤，使用正则匹配，在路由之前，判断session，否则重定向
	beego.InsertFilter("/Article/*", beego.BeforeRouter, filterFunc)
	//用户注册
	beego.Router("/register", &controllers.UserController{}, "get:ShowReg;post:HandleReg")
	//用户登录
	beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")
	beego.Router("/", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")
	//商品列表
	beego.Router("/Article/ShowArticle", &controllers.ArticleController{}, "get:ShowArticleList;post:HandleTypeSelected")
	beego.Router("/Article/AddArticle", &controllers.ArticleController{}, "get:ShowAddArticle;post:HandleAddArticle")
	beego.Router("/Article/content", &controllers.ArticleController{}, "get:ShowContent")
	beego.Router("/Article/DeleteArticle", &controllers.ArticleController{}, "get:HandleDelete")
	beego.Router("/Article/UpdateArticle", &controllers.ArticleController{}, "get:ShowUpdate;post:HandleUpdate")
	//添加文章类型
	beego.Router("/Article/AddArticleType", &controllers.ArticleController{}, "get:ShowAddType;post:HandleAddType")
	//删除文章类型
	beego.Router("/Article/DeleteArticleType", &controllers.ArticleController{}, "get:HandleDeleteType")
	//用户退出登录
	beego.Router("/logout", &controllers.UserController{}, "get:HandleLogout")

}

//检查session，此函数由正则匹配的路由，在寻找路由之前执行
func filterFunc(ctx *context.Context) {
	if name := ctx.Input.Session("userName"); name == nil {
		ctx.Redirect(302, "/login")
		return
	}

}
