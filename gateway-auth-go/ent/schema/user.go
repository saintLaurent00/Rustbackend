package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").Unique(),
		field.String("password_hash").Sensitive(),
		field.String("first_name").Optional(),
		field.String("last_name").Optional(),
		field.Bool("is_active").Default(true),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("groups", Group.Type),
	}
}
