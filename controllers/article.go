package controllers

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"math"
	"math/rand"
	"path"
	"shFreshBMS/models"
	"shFreshBMS/redispool"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
)

type ArticleController struct {
	beego.Controller
}

func (c *ArticleController) ShowArticleList() {
	c.Data["userName"] = c.GetSession("userName")
	c.TplName = "index.html"
	o := orm.NewOrm()
	//创建文章表查询器，但不查询
	qs := o.QueryTable("GoodsSKU")
	var goods []models.GoodsSKU

	//先从redis中读取需要的数据
	log.Println("【a】尝试从redis查询数据，若查询成功直接显示")
	goodsType := []models.GoodsType{}
	//1. 从redis连接池中获取1个连接
	conn := redispool.Redisclient.Get()
	defer conn.Close()

	//2. 将类型从redis中取出并打印
	//正常情况下，存进去什么类型，就利用什么类型的回复助手函数。但是自定义类型不支持，需使用字节流存入和取出。
	relbytes, err := redis.Bytes(conn.Do("get", "GoodsType"))
	if err != nil {
		log.Println("get错误：", err)
	}
	dec := gob.NewDecoder(bytes.NewReader(relbytes))
	dec.Decode(&goodsType)
	if len(goodsType) != 0 {
		log.Println("【a】从redis中成功查询到数据", goodsType)
	} else {
		//如果以上操作没有从redis中读出数据，则去数据库中查询，并存入redis
		log.Println("【a】redis未读取到缓存，本次去mysql中查询数据")
		o.QueryTable("GoodsType").All(&goodsType)
		log.Println("【a】从mysql中成功查到数据查询数据", goodsType)

		//将文章类型存入redis数据库
		//正常情况下，存进去什么类型，就利用什么类型的回复助手函数。但是自定义类型不支持，需使用字节流存入和取出。
		//首先序列化内容

		// 1. 初始化一个buffer类型内存，用来存储编码的结果。（造一个内存卡）
		// 2. 获取一个编码器对象，并给他刚刚创建的buffer内存（得到一个播放器，把内存卡插进去）
		// 3. 使用编码器对象的编码方法开始编码，输入参数为要编码的内容，buffer存储编码结果，返回值为是否出错。

		var buffer bytes.Buffer
		enc := gob.NewEncoder(&buffer)
		enc.Encode(goodsType)
		if _, err := conn.Do("set", "GoodsType", buffer.String()); err != nil {
			log.Println("set错误：", err)
			return
		}
		if _, err := conn.Do("EXPIRE", "GoodsType", 120); err != nil {
			log.Println("过期时间错误：", err)
			return
		}
		log.Println("【a】从将mysql查询dao的数据成功存入redis", goodsType)
	}

	//获取本次查询的页码
	pageIndex, err := c.GetInt("pageIndex")
	if err != nil {
		//若未获取到页码，设置默认页码1
		pageIndex = 1
	}
	//定义每页大小，即本次请求的条数
	pageSize := 6
	//根据以上信息，获取开始查询的位置
	start := pageSize * (pageIndex - 1)

	//使用文章查询器，简单获得记录总数
	count, err := qs.RelatedSel("GoodsType").Count()
	if err != nil {
		log.Println("获取记录数错误：", err)
		return
	}
	//根据查询头和查询量，开始查询数据
	//参数1：限制获取的条数，参数2，偏移量，即开始位置
	qs.Limit(pageSize, start).RelatedSel("GoodsType").All(&goods)

	//加入文章类型筛选，默认全部,选择类型后，再次筛选。
	selectedtype := c.GetString("select")
	if selectedtype == "" || selectedtype == "全部类型" {
		log.Println("本次GET请求全部,未加入select参数,默认全部")
	} else {
		count, err = qs.RelatedSel("GoodsType").Filter("GoodsType__Name", selectedtype).Count()
		if err != nil {
			log.Println("获取记录数错误：", err)
			return
		}
		qs.Limit(pageSize, start).RelatedSel("GoodsType").Filter("GoodsType__Name", selectedtype).All(&goods)
	}
	//得出总页数
	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))
	//定义页码按钮启用状态
	enablelast, enablenext := true, true
	if pageIndex == 1 {
		enablelast = false
	}
	if pageIndex == pageCount {
		enablenext = false
	}
	c.Data["username"] = c.GetSession("userName")
	c.Data["typeName"] = selectedtype
	c.Data["GoodsType"] = goodsType
	c.Data["EnableNext"] = enablenext
	c.Data["EnableLast"] = enablelast
	c.Data["count"] = count
	c.Data["pageCount"] = pageCount
	c.Data["pageIndex"] = pageIndex
	c.Data["articles"] = goods

}
func (c *ArticleController) HandleTypeSelected() {
	/*
		selectedtype := c.GetString("select")
		articles := []models.Article{}
		o := orm.NewOrm()
		o.QueryTable("article").RelatedSel("ArticleType").Filter("ArticleType__TypeName", selectedtype).All(&articles)
		c.Data["artciles"] = articles

		//文章类型下拉
		GoodsType := []models.ArticleType{}
		o.QueryTable("article_type").All(&GoodsType)
		c.Data["GoodsType"] = GoodsType
		c.Data["username"] = c.GetSession("username")
		c.TplName = "index.html"
	*/
}

func (c *ArticleController) ShowAddArticle() {
	/*
		//文章类型下拉
		o := orm.NewOrm()
		GoodsType := []models.ArticleType{}
		o.QueryTable("article_type").All(&GoodsType)
		c.Data["GoodsType"] = GoodsType
		c.Data["username"] = c.GetSession("username")
		c.TplName = "add.html"
	*/
}
func (c *ArticleController) HandleAddArticle() {
	/*
		// c.Layout = "layout.html"
		c.TplName = "add.html"

		//取得post数据，使用getfile取得文件，注意设置enctype
		name := c.GetString("articleName")
		content := c.GetString("content")
		//取得上传文件，需判断是否传了文件
		var filename string
		f, h, err := c.GetFile("uploadname")
		if err != nil {
			log.Println("文件上传失败:", err)
		} else {
			// 1.校验文件类型
			// 2.校验文件大小
			// 3.防止重名，重新命名
			ext := path.Ext(h.Filename)
			log.Println(ext)
			if ext != ".jpg" && ext != ".png" && ext != "jpeg" {
				log.Println("文件类型错误")
				return
			}

			if h.Size > 5000000 {
				log.Println("文件超出大小")
				return
			}
			filename = time.Now().Format("20060102150405") + ext

			//保存文件到某路径下，程序默认当前路由的路径，故注意相对路径
			err = c.SaveToFile("uploadname", "../static/img/"+filename)
			if err != nil {
				log.Println("文件保存失败：", err)
				return
			}
			defer f.Close()

		}

		o := orm.NewOrm()
		//取得文章类型
		selectedtype := c.GetString("select")
		//利用此类型获取完整对象
		articletype := models.ArticleType{TypeName: selectedtype}
		o.Read(&articletype, "TypeName")
		//已知某个字段，查询所有字段时，如果字段为主键，则可省略，否则必须填列名。

		log.Println("aaaaaaaaa:", articletype.Id)
		article := models.Article{Title: name, Content: content, ArticleType: &articletype}
		//根据文件上传情况，判断是否更新路径
		if filename != "" {
			article.Img = "../static/img/" + filename
		}
		//插入数据库

		_, err = o.Insert(&article)
		if err != nil {
			log.Println("插入错误:", err)
			return
		}

		c.Redirect("/Article/ShowArticle", 302)
	*/
}
func (c *ArticleController) ShowContent() {
	/*
		id, err := c.GetInt("id")
		if err != nil {
			log.Println("获取ID失败：", err)
			return
		}
		content := models.Article{Id: id}
		o := orm.NewOrm()
		err = o.Read(&content)
		if err != nil {
			log.Println("查询数据失败：", err)
			return
		}
		//阅读量+1并写回数据库
		content.Count++
		o.Update(&content)

		/*处理最近浏览,
		1. 首先需确定当前浏览者登录状态,获取浏览者信息
		2. 将浏览者信息插入数据表
		3. 将历史浏览者信息从表中读出，去重，显示*/
	/*
		if username := c.GetSession("username"); username != nil {
			user := models.User{Name: username.(string)}
			o.Read(&user, "Name")
			//目的：构造多对多查询器,并执行添加插入方法
			o.QueryM2M(&content, "Users").Add(&user)
		}
		//开始读出历史浏览者信息
		users := []models.User{}
		o.QueryTable("User").Filter("Articles__Article__Id", content.Id).Distinct().All(&users)
		c.Data["users"] = users
		c.Data["content"] = content
		c.Data["username"] = c.GetSession("username")
		c.TplName = "content.html"
	*/
}
func (c *ArticleController) HandleDelete() {
	/*思路
	1.被点击的url传值
	2.执行对应的删除操作
	*/
	/*
		c.TplName = ""
		id, err := c.GetInt("id")
		if err != nil {
			log.Println("获取ID失败：", err)
			return
		}
		article := models.Article{Id: id}
		o := orm.NewOrm()
		_, err = o.Delete(&article)
		if err != nil {
			log.Println("删除数据失败：", err)
			return
		}
		//c.TplName = "ShowArticle.html"
		c.Redirect("/Article/ShowArticle", 302)
	*/
}

func (c *ArticleController) ShowUpdate() {
	/*思路
	1. 获取数据，填充数据
	2. 更新数据，更新数据库，返回列表页
	*/
	// c.Layout = "layout.html
	/*
		c.TplName = "update.html"
		id, err := c.GetInt("id")
		if err != nil {
			log.Println("id获取失败：", err)
			return
		}
		article := models.Article{Id: id}
		o := orm.NewOrm()
		err = o.ReadForUpdate(&article)
		if err != nil {
			log.Println("读取失败：", err)
			return
		}
		c.Data["article"] = article
		c.Data["username"] = c.GetSession("username")
	*/
}

// HandleUpdate 处理更新
func (c *ArticleController) HandleUpdate() {
	/*
		c.TplName = "update.html"
		//取得post数据，使用getfile取得文件，注意设置enctype
		name := c.GetString("articleName")
		content := c.GetString("content")
		oldimagepath := c.GetString("oldimagepath")

		var filename string
		id, err := c.GetInt("id")
		if err != nil {
			log.Println("id获取失败：", err)
			return
		}
		article := models.Article{Id: id, Title: name, Content: content, Img: oldimagepath}
		c.Data["article"] = article
		f, h, err := c.Ge tFile("uploadname")
		if err != nil {
			c.Data["errmsg"] = "文件上传失败"
		} else {
			/*保存之前先做校验处理:
			1.校验文件类型
			2.校验文件大小
			3.防止重名，重新命名
	*/
	/*
			ext := path.Ext(h.Filename)
			//log.Println(ext)
			if ext != ".jpg" && ext != ".png" && ext != "jpeg" {
				log.Println(err)
				c.Data["errmsg"] = "文件类型错误"
				return
			}

			if h.Size > 5000000 {
				log.Println(err)
				c.Data["errmsg"] = "文件超出大小"
				return
			}
			filename = time.Now().Format("20060102150405") + ext

			//保存文件到某路径下，程序默认当前在项目的根目录，故注意相对路径
			err = c.SaveToFile("uploadname", "./static/img/"+filename)
			if err != nil {
				log.Println("文件保存失败：", err)
				c.Data["errmsg"] = "文件保存失败"
				return
			}
			defer f.Close()
		}

		//若上传了新文件，则使用新文件路径，否则使用旧路径不变
		if filename != "" {
			article.Img = "../static/img/" + filename
		}

		//更新数据库
		o := orm.NewOrm()
		_, err = o.Update(&article, "title", "content", "img", "create_time", "update_time")
		if err != nil {
			log.Println("更新错误:", err)
			c.Data["errmsg"] = "更新失败"
			return
		}
		c.Redirect("/Article/ShowArticle", 302)
	*/
}

func (c *ArticleController) ShowAddType() {
	c.TplName = "addType.html"
	var types []models.GoodsType
	o := orm.NewOrm()
	o.QueryTable("GoodsType").All(&types)
	c.Data["types"] = types
	c.Data["userName"] = c.GetSession("userName")
	//刷新页面时更新缓存。
	err := updateRedisDate("set", "GoodsType", types, 300)
	if err != nil {
		log.Println("更新缓存失败：", err)
	}
}

//处理更新redis的功能函数:将自定义类型变量序列化存储到redis
//handlestr为操作名：如get set等
//key为redis中的key
//cont为需要序列号写入的自定义类型变量，需要传指针类型
//time为更新后的过期时间（秒），-1代表永不过期
func updateRedisDate(handlestr string, key string, cont interface{}, time int) error {
	/*
		log.Println("【b】准备序列化：", cont)
		var buffer bytes.Buffer
		enc := gob.NewEncoder(&buffer)
		err := enc.Encode(cont)
		if err != nil {
			return err
		}
		//	log.Println("【b】准备写入redis", buffer.String())
		conn := redispool.Redisclient.Get()
		_, err = conn.Do(handlestr, key, buffer.String())
		if err != nil {
			return err
		}
		log.Println("time")
		_, err = conn.Do("EXPIRE", key, time)
		if err != nil {
			return err
		}
		log.Println("xier")
	*/
	return nil

}
func (c *ArticleController) HandleAddType() {
	c.TplName = "addType.html"
	extLimt := []string{
		".jpg",
		".png",
		"jpeg",
	}

	logoPath, err := UploadFile(&c.Controller, "uploadlogo", extLimt, 5000000)
	if err != nil {
		log.Println(err)
		c.Data["errmsg"] = err
		return
	}
	typeImagePath, err := UploadFile(&c.Controller, "uploadTypeImage", extLimt, 5000000)
	if err != nil {
		log.Println(err)
		c.Data["errmsg"] = err
		return
	}
	if c.GetString("typeName") == "" || logoPath == "" || typeImagePath == "" {
		c.Data["errmsg"] = "数据不完整，无法提交"
		return
	}
	var goodsType models.GoodsType
	goodsType.Name = c.GetString("typeName")
	goodsType.Logo = logoPath
	goodsType.Image = typeImagePath

	o := orm.NewOrm()
	_, err = o.Insert(&goodsType)
	if err != nil {
		log.Println("插入数据库失败：", err)
		c.Data["errmsg"] = err
		return
	}
	c.Redirect("/Article/AddArticleType", 302)

	//插入数据库成功后，此处不更新缓存，否则需要再次请求所有类型，刷新页面时更新更合适。
}
func (c *ArticleController) HandleDeleteType() {
	/*
		id, err := c.GetInt("id")
		if err != nil {
			log.Println("获取ID失败：", err)
			return
		}
		articleType := models.ArticleType{Id: id}
		o := orm.NewOrm()
		o.Delete(&articleType)
		c.Redirect("/Article/AddArticleType", 302)
	*/
}

//上传文件/图片到服务器的imge文件夹下
// 传入controller,模板中input的filekey
// 得到保存地址，和错误
//支持文件类型校验、文件大小限制（单位byte）
func UploadFile(c *beego.Controller, fileKey string, extLimt []string, sizeLimit int64) (filePath string, err error) {
	if sizeLimit <= 0 {
		sizeLimit = 50000 //默认设置为5M
	}
	f, h, err := c.GetFile(fileKey)
	if err != nil {
		log.Println("文件上传失败:", err)
		return "", err
	}
	defer f.Close()
	// 1.校验文件类型
	// 2.校验文件大小
	// 3.防止重名，重新命名
	ext := path.Ext(h.Filename)
	log.Println("用户上传文件的拓展名：", ext)
	for i := 0; i < len(extLimt); i++ {
		if extLimt[i] == ext {
			break
		}
		if i == len(extLimt)-1 {
			log.Println("用户上传文件类型错误：")
			return "", errors.New("文件类型错误")
		}
	}

	if h.Size > sizeLimit {
		log.Println("文件超出大小")
		return "", errors.New("文件大小超过限制")
	}
	//这里注意重命名的粒度，否则容易重复导致文件覆盖
	// 这里由于连续调用两次上传文件，导致以秒命名的文件名出现覆盖
	rand.Seed(time.Now().UnixNano())
	num := strconv.Itoa(rand.Intn(100))
	fileName := time.Now().Format("20060102150405") + num + ext
	filePath = path.Join("static", "upload", fileName)
	//保存文件到某路径下，程序默认当前路由的路径，故注意相对路径
	err = c.SaveToFile(fileKey, filePath)
	if err != nil {
		log.Println("文件保存失败：", err)
		return "", err
	}
	log.Println("文件上传成功!", filePath)
	return filePath, nil
}
