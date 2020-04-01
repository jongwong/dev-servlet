package main

import (
	"github.com/kataras/iris/v12"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	workPath = kingpin.Flag("workPath", "set workPath for work environment.").Default("/temp/devlet/workplace").String()
)

type fileBody struct {
	Files []Files `json:"files"`
}

var app *iris.Application

func main() {

	kingpin.Parse()
	app := iris.Default()
	// before run request handler
	app.Use(errorMiddleWare)

	app.Any("/ws", iris.FromStd(wsHandler))
	app.Post("/files", func(ctx iris.Context) {
		var c fileBody
		if err := ctx.ReadJSON(&c); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.WriteString(err.Error())
			return
		}
		writeFiles(c.Files)
		var data = true
		ctx.JSON(generateSuccessResultMap(genSuccessCodeMsg(data)))
	})
	app.Delete("/files", func(ctx iris.Context) {
		var c fileBody
		if err := ctx.ReadJSON(&c); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.WriteString(err.Error())
			return
		}
		deleteFiles(c.Files)
		var data = true
		ctx.JSON(generateSuccessResultMap(genSuccessCodeMsg(data)))
	})
	app.Listen(":8080")
}

type responseFailedResult struct {
	msg        string
	status     bool
	statusCode int
}
type responseSuccessResult struct {
	data       interface{}
	status     bool
	statusCode int
}

func generateFailedResultMap(result responseFailedResult) iris.Map {
	return iris.Map{"message": result.msg, "status": result.status, "statusCode": result.statusCode}
}
func generateSuccessResultMap(result responseSuccessResult) iris.Map {
	return iris.Map{"data": result.data, "status": result.status, "statusCode": result.statusCode}
}
func genFailedCodeMsg(serverErrorCode int, errMsg string) responseFailedResult {
	var re responseFailedResult
	re.msg = errMsg
	re.statusCode = 500
	re.status = false
	return re
}
func genSuccessCodeMsg(data interface{}) responseSuccessResult {
	var re responseSuccessResult
	re.data = data
	re.statusCode = 200
	re.status = true
	return re
}
