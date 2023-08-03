package nats_queueing

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"hezzl_task_5/clickhouse_logging"
	"hezzl_task_5/types"
	"log"
	"os"
)

var NatsConn *nats.Conn

func NatsConnect() error {
	conn, err := nats.Connect(fmt.Sprintf("nats://%s:4222", os.Getenv("NATS_HOST")))
	if err != nil {
		return err
	}

	NatsConn = conn
	return nil
}

func NatsSubscribe() error {
	_, err := NatsConn.Subscribe(os.Getenv("NATS_QUEUE"), func(msg *nats.Msg) {
		itemJSON := string(msg.Data)
		var goods types.GoodsLog
		_ = json.Unmarshal([]byte(itemJSON), &goods)

		_, err := clickhouse_logging.ClickhouseDBConn.Query(`INSERT INTO GOODS (Id, Name, ProjectId, Description, Priority, Removed, EventTime) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			goods.ID,
			goods.Name,
			goods.ProjectID,
			goods.Description,
			goods.Priority,
			goods.Removed,
			goods.EventTime)
		if err != nil {
			log.Printf("Failed to send a log to CH: %s\n", err.Error())
		}
	})
	if err != nil {
		return err
	}
	return nil
}
