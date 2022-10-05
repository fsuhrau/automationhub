package migrations

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"time"
)

var InitialMigration *gormigrate.Migration

func init() {
	InitialMigration = &gormigrate.Migration{
		ID: "InitialMigration",
		Migrate: func(tx *gorm.DB) error {

			// create company model
			type Company struct {
				gorm.Model
				Token string
				Name  string
			}
			if err := tx.AutoMigrate(&Company{}); err != nil {
				return err
			}

			// create UserAuthModel
			type UserAuth struct {
				gorm.Model
				UserID       uint
				Provider     string
				Token        string
				Username     string
				Email        string
				PasswordHash string
			}
			if err := tx.AutoMigrate(&UserAuth{}); err != nil {
				return err
			}

			// create User
			type User struct {
				gorm.Model
				Name      string
				Role      string
				CompanyID uint
			}
			if err := tx.AutoMigrate(&User{}); err != nil {
				return err
			}

			type AccessToken struct {
				gorm.Model
				ProjectID uint
				Name      string
				Token     string
				ExpiresAt *time.Time
			}
			if err := tx.AutoMigrate(&AccessToken{}); err != nil {
				return err
			}

			type AppFunction struct {
				gorm.Model
				AppID    string
				Assembly string
				Class    string
				Method   string
			}
			if err := tx.AutoMigrate(&AppFunction{}); err != nil {
				return err
			}

			type App struct {
				gorm.Model
				AppID          string
				Name           string
				Platform       string
				Version        string
				AppPath        string
				Identifier     string
				LaunchActivity string
				Additional     string
				Hash           string
				Size           int
				AppFunctions   []AppFunction
				Tags           string
			}
			if err := tx.AutoMigrate(&App{}); err != nil {
				return err
			}

			type Device struct {
				gorm.Model
				CompanyID        uint
				DeviceIdentifier string
				DeviceType       uint
				Name             string
				Manager          string
				HardwareModel    string
				RAM              float32
				SOC              string
				DisplaySize      string
				DPI              float32
				OS               string
				OSVersion        string
				GPU              string
				ABI              string
				OpenGLESVersion  float32
				IsAcknowledged   bool
				Status           device.State `gorm:"-"`
				Dev              interface{}  `gorm:"-"`
				Connection       interface{}  `gorm:"-"`
			}
			if err := tx.AutoMigrate(&Device{}); err != nil {
				return err
			}

			type DeviceParameter struct {
				gorm.Model
				DeviceID uint
				Key      string
				Value    string
			}
			if err := tx.AutoMigrate(&DeviceParameter{}); err != nil {
				return err
			}

			type ConnectionParameter struct {
				gorm.Model
				DeviceID       uint
				ConnectionType uint
				IP             string
				Port           int
			}
			if err := tx.AutoMigrate(&ConnectionParameter{}); err != nil {
				return err
			}

			type DeviceLog struct {
				gorm.Model
				DeviceID uint
				Time     time.Time
				Status   device.State
				Payload  string
			}
			if err := tx.AutoMigrate(&DeviceLog{}); err != nil {
				return err
			}

			type ProtocolEntry struct {
				gorm.Model
				TestProtocolID uint
				Timestamp      time.Time
				Source         string
				Level          string
				Message        string
				Data           string
				Runtime        float64
			}
			if err := tx.AutoMigrate(&ProtocolEntry{}); err != nil {
				return err
			}

			type ProtocolPerformanceEntry struct {
				gorm.Model
				TestProtocolID uint
				Checkpoint     string
				FPS            float32 `sql:"type:decimal(10,2);"`
				MEM            float32 `sql:"type:decimal(10,2);"`
				CPU            float32 `sql:"type:decimal(10,2);"`
				Other          string
				Runtime        float64
			}
			if err := tx.AutoMigrate(&ProtocolPerformanceEntry{}); err != nil {
				return err
			}

			type Test struct {
				gorm.Model
				CompanyID uint
				Name      string
			}
			if err := tx.AutoMigrate(&Test{}); err != nil {
				return err
			}

			type TestConfig struct {
				gorm.Model
				TestID        uint
				ExecutionType uint
				Type          uint
				Platform      uint
				AllDevices    bool
			}
			if err := tx.AutoMigrate(&TestConfig{}); err != nil {
				return err
			}

			type TestConfigDevice struct {
				gorm.Model
				TestConfigID uint
				DeviceID     uint
				Device       Device
			}
			if err := tx.AutoMigrate(&TestConfigDevice{}); err != nil {
				return err
			}

			type TestConfigUnity struct {
				gorm.Model
				TestConfigID       uint
				RunAllTests        bool
				UnityTestFunctions []TestConfigDevice
				Categories         string
				PlayMode           bool
			}
			if err := tx.AutoMigrate(&TestConfigUnity{}); err != nil {
				return err
			}

			type TestProtocol struct {
				gorm.Model
				TestRunID  uint
				DeviceID   *uint
				Device     *Device
				TestName   string
				StartedAt  time.Time
				EndedAt    *time.Time
				TestResult uint
			}
			if err := tx.AutoMigrate(&TestProtocol{}); err != nil {
				return err
			}

			type TestRun struct {
				gorm.Model
				TestID    uint
				AppID     uint
				SessionID string
				Parameter string
			}
			if err := tx.AutoMigrate(&TestRun{}); err != nil {
				return err
			}

			type TestRunLogEntry struct {
				gorm.Model
				TestRunID uint
				Level     string
				Log       string
			}
			if err := tx.AutoMigrate(&TestRunLogEntry{}); err != nil {
				return err
			}

			type UnityTestFunction struct {
				gorm.Model
				TestConfigUnityID uint
				Assembly          string
				Class             string
				Method            string
			}
			if err := tx.AutoMigrate(&UnityTestFunction{}); err != nil {
				return err
			}

			return nil
		},
	}
}
