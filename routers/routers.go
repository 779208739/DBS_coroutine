package routers

import (
	"fmt"
	"go_projects/controller"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRouter(fin chan int, info *controller.AllInfo) *gin.Engine {

	r := gin.Default()

	r.Static("static/", "static/")
	r.LoadHTMLGlob("index.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/index", func(c *gin.Context) {
		time.Sleep(time.Millisecond * 500)
		Lock := info.L
		fmt.Println(info.U[0].Status, info.U[1].Status)
		c.JSON(http.StatusOK, gin.H{
			"code":    2000,
			"msg":     "success",
			"data":    Lock,
			"data1":   info.U[0].T,
			"data2":   info.U[1].T,
			"data3":   info.U[2].T,
			"status1": info.U[0].Status,
			"status2": info.U[1].Status,
			"status3": info.U[2].Status,
		})
	})

	FinGroup := r.Group("index")
	FinGroup.GET("/finish1", func(c *gin.Context) {
		fin <- 0
		c.JSON(http.StatusOK, gin.H{
			"msg": "ok",
		})
	})
	FinGroup.GET("/finish2", func(c *gin.Context) {
		fin <- 1
		c.JSON(http.StatusOK, gin.H{
			"msg": "ok",
		})
	})
	FinGroup.GET("/finish3", func(c *gin.Context) {
		fin <- 2
		c.JSON(http.StatusOK, gin.H{
			"msg": "ok",
		})
	})

	return r
}
