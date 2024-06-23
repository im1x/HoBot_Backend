package socketio

import (
	tokenService "HoBot_Backend/pkg/service/token"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/socket.io/v2/socket"
	"os"
)

var io *socket.Server

type SocketEvent string

const (
	SongRequestAdded      SocketEvent = "SongRequestAdded"
	SongRequestSetVolume  SocketEvent = "SongRequestSetVolume"
	SongRequestSkipSong   SocketEvent = "SongRequestSkipSong"
	SongRequestPlayPause  SocketEvent = "SongRequestPlayPause"
	SongRequestDeleteSong SocketEvent = "SongRequestDeleteSong"
	VotingStart           SocketEvent = "VotingStart"
	VotingVote            SocketEvent = "VotingVote"
	VotingStop            SocketEvent = "VotingStop"
	VotingDelete          SocketEvent = "VotingDelete"
)

func Start() {
	httpServer := types.NewWebServer(nil)

	serverOptions := socket.DefaultServerOptions()
	cors := &types.Cors{
		Origin:         "*",
		Methods:        "GET,POST",
		AllowedHeaders: "Content-Type",
		Credentials:    true,
	}

	serverOptions.SetCors(cors)
	io = socket.NewServer(httpServer, serverOptions)

	io.Use(func(s *socket.Socket, next func(*socket.ExtendedError)) {
		auth := s.Handshake().Auth
		if auth == nil {
			next(socket.NewExtendedError("Unauthorized: auth not found", "401"))
			return
		}
		token, err := getParam(auth, "token")
		if err != nil {
			next(socket.NewExtendedError("Unauthorized: token not found", "401"))
			return
		}
		userDto, err := tokenService.ValidateAccessToken(token)
		if err != nil {
			next(socket.NewExtendedError("Unauthorized: invalid token", "401"))
			return
		}
		s.Join(socket.Room(userDto.Id))

		next(nil)
	})

	/*	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)

		client.On("event", func(datas ...any) {
			fmt.Printf("event %v\n", datas)
		})
		client.On("testEmit", func(datas ...any) {
			fmt.Printf("testEmit %v\n", datas)
			rm := client.Rooms().Keys()[0]
			fmt.Printf("RoomSSSS: %v\n", client.Rooms())
			fmt.Printf("Room: %v\n", rm)
			var scks = io.Sockets().Adapter().Sockets(types.NewSet(rm)).Len()
			fmt.Printf("scks: %v\n", scks)

		})
		client.On("disconnect", func(...any) {
			fmt.Printf("disconnect %v\n", client.Id())
		})
	})*/
	fmt.Println(" ┌───────────────────────────────────────────────────┐ ")
	fmt.Print(" │       Socket.IO Server running on port: " + os.Getenv("WS_PORT") + "      │ ")
	httpServer.Listen(":"+os.Getenv("WS_PORT"), nil)
}

func getParam(mapData any, paramName string) (string, error) {
	paramString, ok := mapData.(map[string]interface{})[paramName].(string)
	if !ok {
		return "", fmt.Errorf("param %s not found", paramName)
	}
	return paramString, nil
}

func Emit(room string, event SocketEvent, data any) {
	err := io.In(socket.Room(room)).Emit(string(event), data)
	if err != nil {
		log.Error("Error while emitting event:", err)
		return
	}
}
