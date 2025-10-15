package main

import (
	"HoBot_Backend/internal/handler"
	"HoBot_Backend/internal/mongodb"
	repoConfig "HoBot_Backend/internal/repository/config"
	repoFeedback "HoBot_Backend/internal/repository/feedback"
	repoPrivilegedLasqaKp "HoBot_Backend/internal/repository/privilegedlasqakp"
	repoSettingsOptions "HoBot_Backend/internal/repository/settingsoptions"
	repoSongRequests "HoBot_Backend/internal/repository/songrequests"
	repoSongRequestsHistory "HoBot_Backend/internal/repository/songrequestshistory"
	repoStatistics "HoBot_Backend/internal/repository/statistics"
	repoToken "HoBot_Backend/internal/repository/token"
	repoUser "HoBot_Backend/internal/repository/user"
	repoUserSettings "HoBot_Backend/internal/repository/usersettings"
	repoVkpl "HoBot_Backend/internal/repository/vkpl"
	"HoBot_Backend/internal/router"
	"HoBot_Backend/internal/schedule"
	"HoBot_Backend/internal/service/chat"
	"HoBot_Backend/internal/service/common"
	"HoBot_Backend/internal/service/settings"
	"HoBot_Backend/internal/service/songRequest"
	"HoBot_Backend/internal/service/user"
	"HoBot_Backend/internal/service/vkplay"
	"HoBot_Backend/internal/service/voting"
	"HoBot_Backend/internal/socketio"
	"context"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	//env
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found")
	}
	if ok := testEnvs([]string{
		"PORT",
		"IPV6_ONLY",
		"MONGODB_URI",
		"DB_NAME",
		"JWT_ACCESS_SECRET",
		"JWT_REFRESH_SECRET",
		"WS_PORT",
		"VKPL_LOGIN",
		"VKPL_PASSWORD",
		"VKPL_APP_CREDEANTIALS",
		"BOT_VKPL_ID"}); !ok {
		log.Fatalln("Please add required envs")
	}

	ctx := context.Background()

	//MongoDB
	db, err := mongodb.NewClient(ctx, os.Getenv("MONGODB_URI"), os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatal("Error while connecting to MongoDB: ", err)
	}
	defer db.Close(ctx)

	//Repositories
	vkplRepo := repoVkpl.NewVkplRepository(db)
	configRepo := repoConfig.NewConfigRepository(db)
	userSettingsRepo := repoUserSettings.NewUserSettingsRepository(db)
	feedbackRepo := repoFeedback.NewFeedbackRepository(db)
	settingsOptionsRepo := repoSettingsOptions.NewSettingsOptionsRepository(db)
	songRequestsRepo := repoSongRequests.NewSongRequestsRepository(db)
	songRequestsHistoryRepo := repoSongRequestsHistory.NewSongRequestsHistoryRepository(db)
	userRepo := repoUser.NewUserRepository(db)
	tokenRepo := repoToken.NewTokenRepository(db)
	statisticsRepo := repoStatistics.NewStatisticsRepository(db)
	lasqaKpRepo := repoPrivilegedLasqaKp.NewLasqaKpRepository(db)

	//Services
	socketioServer := socketio.NewSocketServer()

	vkplService := vkplay.NewVkplService(ctx, userSettingsRepo, vkplRepo)
	commonService := common.NewCommonService(ctx, feedbackRepo)
	settingsService := settings.NewSettingsService(ctx, userSettingsRepo, settingsOptionsRepo, vkplService)
	songRequestsService := songRequest.NewSongRequestService(ctx, songRequestsRepo, songRequestsHistoryRepo, settingsService)
	votingService := voting.NewVotingService(ctx, statisticsRepo, socketioServer)

	chatService := chat.NewChatService(ctx, socketioServer, vkplRepo, userRepo, configRepo, vkplService, votingService)
	lasqaService := chat.NewLasqaService(lasqaKpRepo, userRepo, chatService)
	commandService := chat.NewCommandService(ctx, userRepo, songRequestsRepo, statisticsRepo, userSettingsRepo, settingsService, songRequestsService, socketioServer, chatService, lasqaService)
	chatService.SetCommandService(commandService)
	userService := user.NewUserService(ctx, userRepo, userSettingsRepo, songRequestsRepo, songRequestsHistoryRepo, tokenRepo, vkplService, settingsService, chatService)

	validate := validator.New()
	// handlers
	commonHandler := handler.NewCommonHandler(commonService)
	settingHandler := handler.NewSettingHandler(validate, userSettingsRepo, settingsService)
	songRequestHandler := handler.NewSongRequestHandler(songRequestsRepo, userRepo, songRequestsService, chatService)
	userHandler := handler.NewUserHandler(validate, userService, userRepo)
	votingHandler := handler.NewVotingHandler(votingService)

	go socketioServer.Start()
	chatService.Start()
	schedule.Start(lasqaService)

	//Http server
	var app *fiber.App
	if os.Getenv("IPV6_ONLY") == "true" {
		app = fiber.New(fiber.Config{
			Network: "tcp6",
		})
	} else {
		app = fiber.New()
	}

	router.Register(app, commonHandler, settingHandler, songRequestHandler, userHandler, votingHandler)

	log.Fatal(app.Listen(":" + os.Getenv("PORT")))
}

func testEnvs(enums []string) bool {
	successful := true
	for _, enum := range enums {
		if _, ok := os.LookupEnv(enum); !ok {
			successful = false
			log.Printf("Env \"%s\" not found\n", enum)
		}
	}
	return successful
}
