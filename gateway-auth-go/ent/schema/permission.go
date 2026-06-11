package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Permission holds the schema definition for the Permission entity.
type Permission struct {
	ent.Schema
}

// Fields of the Permission.
func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("access_level").
			Values("CAN_READ", "CAN_EDIT", "CAN_SHARE", "CAN_GRANT").
			Default("CAN_READ"),
        field.Int("object_id").Optional(), // ID de l'objet (Chart, Dashboard, etc.)
        field.String("object_type").Optional(), // Type de l'objet
	}
}

// Edges of the Permission.
func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roles", Role.Type).Ref("permissions"),
	}
}
