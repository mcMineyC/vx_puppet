package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mcMineyC/vx_puppet"
)

func main() {
	ctx := context.Background()

	// CLI flags
	var (
		schoolID = flag.String("school", "", "School ID")
		username = flag.String("user", "", "Username / student ID")
		password = flag.String("pass", "", "Password")
		port     = flag.Int("port", 9222, "Chromedp-compatible DevTools browser port")
		timeout  = flag.Duration("timeout", 90*time.Second, "Request timeout (e.g. 30s, 2m)")
		remote   = flag.Bool("external", false, "Use external browser (assumes one is already running on the specified port)")
		machine  = flag.Bool("json", false, "Output JSON")
	)

	flag.Parse()

	if *username == "" || *password == "" || *schoolID == "" {
		log.Fatal("schoolID, username, and password are required (-school, -user, -pass)")
	}

	// Download Lightpanda automatically
	if !(*remote) {
		binaryPath, err := vx_puppet.DownloadLightpanda(ctx, "lightpanda")
		if err != nil {
			log.Fatal(err)
		}

		// Start browser
		browser, err := vx_puppet.StartLightpanda(ctx, binaryPath, *port)
		if err != nil {
			log.Fatal(err)
		}
		defer browser.Close()
	}

	client, err := vx_puppet.New(vx_puppet.Options{
		WebSocketURL: fmt.Sprintf("ws://127.0.0.1:%d/devtools/browser", *port),
		SchoolID:     *schoolID,
		Timeout:      *timeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	err = client.Login(*username, *password)
	if err != nil {
		log.Fatalf("Fatal error logging in: %s", err)
	}

	if !(*machine) {fmt.Println("Fetching grades...\n")}
	grades, err := client.GetGrades()
	if err != nil {
		log.Fatalf("Fatal error in grades: %s", err)
	}
	if !(*machine){
		for _, class := range grades {
			fmt.Printf("%s:\n - %s (%s)\n", class.Class, class.Grade, class.GradeLetter)
		}
	}

	if (*machine) == true {
		jsonData, err := json.MarshalIndent(grades, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(jsonData))
	}
}
