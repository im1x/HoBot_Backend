package schedule

import (
	"HoBot_Backend/internal/service/chat"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/gofiber/fiber/v2/log"
)

func Start(lasqaService *chat.LasqaService) {
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Error("Error while creating scheduler:", err)
	}

	_, err = s.NewJob(
		gocron.DurationJob(30*time.Second),
		gocron.NewTask(lasqaService.CheckDonationAlertsStatus),
	)
	if err != nil {
		log.Error("Error while creating job:", err)
	}

	s.Start()
}
