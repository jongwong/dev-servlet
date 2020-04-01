package main

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"runtime"
)

func errorMiddleWare(ctx iris.Context) {
	defer func() {
		if err := recover(); err != nil {
			if ctx.IsStopped() {
				return
			}

			var stacktrace string
			for i := 1; ; i++ {
				_, f, l, got := runtime.Caller(i)
				if !got {
					break
				}
				stacktrace += fmt.Sprintf("%s:%d\n", f, l)
			}

			errMsg := fmt.Sprintf("error message: %s", err)
			// when stack finishes
			logMessage := fmt.Sprintf("从错误中回复：('%s')\n", ctx.HandlerName())
			logMessage += errMsg + "\n"
			logMessage += fmt.Sprintf("\n%s", stacktrace)
			// 打印错误日志
			ctx.Application().Logger().Warn(logMessage)
			// 返回错误信息
			var serverErrorCode = iris.StatusInternalServerError
			ctx.JSON(generateFailedResultMap(genFailedCodeMsg(serverErrorCode, errMsg)))
			ctx.StatusCode(serverErrorCode)
			ctx.StopExecution()
		}
	}()
	ctx.Next()
}

func throwError(e error) {
	if e != nil {
		panic(e)
	}
}
