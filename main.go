package main

import (
	"codeforge-observer/daemon"
	"codeforge-observer/proxy"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printRootUsage()
		os.Exit(1)
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help" {
		printRootUsage()
		return
	}

	hostCmd := flag.NewFlagSet("host", flag.ExitOnError)
	action := hostCmd.String("action", "", "add | remove | list hosts")
	hostName := hostCmd.String("name", "", "set the host name")
	upstream := hostCmd.String("upstream", "", "host's upstream URL")
	contractFile := hostCmd.String("contract", "", "API contract file")
	resourceFile := hostCmd.String("resource", "", "Resource contract file")

	cmdArg := os.Args[1]
	switch cmdArg {
	// NOTE: Run the daemon in the background when the project is ready for deployment
	case "start":
		if err := daemon.RunDaemon(); err != nil {
			log.Fatal(err)
		}
	case "stop":
		if err := daemon.StopDaemon(); err != nil {
			log.Fatal(err)
		}
	case "host":
		if err := hostCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
		}

		switch *action {
		case "add":
			if *hostName == "" || *upstream == "" {
				log.Fatal("host add requires -name and -upstream")
			}

			if err := proxy.AddHostCommand(*hostName, *upstream, *contractFile, *resourceFile); err != nil {
				log.Fatal(err)
			}
		case "remove":
			if *hostName == "" {
				log.Fatal("host remove requires -name")
			}

			if err := proxy.RemoveHostCommand(*hostName); err != nil {
				log.Fatal(err)
			}
		case "list":
			if err := proxy.ListHostsCommand(); err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatal("host requires -action add|remove|list")
		}
	default:
		printRootUsage()
		os.Exit(1)
	}

}

func printRootUsage() {
	fmt.Println("Usage: cf-observer <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  start     Start the observer daemon")
	fmt.Println("  stop      Stop the observer daemon")
	fmt.Println("  host      Host management")
	fmt.Println()
	fmt.Println("Run 'cf-observer <command> -h' for command-specific help")
}
