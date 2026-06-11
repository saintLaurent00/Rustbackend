package provider

import (
	"bi-gateway/ent"
	"bi-gateway/ent/user"
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// LocalProvider gère l'authentification native.
type LocalProvider struct {
	client *ent.Client
}

func NewLocalProvider(client *ent.Client) *LocalProvider {
	return &LocalProvider{client: client}
}

func (p *LocalProvider) Authenticate(ctx context.Context, email, password string) (bool, error) {
	u, err := p.client.User.
		Query().
		Where(user.Email(email)).
		Only(ctx)
	if err != nil {
		return false, fmt.Errorf("utilisateur non trouvé")
	}

	if !u.IsActive {
		return false, fmt.Errorf("compte inactif")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return false, fmt.Errorf("mot de passe incorrect")
	}

	return true, nil
}
