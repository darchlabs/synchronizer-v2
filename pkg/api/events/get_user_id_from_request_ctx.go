package events

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

func getUserIDFromRequestCtx(c *fiber.Ctx) (string, error) {
	id := c.Locals("user_id")
	userID, ok := id.(string)
	if !ok {
		return "", errors.New("unrecognized id type")
	}

	return userID, nil
}
