package internal

import "github.com/deevus/truenas-go"

// Services holds initialized truenas-go service interfaces for one server.
type Services struct {
	Datasets   truenas.DatasetServiceAPI
	Snapshots  truenas.SnapshotServiceAPI
	System     truenas.SystemServiceAPI
	Reporting  truenas.ReportingServiceAPI
	Interfaces truenas.InterfaceServiceAPI
	Apps       truenas.AppServiceAPI
}

// NewServices creates a Services container from the given service interfaces.
func NewServices(
	ds truenas.DatasetServiceAPI,
	ss truenas.SnapshotServiceAPI,
	sys truenas.SystemServiceAPI,
	rep truenas.ReportingServiceAPI,
	ifaces truenas.InterfaceServiceAPI,
	apps truenas.AppServiceAPI,
) *Services {
	return &Services{
		Datasets:   ds,
		Snapshots:  ss,
		System:     sys,
		Reporting:  rep,
		Interfaces: ifaces,
		Apps:       apps,
	}
}
