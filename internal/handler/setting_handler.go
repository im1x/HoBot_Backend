package handler

import (
	"HoBot_Backend/internal/model"
	repoUserSettings "HoBot_Backend/internal/repository/usersettings"
	"HoBot_Backend/internal/service/settings"
	"errors"
	"net/url"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type SettingHandler struct {
	validate         *validator.Validate
	userSettingsRepo repoUserSettings.Repository
	settingsService  *settings.SettingsService
}

func NewSettingHandler(validate *validator.Validate, userSettingsRepo repoUserSettings.Repository, settingsService *settings.SettingsService) *SettingHandler {
	return &SettingHandler{validate: validate, userSettingsRepo: userSettingsRepo, settingsService: settingsService}
}

func (s *SettingHandler) GetCommands(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	commands, err := s.settingsService.GetCommands(c.Context(), userId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(commands)
}
func (s *SettingHandler) GetCommandsDropdown(c *fiber.Ctx) error {
	commandsList, err := s.settingsService.GetCommandsList()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(commandsList.Commands)
}

func (s *SettingHandler) AddCommandAndAlias(c *fiber.Ctx) error {
	newCommand := new(model.CommonCommand)

	if err := c.BodyParser(&newCommand); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := s.validate.Struct(newCommand); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	userId := parseUserIdFromRequest(c)

	commandList, err := s.settingsService.AddCommandForUser(c.Context(), userId, newCommand)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(commandList)
}

func (s *SettingHandler) EditCommand(c *fiber.Ctx) error {
	editCommand := new(model.CommonCommand)

	if err := c.BodyParser(&editCommand); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := s.validate.Struct(editCommand); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	alias, err := url.QueryUnescape(c.Params("alias"))
	if alias == "" || err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Alias is empty")
	}

	userId := parseUserIdFromRequest(c)
	commands, err := s.settingsService.EditCommandForUser(c.Context(), userId, alias, editCommand)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(commands)
}

func (s *SettingHandler) DeleteCommand(c *fiber.Ctx) error {
	alias, err := url.QueryUnescape(c.Params("alias"))
	if alias == "" || err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Alias is empty")
	}

	userId := parseUserIdFromRequest(c)
	commands, err := s.settingsService.DeleteCommandForUser(c.Context(), userId, alias)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(commands)
}

func (s *SettingHandler) SaveVolume(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	volume, err := c.ParamsInt("volume")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid volume")
	}

	err = s.userSettingsRepo.SaveVolume(c.Context(), userId, volume)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return nil
}

func (s *SettingHandler) GetVolume(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	volume, err := s.userSettingsRepo.GetVolume(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(volume)
}

func (s *SettingHandler) ChangeSongRequestsSettings(c *fiber.Ctx) error {

	var newSettings model.SongRequestsSettings
	err := c.BodyParser(&newSettings)
	if err != nil {
		log.Error("Error while parsing request body:", err)
		return err
	}

	if err = s.validate.Struct(newSettings); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	userId := parseUserIdFromRequest(c)

	err = s.settingsService.SaveSongRequestSettings(c.Context(), userId, newSettings)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return nil
}

func (s *SettingHandler) GetSongRequestsSettings(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)

	if userSettings, ok := s.settingsService.UsersSettings[userId]; ok {
		return c.Status(fiber.StatusOK).JSON(userSettings.SongRequests)
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(errors.New("user not found"))
	}
}
