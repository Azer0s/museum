package main

import (
	"fmt"
	"museum/cmd/proxy"
	"os"
)

func printUsage() {
	fmt.Println("Usage: museum <command>")
	fmt.Println("Commands:")
	fmt.Println("\tproxy")
	fmt.Println("\t- Starts the proxy server")
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
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "proxy":
		proxy.Run()
	default:
		printUsage()
	}
}
