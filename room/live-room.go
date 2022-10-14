package room

import (
	"encoding/json"
	"fmt"
	"net/http"

	socketIO "github.com/googollee/go-socket.io"
	"github.com/labstack/echo/v4"
)

// UserRoomInfo 用户房间信息
type UserRoomInfo struct {
	RoomId   string `json:"roomId"`
	UserName string `json:"userName"`
}

// LiveServer 直播服务
type LiveServer struct {
	server       *socketIO.Server
	port         string
	httpNSpace   string
	serverNSpace string

	e *echo.Echo
}

func BuildLiveServer() *LiveServer {
	liveServer := &LiveServer{}
	liveServer.port = "127.0.0.1:8085"
	liveServer.httpNSpace = "/socket.io/"
	liveServer.serverNSpace = "/socket.io"
	liveServer.e = echo.New()

	liveServer.e.Group("/socket.io").Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			liveServer.server.ServeHTTP(c.Response(), c.Request())
			return nil
		}
	})

	liveServer.e.Static("/node_modules/", "./node_modules/")
	liveServer.e.Static("/", "./static/")

	return liveServer
}

func (l *LiveServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	l.server.ServeHTTP(w, r)
}

// StartHost 开始host
func (l *LiveServer) StartHost() {
	l.server = socketIO.NewServer(nil)
	l.server.OnConnect("/", func(conn socketIO.Conn) error {
		fmt.Printf("connected. sid:%v", conn.ID())
		return nil
	})
	l.server.OnDisconnect(l.serverNSpace, func(conn socketIO.Conn, s string) {
		fmt.Printf("disconnect sid \n :%v", conn.ID())
	})
	l.server.OnError(l.serverNSpace, func(conn socketIO.Conn, err error) {
		fmt.Printf("error \n %v", err)
	})
	// 加入房间
	l.server.OnEvent(l.serverNSpace, "join-room", l.joinRoom)
	// peer 广播
	l.server.OnEvent(l.serverNSpace, "broadcast", l.broadcast)

	go l.server.Serve()
	defer l.server.Close()

	// http.Handle(l.httpNSpace, l)
	// http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(l.staticPath))))

	fmt.Printf("serving ok.port:%s", l.port)

	err := http.ListenAndServe(l.port, l.e)
	if err != nil {
		fmt.Printf("listen err. %v", err)
	}
}

// 加入房间
func (l *LiveServer) joinRoom(conn socketIO.Conn, msg string) {
	var userRoomInfo UserRoomInfo
	err := json.Unmarshal([]byte(msg), &userRoomInfo)
	if err != nil {
		fmt.Printf("join room .%v", err)
		return
	}
	fmt.Printf("jon room.%v", userRoomInfo)
	l.server.JoinRoom("/socket.io", userRoomInfo.RoomId, conn)
	l.broadcastTo(l.server, conn.Rooms(), "user-joined", userRoomInfo.UserName)
}

// 处理广播
func (l *LiveServer) broadcast(conn socketIO.Conn, msg interface{}) {
	l.broadcastTo(l.server, conn.Rooms(), "broadcast", msg)
}

// 广播房间事件
func (l *LiveServer) broadcastTo(server *socketIO.Server, rooms []string, event string, msg interface{}) {
	fmt.Printf("broadcast to .\n %v \n", msg)
	if len(rooms) == 0 {
		fmt.Println("broadcast rooms is null.")
		return
	}
	for _, room := range rooms {
		server.BroadcastToRoom("/socket.io", room, event, msg)
	}
}
