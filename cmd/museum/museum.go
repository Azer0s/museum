package main

import (
	"fmt"
	"museum/cmd/server"
	"museum/cmd/tool"
	"os"
)

func printUsage() {
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
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		server.Run()
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("❌ missing file argument")
			os.Exit(1)
		}
		exhibit, url, err := tool.Create(os.Args[2])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("🧑‍🎨 exhibit " + exhibit.Name + " created successfully")
		fmt.Println("‎‎‎👉 " + url)
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("❌ missing id argument")
			os.Exit(1)
		}
		err := tool.Delete(os.Args[2])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("‎‎‎🗑️ exhibit deleted successfully")
	case "list":
		_, err := tool.List()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	case "warmup":
		if len(os.Args) < 3 {
			fmt.Println("❌ missing id argument")
			os.Exit(1)
		}
		url, err := tool.Warmup(os.Args[2])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println("‎‎‎🔥 exhibit warmed up successfully")
		fmt.Println("‎‎‎👉 " + url)
	default:
		printUsage()
	}
}
