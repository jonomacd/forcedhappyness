package dao

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"cloud.google.com/go/datastore"
)

var (
	ds *datastore.Client

	ErrNotFound = datastore.ErrNoSuchEntity
)

func Init(index []byte) error {
	ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
	defer cf()
	var err error
	ds, err = datastore.NewClient(ctx, "")
	if err != nil {
		return err
	}
	return createIndexes(index)

}

func createIndexes(index []byte) error {
	indexFile := "/etc/iwillbenice/index.xml"
	err := ioutil.WriteFile(indexFile, index, 0644)
	if err != nil {
		return err
	}

	return execute("gcloud", "--quiet", "datastore", "create-indexes", indexFile)
}

func execute(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type perspectivekey struct {
	Key string
}

func GetPerspectiveKey() string {
	key := datastore.NameKey("perspective", "key", nil)
	pk := &perspectivekey{}
	err := ds.Get(context.Background(), key, pk)
	if err != nil {
		log.Printf("Unable to read perspective key: %v", err)
	}
	return pk.Key
}

type vapidkey struct {
	PublicKey  string
	PrivateKey string
}

func GetVAPIDKey() *vapidkey {
	key := datastore.NameKey("vapid", "key", nil)
	pk := &vapidkey{}
	err := ds.Get(context.Background(), key, pk)
	if err != nil {
		log.Printf("Unable to read VAPID Private key: %v", err)
	}
	return pk
}
