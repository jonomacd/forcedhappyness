package dao

import (
	"context"
	"io/ioutil"
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
