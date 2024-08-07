package server

import (
	"os"
	"fmt"
	"io"
	"net/http"
	"net"
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

func (serverData ServerData) LoadPassword() error {
	file, err := os.Open("password.txt")
	if err != nil { fmt.Println("Can't find password.txt!"); return err }
	defer file.Close()
	result, err := io.ReadAll(file)
	if err != nil { return err }
	serverData.password = string(result[:])
	return nil
}

func DataHandler(writer http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(writer, "hi!\n")
}

func Init() error {
	fmt.Println("Server starting...")
	var data ServerData
	err := data.LoadPassword()
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
