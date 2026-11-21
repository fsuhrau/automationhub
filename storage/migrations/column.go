package migrations

import "gorm.io/gorm"

// addColumnIfMissing adds field to the table of model when the table exists
// and has no such column yet, so reruns and fresh schemas are harmless.
func addColumnIfMissing(g *gorm.DB, model any, field string) error {
	if !g.Migrator().HasTable(model) || g.Migrator().HasColumn(model, field) {
		return nil
	}
	return g.Migrator().AddColumn(model, field)
}
