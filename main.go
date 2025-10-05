package main

import (
	_ "template-golang/docs"

	"template-golang/cmd"
)

// @title UT School
// @version 1.0
// @description This is a sample swagger for Fiber
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @termsOfService http://swagger.io/terms/
// @contact.name Adi Kurniawan
// @contact.email kurniawanadi4556@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:10010
// @BasePath /api/v1

func main() {
	cmd.Execute()
}
