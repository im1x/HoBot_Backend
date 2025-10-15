package socketio

import (
	tokenService "HoBot_Backend/internal/service/token"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2/log"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/socket.io/v2/socket"
)

type SocketEvent string

const (
	SongRequestAdded      SocketEvent = "SongRequestAdded"
	SongRequestSetVolume              = "SongRequestSetVolume"
	SongRequestSkipSong               = "SongRequestSkipSong"
	SongRequestPlayPause              = "SongRequestPlayPause"
	SongRequestDeleteSong             = "SongRequestDeleteSong"
	VotingStart                       = "VotingStart"
	VotingVote                        = "VotingVote"
	VotingStop                        = "VotingStop"
	VotingDelete                      = "VotingDelete"
)

type SocketServer struct {
	io *socket.Server
}

func NewSocketServer() *SocketServer {
	s := &SocketServer{}
	return s
}

func (s *SocketServer) Start() {
	httpServer := types.NewWebServer(nil)

	serverOptions := socket.DefaultServerOptions()
	cors := &types.Cors{
		Origin:         "*",
		Methods:        "GET,POST",
		AllowedHeaders: "Content-Type",
		Credentials:    true,
	}

	serverOptions.SetCors(cors)
	s.io = socket.NewServer(httpServer, serverOptions)

	s.io.Use(func(s *socket.Socket, next func(*socket.ExtendedError)) {
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

func (s *SocketServer) Emit(room string, event SocketEvent, data any) {
	err := s.io.In(socket.Room(room)).Emit(string(event), data)
	if err != nil {
		log.Error("Error while emitting event:", err)
		return
	}
}
