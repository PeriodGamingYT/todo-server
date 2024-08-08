package server

import (
	"os"
	"io"
	"fmt"
	"net"
	"errors"
	"net/http"
	"encoding/json"
)

type ChecklistItem struct {
	name string
	checked bool
	index int
}

type InventoryItem struct {
	name string
	current int
	max int
	index int
}

type ServerData struct {
	password string
	checklist map[string]ChecklistItem
	inventory map[string]InventoryItem
}

type checklistItemJson struct {
	Checked bool `json:"checked"`
	Index int `json:"index"`
}

type inventoryItemJson struct {
	Current int `json:"current"`
	Max int `json:"max"`
	Index int `json:"index"`
}

type serverDataJson struct {
	Checklist map[string]checklistItemJson `json:"checklist"`
	Inventory map[string]inventoryItemJson `json:"inventory"`
}

func (serverData *ServerData) Clear() {
	for key := range serverData.checklist { delete(serverData.checklist, key) }
	for key := range serverData.inventory { delete(serverData.inventory, key) }
	serverData.checklist = make(map[string]ChecklistItem)
	serverData.inventory = make(map[string]InventoryItem)
}

func (serverData *ServerData) LoadJSONBytes(bytes []byte) error {
	var serverJson serverDataJson
	err := json.Unmarshal(bytes, &serverJson)
	if err != nil { return err }
	serverData.Clear()
	for key := range serverJson.Checklist {
		item := serverJson.Checklist[key]
		serverData.checklist[key] = ChecklistItem {
			key,
			item.Checked,
			item.Index,
		}
	}

	for key := range serverJson.Inventory {
		item := serverJson.Inventory[key]
		serverData.inventory[key] = InventoryItem {
			key,
			item.Current,
			item.Max,
			item.Index,
		}
	}

	return nil
}

func (serverData *ServerData) LoadJSON(filename string) error {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	file, err := os.Open(filename)
	if err != nil { return err }
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil { return err }
	err = serverData.LoadJSONBytes(bytes)
	if err != nil { return err }
	return nil
}

func (serverData *ServerData) LoadPassword() error {
	file, err := os.Open("password.txt")
	if err != nil { fmt.Println("Can't find password.txt!"); return err }
	defer file.Close()
	result, err := io.ReadAll(file)
	if err != nil { return err }
	serverData.password = string(result[:])
	return nil
}

type dataRequestType int
const(
	ClientSave dataRequestType = iota
	ClientLoad
)

type dataRequest struct {
	Type dataRequestType `json:"type"`
	Checklist []ChecklistItem `json:"checklist"`
	Inventory []InventoryItem `json:"inventory"`
}

func DataHandler(writer http.ResponseWriter, req *http.Request) error {
	var data dataRequest
	err := json.NewDecoder(req.Body).Decode(&data)
	if err := nil { return err }

	// TODO(ElkElan) handle that request
	return nil
}

func Init() error {
	fmt.Println("Server starting...")
	var serverData ServerData
	err := serverData.LoadPassword()
	if err != nil { return err }
	err = serverData.LoadJSON("data.json")
	if err != nil { return err }
	var server http.Server
	server.Handler = http.HandlerFunc(DataHandler)
	fmt.Println("Listening and serving...")
	listen, err := net.Listen("tcp", "localhost:0")
	if err != nil { return err }
	fmt.Printf("Listening on %s\n", listen.Addr().String())
	server.Serve(listen)
	fmt.Println("Shutting down...")
	return nil
}
