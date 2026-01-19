<div align="center">

  [![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/DenisKozarezov/ws-battleship/blob/master/README.md)
  [![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://golang.org/)
  [![WebSocket](https://img.shields.io/badge/Protocol-WebSocket-yellow)](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)

  <h1>Battleship Websocket Game</h1>

</div>

Real-time multiplayer naval combat in classic way. Gaming programming patterns are widely used in this project, such as Command, Event Bus, concurrency via Goroutines, Model/View, etc...

Implemented Features:
* Dedicated Go server for game logic and matchmaking
* WebSocket connections for real-time sync
* Terminal clients with interactive UI. No ReactJS, no Angular or any frontend. Just poor terminal.

## Usage

If you want to play straight in IDE/editor, then download the project and run `server` and `client` Go-files. Make sure you have installed Go with a minimum version 1.25:
```shell
# git clone
git clone https://github.com/DenisKozarezov/ws-battleship.git

# open the project
cd ws-battleship

# start Go-files
go run ws-battleship-server/cmd/main.go
go run ws-battleship-client/cmd/main.go
```

Local Multiplayer is available at your `localhost:8080`, so try to connect to `127.0.0.1` or `0.0.0.0` IP-address of the running server.

> [!NOTE]
> Server is configured at port :8080 by default.

## Demo

### 1. Main Menu

<img width="1280" src="docs/main-menu-connect-error.gif" />

### 2. Multiplayer Game

<img width="1280" src="docs/multiplayer.gif" />

### 3. Chat

<img width="1280" src="docs/chat.gif" />