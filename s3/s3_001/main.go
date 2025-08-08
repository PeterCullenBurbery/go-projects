package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	ctx := context.Background()

	// Force the named CLI profile you configured
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile("peter-s3-uploader"),
		config.WithRegion("us-east-2"), // default base region
	)
	if err != nil { log.Fatal(err) }

	bucket := "peter-cullen-burbery-my-first-aws-bucket"

	// Discover the bucket's actual region (api returns "" for us-east-1)
	tmp := s3.NewFromConfig(cfg)
	loc, err := tmp.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: &bucket})
	if err != nil { log.Fatal(err) }
	bucketRegion := string(loc.LocationConstraint)
	if bucketRegion == "" { bucketRegion = "us-east-1" }

	// Pin client to the bucket’s region
	client := s3.NewFromConfig(cfg, func(o *s3.Options) { o.Region = bucketRegion })
	log.Printf("using bucket region: %s", bucketRegion)

	// Create a random temp file and write random content (kept locally)
	dir := os.TempDir()
	name := randomHex(8) + ".txt"
	path := filepath.Join(dir, "upload_"+name)
	content := []byte("random upload id: " + randomHex(16) + "\n")
	if err := os.WriteFile(path, content, 0o600); err != nil { log.Fatal(err) }

	// Upload it
	f, err := os.Open(path)
	if err != nil { log.Fatal(err) }
	defer f.Close()

	key := "uploads/" + filepath.Base(path)
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	})
	if err != nil { log.Fatal(err) }

	log.Printf("uploaded s3://%s/%s", bucket, key)
	log.Printf("local file kept at: %s", path)
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil { panic(err) }
	return hex.EncodeToString(b)
}