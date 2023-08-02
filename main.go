package main

import (
	"github.com/gorilla/mux"
	"hezzl_task_5/clickhouse_logging"
	"hezzl_task_5/nats_queueing"
	"hezzl_task_5/postgresql"
	"hezzl_task_5/redis_caching"
	"hezzl_task_5/routes"
	"log"
	"net/http"
)

func init() {
	var err error
	err = postgresql.DBConnect()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s\n", err)
	}

	err = clickhouse_logging.CHConnect()
	if err != nil {
		log.Fatalf("Failed to connect to CH: %s\n", err)
	}

	err = nats_queueing.NatsConnect()
	if err != nil {
		log.Fatalf("Failed to connect to Nats: %s\n", err)
	}

	err = redis_caching.RedisConnect()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %s\n", err)
	}

	go func() {
		err := nats_queueing.NatsSubscribe()
		if err != nil {
			log.Printf("Failed to subscribe to CH: %s\n", err)
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
