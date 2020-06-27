package main

import (
	"fmt"
	"net/http"

	"mono/config"
	"mono/gc"
	"mono/recordstore"

	"github.com/labstack/echo/v4"
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

	e.GET("/", index)
	e.GET("/order/", provide)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.Get().Port)))
}

func index(c echo.Context) error {
	return c.String(http.StatusOK, "mono.")
}
