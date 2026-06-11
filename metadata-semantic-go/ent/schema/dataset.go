package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Dataset struct {
	ent.Schema
}

func (Dataset) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Enum("source_type").Values("DB", "API", "FILE"),
		field.String("connection_details").Sensitive(), // JSON chiffré
	}
}

func (Dataset) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("charts", Chart.Type),
		edge.To("rls_rules", RlsRule.Type),
	}
}
