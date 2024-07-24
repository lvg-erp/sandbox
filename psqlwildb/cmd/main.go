package main

import (
	"database/sql"
	_ "database/sql"
	"fmt"

	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"psqlwildb/models"
)

// short error checking
func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

type Server struct {
	db     *sql.DB
	router *mux.Router
}

func (s Server) Routes() {

}

func main() {
	//getting database connection
	myDb, err := models.NewDB(models.GetConnectionString("configuration.json"))
	if err != nil {
		fmt.Println(err, "exiting programm")
		os.Exit(1)
	}

	fmt.Println("Connected to DB")

	myRouter := mux.NewRouter()
	//
	myServer := &Server{
		db:     myDb,
		router: myRouter,
	}

	//function with all routes
	myServer.Routes()

	server := &http.Server{
		Addr: ":8099",

		Handler: myRouter,
	}
	err = server.ListenAndServe()
	checkError(err)
}

//curl -d {"txt":"cheers"} -H "Content-Type: application/json" -X POST http://127.0.0.1:8080/api/v1/user/5/comment
