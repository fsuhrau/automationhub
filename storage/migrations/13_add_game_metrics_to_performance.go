package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddGameMetricsToPerformance *gormigrate.Migration

func init() {
	AddGameMetricsToPerformance = &gormigrate.Migration{
		ID: "AddGameMetricsToPerformance",
		Migrate: func(tx *gorm.DB) error {
			type ProtocolPerformanceEntry struct {
				gorm.Model
				VertexCount float64 `sql:"type:decimal(10,2);"`
				Triangles   float64 `sql:"type:decimal(10,2);"`
			}

			if err := tx.Migrator().AddColumn(&ProtocolPerformanceEntry{}, "VertexCount"); err != nil {
				return err
			}
			if err := tx.Migrator().AddColumn(&ProtocolPerformanceEntry{}, "Triangles"); err != nil {
				return err
			}

			type TestProtocol struct {
				gorm.Model
				AvgVertexCount float64 `sql:"type:decimal(10,2);" json:",omitempty"`
				AvgTriangles   float64 `sql:"type:decimal(10,2);" json:",omitempty"`
			}

			if err := tx.Migrator().AddColumn(&TestProtocol{}, "AvgVertexCount"); err != nil {
				return err
			}
			if err := tx.Migrator().AddColumn(&TestProtocol{}, "AvgTriangles"); err != nil {
				return err
			}

			return nil
		},
	}
}
