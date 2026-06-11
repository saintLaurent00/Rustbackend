package auth

import (
	"context"
)

// AuthStrategy définit le contrat d'interface pour tout nouveau module de connexion.
type AuthStrategy interface {
	Authenticate(ctx context.Context, username, password string) (bool, error)
}
