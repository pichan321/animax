package animax

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	logger "github.com/sirupsen/logrus"
)

func VerifyFilePath(filePath string) (err error) {
	file, err := os.Stat(filePath)
	if err != nil {
		return errors.New("file does not exist")
	}	

	if file.IsDir() {
		return errors.New("filePath specified is a directory")
	}
	return nil
}

/***
	Automatically recale video to 9:16 for Shorts Video. This is intentional.
***/
func AddOverlayBackground(video Video, outputPath string) (err error) {

	cmd := exec.Command("ffmpeg", "-i", video.FilePath, "-i", video.FilePath, "-filter_complex", "[1]scale=1080:600[vid]; [0]scale=1080:1920[img]; [img][vid] overlay=(W-w)/2:(H-h)/2", "-acodec", "copy", outputPath)
	err = cmd.Run()
	if err != nil {
		logger.Error("failed to add background")
		return err
	}
	logger.Info(fmt.Sprintf(`Overlay background added for %s | Output: %s`, video.FilePath, outputPath))
	return nil
}

/***
	Automatically recale video to 9:16 for Shorts Video. This is intentional.
***/
func AddOverlayBackgroundAndLogo(video Video, logoPath string, outputPath string) (err error) {
	cmd := exec.Command("ffmpeg", "-i", video.FilePath, "-i", video.FilePath, "-i", logoPath, "-filter_complex", "[1]scale=1080:600[vid]; [0]crop=540:960:(in_w-540)/2:(in_h-960)/2,scale=1080:1920[img]; [img]boxblur=15[blurred]; [blurred][vid]overlay=(W-w)/2:(H-h)/2[with_logo]; [2:v]scale=320:220[logo_resized]; [with_logo][logo_resized]overlay=20:H-h-300", "-acodec", "copy", outputPath)

	_, err = cmd.CombinedOutput()
	if err != nil {
		logger.Error("failed to add background")
		return err
	}
	logger.Info(fmt.Sprintf(`Overlay background added for %s | Output: %s`, video.FilePath, outputPath))
	return nil
}

func ConcatenateVideos(videos []Video, outputPath string) (err error) {
	


	return nil
}
