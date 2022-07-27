package function

import (
	"context"
	"io"
	"log"

	"cloud.google.com/go/storage"
)

func loadBucket() {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Error opening storage.NewClient: %s.", err)
		return
	}
	rc, err := client.Bucket("gomercury-bucket356415").Object("GeoLite2-City.mmdb").NewReader(ctx)
	if err != nil {
		log.Fatalf("Error opening storage bucket: %s.", err)
		return
	}
	defer rc.Close()
	geoIPData, err = io.ReadAll(rc)
	if err != nil {
		log.Fatalf("Error with reading geoIPData: %s.", err)
		return
	}
}

func initDatabase() (geoIPData []byte) {
	loadGeoIpOnce.Do(loadBucket)
	return geoIPData
}
