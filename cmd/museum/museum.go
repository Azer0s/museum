package main

import (
	"encoding/json"
	"fmt"
	"github.com/hako/durafmt"
	"museum/cmd/server"
	"museum/cmd/tool"
	"museum/domain"
	"museum/util"
	"os"
	"time"
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
	fmt.Println("\tlist (--json)")
	fmt.Println("\t- Lists all exhibits")
	fmt.Println("\trenew <name> <lease>")
	fmt.Println("\t- Renews a lease on an exhibit")
	fmt.Println("\twarmup <name>")
	fmt.Println("\t- Warms up an exhibit")
}

func printSeparator() {
	w := util.GetTerminalWidth() / 2
	for i := 0; i < w; i++ {
		fmt.Print("─")
	}
	fmt.Println()
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
		baseUrl, exhibits, err := tool.List()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if len(os.Args) > 2 && os.Args[2] == "--json" {
			b, err := json.Marshal(exhibits)
			if err != nil {
				os.Exit(1)
			}
			fmt.Println(string(b))
			os.Exit(0)
		}

		for _, e := range exhibits {
			printSeparator()

			fmt.Println("🧮  " + e.Name)
			fmt.Print("    ")
			if e.RuntimeInfo.Status == domain.Running {
				fmt.Print("🟢 ")
			} else {
				fmt.Print("🔴 ")
			}
			fmt.Println(" " + baseUrl + "/exhibit/" + e.Id)

			if e.RuntimeInfo.Status == domain.Running {
				d, err := time.ParseDuration(e.Lease)
				if err != nil {
					panic(err)
				}
				fmt.Println("    ⏰‎  Expires in " + durafmt.Parse(time.Until(time.Unix(e.RuntimeInfo.LastAccessed, 0).Add(d)).Truncate(time.Second)).String() + " from now")
			} else {
				fmt.Println("    ⏰‎  Expired " + durafmt.Parse(time.Since(time.Unix(e.RuntimeInfo.LastAccessed, 0)).Truncate(time.Second)).String() + " ago")
			}

			fmt.Println("    🧺  exhibits:")
			for _, o := range e.Objects {
				fmt.Println("        📜  " + o.Name + " (" + o.Image + ")")
			}
		}

		printSeparator()
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
