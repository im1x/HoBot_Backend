package socketio

import (
	"fmt"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/socket.io/v2/socket"
	"os"
)

func Start() {
	httpServer := types.CreateServer(nil)

	serverOptions := socket.DefaultServerOptions()
	cors := &types.Cors{
		Origin:               "*",            // Replace with your allowed origin(s)
		Methods:              "GET,POST",     // Replace with your allowed HTTP methods
		AllowedHeaders:       "Content-Type", // Replace with your allowed headers
		ExposedHeaders:       "",             // Replace with your exposed headers
		MaxAge:               "3600",         // Replace with your max age
		Credentials:          true,           // Set to true if you want to allow credentials (cookies, etc.)
		PreflightContinue:    false,          // Set to true if you want to continue with preflight requests
		OptionsSuccessStatus: 204,            // Replace with your desired status for preflight success
	}

	serverOptions.SetCors(cors)
	io := socket.NewServer(httpServer, serverOptions)

	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		fmt.Printf("connection %v\n", client.Id())
		client.On("event", func(datas ...any) {
			fmt.Printf("event %v\n", datas)
		})
		client.On("disconnect", func(...any) {
			fmt.Printf("disconnect %v\n", client.Id())
		})
	})
	fmt.Println(" ┌───────────────────────────────────────────────────┐ ")
	fmt.Print(" │       Socket.IO Server running on port: " + os.Getenv("WS_PORT") + "      │ ")
	httpServer.Listen(":"+os.Getenv("WS_PORT"), nil)
}
