package main

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

type CaptchaResponse struct {
	CaptchaId string `json:"captchaId"` //验证码Id
	ImageUrl  string `json:"imageUrl"`  //验证码图片url
}

func main() {
	r := gin.Default()

	//1.获取验证码
	//http://localhost:8080/captcha
	r.GET("/captcha", func(c *gin.Context) {
		length := captcha.DefaultLen
		captchaId := captcha.NewLen(length)
		var captcha CaptchaResponse
		captcha.CaptchaId = captchaId
		captcha.ImageUrl = "/captcha/" + captchaId + ".png"
		c.JSON(http.StatusOK, captcha)
	})

	//2.获取验证码图片
	//http://localhost:8080/captcha/gHEIwh7nWreTFb53MkVk.png
	r.GET("/captcha/:captchaId", func(c *gin.Context) {
		captchaId := c.Param("captchaId")
		fmt.Println("GetCaptchaPng : " + captchaId)
		ServeHTTP(c.Writer, c.Request)
	})

	//3.验证
	//http://localhost:8080/verify/dVCqYbq7r2olKZfEtTvo/647489
	r.GET("/verify/:captchaId/:value", func(c *gin.Context) {
		captchaId := c.Param("captchaId")
		value := c.Param("value")
		if captchaId == "" || value == "" {
			c.String(http.StatusBadRequest, "参数错误")
		}
		if captcha.VerifyString(captchaId, value) {
			c.JSON(http.StatusOK, "验证成功")
		} else {
			c.JSON(http.StatusOK, "验证失败")
		}
	})
	r.Run()
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dir, file := path.Split(r.URL.Path)
	ext := path.Ext(file)
	id := file[:len(file)-len(ext)]
	fmt.Println("file : " + file)
	fmt.Println("ext : " + ext)
	fmt.Println("id : " + id)
	if ext == "" || id == "" {
		http.NotFound(w, r)
		return
	}
	fmt.Println("reload : " + r.FormValue("reload"))
	if r.FormValue("reload") != "" {
		captcha.Reload(id)
	}
	lang := strings.ToLower(r.FormValue("lang"))
	download := path.Base(dir) == "download"
	if Serve(w, r, id, ext, lang, download, captcha.StdWidth, captcha.StdHeight) == captcha.ErrNotFound {
		http.NotFound(w, r)
	}
}

func Serve(w http.ResponseWriter, r *http.Request, id, ext, lang string, download bool, width, height int) error {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	var content bytes.Buffer
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
		captcha.WriteImage(&content, id, width, height)
	case ".wav":
		w.Header().Set("Content-Type", "audio/x-wav")
		captcha.WriteAudio(&content, id, lang)
	default:
		return captcha.ErrNotFound
	}

	if download {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	http.ServeContent(w, r, id+ext, time.Time{}, bytes.NewReader(content.Bytes()))
	return nil
}
