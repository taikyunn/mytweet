package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type Tweet struct {
	gorm.Model
	// bindingはバリデーションを実施することができる
	Content string `form:"content" binding:"required"`
}

// gormへの接続を行う関数
func gormConnect() *gorm.DB {
	DBMS := "mysql"
	USER := "root"
	DBNAME := "test"
	CONNECT := USER + ":" + "@/" + DBNAME + "?parseTime=true"
	db, err := gorm.Open(DBMS, CONNECT)

	if err != nil {
		panic(err.Error())
	}
	return db
}

func dbInit() {
	db := gormConnect()

	// 遅延させる。つまり関数の一番最後にDBとの接続を解除し、マイグレートすると言うこと
	defer db.Close()
	db.AutoMigrate(&Tweet{})
}

func dbInsert(content string) {
	db := gormConnect()

	defer db.Close()

	db.Create(&Tweet{Content: content})
}

func dbUpdate(id int, tweetText string) {
	db := gormConnect()
	var tweet Tweet
	// Tweet構造体にアクセスし、idが一致する値を取得する値を取得する
	db.First(&tweet, id)
	tweet.Content = tweetText
	db.Save(&tweet)
	db.Close()
}

// この関数の戻り値はTweet配列になる
func dbGetAll() []Tweet {
	// gorm接続の関数をdbに入れる
	db := gormConnect()
	// DBとの接続を解除
	defer db.Close()
	// 変数tweetsに配列を定義する
	var tweets []Tweet
	// dbからレコードを取得する際の順序を指定する
	// Find :条件にマッチする値を取得する
	// &tweetsはvar tweetのことである。これはContentというstring型のModel
	db.Order("created_at desc").Find(&tweets)
	return tweets
}

func dbGetOne(id int) Tweet {
	db := gormConnect()
	var tweet Tweet
	// idに紐づくtweetの中身を取得する
	db.First(&tweet, id)
	db.Close()
	return tweet
}

func dbDelete(id int) {
	db := gormConnect()
	var tweet Tweet
	db.First(&tweet, id)
	db.Delete(&tweet)
	db.Close()
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*.html")

	dbInit()

	//一覧
	r.GET("/", func(c *gin.Context) {
		tweets := dbGetAll()
		c.HTML(200, "index.html", gin.H{"tweets": tweets})
	})

	//登録
	r.POST("/new", func(c *gin.Context) {
		var form Tweet
		// ここがバリデーション部分.form=Tweet=Structなので
		if err := c.Bind(&form); err != nil {
			tweets := dbGetAll()
			c.HTML(http.StatusBadRequest, "index.html", gin.H{"tweets": tweets, "err": err})
			c.Abort()
		} else {
			content := c.PostForm("content")
			dbInsert(content)
			c.Redirect(302, "/")
		}
	})

	//投稿詳細
	r.GET("/detail/:id", func(c *gin.Context) {
		// 選択された編集ボタンが付随するcontentのidを取得している
		// .Paramでパラメータの値を取得できる
		n := c.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic(err)
		}
		tweet := dbGetOne(id)
		c.HTML(200, "detail.html", gin.H{"tweet": tweet})
	})

	//更新
	r.POST("/update/:id", func(c *gin.Context) {
		n := c.Param("id")
		// 変数n(string型)をint型に変更して変数idに代入する
		id, err := strconv.Atoi(n)
		if err != nil {
			panic("ERROR")
		}
		tweet := c.PostForm("tweet")
		dbUpdate(id, tweet)
		c.Redirect(302, "/")
	})

	//削除確認
	r.GET("/delete_check/:id", func(c *gin.Context) {
		n := c.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic("ERROR")
		}
		tweet := dbGetOne(id)
		c.HTML(200, "delete.html", gin.H{"tweet": tweet})
	})

	//削除
	r.POST("/delete/:id", func(c *gin.Context) {
		n := c.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic("ERROR")
		}
		dbDelete(id)
		c.Redirect(302, "/")

	})

	r.Run()
}
