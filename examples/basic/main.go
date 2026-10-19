package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/mcMineyC/vx_puppet"
)

func main() {
	ctx := context.Background()
	const lightpandaPort = 9222

	// CLI flags
	var (
		schoolID      = flag.String("school", "", "School ID")
		username      = flag.String("user", "", "Username / student ID")
		password      = flag.String("pass", "", "Password")
		lightpandaURL = flag.String("lightpandaUrl", "", "Lightpanda DevTools WebSocket URL")
		timeout       = flag.Duration("timeout", 90*time.Second, "Request timeout (e.g. 30s, 2m)")
		machine       = flag.Bool("json", false, "Output JSON")
	)

	flag.Parse()

	if *username == "" || *password == "" || *schoolID == "" {
		log.Fatal("schoolID, username, and password are required (-school, -user, -pass)")
	}

	webSocketURL := *lightpandaURL
	if webSocketURL != "" {
		if !strings.Contains(webSocketURL, "://") {
			webSocketURL = "ws://" + webSocketURL
		}

		parsedURL, err := url.Parse(webSocketURL)
		if err != nil || parsedURL.Host == "" {
			log.Fatalf("invalid -lightpandaUrl %q: expected a host:port or WebSocket URL", *lightpandaURL)
		}
		if parsedURL.Scheme != "ws" && parsedURL.Scheme != "wss" && parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			log.Fatalf("invalid -lightpandaUrl scheme %q: expected ws, wss, http, or https", parsedURL.Scheme)
		}
		if parsedURL.Path == "" {
			parsedURL.Path = "/"
		}
		webSocketURL = parsedURL.String()
	}

	if webSocketURL == "" {
		if runtime.GOOS == "windows" {
			log.Fatal("the Lightpanda executable is not compiled for Windows; provide -lightpandaUrl to connect to a running instance, or run this program on Linux or macOS")
		}

		webSocketURL = fmt.Sprintf("ws://127.0.0.1:%d/", lightpandaPort)
		binaryPath, err := vx_puppet.DownloadLightpanda(ctx, "lightpanda")
		if err != nil {
			log.Fatal(err)
		}

		browser, err := vx_puppet.StartLightpanda(ctx, binaryPath, lightpandaPort)
		if err != nil {
			log.Fatal(err)
		}
		defer browser.Close()
	}

	client, err := vx_puppet.New(vx_puppet.Options{
		WebSocketURL: webSocketURL,
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

	if !(*machine) {
		fmt.Println("Fetching grades...\n")
	}
	grades, err := client.GetGrades()
	if err != nil {
		log.Fatalf("Fatal error in grades: %s", err)
	}
	if !(*machine) {
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
