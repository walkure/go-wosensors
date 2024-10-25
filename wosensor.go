package wosensors

import (
	"log/slog"

	"github.com/walkure/gatt"
	"github.com/walkure/gatt/logger"
)

// The Member Service UUID of Woan Technology (Shenzhen) Co., Ltd.
var memberUUID = gatt.MustParseUUID("fd3d")

// The Company Identifier of Woan Technology (Shenzhen) Co., Ltd.
const companyID = 0x0969

// SetLogger sets the logger for the package
func SetLogger(newLogger *slog.Logger) {
	logger.SetLogger(newLogger)
}
