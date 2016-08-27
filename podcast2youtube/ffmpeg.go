package podcast2youtube

import (
	"os"
	"os/exec"
)

// CreateVideo creates a video that plays the given audio with the given image
// as fixed background in the requested path.
func CreateVideo(img, mp3, vid string) error {
	const (
		slidePath = "slide.png"
		mp3Path   = "audio.mp3"
		vidPath   = "vid.mp4"
	)

	// ffmpeg -y -i slide.png -i audio.mp3 -pix_fmt yuv420p -c:a aac -c:v libx264 -crf 18 out.mp4
	cmd := exec.Command("ffmpeg", "-y", "-loop", "1", "-i", img, "-i", mp3, "-shortest",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-crf", "18",
		vid)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
