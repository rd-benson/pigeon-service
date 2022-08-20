package pigeon

import (
	"context"
	"fmt"
	"os"

	uuid "github.com/google/uuid"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// NewInfluxDB returns read and write clients and checks validity of tokens
func NewInfluxDB() (writeClient *influxdb2.Client) {
	write := influxdb2.NewClient(cfg.InfluxDB.URI(), cfg.InfluxDB.TokenWrite)

	// Check API token validity,
	writeBucket := write.BucketsAPI()
	ctx := context.Background()
	bucketName := fmt.Sprintf("pigeon-auth-test-%v", uuid.NewString())
	// 1. create a new bucket with writeBucket
	bucket, err := writeBucket.CreateBucketWithNameWithID(ctx, cfg.InfluxDB.OrgId, bucketName)
	if err != nil {
		fmt.Printf("influxdb %v\n", err.Error())
		os.Exit(2)
	}
	// 2. delete with writeBucket
	writeBucket.DeleteBucketWithID(ctx, *bucket.Id)

	return &write
}

// CreateBucketSafe creates bucket only if it does not already exist
func CreateBucketSafe(name string, writeClient *influxdb2.Client) bool {
	_, err := (*writeClient).BucketsAPI().CreateBucketWithNameWithID(
		context.Background(), cfg.InfluxDB.OrgId, name)
	return err == nil
}
