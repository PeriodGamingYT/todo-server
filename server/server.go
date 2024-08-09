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
	checked bool
	index int
}

type InventoryItem struct {
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
			item.Checked,
			item.Index,
		}
	}

	for key := range serverJson.Inventory {
		item := serverJson.Inventory[key]
		serverData.inventory[key] = InventoryItem {
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

func (serverData *ServerData) SaveJSON() error {
	var data serverDataJson
	data.Checklist = make(map[string]checklistItemJson)
	for key, value := range serverData.checklist {
		data.Checklist[key] = checklistItemJson {
			value.checked,
			value.index,
		}
	}

	data.Inventory = make(map[string]inventoryItemJson)
	for key, value := range serverData.inventory {
		data.Inventory[key] = inventoryItemJson {
			value.current,
			value.max,
			value.index,
		}
	}

	bytes, err := json.Marshal(data)
	if err != nil { return err }
	err = os.WriteFile("data.json", bytes, 0644)
	if err != nil { return err }
	return nil
}

type dataRequestType int
const(
	ClientSave dataRequestType = iota
	ClientLoad
	ClientTest
)

type dataRequest struct {
	Type dataRequestType `json:"type"`
	Password string `json:"password"`
	Checklist map[string]checklistItemJson `json:"checklist"`
	Inventory map[string]inventoryItemJson `json:"inventory"`
}

type dataResponse struct {
	IsSuccess bool `json:"success"`
	Checklist map[string]checklistItemJson `json:"checklist"`
	Inventory map[string]inventoryItemJson `json:"inventory"`
}

// not the best, wish i could just pass this into data handler via some
// pointer to user data
var globalServerData *ServerData
func sendResponse(
	writer http.ResponseWriter,
	isSuccess bool,
	includeData bool,
) error {
	response := dataResponse {
		isSuccess,
		nil,
		nil,
	}

	if isSuccess && includeData {
		response.Checklist = make(map[string]checklistItemJson)
		for key, value := range globalServerData.checklist {
			response.Checklist[key] = checklistItemJson {
				value.checked,
				value.index,
			}
		}

		response.Inventory = make(map[string]inventoryItemJson)
		for key, value := range globalServerData.inventory {
			response.Inventory[key] = inventoryItemJson {
				value.current,
				value.max,
				value.index,
			}
		}
	}

	bytes, err := json.Marshal(response)
	if err != nil { return err }
	_, err = writer.Write(bytes)
	if err != nil { return err }
	return nil
}

// the fact that there is no way to return errors because http.HandlerFunc doesn't
// have error as a return type is not ideal at all. so this is the next best thing to
// do
func DataHandler(writer http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling request...")
	bytes, err := io.ReadAll(req.Body)
	if err != nil { return }
	fmt.Printf("Bytes: %+v\n", string(bytes[:]))
	var data dataRequest
	err = json.Unmarshal(bytes, &data)
	fmt.Printf("Err: %+v\n", err)
	if err != nil { return }
	fmt.Printf("Data request value: %+v\n", data)

	// cybersecurity is my passion
	if data.Password != globalServerData.password {
		err = sendResponse(writer, false, false)
		if err != nil { return }
		return
	}

	switch(data.Type) {
		case ClientSave:
			if globalServerData.checklist == nil {
				globalServerData.checklist = make(map[string]ChecklistItem)
			}

			for key, value := range data.Checklist {
				globalServerData.checklist[key] = ChecklistItem {
					value.Checked,
					value.Index,
				}
			}

			if globalServerData.inventory == nil {
				globalServerData.inventory = make(map[string]InventoryItem)
			}

			for key, value := range data.Inventory {
				globalServerData.inventory[key] = InventoryItem {
					value.Current,
					value.Max,
					value.Index,
				}
			}

			err = sendResponse(writer, true, false)
			if err != nil { return }
			err = globalServerData.SaveJSON()
			if err != nil { return }

		case ClientLoad:
			err = sendResponse(writer, true, true)
			if err != nil { return }

		case ClientTest:
			err = sendResponse(writer, true, false)
			if err != nil { return }
	}

}

func Init() error {
	fmt.Println("Server starting...")
	var serverData ServerData
	err := serverData.LoadPassword()
	if err != nil { return err }
	err = serverData.LoadJSON("data.json")
	if err != nil { return err }
	serverData.checklist = make(map[string]ChecklistItem)
	serverData.inventory = make(map[string]InventoryItem)
	globalServerData = &serverData
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
