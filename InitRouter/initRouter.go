package InitRouter

import (
	"github.com/gin-gonic/gin"
	"github.com/tddey01/blogs/handler"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	if mode := gin.Mode(); mode == gin.TestMode {
		router.LoadHTMLGlob("./../templates/*")
	} else {
		router.LoadHTMLGlob("templates/*")
	}
	router.StaticFile("/favicon.ico", "./favicon.ico")
	router.Static("/statics", "./statics/")
	//router.StaticFS("/avatar", http.Dir(filepath.Join(utils.RootPath(), "avatar")))
	index := router.Group("/")
	{
		index.Any("", handler.Index)
	}

	//index := router.Group("/")
	//{
	//	index.Any("", handler.Index)
	//}

	return router
}
