package controller

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

/*
返回的数据格式
{
	"code": 10000, // 程序中的错误码
	"msg": xx,     // 提示信息
	"data": {},    // 数据
}
*/

// ResponseData 返回的数据结构
type ResponseData struct {
	Status ResCode     `json:"status"`
	Msg    interface{} `json:"msg"`
	Data   interface{} `json:"data"`
}

// ResponseErrorAndLog 返回响应错误并记录日志
func ResponseErrorAndLog(c *gin.Context, code ResCode, err error) {
	// 获取url 参数
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery

	// 记录错误信息
	zap.L().Error(
		err.Error(),
		zap.String("method", c.Request.Method),
		zap.String("path", path),
		zap.String("query", query),
		zap.String("ip", c.ClientIP()),
		zap.Error(err))

	// 返回错误数据
	c.JSON(http.StatusOK, &ResponseData{
		Status: code,
		Msg:    code.Msg(),
		Data:   "",
	})
	// 停止之后中间件
	c.Abort()
}

// ResponseError  返回响应错误
func ResponseError(c *gin.Context, code ResCode) {
	c.JSON(http.StatusOK, &ResponseData{
		Status: code,
		Msg:    code.Msg(),
		Data:   "",
	})
	// 停止之后中间件
	c.Abort()
}

// ResponseSuccess 返回成功数据
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Status: 200,
		Msg:    CodeSuccess.Msg(),
		Data:   data,
	})
}

// Response 自定义返回数据
func Response(c *gin.Context, code ResCode, data interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Status: code,
		Msg:    code.Msg(),
		Data:   data,
	})
}
