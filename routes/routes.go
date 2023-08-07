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
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func Create(w http.ResponseWriter, r *http.Request) {
	var errmsg string
	var payload map[string]string
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
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
	var projectId int
	if projectId, err = strconv.Atoi(URLVars["projectId"]); err != nil {
		errmsg = "Failed to convert the project ID from string to int\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	rows, err := conn.Query(`SELECT * FROM GOODS WHERE project_id = ? and name = ?`, projectId, payload["name"])
	if err != nil {
		errmsg = "Failed to get rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if !rows.Next() {
		errmsg = "No selected rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	_, err = conn.Query(`SELECT * FROM GOODS WHERE project_id = ? and name = ?`, projectId, payload["name"])
	if err != nil {
		errmsg = "Failed to get rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s}", errmsg)))
		return
	}
	_, err = conn.Query(`INSERT INTO GOODS (project_id, name, description, priority) VALUES (?,?,?, (SELECT COALESCE(MAX(priority), 0) + 1 FROM GOODS))`,
		projectId,
		payload["name"],
		fmt.Sprintf("Entry with id \"%s\"", projectId))
	if err != nil {
		errmsg = "Failed to add a row\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	rows, err = conn.Query(`SELECT * FROM GOODS WHERE project_id = ? AND name = ? AND description = ? ORDER BY created_at DESC LIMIT 1`,
		projectId,
		payload["name"],
		fmt.Sprintf("Entry with id \"%s\"", projectId))
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
	err = rows.Scan(&resp.ID, &resp.ProjectID, &resp.Name, &resp.Description, &resp.Priority, &resp.Removed, &resp.CreatedAt)
	if err != nil {
		errmsg = "Failed to scan rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	_ = rows.Close()

	_ = redis_caching.InvalidateGoodsList()

	goodLog := types.GoodsLog{
		ID:          resp.ID,
		ProjectID:   resp.ProjectID,
		Name:        resp.Name,
		Description: resp.Description,
		Priority:    resp.Priority,
		Removed:     resp.Removed,
		EventTime:   time.Now(),
	}
	goodJSON, _ := json.Marshal(goodLog)
	_ = nats_queueing.NatsConn.Publish(os.Getenv("NATS_QUEUE"), goodJSON)

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
	var errmsg string
	var err error
	var payload map[string]string
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		errmsg = "Failed to decode payload\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	defer r.Body.Close()
	if err != nil {
		errmsg = "Failed to get a payload\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if _, ok := payload["name"]; ok == false || payload["name"] == "" {
		errmsg = "No name key in payload or name value is empty\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg)))
		return
	}
	URLVars := mux.Vars(r)
	tx, err := postgresql.DBConn.Begin()
	if err != nil {
		errmsg = "Failed to start a transaction\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	var id int
	var projectId int
	if id, err = strconv.Atoi(URLVars["id"]); err != nil {
		errmsg = "Failed to convert the ID from string to int\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if projectId, err = strconv.Atoi(URLVars["projectId"]); err != nil {
		errmsg = "Failed to convert the project ID from string to int\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	rows, err := tx.Query(`SELECT * FROM GOODS WHERE id = ? AND project_id = ? FOR UPDATE`, id, projectId)
	if err != nil {
		_ = tx.Rollback()
		errmsg = "Failed to select the row for update\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if !rows.Next() {
		_ = tx.Rollback()
		errmsg = "No selected rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	_ = rows.Close()
	if _, ok := payload["description"]; !ok {
		_, err = tx.Exec("UPDATE GOODS SET name = ? WHERE id = ? AND project_id = ?",
			payload["name"],
			id,
			projectId)
	} else {
		_, err = tx.Exec("UPDATE GOODS SET name = ?, description = ? WHERE id = ? AND project_id = ?",
			payload["name"],
			payload["description"],
			id,
			projectId)
	}
	if err != nil {
		_ = tx.Rollback()
		errmsg = "Failed to update the row\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	rows, err = tx.Query(`SELECT * FROM GOODS WHERE id = ? AND project_id = ?`, id, projectId)
	if err != nil {
		_ = tx.Rollback()
		errmsg = "Failed to select the row\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	rows.Next()
	resp := types.Goods{}
	err = rows.Scan(&resp.ID, &resp.ProjectID, &resp.Name, &resp.Description, &resp.Priority, &resp.Removed, &resp.CreatedAt)
	if err != nil {
		_ = tx.Rollback()
		errmsg = "Failed to scan rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	_ = rows.Close()

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		errmsg = "Failed to commit transaction\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}

	_ = redis_caching.InvalidateGoodsList()

	goodLog := types.GoodsLog{
		ID:          resp.ID,
		ProjectID:   resp.ProjectID,
		Name:        resp.Name,
		Description: resp.Description,
		Priority:    resp.Priority,
		Removed:     resp.Removed,
		EventTime:   time.Now(),
	}
	goodJSON, _ := json.Marshal(goodLog)
	_ = nats_queueing.NatsConn.Publish(os.Getenv("NATS_QUEUE"), goodJSON)

	JSONResp, err := json.Marshal(resp)
	if err != nil {
		_ = tx.Rollback()
		errmsg = "Row updated, yet failed to marshal row data into JSON\n"
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

func Remove(w http.ResponseWriter, r *http.Request) {
	var errmsg string
	var err error
	URLVars := mux.Vars(r)
	conn := postgresql.DBConn
	var id int
	var projectId int
	if id, err = strconv.Atoi(URLVars["id"]); err != nil {
		errmsg = "Failed to convert the ID from string to int\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if projectId, err = strconv.Atoi(URLVars["projectId"]); err != nil {
		errmsg = "Failed to convert the project ID from string to int\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	rows, err := conn.Query(`SELECT * FROM GOODS WHERE id = ? AND project_id = ?`, id, projectId)
	if err != nil {
		errmsg = "Failed to select the row for update (deletion)\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if !rows.Next() {
		errmsg = "No selected rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"code\": 3, \"message\": errors.good.notFound, \"details\": {}}")))
		return
	}
	_, err = conn.Exec("UPDATE GOODS SET removed = true WHERE id = ? AND project_id = ?",
		id,
		projectId)
	if err != nil {
		errmsg = "Failed to update (delete) the row\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	resp := types.Goods{}
	err = rows.Scan(&resp.ID, &resp.ProjectID, &resp.Name, &resp.Description, &resp.Priority, &resp.Removed, &resp.CreatedAt)
	if err != nil {
		errmsg = "Failed to scan rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	_ = rows.Close()

	_ = redis_caching.InvalidateGoodsList()

	goodLog := types.GoodsLog{
		ID:          resp.ID,
		ProjectID:   resp.ProjectID,
		Name:        resp.Name,
		Description: resp.Description,
		Priority:    resp.Priority,
		Removed:     true,
		EventTime:   time.Now(),
	}
	goodJSON, _ := json.Marshal(goodLog)
	_ = nats_queueing.NatsConn.Publish(os.Getenv("NATS_QUEUE"), goodJSON)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\"id\": %d, \"projectId\": %d, \"removed\": %t}", resp.ID, resp.ProjectID, resp.Removed)))
}

func List(w http.ResponseWriter, r *http.Request) {
	var errmsg string
	var err error
	URLVars := mux.Vars(r)
	conn := postgresql.DBConn
	var limit int
	var offset int
	var redis_limit int
	var redis_offset int
	var ok bool
	var resp types.List
	if limit, err = strconv.Atoi(URLVars["limit"]); err != nil {
		errmsg = "Failed to convert the ID from string to int\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if offset, err = strconv.Atoi(URLVars["offset"]); err != nil {
		errmsg = "Failed to convert the project ID from string to int\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if ok, resp, redis_limit, redis_offset, err = redis_caching.GetGoodsList(); ok && redis_limit == limit && redis_offset == offset && err == nil {
		JSONResp, err := json.Marshal(resp)
		if err != nil {
			errmsg = "Failed to marshal row data into JSON\n"
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
		return
	}

	resp.Meta.Limit = limit
	resp.Meta.Offset = offset
	rows, err := conn.Query(`SELECT count(*) FROM GOODS`)
	if err != nil {
		errmsg = "Failed to select the row count\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if !rows.Next() {
		errmsg = "No selected rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	err = rows.Scan(&resp.Meta.Total)
	_ = rows.Close()
	rows, err = conn.Query(`SELECT count(*) FROM GOODS WHERE removed = true`)
	if err != nil {
		errmsg = "Failed to select the removed row count\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if !rows.Next() {
		errmsg = "No selected rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	err = rows.Scan(&resp.Meta.Removed)
	_ = rows.Close()
	rows, err = conn.Query(`SELECT * FROM GOODS ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		errmsg = "Failed to select the removed row count\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	goods := types.Goods{}
	for rows.Next() {
		err = rows.Scan(&goods.ID, &goods.ProjectID, &goods.Name, &goods.Description, &goods.Priority, &goods.Removed, &goods.CreatedAt)
		if err != nil {
			errmsg = "Failed to scan rows\n"
			log.Printf(errmsg)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
			return
		}
		resp.Goods = append(resp.Goods, goods)
	}
	redis_caching.SetGoodsList(resp, limit, offset)
	JSONResp, err := json.Marshal(resp)
	if err != nil {
		errmsg = "Failed to marshal row data into JSON\n"
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

func Reprioritize(w http.ResponseWriter, r *http.Request) {
	var errmsg string
	var err error
	var payload map[string]int
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		errmsg = "Failed to decode payload\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	defer r.Body.Close()
	URLVars := mux.Vars(r)
	conn := postgresql.DBConn
	var id int
	var projectId int
	if id, err = strconv.Atoi(URLVars["id"]); err != nil {
		errmsg = "Failed to convert the ID from string to int\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if projectId, err = strconv.Atoi(URLVars["projectId"]); err != nil {
		errmsg = "Failed to convert the project ID from string to int\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if _, ok := payload["newPriority"]; ok == false || payload["newPriority"] == 0 {
		errmsg = "No name key in payload or name value is zero\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg)))
		return
	}
	rows, err := conn.Query(`SELECT * FROM GOODS WHERE id = ? AND project_id = ?`, id, projectId)
	if err != nil {
		errmsg = "Failed to select the row for reprioritizing\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	if !rows.Next() {
		errmsg = "No selected rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"code\": 3, \"message\": errors.good.notFound, \"details\": {}}")))
		return
	}
	_, err = conn.Exec("UPDATE GOODS SET priority = ? WHERE id = ? AND project_id = ?",
		payload["newPriority"],
		id,
		projectId)
	if err != nil {
		errmsg = "Failed to reprioritize the row\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	good := types.Goods{}
	resp := types.Priorities{}
	priority := types.Priority{}
	err = rows.Scan(&good.ID, &good.ProjectID, &good.Name, &good.Description, &good.Priority, &good.Removed, &good.CreatedAt)
	if err != nil {
		errmsg = "Failed to scan rows\n"
		log.Printf(errmsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("{\"message\": %s, %s}", errmsg, err.Error())))
		return
	}
	_ = rows.Close()

	_ = redis_caching.InvalidateGoodsList()

	goodLog := types.GoodsLog{
		ID:          good.ID,
		ProjectID:   good.ProjectID,
		Name:        good.Name,
		Description: good.Description,
		Priority:    payload["newPriority"],
		Removed:     good.Removed,
		EventTime:   time.Now(),
	}
	goodJSON, _ := json.Marshal(goodLog)
	_ = nats_queueing.NatsConn.Publish(os.Getenv("NATS_QUEUE"), goodJSON)

	priority.ID = id
	priority.Priority = payload["newPriority"]
	resp.Priorities = append(resp.Priorities, priority)

	JSONResp, err := json.Marshal(resp)
	if err != nil {
		errmsg = "Failed to marshal row data into JSON\n"
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
