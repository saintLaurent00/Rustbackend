package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Dashboard struct {
	ent.Schema
}

func (Dashboard) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.JSON("layout", map[string]interface{}{}),
	}
}

func (Dashboard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("charts", Chart.Type).Ref("dashboards"),
	}
}
