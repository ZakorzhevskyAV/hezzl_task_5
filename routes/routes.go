package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"hezzl_task_5/nats_queueing"
	"hezzl_task_5/postgresql"
	"hezzl_task_5/redis_caching"
	"hezzl_task_5/types"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func Create(w http.ResponseWriter, r *http.Request) {
	var errmsg string
	name, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		errmsg = "Failed to get a payload\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	URLVars := mux.Vars(r)
	conn := postgresql.DBConn
	rows, err := conn.Query(`SELECT * FROM GOODS WHERE project_id = ? and name = ?`, URLVars["projectId"], string(name))
	if err != nil {
		errmsg = "Failed to get rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if !rows.Next() {
		errmsg = "No rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s}", errmsg)))
		return
	}
	_, err = conn.Query(`SELECT * FROM GOODS WHERE project_id = ? and name = ?`, URLVars["projectId"], string(name))
	if err != nil {
		errmsg = "Failed to get rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s}", errmsg)))
		return
	}
	_, err = conn.Query(`INSERT INTO GOODS (project_id, name, description, priority) VALUES (?,?,?, (SELECT COALESCE(MAX(priority), 0) + 1 FROM GOODS))`,
		URLVars["projectId"],
		string(name),
		fmt.Sprintf("Entry with id \"%s\"", URLVars["projectId"]))
	if err != nil {
		errmsg = "Failed to add a row\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	rows, err = conn.Query(`SELECT * FROM GOODS WHERE project_id = ? AND name = ? AND description = ? ORDER BY created_at DESC LIMIT 1`,
		URLVars["projectId"],
		string(name),
		fmt.Sprintf("Entry with id \"%s\"", URLVars["projectId"]))
	if err != nil {
		errmsg = "Failed to select the created row\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	rows.Next()
	resp := types.Goods{}
	rows.Scan(&resp.ID, &resp.ProjectID, &resp.Name, &resp.Description, &resp.Priority, &resp.Removed, &resp.CreatedAt)
	_ = rows.Close()

	_ = redis_caching.InvalidateItems()

	itemLog := types.GoodsLog{
		ID:          resp.ID,
		ProjectID:   resp.ProjectID,
		Name:        resp.Name,
		Description: resp.Description,
		Priority:    resp.Priority,
		Removed:     resp.Removed,
		EventTime:   time.Now(),
	}
	itemJSON, _ := json.Marshal(itemLog)
	_ = nats_queueing.NatsConn.Publish(os.Getenv("NATS_QUEUE"), itemJSON)

	JSONResp, err := json.Marshal(resp)
	if err != nil {
		errmsg = "Row created, yet failed to marshal row data into JSON\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s}", errmsg)))
		return
	}
	var out bytes.Buffer
	json.Indent(&out, JSONResp, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out.Bytes())

}

func Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "get called"}`))
}

func Remove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "get called"}`))
}

func List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "get called"}`))
}

func Reprioritize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "get called"}`))
}
