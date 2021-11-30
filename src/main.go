package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"src/config"
	. "src/util"
)

type returnData struct {
	ShortUrl string `yaml:"short_url"`
}

func initDB() {
	initPgDB()
}

func main() {
	initDB()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	gAPI := e.Group(config.C.App.Prefix)
	gAPI.GET("/api/short_url", GetShortUrl)
	gAPI.GET("/:shorter",RedirectShortUrl)
	Logger.Fatal(e.Start(config.C.App.Addr))
}

//GetShortUrl POST /api/short_url?url=
func GetShortUrl(c echo.Context) error {
	longUrl := c.QueryParam("url")
    m := GetModel()
    defer m.Close()

	//intID, err := strconv.ParseInt(id.Hex(), 16, 64)
	//if err!=nil {
	//	Logger.Error(err)
	//	return c.JSON(http.StatusInternalServerError,"ParseInt error")
	//}

	var url = Url{
		LongUrl:  longUrl,
		//ShortUrl: shortUrl,
	}

	rst,id,err:=m.CreateUrl(url)
	Logger.Debug(rst)
	if err!=nil {
		Logger.Error(err)
		return c.JSON(http.StatusInternalServerError,"CreateUrl error")
	}
	//id := int64(rst.RowsReturned())
	Logger.Debug(id)
	shortUrl := base10ToBase62(id)

	rst,err=m.UpdateUrl(id,shortUrl)
	Logger.Debug(rst)
	if err!=nil {
		Logger.Error(err)
		return c.JSON(http.StatusInternalServerError,"UpdateUrl error")
	}

	shorter := config.C.Web.Addr + "/" + shortUrl
	var rsp = returnData{ShortUrl: shorter}
	return c.JSON(http.StatusOK,rsp)
}

// RedirectShortUrl GET /:shorter
func RedirectShortUrl(c echo.Context) error {
	shorter:=c.Param("shorter")
	if len(shorter)!=5 {
		return c.JSON(http.StatusNotFound,"not found")
	}

	m := GetModel()
	defer m.Close()
	url,err:=m.RetrieveUrlByShorter(shorter)
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
