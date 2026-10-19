# vx-puppet

Programmatic Veracross access built for students
No OAuth, it only requires your standard Veracross login

## Features

- Login
- Grade retrieval
- TODO: Fetch individual assignment grades

## Quick Start
The examples/basic directory has a great starter program that auto-downloads the lightpanda binary (on linux and mac, only; no windows sorry), starts it, and connects to it. Below is an example execution that connects to a lightpanda Docker container running on the remote machine `192.168.40.167:9222`:
```
go run main.go -school <redacted, replace with yours> -user 123456 -pass "MyPassword" -lightpandaUrl "192.168.40.167:9222"
```

## Installation

```bash
go get github.com/mcMineyC/vx-puppet
```

## Start ChromeDP-compatible server

```bash
google-chrome \
  --remote-debugging-port=9222
```
Note: anything remotely based on Chromium (eg Brave, Chromium, Helium, Arc, Vivaldi, etc) should work, just replace google-chrome with its executable name

Personally, I'm a fan of [lightpanda](https://lightpanda.io/docs/open-source/installation)
```bash
./lightpanda
```

I tried using [obscura](https://github.com/h4ckf0r0day/obscura), but it had trouble parsing the login form.

## Usage

```go
client, err := veracross.New(
	veracross.Options{
		WebSocketURL: "ws://127.0.0.1:9222/devtools/browser",
	},
)
```

## Login

```go
err = client.Login(
	context.Background(),
	"username",
	"password",
)
```

## Get Grades

```go
grades, err := client.GetGrades(
	context.Background(),
)
```

## Example Output

```json
[
  {
    "class": "English",
    "gradeLetter": "A",
    "grade": "97",
    "new_updates": 2
  }
]
```