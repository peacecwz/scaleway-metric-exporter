package scaleway

import (
	"context"
	"github.com/peacecwz/scaleway-metric-exporter/config"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	log "github.com/sirupsen/logrus"
)

type ScalewayClient struct {
	client *scw.Client

	rdb     *rdb.API
	Regions []scw.Region
}

func EnsureCreateClient(cfg *config.Config) *ScalewayClient {
	if cfg.ScalewayAccessKey == "" || cfg.ScalewaySecretKey == "" {
		log.Fatal("Scaleway access key and secret key must be set")
	}

	var regions []scw.Region
	if cfg.ScalewayRegion == "" {
		regions = scw.AllRegions
		log.Warnf("scaleway region is not set, default setting to ALL regions")
	} else {
		regions = []scw.Region{scw.Region(cfg.ScalewayRegion)}
	}

	var zones []scw.Zone
	if cfg.ScalewayZone == "" {
		log.Warnf("scaleway zone is not set, default setting to ALL zones")
		zones = scw.AllZones
	} else {
		zones = []scw.Zone{scw.Zone(cfg.ScalewayZone)}
	}

	client, err := scw.NewClient(
		scw.WithAuth(cfg.ScalewayAccessKey, cfg.ScalewaySecretKey),
		scw.WithDefaultRegion(regions[0]),
		scw.WithDefaultZone(zones[0]),
	)
	if err != nil {
		log.Fatalf("failed to create scaleway client %v", err)
	}

	return &ScalewayClient{
		client:  client,
		rdb:     rdb.NewAPI(client),
		Regions: regions,
	}
}

func (c ScalewayClient) ListRDBInstance(ctx context.Context, req *rdb.ListInstancesRequest, pages scw.RequestOption) (*rdb.ListInstancesResponse, error) {
	return c.rdb.ListInstances(req, pages, scw.WithContext(ctx))
}

func (c ScalewayClient) GetInstanceMetrics(ctx context.Context, req *rdb.GetInstanceMetricsRequest) (*rdb.InstanceMetrics, error) {
	return c.rdb.GetInstanceMetrics(req, scw.WithContext(ctx))
}
