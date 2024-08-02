package app

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"log"
	"testjunior/internal/app/endpoint"
	"testjunior/internal/app/mw"
	"testjunior/internal/app/service"
)

type App struct {
	e    *endpoint.Endpoint
	s    *service.Service
	echo *echo.Echo
}

func NewApp() (*App, error) {
	a := App{}

	a.s = service.NewService()
	a.e = endpoint.NewEndpoint(a.s)
	a.echo = echo.New()

	a.echo.Use(mw.MiddlewareRoleCheck)
	a.echo.GET("/status", a.e.Status)

	return &a, nil

}

func (a *App) Run() error {
	fmt.Println("Server running")
	err := a.echo.Start(":8100")
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
