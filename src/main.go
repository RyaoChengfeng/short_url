package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"src/config"
	. "src/util"
	"strconv"
)

type returnData struct {
	ShortUrl string `yaml:"short_url"`
}

func initDB() {
	initMongoMongoDB()
}

func main() {
	initDB()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	gAPI := e.Group(config.C.App.Prefix)
	gAPI.POST("/short_url", GetShortUrl)
	gAPI.GET("/:shorter",RedirectShortUrl)
	Logger.Fatal(e.Start(config.C.App.Addr))
}

//GetShortUrl POST /short_url?url=
func GetShortUrl(c echo.Context) error {
	longUrl := c.QueryParam("url")
    m := GetModel()
    defer m.Close()

    id := primitive.NewObjectID()
	intID, err := strconv.ParseInt(id.Hex(), 10, 64)
	if err!=nil {
		Logger.Error(err)
		return c.JSON(http.StatusInternalServerError,"ParseInt error")
	}
    shortUrl := base10ToBase62(intID)
	var url = Url{
		ID:       id,
		LongUrl:  longUrl,
		ShortUrl: shortUrl,
	}
	rst,err:=m.InsertUrl(url)
	Logger.Debug(rst)
	if err!=nil {
		Logger.Error(err)
		return c.JSON(http.StatusInternalServerError,"InsertUrl error")
	}

	shorter := config.C.Web.Addr + shortUrl
	var rsp = returnData{ShortUrl: shorter}
	return c.JSON(http.StatusOK,rsp)
}

// RedirectShortUrl GET /:shorter
func RedirectShortUrl(c echo.Context) error {
	shorter:=c.Param("shorter")
	if len(shorter)!=6 {
		return c.JSON(http.StatusNotFound,"not found")
	}

	m := GetModel()
	defer m.Close()
	url,err:=m.RetrieveUrlWithShortUrl(shorter)
	Logger.Debug(url)
	if err!=nil {
		Logger.Error(err)
		return c.JSON(http.StatusInternalServerError,"RetrieveUrl error")
	}
	if url.LongUrl=="" {
		return c.JSON(http.StatusNotFound,"invalid url")
	}

	return c.Redirect(301,url.LongUrl)
}
