// package main

// import (
// 	"bytes"
// 	"fmt"
// 	"net/http"
// 	"path"
// 	"strings"
// 	"time"

// 	"github.com/dchest/captcha"
// 	"github.com/gin-gonic/gin"
// )

// type CaptchaResponse struct {
// 	CaptchaId string `json:"captchaId"` //验证码Id
// 	ImageUrl  string `json:"imageUrl"`  //验证码图片url
// }

// func main() {
// 	r := gin.Default()

// 	//1.获取验证码
// 	//http://localhost:8080/captcha
// 	r.GET("/captcha", func(c *gin.Context) {
// 		length := captcha.DefaultLen
// 		captchaId := captcha.NewLen(length)
// 		var captcha CaptchaResponse
// 		captcha.CaptchaId = captchaId
// 		captcha.ImageUrl = "/captcha/" + captchaId + ".png"
// 		c.JSON(http.StatusOK, captcha)
// 	})

// 	//2.获取验证码图片
// 	//http://localhost:8080/captcha/gHEIwh7nWreTFb53MkVk.png
// 	r.GET("/captcha/:captchaId", func(c *gin.Context) {
// 		captchaId := c.Param("captchaId")
// 		fmt.Println("GetCaptchaPng : " + captchaId)
// 		ServeHTTP(c.Writer, c.Request)
// 	})

// 	//3.验证
// 	//http://localhost:8080/verify/dVCqYbq7r2olKZfEtTvo/647489
// 	r.GET("/verify/:captchaId/:value", func(c *gin.Context) {
// 		captchaId := c.Param("captchaId")
// 		value := c.Param("value")
// 		if captchaId == "" || value == "" {
// 			c.String(http.StatusBadRequest, "参数错误")
// 		}
// 		if captcha.VerifyString(captchaId, value) {
// 			c.JSON(http.StatusOK, "验证成功")
// 		} else {
// 			c.JSON(http.StatusOK, "验证失败")
// 		}
// 	})
// 	r.Run()
// }

// func ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	dir, file := path.Split(r.URL.Path)
// 	ext := path.Ext(file)
// 	id := file[:len(file)-len(ext)]
// 	fmt.Println("file : " + file)
// 	fmt.Println("ext : " + ext)
// 	fmt.Println("id : " + id)
// 	if ext == "" || id == "" {
// 		http.NotFound(w, r)
// 		return
// 	}
// 	fmt.Println("reload : " + r.FormValue("reload"))
// 	if r.FormValue("reload") != "" {
// 		captcha.Reload(id)
// 	}
// 	lang := strings.ToLower(r.FormValue("lang"))
// 	download := path.Base(dir) == "download"
// 	if Serve(w, r, id, ext, lang, download, captcha.StdWidth, captcha.StdHeight) == captcha.ErrNotFound {
// 		http.NotFound(w, r)
// 	}
// }

// func Serve(w http.ResponseWriter, r *http.Request, id, ext, lang string, download bool, width, height int) error {
// 	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
// 	w.Header().Set("Pragma", "no-cache")
// 	w.Header().Set("Expires", "0")

// 	var content bytes.Buffer
// 	switch ext {
// 	case ".png":
// 		w.Header().Set("Content-Type", "image/png")
// 		captcha.WriteImage(&content, id, width, height)
// 	case ".wav":
// 		w.Header().Set("Content-Type", "audio/x-wav")
// 		captcha.WriteAudio(&content, id, lang)
// 	default:
// 		return captcha.ErrNotFound
// 	}

// 	if download {
// 		w.Header().Set("Content-Type", "application/octet-stream")
// 	}
// 	http.ServeContent(w, r, id+ext, time.Time{}, bytes.NewReader(content.Bytes()))
// 	return nil
// }

package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	ErrorServerBusy = "server is busy"
	ErrorReLogin    = "relogin"
)

type User struct {
	Id   int
	Name string
}
type JWTClaims struct {
	jwt.StandardClaims
	User User
}

var (
	Secret     = "123#111" //salt
	ExpireTime = 3600      //token expire time
)

//生成 jwt token
func genToken(user User) (string, error) {
	claims := &JWTClaims{
		User: user,
	}
	claims.IssuedAt = time.Now().Unix()
	claims.ExpiresAt = time.Now().Add(time.Second * time.Duration(ExpireTime)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(Secret))
	if err != nil {
		return "", errors.New(ErrorServerBusy)
	}
	return signedToken, nil
}

//验证jwt token
func verifyToken(ctx *gin.Context) (*JWTClaims, error) {
	strToken := ctx.Request.Header.Get("token")
	token, err := jwt.ParseWithClaims(strToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(Secret), nil
	})
	if err != nil {
		return nil, errors.New(ErrorServerBusy)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New(ErrorReLogin)
	}
	if err := token.Claims.Valid(); err != nil {
		return nil, errors.New(ErrorReLogin)
	}
	return claims, nil
}

// 更新token
func refresh(c *gin.Context) (string, error) {
	claims, _ := verifyToken(c)
	return genToken(claims.User)
}

func jwtAuth(ctx *gin.Context) {
	if _, err := verifyToken(ctx); err == nil {
		ctx.Next()
	} else {
		ctx.JSON(http.StatusOK, gin.H{"code": 4001})
		ctx.Abort()
	}
}

func main() {
	router := gin.Default()
	//在web开发中，浏览器处于安全考虑会限制跨域请求。
	//我们采用前后端分离的方式写接口的时候服务器端要允许跨域请求
	//使用 cors中间件 允许跨域
	router.Use(cors.Default())
	router.GET("/login", func(ctx *gin.Context) {
		user := User{1, "hanyun"}
		singedToken, err := genToken(user)
		if err == nil {
			ctx.JSON(http.StatusOK, gin.H{"code": 0, "token": singedToken})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"code": 1, "token": singedToken})
		}
	})
	//使用自定义的jwtAuth中间件
	router.Use(jwtAuth)
	router.GET("/user", func(context *gin.Context) {
		claims, _ := verifyToken(context)
		context.JSON(http.StatusOK, gin.H{"code": 0, "user": claims.User})
	})
	router.GET("/refresh", func(context *gin.Context) {
		singedToken, err := refresh(context)
		if err == nil {
			context.JSON(http.StatusOK, gin.H{"code": 0, "token": singedToken})
		} else {
			context.JSON(http.StatusOK, gin.H{"code": 1, "token": singedToken})
		}
	})
	router.Run()
}



// curl --location --request GET 'http://127.0.0.1:8080/user' \
// --header 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODM2NzQ2MzEsImlhdCI6MTU4MzY3MTAzMSwiVXNlciI6eyJJZCI6MSwiTmFtZSI6Imhhbnl1biJ9fQ.kMmE3DWXvNOUVsuHWgrlbm2pbsOHmbMtyr-V6hVjQ4s'