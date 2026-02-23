package internal

import "github.com/deevus/truenas-go"

// Services holds initialized truenas-go service interfaces for one server.
type Services struct {
	Datasets  truenas.DatasetServiceAPI
	Snapshots truenas.SnapshotServiceAPI
}

// NewServices creates a Services container from the given service interfaces.
func NewServices(ds truenas.DatasetServiceAPI, ss truenas.SnapshotServiceAPI) *Services {
	return &Services{
		Datasets:  ds,
		Snapshots: ss,
	}
}
