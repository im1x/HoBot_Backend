package handler

import (
	"HoBot_Backend/internal/model"
	repoUser "HoBot_Backend/internal/repository/user"
	"HoBot_Backend/internal/service/token"
	userService "HoBot_Backend/internal/service/user"
	"HoBot_Backend/internal/service/vkplay"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	validate    *validator.Validate
	userRepo    repoUser.Repository
	userService *userService.UserService
}

func NewUserHandler(validate *validator.Validate, userService *userService.UserService, userRepo repoUser.Repository) *UserHandler {
	return &UserHandler{validate: validate, userService: userService, userRepo: userRepo}
}

func (s *UserHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refreshToken")
	if refreshToken == "" {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	err := s.userService.Logout(c.Context(), refreshToken)
	if err != nil {
		return err
	}
	c.ClearCookie()

	return nil
}

func (s *UserHandler) Refresh(c *fiber.Ctx) error {
	refreshTokenCookie := c.Cookies("refreshToken")
	if refreshTokenCookie == "" {
		return fiber.NewError(fiber.StatusUnauthorized)
	}
	accessToken, refreshToken, err := s.userService.RefreshToken(c.Context(), refreshTokenCookie)
	if err != nil {
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "refreshToken"
	cookie.Value = refreshToken
	cookie.Expires = time.Now().Add(token.RefreshTokenExpireHour * time.Hour)
	cookie.HTTPOnly = true
	c.Cookie(cookie)

	return c.JSON(model.AccessToken{AccessToken: accessToken})
}

func (s *UserHandler) VkplAuth(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	// codeToToken
	accessToken, err := vkplay.CodeToToken(code)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	// getUserInfo
	userInfo, err := vkplay.GetCurrentUserInfo(accessToken)
	if err != nil {
		log.Error("Error while getting user info:", err)
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	// is streamer?
	if !userInfo.Data.User.IsStreamer {
		return c.Redirect(os.Getenv("CLIENT_URL") + "?s=1")
	}

	refreshToken, err := s.userService.LoginVkpl(c.Context(), userInfo)
	if err != nil {
		log.Error(err)
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "refreshToken"
	cookie.Value = refreshToken
	cookie.Expires = time.Now().Add(token.RefreshTokenExpireHour * time.Hour)
	cookie.HTTPOnly = true
	c.Cookie(cookie)

	return c.Redirect(os.Getenv("CLIENT_URL"))
}

func (s *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	user, err := s.userRepo.GetUser(c.Context(), userId)
	if err != nil {
		log.Error("Error while getting user info:", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}
	return c.JSON(user)
}

func (s *UserHandler) WipeUser(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	err := s.userService.WipeUser(c.Context(), userId)
	if err != nil {
		log.Error("Error while wiping user:", err)
		return err
	}

	return s.Logout(c)
}

func parseUserIdFromRequest(c *fiber.Ctx) string {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["id"].(string)
}
