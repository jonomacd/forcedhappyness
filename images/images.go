package images

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
)

type BucketName string

const (
	BucketAvatar BucketName = "iwbn-avatar"
	BucketImages BucketName = "iwbn-images"
)

var (
	projectID = os.Getenv("DATASTORE_PROJECT_ID")

	chars = "abcdefghijklmnopqrstuvwxyz-0123456789"
)

func Init() error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Unable to get bucket %v", err)
		return err
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Unable to close gcs client: %v", err)
		}
	}()

	if err := client.Bucket(string(BucketAvatar)).Create(ctx, projectID, &storage.BucketAttrs{
		Location: "EU",
	}); err != nil {
		gerr, ok := err.(*googleapi.Error)
		if !ok || gerr.Code != 409 {
			log.Fatalf("Failed to create bucket %s: %v", BucketAvatar, err)
			return err
		}
	}

	if err := client.Bucket(string(BucketImages)).Create(ctx, projectID, &storage.BucketAttrs{
		Location: "EU",
	}); err != nil {
		gerr, ok := err.(*googleapi.Error)
		if !ok || gerr.Code != 409 {
			log.Fatalf("Failed to create bucket %s: %v", BucketAvatar, err)
			return err
		}
	}

	return nil
}

func UploadImage(ctx context.Context, bucketName BucketName, contentType, fileExtention string, r io.Reader) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}

	name := newFilename("", 30) + "." + fileExtention
	bucket := client.Bucket(string(bucketName))
	w := bucket.Object(name).NewWriter(ctx)
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.ContentType = contentType

	// Entries are immutable, be aggressive about caching (1 day).
	w.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(w, r); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, name), nil
}

func newFilename(prefix string, length int) string {
	token := make([]byte, length)
	for ii := 0; ii < length; ii++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			// This should never happen
			panic(err)
		}
		token[ii] = chars[n.Uint64()]
	}

	return prefix + string(token)
}
