package devicestore

import (
	"context"
	"fmt"

	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// Apply writes a change of the alias and the custom parameters of the device
// in one transaction.
func (s *Store) Apply(ctx context.Context, deviceID uint, change svc.Change) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Device{}).Where("id = ?", deviceID).Update("alias", change.Alias).Error; err != nil {
			return err
		}
		if err := updateParams(tx, deviceID, change.Update); err != nil {
			return err
		}
		if err := addParams(tx, deviceID, change.Add); err != nil {
			return err
		}
		return removeParams(tx, deviceID, change.Remove)
	})
	return mapErr(err, fmt.Sprintf("device %d", deviceID))
}

func updateParams(tx *gorm.DB, deviceID uint, params []svc.Param) error {
	for _, p := range params {
		err := tx.Model(&models.CustomParameter{}).Where("id = ? AND device_id = ?", p.ID, deviceID).
			Updates(map[string]any{"key": p.Key, "value": p.Value}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func addParams(tx *gorm.DB, deviceID uint, params []svc.Param) error {
	for _, p := range params {
		if err := tx.Create(&models.CustomParameter{DeviceID: deviceID, Key: p.Key, Value: p.Value}).Error; err != nil {
			return err
		}
	}
	return nil
}

func removeParams(tx *gorm.DB, deviceID uint, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return tx.Delete(&models.CustomParameter{}, "id IN ? AND device_id = ?", ids, deviceID).Error
}
