package provider

import (
	"bi-gateway/ent"
	"bi-gateway/ent/user"
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// LocalProvider gère l'authentification native par email/mot de passe.
// Il utilise la base de données PostgreSQL gérée par Ent pour vérifier les identifiants.
type LocalProvider struct {
	client *ent.Client
}

// NewLocalProvider crée une nouvelle instance du fournisseur local avec le client DAO.
func NewLocalProvider(client *ent.Client) *LocalProvider {
	return &LocalProvider{client: client}
}

// Authenticate vérifie les informations de connexion d'un utilisateur.
// 1. Recherche l'utilisateur par son email.
// 2. Vérifie si le compte est actif.
// 3. Compare le hash du mot de passe stocké avec le mot de passe fourni.
func (p *LocalProvider) Authenticate(ctx context.Context, email, password string) (bool, error) {
	// Récupération de l'utilisateur via le DAO Ent
	u, err := p.client.User.
		Query().
		Where(user.Email(email)).
		Only(ctx)
	if err != nil {
		return false, fmt.Errorf("utilisateur non trouvé")
	}

	// Vérification de l'état du compte
	if !u.IsActive {
		return false, fmt.Errorf("compte inactif")
	}

	// Comparaison sécurisée du mot de passe avec Bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return false, fmt.Errorf("mot de passe incorrect")
	}

	return true, nil
}
