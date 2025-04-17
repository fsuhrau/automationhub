package migrations

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"time"
)

var (
	deviceRegex = regexp.MustCompile(`Run test '.*' on device '(.*)'`)
)

var AddPerformanceAverage *gormigrate.Migration

func init() {
	AddPerformanceAverage = &gormigrate.Migration{
		ID: "AddPerformanceAverage",
		Migrate: func(tx *gorm.DB) error {

			type TestProtocol struct {
				gorm.Model
				AvgFPS float64 `sql:"type:decimal(10,2);"`
				AvgMEM float64 `sql:"type:decimal(10,2);"`
				AvgCPU float64 `sql:"type:decimal(10,2);"`
			}

			if err := tx.Migrator().AddColumn(&TestProtocol{}, "AvgFPS"); err != nil {
				return err
			}
			if err := tx.Migrator().AddColumn(&TestProtocol{}, "AvgMEM"); err != nil {
				return err
			}
			if err := tx.Migrator().AddColumn(&TestProtocol{}, "AvgCPU"); err != nil {
				return err
			}
			var protocols []models.TestProtocol
			if err := tx.Preload("Performance").Find(&protocols).Error; err != nil {
				return err
			}

			for _, protocol := range protocols {
				var sumFPS float64
				var sumMEM float64
				var sumCPU float64
				for _, pe := range protocol.Performance {
					sumFPS += pe.FPS
					sumMEM += pe.MEM
					sumCPU += pe.CPU
				}
				numEntries := len(protocol.Performance)
				protocol.AvgFPS = sumFPS / float64(numEntries)
				protocol.AvgMEM = sumMEM / float64(numEntries)
				protocol.AvgCPU = sumCPU / float64(numEntries)
				if err := tx.Save(&protocol).Error; err != nil {
					return err
				}
			}

			type TestRunDeviceStatus struct {
				gorm.Model
				TestRunID   uint
				DeviceID    uint
				StartupTime uint
			}
			if err := tx.AutoMigrate(&TestRunDeviceStatus{}); err != nil {
				return err
			}

			var devices []models.Device
			var deviceMap map[string]models.Device
			deviceMap = make(map[string]models.Device)
			if err := tx.Find(&devices).Error; err != nil {
				return err
			}
			for _, d := range devices {
				deviceMap[d.DeviceIdentifier] = d
			}

			var testRuns []models.TestRun
			if err := tx.Preload("Log").Find(&testRuns).Error; err != nil {
				return err
			}

			for _, run := range testRuns {
				var startTime time.Time
				hasStartup := false
				var processedDevices map[string]bool
				processedDevices = make(map[string]bool)
				for _, log := range run.Log {
					if log.Log == "start app on devices and wait for connection" {
						startTime = log.CreatedAt
						hasStartup = true
						continue
					}

					if strings.HasPrefix(log.Log, "Run test") && hasStartup {
						ms := log.CreatedAt.Sub(startTime).Milliseconds()

						matches := deviceRegex.FindStringSubmatch(log.Log)
						if len(matches) == 2 {
							deviceIdentifier := matches[1]

							if _, ok := processedDevices[deviceIdentifier]; ok {
								continue
							}

							processedDevices[deviceIdentifier] = true

							if dev, ok := deviceMap[deviceIdentifier]; ok {
								deviceStatus := models.TestRunDeviceStatus{DeviceID: dev.ID, TestRunID: run.ID, StartupTime: uint(ms)}
								if err := tx.Create(&deviceStatus).Error; err != nil {
									return err
								}
							}
						}
					}
				}
			}

			return nil
		},
	}
}
