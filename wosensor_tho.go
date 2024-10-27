package wosensors

import (
	"fmt"
	"strings"

	"log/slog"

	"github.com/walkure/gatt"
)

// device type of WoSensorTHO(W3400010)
const devTypeWoSensorTHO = 0x77 // 'w'

// The THOData struct contains the data of the WoSensorTHO device.
type THOData struct {
	// DeviceId the id of the device
	DeviceId string
	// Temperature in Celsius. 0.1 degree resolution.
	Temperature float32
	// Humidity in percentage (0-100%)
	Humidity uint8
	// Sequence Number (1-255)
	SequenceNumber uint8
	// Battery Level in percentage (0-100%)
	BatteryPercent uint8
	// Received Signal Strength Indicator
	RSSI int
}

// LogValue returns the slog.Value of the THOData.
func (d THOData) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("DeviceId", d.DeviceId),
		slog.String("Temperature", fmt.Sprintf("%.1f°C", d.Temperature)),
		slog.Uint64("Humidity", uint64(d.Humidity)),
		slog.Uint64("SequenceNumber", uint64(d.SequenceNumber)),
		slog.Uint64("BatteryPercent", uint64(d.BatteryPercent)),
		slog.Int64("RSSI", int64(d.RSSI)),
	)
}

// HandleWoSensorTHO returns a callback function for gatt.PeripheralDiscovered that can be used to handle the WoSensorTHO device.
//
// The deviceId is the device id of the target WoSensorTHO device. if it is empty, all WoSensorTHO devices will be handled.
// The allowWithoutBatteryPercent is true if the battery level is unknown(value=255).
// The cb function will be called with new goroutine when receives the advertisement packet from a WoSensorTHO device.
// The next function will be called if the device is not a target WoSensorTHO device.
func HandleWoSensorTHO(deviceId string, allowWithoutBatteryPercent bool,
	cb func(d THOData),
	next func(gatt.Peripheral, *gatt.Advertisement, int)) func(gatt.Peripheral, *gatt.Advertisement, int) {

	if next == nil {
		next = func(gatt.Peripheral, *gatt.Advertisement, int) {}
	}

	if cb == nil {
		panic("cb is nil")
	}

	deviceId = strings.ToUpper(deviceId)

	return func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {

		if a.CompanyID != companyID {
			// Another Manufacturer's Device
			next(p, a, rssi)
			return
		}

		devId := strings.ToUpper(p.ID())

		if deviceId != "" && devId != deviceId {
			// Another UUID Device
			next(p, a, rssi)
			return
		}

		datum := THOData{
			RSSI:           rssi,
			BatteryPercent: 255,
		}

		//https://github.com/OpenWonderLabs/SwitchBotAPI-BLE/blob/latest/devicetypes/meter.md#outdoor-temperaturehumidity-sensor

		for _, d := range a.ServiceData {
			if !d.UUID.Equal(memberUUID) {
				// Another Manufacturer's Device
				next(p, a, rssi)
				return
			}
			if len(d.Data) > 2 {
				if d.Data[0] != devTypeWoSensorTHO {
					// Another Device Type
					next(p, a, rssi)
					return
				}
				datum.BatteryPercent = d.Data[2] & 0x7F
			}
		}

		if !allowWithoutBatteryPercent && datum.BatteryPercent > 128 {
			// Passive Scan. Battery Level is unknown.
			next(p, a, rssi)
			return
		}

		if a.ManufacturerData == nil {
			// Manufacturer Data is nil
			next(p, a, rssi)
			return
		}

		if len(a.ManufacturerData) < 13 {
			// Truncated Data. Discard.
			next(p, a, rssi)
			return
		}

		datum.DeviceId = devId

		temperature := int16(a.ManufacturerData[10]&0x0F) + int16(a.ManufacturerData[11]&0x7F)*10
		if a.ManufacturerData[11]&0x80 == 0 {
			temperature = -temperature
		}

		datum.Temperature = float32(temperature) / 10.0
		datum.SequenceNumber = a.ManufacturerData[8]
		datum.Humidity = a.ManufacturerData[12] & 0x7F

		cb(datum)

	}
}
