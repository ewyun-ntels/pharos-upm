package transmission

import (
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/server/udp"
	transmission_udp "ntels.com/pharos/extensions/catv/business/pkg/collect/transmission/udp"
)

func Load(config common.Config) error {
	if !config.Catv.Collect.Transmission.Use {
		return nil
	}

	udp.AddGnetServer("daily", config.Catv.Collect.Transmission.Daily.Port, transmission_udp.NewEventHandler(config, transmission_udp.TransmissionTypeDaily))
	udp.AddGnetServer("diagnostic", config.Catv.Collect.Transmission.Diagnostic.Port, transmission_udp.NewEventHandler(config, transmission_udp.TransmissionTypeDiagnostic))
	udp.AddGnetServer("network_quality_transition", config.Catv.Collect.Transmission.NetworkQualityTransition.Port, transmission_udp.NewEventHandler(config, transmission_udp.TransmissionTypeNetworkQualityTransition))
	udp.AddGnetServer("periodic", config.Catv.Collect.Transmission.Periodic.Port, transmission_udp.NewEventHandler(config, transmission_udp.TransmissionTypePeriodic))
	udp.AddGnetServer("quality_measurement", config.Catv.Collect.Transmission.QualityMeasurement.Port, transmission_udp.NewEventHandler(config, transmission_udp.TransmissionTypeQualityMeasurement))

	return nil
}

func Unload() {
}
