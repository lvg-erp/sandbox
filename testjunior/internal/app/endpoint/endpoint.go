package endpoint

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Service interface {
	DaysLeft() int64
}

type Endpoint struct {
	s Service
}

func NewEndpoint(svc Service) *Endpoint {
	return &Endpoint{
		s: svc,
	}
}

func (e *Endpoint) Status(ctx echo.Context) error {

	ld := e.s.DaysLeft()

	s := fmt.Sprintf("Count days %d", ld)

	err := ctx.String(http.StatusOK, s)
	if err != nil {
		return err
	}
	return nil
}
