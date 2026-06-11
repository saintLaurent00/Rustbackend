package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type RlsRule struct {
	ent.Schema
}

func (RlsRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("target_table"),
		field.String("sql_predicate"),
	}
}

func (RlsRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("dataset", Dataset.Type).Ref("rls_rules").Unique(),
	}
}
