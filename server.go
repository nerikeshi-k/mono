package main

import (
	"fmt"

	"github.com/nerikeshi-k/mono/config"
	"github.com/nerikeshi-k/mono/gc"
	"github.com/nerikeshi-k/mono/handler"
	"github.com/nerikeshi-k/mono/recordstore"

	echo "github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	serverHeader := func(hf echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			return hf(c)
		}
	}
	e.Use(serverHeader)

	go gc.Start()
	defer recordstore.Close()

	e.GET("/order/", handler.HandleOrder)
	e.GET("/*", handler.Handle)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.Get().Port)))
}
