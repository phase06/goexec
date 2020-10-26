package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	time "time"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/gofiber/fiber/v2"
)

// Task for parse resquest from goexec
type Task struct {
	Executable string `json:"executable"`
	Parameters string `json:"parameters"`
	Status     string
	Result     string
	Output     string
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

			log.Println(err)
		}

		var data []byte
		err = item.Value(func(val []byte) error {

			data = append([]byte{}, val...)

			return nil
		})
		if err != nil {

			log.Println(err)
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

// ProcessQueue to find new task in db
func ProcessQueue() error {

	err := TaskDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			var data []byte
			err := item.Value(func(v []byte) error {
				fmt.Printf("key=%s, value=%s\n", k, v)
				data = append([]byte{}, v...)
				return nil
			})
			if err != nil {
				return err
			}

			p := new(Task)
			err = json.Unmarshal(data, &p)
			if err != nil {
				return err
			}

			if p.Status == "NEW" {
				fmt.Printf("Running exec_id=%s, executable=%s, parameters=%s  \n", k, p.Executable, p.Parameters)
				err = updateTask(string(k), "WIP", "", "")
				if err != nil {
					return err
				}
				result, output := run(p.Executable, p.Parameters)
				if result == true {
					updateTask(string(k), "DONE", "PASS", output)
				} else {
					updateTask(string(k), "DONE", "FAIL", output)
				}

			}
		}
		return nil
	})

	return err
}

func updateTask(taskID string, status string, result string, output string) error {

	err := TaskDB.Update(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte(taskID))
		if err != nil {

			log.Println(err)
		}

		var data []byte
		err = item.Value(func(val []byte) error {

			data = append([]byte{}, val...)

			return nil
		})
		if err != nil {

			log.Println(err)
		}

		log.Printf("values: %s\n", data)
		p := new(Task)
		err = json.Unmarshal(data, &p)

		p.Status = status
		p.Result = result
		p.Output = output

		enc, err := json.Marshal(p)
		if err != nil {
			return err
		}

		e := badger.NewEntry([]byte(taskID), []byte(enc))
		err = txn.SetEntry(e)

		return err
	})
	return err
}

func generateKey() string {
	t := time.Now()
	key := t.Format("20060102") + "_" + t.Format("15:04:05.0000")
	key = strings.ReplaceAll(key, ":", "")
	key = strings.ReplaceAll(key, ".", "")

	return key
}

func run(executable string, parameters string) (bool, string) {

	cmd := exec.Command(executable, parameters)
	// cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// log.Fatal("here")

		log.Println(err)
		return false, err.Error()
	}

	return (err == nil), out.String()

}
