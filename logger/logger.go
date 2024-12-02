package logger

import (
	"net"
	"net/http"
	"net/http/httputil"
	"onpaper-api-go/settings"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(config *settings.LogConfig) (err error) {
	// 1.zap.New(core,options)方法来手动自定义日志对象
	// 2.zap.New(core zapcore.Core, options ...Option) *Logger 需要传入一个core
	// zapcore.NewCore 返回 zapcore.Core
	// 3.而zapcore.NewCore 需要三个配置——Encoder，WriteSyncer，LogLevel

	//指定日志将写到哪里
	writeSyncer := getLogWriter(
		config.Filename,
		config.MaxSize,
		config.MaxBackups,
		config.MaxAge,
	)
	// 获取 日记写入的格式
	encoder := getEncoder()
	var level = new(zapcore.Level)
	// 将配置中的 level字符串 转成 zap 库里面的 level 类型
	err = level.UnmarshalText([]byte(config.Level))
	if err != nil {
		return
	}
	core := zapcore.NewCore(encoder, writeSyncer, level)
	// 添加一个option 将调用函数信息记录到日志中的功能 zap.AddCaller()
	logger := zap.New(core, zap.AddCaller())
	// 替换zap库中全局的logger
	zap.ReplaceGlobals(logger)
	return

}

func getEncoder() zapcore.Encoder {
	// 覆盖默认的ProductionConfig()
	encoderConfig := zap.NewProductionEncoderConfig()
	// 返回正常的时间而不是时间戳
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	// 日记级别大写
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 浮点化秒数。
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewJSONEncoder(encoderConfig) // json 格式
}
func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	//我们使用lumberjack 这个包打开文件，可以实现分割日记。
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,   // 一个日志包最大大小 M
		MaxBackups: maxBackup, // 到达最大size 会进行切割，可以选择备份数量
		MaxAge:     maxAge,    // 最大备份天数
		Compress:   false,     // 是否压缩
	}
	//使用zapcore.AddSync()函数并且将打开的文件句柄传进去。
	// 写入日记文件 并打印到终端
	//zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
	return zapcore.AddSync(lumberJackLogger)

}

// GinLogger 定义一个中间件 记录日记
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// 获取url 参数
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		// 计算花费时间
		cost := time.Since(start)
		//zap.L() 返回全局Logger，可以用ReplaceGlobals 重新配置。
		//zap.XXX 比如zap.int 是告诉 zap库 传入的类型 可以更高性能
		zap.L().Debug(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery 出现的panic恢复项目，并使用zap记录相关日志
// 重写 官方的 gin.Recovery()  中间件
// stack 是否记录堆栈的信息
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					zap.L().Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}
				// stack 是否记录堆栈的信息
				if stack {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// ErrZapLog 记录错误信息
func ErrZapLog(err error, meta interface{}) {
	zap.L().Error(err.Error(), zap.Any("info", meta), zap.Error(err))
}

// InfoZapLog 记录普通信息
func InfoZapLog(message string, meta interface{}) {
	// 记录更新日志
	zap.L().Info(message, zap.Any("info", meta))
}
