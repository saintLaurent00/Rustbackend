package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Chart struct {
	ent.Schema
}

func (Chart) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.JSON("config", map[string]interface{}{}),
		field.JSON("column_styles", map[string]interface{}{}),
	}
}

func (Chart) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("dataset", Dataset.Type).Ref("charts").Unique(),
		edge.To("dashboards", Dashboard.Type),
	}
}
