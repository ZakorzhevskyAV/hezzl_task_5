package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"hezzl_task_5/postgresql"
	"hezzl_task_5/routes"
	"log"
	"net/http"
	"strings"
)

func init() {
	var err error
	err = postgresql.DBConnect()
	if err != nil {
		log.Fatal("Failed to connect to DB")
	}

	err := cache.Connect()
	if err != nil {
		initErrors = append(initErrors, err.Error())
	}

	err = logs.Connect()
	if err != nil {
		initErrors = append(initErrors, err.Error())
	}

	err = nats.Connect()
	if err != nil {
		initErrors = append(initErrors, err.Error())
	}

	if len(initErrors) != 0 {
		panic(fmt.Sprintf("Запуск приложения невозможен из-за следующих ошибок инициализации %s", strings.Join(initErrors, ",\n")))
	}

	go func() { // работа без логирования возможна, но на эту ошибку нужно будет обратить внимание
		err := nats.Subscribe()
		if err != nil {
			log.Printf("Получение логов невозможно, Nats возвращает в подписчике: %s\n", err.Error())
		}
	}()
}

func main() {
	r := mux.NewRouter()
	r.Path("/good/create/{projectId:[0-9]+}").Methods(http.MethodPost).HandlerFunc(routes.Create)
	r.Path("/good/update/{id:[0-9]+}/{projectId:[0-9]+}").Methods(http.MethodPatch).HandlerFunc(routes.Update)
	r.Path("/good/remove/{id:[0-9]+}/{projectId:[0-9]+}").Methods(http.MethodDelete).HandlerFunc(routes.Remove)
	r.Path("/good/list/{limit:[0-9]+}/{offset:[0-9]+}").Methods(http.MethodGet).HandlerFunc(routes.List)
	r.Path("/good/reprioritize/{id:[0-9]+}/{projectId:[0-9]+}").Methods(http.MethodPatch).HandlerFunc(routes.Reprioritize)
	server := http.Server{
		Addr:    ":8000",
		Handler: r,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic("Server failed to listen and serve")
	}
}
