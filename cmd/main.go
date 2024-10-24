package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/walkure/gatt"
	"github.com/walkure/go-wosensors"
)

func main() {
	/*
		d, err := gatt.NewDevice(gatt.LnxSetScanParameters(
			&cmd.LESetScanParameters{
				LEScanType:           0x01,  // 0x00: Passive, 0x01: Active
				LEScanInterval:       0xf00, // 3840 * 0.625m = 2400ms = 2s   ; 0x0004 - 0x4000
				LEScanWindow:         0x800, // 2048 * 0.625m = 1280ms = 1.3s ; 0x0004 - 0x4000
				OwnAddressType:       0x00,  // 0x00: Public, 0x01: Random
				ScanningFilterPolicy: 0x00,  // 0x00: accept all, 0x01: ignore non-white-listed.
			}))
	*/

	d, err := gatt.NewDevice()

	if err != nil {
		panic(err)
	}

	var mu sync.Mutex
	seqno := uint8(0)

	d.Handle(gatt.PeripheralDiscovered(
		wosensors.HandleWoSensorTHO("", // "aa:zz:pp:ff:dd:cc"
			func(d wosensors.THOData) {

				// GATT lib. calls this callback function with a new goroutine.
				mu.Lock()
				defer mu.Unlock()
				if seqno != d.SequenceNumber {
					slog.Info("WoSensorTHO", "Data", d.LogValue())
					seqno = d.SequenceNumber
				}
			}, nil)))
	d.Init(onStateChanged)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	<-ctx.Done()
	fmt.Println("interrupted. Bye~")

	d.StopScanning()
	fmt.Printf("scan stopped: %+v\n", d.Stop())
}

func onStateChanged(d gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		// allow duplicate
		d.Scan([]gatt.UUID{}, true)
		return
	default:
		d.StopScanning()
	}
}
