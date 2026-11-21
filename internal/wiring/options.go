package wiring

import (
	"fmt"
	"os"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Option replaces a default of the build, mostly for tests.
type Option func(*options)

type options struct {
	db       *gorm.DB
	logger   *logrus.Logger
	handlers []device.Handler
}

// WithDeviceHandlers adds device handlers to the ones of the config.
func WithDeviceHandlers(handlers ...device.Handler) Option {
	return func(o *options) { o.handlers = append(o.handlers, handlers...) }
}

// WithDB uses an open database instead of the one of the config.
func WithDB(db *gorm.DB) Option { return func(o *options) { o.db = db } }

// WithLogger uses the logger instead of a new one.
func WithLogger(logger *logrus.Logger) Option { return func(o *options) { o.logger = logger } }

// resolve applies the options and creates the folders of the config.
func resolve(cfg config.Service, opts []Option) (options, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}
	if o.logger == nil {
		o.logger = logrus.New()
	}
	return o, ensureFolders(cfg.Paths)
}

// openDB opens the database of the config unless the options brought one.
func (o *options) openDB(cfg config.Database) error {
	if o.db != nil {
		return nil
	}
	db, err := storage.Open(cfg)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	o.db = db
	return nil
}

// ensureFolders creates the folders of the paths.
func ensureFolders(paths config.Paths) error {
	for _, dir := range paths.All() {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("create folder %q: %w", dir, err)
		}
	}
	return nil
}
