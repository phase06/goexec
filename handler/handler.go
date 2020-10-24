package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/gofiber/fiber/v2"
)

// Task for parse resquest from goexec
type Task struct {
	Executable string `json:"executable"`
	Parameters string `json:"parameters"`
	Status     string
	Result     string
}

// TaskDB reference in main
var TaskDB *badger.DB

// QueueTask to queue incoming request in temp db
func QueueTask(c *fiber.Ctx) error {
	p := new(Task)

	if err := c.BodyParser(p); err != nil {
		return err
	}

	p.Status = "NEW"

	enc, err := json.Marshal(p)
	if err != nil {
		return err
	}

	log.Println("Task receiced successfully from " + c.IP())
	taskID := generateKey()
	err = TaskDB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(taskID), []byte(enc))
		return err
	})

	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"exec_id": taskID,
	})
}

//CheckStatus to return task status and result
func CheckStatus(c *fiber.Ctx) error {
	taskID := c.Params("id")
	err := TaskDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(taskID))
		if err != nil {

			log.Fatal(err)
		}

		var data []byte
		err = item.Value(func(val []byte) error {

			data = append([]byte{}, val...)

			return nil
		})
		if err != nil {

			log.Fatal(err)
		}

		log.Printf("values: %s\n", data)
		p := new(Task)
		err = json.Unmarshal(data, &p)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"exec_id": taskID,
			"result":  p.Result,
			"status":  p.Status,
		})
	})

	if err != nil {
		return err
	}

	return nil
}

func generateKey() string {
	t := time.Now()
	key := t.Format("20060102") + "_" + t.Format("15:04:05.0000")
	key = strings.ReplaceAll(key, ":", "")
	key = strings.ReplaceAll(key, ".", "")

	return key
}

func run(executable string, parameters string) string {

	cmd := exec.Command(executable, parameters)
	// cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {

		log.Fatal(err)
	}

	return out.String()

}
