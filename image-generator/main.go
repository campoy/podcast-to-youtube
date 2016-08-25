package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
)

const (
	projectID  = "podcast-to-youtube"
	topicName  = "image-generation"
	subsName   = "worker"
	bucketName = "podcast-to-youtube"
)

func main() {
	ctx := context.Background()

	ps, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("could not create pubsub client: %v", err)
	}

	subs, _ := ps.NewSubscription(ctx, subsName, ps.Topic(topicName), 10*time.Second, nil)
	tasks, err := subs.Pull(ctx)
	if err != nil {
		log.Fatalf("could not iterate over pubsub tasks: %v", err)
	}

	for {
		task, err := tasks.Next()
		if err != nil {
			log.Printf("could not get next task: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		var data struct{ Number, Title, MP3 string }
		if err := json.Unmarshal(task.Data, &data); err != nil {
			log.Printf("could not decode task: %v", err)
			task.Done(true)
			continue
		}

		log.Printf("processing task: %v %v", data.Number, data.Title)

		if err := processTask(ctx, data.Number, data.Title, data.MP3); err != nil {
			log.Printf("could not process task: %v", err)
			task.Done(false)
			continue
		}

		log.Printf("taks processed successful")
		task.Done(true)
	}
}

func processTask(ctx context.Context, number, title, mp3 string) error {
	// create a new file, will truncate if existing.
	f, err := os.Create("slide.png")
	if err != nil {
		return fmt.Errorf("could not create slide.png: %v", err)
	}
	defer f.Close()

	// create the background image for the video and writing to slide.png.
	m, err := createImage(number, title)
	if err != nil {
		return err
	}
	if err := png.Encode(f, m); err != nil {
		return fmt.Errorf("could not encode image: %v", err)
	}

	// download the mp3 and save to audio.mp3
	res, err := http.Get(mp3)
	if err != nil {
		return fmt.Errorf("could not download audio %s: %v", mp3, err)
	}
	defer res.Body.Close()

	f, err = os.Create("audio.mp3")
	if err != nil {
		return fmt.Errorf("could not create audio.mp3: %v", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, res.Body); err != nil {
		return fmt.Errorf("could not write to audio.mp3: %v", err)
	}

	vid, err := createVideo("slide.png", "audio.mp3")
	if err != nil {
		return fmt.Errorf("could not create video: %v", err)
	}

	return upload(ctx, number, title, vid)
}

func upload(ctx context.Context, number, title, path string) error {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("could not create storage client: %v", err)
	}

	w := c.Bucket(bucketName).Object(fmt.Sprintf("%s-%s.mp4", number, title)).NewWriter(ctx)
	w.ContentType = "video/mp4"

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %s: %v", path, err)
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("could not write to storage: %v", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("could not close storage writer: %v", err)
	}
	return nil
}
