package animax

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pichan321/animax"
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
	Automatically rescale video to 9:16 for Shorts Video. This is intentional.
	Returns nil if successful and an error otherwise.
***/
func AddOverlayBackground(video animax.Video, outputPath string) (err error) {
	err = VerifyFilePath(video.FilePath)
	if err != nil {
		return err
	}

	err = VerifyFilePath(outputPath)
	if err == nil {
		os.Remove(outputPath)
	}
	
	animax.Logger.Info(fmt.Sprintf(`Adding overlay background for %s | Output: %s`, video.FilePath, outputPath))
	cmd := exec.Command("ffmpeg", "-i", video.FilePath, "-i", video.FilePath, "-filter_complex", "[1]scale=1080:600[vid]; [0]scale=1080:1920[img]; [img][vid] overlay=(W-w)/2:(H-h)/2", "-c:v", "libx264", "-c:a", "aac", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		animax.Logger.Errorf("Failed to add background | Error: %s", string(output))
		return err
	}
	animax.Logger.Info(fmt.Sprintf(`Overlay background added for %s | Output: %s`, video.FilePath, outputPath))
	return nil
}

/***
	Automatically rescale video to 9:16 for Shorts Video. This is intentional.
	Returns nil if successful and an error otherwise.
***/
func AddOverlayBackgroundAndLogo(video animax.Video, logoPath string, outputPath string) (err error) {
	err = VerifyFilePath(logoPath)
	if err != nil {
		return err
	}
	err = VerifyFilePath(video.FilePath)
	if err != nil {
		return err
	}

	err = VerifyFilePath(outputPath)
	if err == nil {
		os.Remove(outputPath)
	}
	animax.Logger.Info(fmt.Sprintf(`Adding overlay background with logo for %s | Output: %s`, video.FilePath, outputPath))
	cmd := exec.Command("ffmpeg", "-i", video.FilePath, "-i", video.FilePath, "-i", logoPath, "-filter_complex", "[1]scale=1080:600[vid]; [0]crop=540:960:(in_w-540)/2:(in_h-960)/2,scale=1080:1920[img]; [img]boxblur=15[blurred]; [blurred][vid]overlay=(W-w)/2:(H-h)/2[with_logo]; [2:v]scale=320:220[logo_resized]; [with_logo][logo_resized]overlay=20:H-h-300", "-c:v", "libx264", "-c:a", "aac", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		animax.Logger.Errorf("Failed to add background | Error: %s", string(output))
		return err
	}
	animax.Logger.Info(fmt.Sprintf(`Overlay background added for %s | Output: %s`, video.FilePath, outputPath))
	return nil
}

/***
	Takes in a slice of containing Video structs and concatenates all those structs into a single video output file.
	Returns nil if successful and an error otherwise.
***/
func ConcatenateVideos(videos []animax.Video, encode bool, outputPath string) (err error) {
	err = VerifyFilePath(outputPath)
	if err == nil {
		os.Remove(outputPath)
	}

	inputTextFileName := fmt.Sprintf(`%s-temp-input.txt`, uuid.New().String()[0:8])
	inputTextFile, err := os.Create(inputTextFileName)
	if err != nil {
		return errors.New("unable to create a temp ")
	}

	defer os.Remove(inputTextFileName)

	for _, video := range videos {
		inputTextFile.WriteString(fmt.Sprintf(`file '%s'`, video.FilePath) + "\n")
	}
	inputTextFile.Close()

	if encode {
		cmd := exec.Command("ffmpeg", "-f", "concat", "-i", inputTextFileName, "-c:v", "libx264", "-c:a", "aac", outputPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			animax.Logger.Info(fmt.Sprintf(`Error: %s`, string(output)))
			return err
		}
		return nil
	}

	cmd := exec.Command("ffmpeg", "-f", "concat", "-i", inputTextFileName, "-c:v", "copy", "-c:a", "copy", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		animax.Logger.Info(fmt.Sprintf(`Error: %s`, string(output)))
		return err
	}
	return nil
}

func ConcatenateVideosFromDir(directoryPath string, encode bool, outputPath string) error {
	dir, err := os.Stat(directoryPath); 
	if os.IsNotExist(err) {
		animax.Logger.Errorf("%s does not exist", directoryPath)
		return errors.New("path does not exist")
	}

	if !dir.IsDir() {
		animax.Logger.Error("Invalid directoryPath specified")
		return errors.New("invalid directory path")
	}

	filesInDir, err :=os.ReadDir(directoryPath)
	if err != nil {
		animax.Logger.Errorf("Could not read files from directory %s", directoryPath)
		return err
	}
	
	videosInDir := []animax.Video{}
	for _, file := range filesInDir {
		videoPath := fmt.Sprintf(`%s/%s`, directoryPath, file.Name())
		extension := filepath.Ext(videoPath)
		if extension == ".mp4" {
			video, err := animax.LoadVideo(videoPath)
			if err != nil {
				animax.Logger.Info(fmt.Sprintf(`file: %s is used in concatenation since it is not a video`, filepath.Base(videoPath)))
				continue
			}
			animax.Logger.Infof("Appending %s for video concatenation", file.Name())
			videosInDir = append(videosInDir, video)
		}
	}

	err = ConcatenateVideos(videosInDir, false, outputPath)
	if err != nil {
		animax.Logger.Errorf("Error during concatenation phase | %s", err)
		return err
	}

	return nil
}

func TrimNoEncode(video animax.Video, startTime int64, endTime int64, outputString string) (animax.Video, error) {
	newStart := video.SeekFrame(startTime)
	if newStart == -1 {newStart = float64(startTime)}
	cmd := exec.Command("ffmpeg", "-ss", fmt.Sprintf("%.5f", newStart), "-i", video.FilePath, "-to", fmt.Sprintf("%.5f", float64(endTime) - float64(startTime)), "-c", "copy", "-y", outputString)
	animax.Logger.Infoln("Command to be executed: " + cmd.String())
	_, err := cmd.CombinedOutput()
	if err != nil {
		return animax.Video{}, errors.New("unable to trim the video")
	}

	outputVideo, err := animax.LoadVideo(outputString)
	if err != nil {
		return animax.Video{}, errors.New("unexpected error")
	}
	return outputVideo, nil
}

func Skipper(video animax.Video, skipDuration float64, skipInterval float64, outputPath string) error {
	nextFrameToSkip := 0.0
	workingDir := uuid.New().String()

	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		os.Mkdir(workingDir, os.ModePerm)
	}
	videoDuration := video.Duration
	originalVideoPath := video.FilePath
	index := 0
	clipsToConcat := []animax.Video{}
	animax.Logger.Infof("Video: %s | Path: %s | Initiating skipper", video.FileName, video.FilePath)
	for i := 0.0; i < float64(videoDuration); i++ {
		currentFrame := i

		if (currentFrame >= float64(videoDuration)) || (nextFrameToSkip + skipDuration > float64(videoDuration)) || (nextFrameToSkip >= float64(videoDuration)) {break}
		if currentFrame < nextFrameToSkip {continue}

		start := nextFrameToSkip
		end := nextFrameToSkip + skipInterval

		if end >= float64(videoDuration) {end = float64(videoDuration)}

		// clipUuid := uuid.New().String()
		clipName := fmt.Sprintf(`%s/%d.mp4`, workingDir, index)
		originalVideo, err := animax.LoadVideo(originalVideoPath)
		if err != nil {
			animax.Logger.Errorf("Video: %s | Path: %s | Error reading file", originalVideo.FileName, originalVideo.FilePath)
			animax.Logger.Infof("Video: %s | Path: %s | Exiting skipper loop", originalVideo.FileName, originalVideo.FilePath)
			break
		}

		video, _ = TrimNoEncode(originalVideo, int64(start), int64(end), clipName)
		// video = originalVideo.Trim(int64(start), int64(end)).Render(clipName, animax.VIDEO_ENCODINGS.Best)
		clipsToConcat = append(clipsToConcat, video)
		animax.Logger.Infof("Video: %s | Start: %f | End: %f |Path: %s | Subclip %s generated", video.FileName, start, end, video.FilePath, clipName)
		index++
		nextFrameToSkip = end + skipDuration

		if nextFrameToSkip > float64(videoDuration) {
			animax.Logger.Infof("Video: %s | Path: %s | Exiting skipper loop", originalVideo.FileName, originalVideo.FilePath)
			break
		}
	}

	animax.Logger.Infof("Video: %s | Path: %s | Concatenating all clips in working directory %s", video.FileName, video.FilePath, workingDir)
	err := ConcatenateVideos(clipsToConcat, true, outputPath)
	if err != nil {
		animax.Logger.Error("Error during concatenation")
		return err
	}
	animax.Logger.Infof("Video: %s | Path: %s | Completed concatenation in working directory %s", video.FileName, video.FilePath, workingDir)
	animax.Logger.Infof("Video: %s | Path: %s | Cleaning up working directory %s", video.FileName, video.FilePath, workingDir)
	defer os.RemoveAll(workingDir)
	return nil
}

func ExtractAudio(videoPath string, outputPath string) error {
	return nil
}

func TakeScreenshot(videoPath string, time float64, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", fmt.Sprintf("%f", time), "-frames:v", "1", "-y", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		animax.Logger.Infof("%s\n", string(output))
		return err
	}
	
	return nil
}