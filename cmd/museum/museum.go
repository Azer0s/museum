package main

import (
	"fmt"
	"museum/cmd/server"
	"museum/cmd/tool"
	"os"
)

func printUsage(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Usage: museum <command>")
	fmt.Println("Commands:")
	fmt.Println("\tserver")
	fmt.Println("\t- Starts the mūsēum API and proxy server")
	fmt.Println("\tcreate <file>")
	fmt.Println("\t- Creates a new exhibit")
	fmt.Println("\tdelete <name>")
	fmt.Println("\t- Deletes a exhibit")
	fmt.Println("\tlist")
	fmt.Println("\t- Lists all exhibits")
	fmt.Println("\trenew <name> <lease>")
	fmt.Println("\t- Renews a lease on an exhibit")
	fmt.Println("\twarmup <name>")
	fmt.Println("\t- Warms up an exhibit")
}

func main() {
	if len(os.Args) < 2 {
		printUsage(nil)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		server.Run()
	case "create":
		err := tool.Create()
		if err != nil {
			printUsage(err)
			os.Exit(1)
		}
	case "delete":
		err := tool.Delete()
		if err != nil {
			printUsage(err)
			os.Exit(1)
		}
	case "list":
		if tool.List() != nil {
			printUsage(nil)
			os.Exit(1)
		}
	}
}
