package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
)

func uploadVideo(ctx context.Context, number int, path string) error {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("could not create storage client: %v", err)
	}

	w := c.Bucket(bucketName).Object(fmt.Sprintf("%d.mp4", number)).NewWriter(ctx)
	w.ContentType = "video/mp4"

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %s: %v", path, err)
	}
	defer f.Close()

	fmt.Println("uploading video to Google Cloud Storage")
	r, err := progressBarReader(f)
	if err != nil {
		log.Printf("could not create progress bar: %v", err)
		r = f
	}

	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("could not write to storage: %v", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("could not close storage writer: %v", err)
	}
	return nil
}

func videoExists(ctx context.Context, number int) bool {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return false
	}
	_, err = c.Bucket(bucketName).Object(fmt.Sprintf("%d.mp4", number)).Attrs(ctx)
	return err == nil
}

func downloadVideo(ctx context.Context, number int) (string, error) {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("could not create storage client: %v", err)
	}

	r, err := c.Bucket(bucketName).Object(fmt.Sprintf("%d.mp4", number)).NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("could not create storage reader: %v", err)
	}
	defer r.Close()

	const vid = "vid.mp4"
	f, err := os.Create(vid)
	if err != nil {
		return "", fmt.Errorf("could not create %s: %v", vid, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("could not write to %s: %v", vid, err)
	}
	return vid, nil
}
