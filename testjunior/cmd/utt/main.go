package main

import (
	"log"
	"testjunior/internal/pkg/app"
)

//func Handler(ctx echo.Context) error {
//	d := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
//	r := time.Until(d)
//
//	s := fmt.Sprintf("Count days %d", int64(r.Hours()/24))
//
//	err := ctx.String(http.StatusOK, s)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//func MW(next echo.HandlerFunc) echo.HandlerFunc {
//	return func(ctx echo.Context) error {
//		val := ctx.Request().Header.Get("User-Role")
//		if val == "admin" {
//			log.Println("Red button user")
//		}
//
//		err := next(ctx)
//		if err != nil {
//			return err
//		}
//
//		return nil
//
//	}
//
//}

func main() {
	a, err := app.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	err = a.Run()
	if err != nil {
		log.Fatal(err)
	}
}
