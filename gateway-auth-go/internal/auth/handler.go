package auth

import (
	"bi-gateway/ent"
	"bi-gateway/ent/user"
	"bi-gateway/internal/auth/provider"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	client *ent.Client
	local  *provider.LocalProvider
}

func NewAuthHandler(client *ent.Client) *AuthHandler {
	return &AuthHandler{
		client: client,
		local:  provider.NewLocalProvider(client),
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	success, err := h.local.Authenticate(c.Context(), req.Email, req.Password)
	if err != nil || !success {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	u, err := h.client.User.Query().Where(user.Email(req.Email)).Only(c.Context())
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
    }

	token, err := GenerateJWT(u.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "could not generate token"})
	}

	return c.JSON(fiber.Map{"token": token})
}
