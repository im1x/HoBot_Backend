package handler

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/service/settings"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"net/url"
)

func GetCommands(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	commands, err := settings.GetCommands(c.Context(), userId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(commands)
}
func GetCommandsDropdown(c *fiber.Ctx) error {
	commandsList, err := settings.GetCommandsList()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(commandsList.Commands)
}

func AddCommandAndAlias(c *fiber.Ctx) error {
	newCommand := new(model.CommonCommand)

	if err := c.BodyParser(&newCommand); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := validate.Struct(newCommand); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	userId := parseUserIdFromRequest(c)

	commandList, err := settings.AddCommandForUser(c.Context(), userId, newCommand)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(commandList)
}

func EditCommand(c *fiber.Ctx) error {
	editCommand := new(model.CommonCommand)

	if err := c.BodyParser(&editCommand); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := validate.Struct(editCommand); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	alias, err := url.QueryUnescape(c.Params("alias"))
	if alias == "" || err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Alias is empty")
	}

	userId := parseUserIdFromRequest(c)
	commands, err := settings.EditCommandForUser(c.Context(), userId, alias, editCommand)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(commands)
}

func DeleteCommand(c *fiber.Ctx) error {
	alias, err := url.QueryUnescape(c.Params("alias"))
	if alias == "" || err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Alias is empty")
	}

	userId := parseUserIdFromRequest(c)
	commands, err := settings.DeleteCommandForUser(c.Context(), userId, alias)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(commands)
}

func SaveVolume(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	volume, err := c.ParamsInt("volume")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid volume")
	}

	err = settings.SaveVolume(c.Context(), userId, volume)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return nil
}

func GetVolume(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	volume, err := settings.GetVolume(c.Context(), userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(volume)
}

func ChangeSongRequestsSettings(c *fiber.Ctx) error {

	var newSettings settings.SongRequestsSettings
	err := c.BodyParser(&newSettings)
	if err != nil {
		log.Error("Error while parsing request body:", err)
		return err
	}

	if err = validate.Struct(newSettings); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	userId := parseUserIdFromRequest(c)

	err = settings.SaveSongRequestSettings(c.Context(), userId, newSettings)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return nil
}

func GetSongRequestsSettings(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)

	if userSettings, ok := settings.UsersSettings[userId]; ok {
		return c.Status(fiber.StatusOK).JSON(userSettings.SongRequests)
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(errors.New("user not found"))
	}
}
