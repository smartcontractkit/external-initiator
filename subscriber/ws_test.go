package subscriber

import (
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options
