package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
)

func main() {
	flag.Parse()

	m := mux.NewRouter()

	client, err := CreateLocalstackS3Client(context.Background(), LocalstackConfig{
		Endpoint: os.Getenv("LOCALSTACK_ENDPOINT"),
	})
	if err != nil {
		log.Fatal(err)
	}

	bucket := os.Getenv("BUCKET")

	if err := EnsureBucketExistsByName(context.Background(), client, bucket); err != nil {
		log.Fatal(err)
	}

	m.Path("/{key}").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		output, err := client.GetObject(r.Context(), &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(mux.Vars(r)["key"]),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%v", err)
		} else {
			io.Copy(w, output.Body)
		}
	})

	m.Path("/{key}").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		response, err := client.PutObject(r.Context(), &s3.PutObjectInput{
			Bucket:        aws.String(bucket),
			Key:           aws.String(mux.Vars(r)["key"]),
			Body:          bytes.NewReader(body),
			ContentLength: int64(len(body)),
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%v", err)
		} else {
			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			enc.Encode(response)
		}
	})

	if err := http.ListenAndServe(os.Getenv("LISTEN_ADDR"), m); err != nil {
		log.Fatal(err)
	}
}
