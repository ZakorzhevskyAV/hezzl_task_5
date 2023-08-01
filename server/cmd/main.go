package main

import (
	"github.com/gorilla/mux"
	"hezzl_task_5/routes"
	"net/http"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/good/create", routes.Create).Methods(http.MethodPost)
	r.HandleFunc("/good/update", routes.Update).Methods(http.MethodPatch)
	r.HandleFunc("/good/remove", routes.Remove).Methods(http.MethodDelete)
	r.HandleFunc("/good/list", routes.List).Methods(http.MethodGet)
	r.HandleFunc("/good/reprioritize", routes.Reprioritize).Methods(http.MethodPatch)
	server := http.Server{
		Addr:    ":8000",
		Handler: r,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic("Server failed to listen and serve")
	}
}
