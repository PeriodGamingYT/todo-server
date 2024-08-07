package main
import (
	"os"
	"fmt"
	"todoserver/server"
)

func main() {
	err := server.Init()
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
