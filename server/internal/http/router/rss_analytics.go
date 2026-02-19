package router

import (
	"context"
	"log"

	apprss "github.com/grtsinry43/grtblog-v2/server/internal/app/rss"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/geoip"
)

func newRSSAccessAnalyticsService(deps Dependencies) *apprss.AccessAnalyticsService {
	var geoResolver *geoip.Resolver
	if deps.Config.GeoIP.DBPath != "" {
		geoip.EnsureDatabasesAsync(
			context.Background(),
			deps.Config.GeoIP.DBPath,
			deps.Config.GeoIP.DownloadURL,
			deps.Config.GeoIP.ASNPath,
			deps.Config.GeoIP.ASNURL,
			log.Printf,
		)
		resolver, err := geoip.NewResolver(deps.Config.GeoIP.DBPath, deps.Config.GeoIP.ASNPath)
		if err != nil {
			log.Printf("geoip resolver init failed: %v", err)
			geoResolver = geoip.NewLazyResolver(deps.Config.GeoIP.DBPath, deps.Config.GeoIP.ASNPath)
		} else {
			geoResolver = resolver
		}
	}
	return apprss.NewAccessAnalyticsService(deps.DB, geoResolver)
}

