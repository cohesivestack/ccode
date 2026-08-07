package ccode

import (
	"strings"

	"ariga.io/atlas/sql/schema"
)

func databaseAttribute[T schema.Attr](attrs []schema.Attr) (T, bool) {
	for _, attr := range attrs {
		if value, ok := attr.(T); ok {
			return value, true
		}
	}
	var zero T
	return zero, false
}

func databaseComment(attrs []schema.Attr) string {
	if comment, ok := databaseAttribute[*schema.Comment](attrs); ok {
		return comment.Text
	}
	return ""
}

func databaseCharset(attrs []schema.Attr) string {
	if charset, ok := databaseAttribute[*schema.Charset](attrs); ok {
		return charset.V
	}
	return ""
}

func databaseCollation(attrs []schema.Attr) string {
	if collation, ok := databaseAttribute[*schema.Collation](attrs); ok {
		return collation.V
	}
	return ""
}

func databaseGeneratedExpression(attrs []schema.Attr) string {
	if generated, ok := databaseAttribute[*schema.GeneratedExpr](attrs); ok {
		return generated.Expr
	}
	return ""
}

func databaseExpression(expression schema.Expr) string {
	if expression == nil {
		return ""
	}
	switch expression := schema.UnderlyingExpr(expression).(type) {
	case *schema.Literal:
		return expression.V
	case *schema.RawExpr:
		return expression.X
	default:
		return ""
	}
}

func databaseIndexPartExpression(part *schema.IndexPart) string {
	if part == nil || part.X == nil {
		return ""
	}
	return databaseExpression(part.X)
}

func databaseColumnNames(columns []*schema.Column) []string {
	names := make([]string, 0, len(columns))
	for _, column := range columns {
		if column != nil {
			names = append(names, column.Name)
		}
	}
	return names
}

func databaseIndexColumnNames(index *schema.Index) []string {
	if index == nil {
		return []string{}
	}
	names := make([]string, 0, len(index.Parts))
	for _, part := range index.Parts {
		if part != nil && part.C != nil {
			names = append(names, part.C.Name)
		}
	}
	return names
}

func databaseReferenceAction(action schema.ReferenceOption) string {
	if action == "" {
		return "no-action"
	}
	return strings.ReplaceAll(strings.ToLower(string(action)), " ", "-")
}

func databaseForeignKeyOptional(columns []*schema.Column) bool {
	for _, column := range columns {
		if column != nil && column.Type != nil && column.Type.Null {
			return true
		}
	}
	return false
}

func databaseForeignKeyCardinality(table *schema.Table, columns []*schema.Column, partialIndex func(*schema.Index) bool) string {
	if table == nil {
		return "many-to-one"
	}
	if databaseColumnsMatchIndex(columns, table.PrimaryKey) {
		return "one-to-one"
	}
	for _, index := range table.Indexes {
		if index != nil && index.Unique && !partialIndex(index) && databaseColumnsMatchIndex(columns, index) {
			return "one-to-one"
		}
	}
	return "many-to-one"
}

func databaseColumnsMatchIndex(columns []*schema.Column, index *schema.Index) bool {
	if index == nil || len(columns) != len(index.Parts) {
		return false
	}
	for position, part := range index.Parts {
		if columns[position] == nil || part == nil || part.C == nil || columns[position].Name != part.C.Name {
			return false
		}
	}
	return true
}

func databaseTableInRealm(realm *schema.Realm, table *schema.Table) bool {
	if realm == nil || table == nil || table.Schema == nil {
		return false
	}
	s, ok := realm.Schema(table.Schema.Name)
	if !ok {
		return false
	}
	_, ok = s.Table(table.Name)
	return ok
}
