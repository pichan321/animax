package animax

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Video struct {
	FileName string
	FilePath string
	Width int64
	Height int64
	Duration int64
	AspectRatio string
	Format string
	args Args
	IsMuted bool
}

type TrimSection struct {
	StartTime int64
	EndTime int64
	OutputName string
}

var videoExtensions = []string{".mp4", ".avi", ".mov", ".mkv"}
var VIDEO_ENCODINGS = struct {
	Best string
	Efficient string
	Compressed string
}{
	Best: "libx264",
	Efficient: "libvpx-vp9",
	Compressed: "libaom-av1",
}

var ASPECT_RATIOS = struct {
	Square float32
	Standard float32
	
	Shorts float32 //Youtube Shorts, Facebok Reels, Instagram Reels, TikTok Videos
	Videos float32 //Youtube Videos, Facebok Videos, General Videos
}{
	Square: 1.0,
	Standard: 16.0/9.0,

	Shorts: 9.0/16.0, //Youtube Shorts, Facebok Reels, Instagram Reels, TikTok Videos
	Videos: 16.0/9.0, //Youtube Videos, Facebok Videos, General Videos
}

type subArg struct {
	Key string
	Value string
}

func (args Args) addArg(flag string, subAr subArg) {
    args[flag] = append(args[flag], subAr)
}

func contains(format string) bool {
	for _, val := range videoExtensions {
		if format == val {return true}
	}
	return false
}

func calculatePts(n int, fps float64) float64 {
	return float64(n) / fps
}

func searchPts(fps float64, frames int, start float64) float64 {
	low, high := 1, frames
	for low <= high {
		mid := (low + high) / 2
		pts := calculatePts(mid, fps)
		if math.Abs(pts-start) <= 1.0 {
			ptsStr := strconv.FormatFloat(pts, 'f', 5, 64)
			pts, _ = strconv.ParseFloat(ptsStr, 64)
			return pts
		} else if pts > start {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return -1
}

func (v Video) GetType() string {
	return video
}

func (v Video) GetFilename() string {
	return v.FileName
}

func (v Video) GetFilePath() string {
	return v.FilePath
}

func (v Video) GetExtension() string {
	return filepath.Ext(v.FilePath)
}

func (video Video) getFramesAndFps() (float64, int){
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=nb_frames,avg_frame_rate", "-of", "default=noprint_wrappers=1", video.FilePath)
	output, _ := cmd.CombinedOutput()
	outputLines := strings.Split(string(output), "\n")

	var fps float64 = -1
	var frames int = 0
	for _, line := range outputLines {
		if strings.HasPrefix(line, "avg_frame_rate") && fps == -1 {
			fpsStr := strings.ReplaceAll(strings.Split(line, "=")[1], "\r", "")
			fpsParts := strings.Split(fpsStr, "/")
			if len(fpsParts) == 2 {
				numerator, _ := strconv.ParseFloat(fpsParts[0], 64)
				denominator, _ := strconv.ParseFloat(fpsParts[1], 64)
				if denominator != 0 {
					fps = numerator / denominator
				}
			}
		}
		if strings.HasPrefix(line, "nb_frames") && frames == 0 {
			framesStr := strings.ReplaceAll(strings.Split(line, "=")[1], "\r", "")
			framesValue, err := strconv.ParseInt(framesStr, 10, 64)
			if err == nil {
				frames = int(framesValue)
			}
		}
	}
	return fps, frames
}

func (video Video) SeekFrame(time int64) float64 {
	fps, frames := video.getFramesAndFps()
	newTime := searchPts(fps, frames, float64(time))
	if newTime == -1 {return float64(time)}
	return newTime
}

func pullVideoStats(videoPath string) (width int, height int, duration int, aspectRatio string) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=width,height,duration,display_aspect_ratio", "-of", "default=noprint_wrappers=1", videoPath)
	output, _ := cmd.CombinedOutput()
	outputLines := strings.Split(string(output), "\n")
	if len(outputLines) < 5 {
		return -1, -1, -1, ""
	}

	for _, v :=range outputLines {
		switch strings.Split(v, "=")[0] {
			case "width":
				width, _ = strconv.Atoi(strings.TrimSuffix(strings.Split(v, "=")[1], "\r"))
			case "height":
				height, _ = strconv.Atoi(strings.TrimSuffix(strings.Split(v, "=")[1], "\r"))
			case "duration":
				duration, _ = strconv.Atoi(strings.Split(strings.TrimSuffix(strings.Split(v, "=")[1], "\r"), ".")[0])
				if strings.HasPrefix(v, "duration") {
					var err error
					duration, err = strconv.Atoi(strings.Split(strings.TrimSuffix(strings.Split(v, "=")[1], "\r"), ".")[0])
					if err != nil {
						duration, _ = strconv.Atoi(strings.Split(strings.TrimSuffix(strings.Split(v, "=")[1], "\r"), ".")[0])
					}
				}
			case "display_aspect_ratio":
				aspectRatio = strings.TrimSuffix(strings.Split(outputLines[2], "=")[1], "\r")
		}
	}

	return width, height, duration, aspectRatio
}

/*
	Takes in the path of the video to be loaded and returns Video struct containing the video's metadata if the videoPath provided is valid.
*/
func LoadVideo(videoPath string) (video Video, err error) {
	file, err := os.Stat(videoPath)
	if err != nil {
		Logger.Error(fmt.Sprintf(`videoPath: %s does not exist`, videoPath))
		return Video{}, err
	}

	if file.IsDir() {
		Logger.Error(fmt.Sprintf(`videoPath: %s is a directory`, videoPath))
		return Video{}, errors.New("videoPath: %s is a directory")
	}

	width, height, duration, aspectRatio := pullVideoStats(videoPath)
	fileFormat :=  filepath.Ext(videoPath)
	if !contains(fileFormat) {
		Logger.Error(fmt.Sprintf(`videoPath: %s | Video format is not supported\n`, videoPath))
		return Video{}, err
	}
	return Video{
		FileName:    filepath.Base(videoPath),
		FilePath:    videoPath,
		Format:   fileFormat,
		Width:       int64(width),
		Height:      int64(height),
		Duration:    int64(duration),
		AspectRatio: aspectRatio,
		args:        make(Args),
	}, nil
}

func (video *Video) Resize(width int64, height int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", 
		subArg{
			Key:   "scale",
			Value: fmt.Sprintf(`scale=%d:%d`, width, height),
	})
	return video
}

func (video *Video) ResizeByWidth(width int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", 
		subArg{
			Key: "scale",
			Value: fmt.Sprintf(`scale=%d:%d`, width, -1),
		})
		
	return video
}

func (video *Video) ResizeByHeight(height int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", 
		subArg{
			Key: "scale",
			Value: fmt.Sprintf(`scale=%d:%d`, -1 , height),
		})
		
	return video
}

func (video *Video) Trim(startTime int64, endTime int64) (modifiedVideo *Video){
	if startTime > endTime {
		Logger.Errorln("Start time cannot be bigger than end time")
		return &Video{}
	}

	newStart := video.SeekFrame(startTime)
	video.args.addArg("-ss", 
	subArg{
		Key: "ss",
		Value: 	fmt.Sprintf(`%f -to %f`, newStart, float64(endTime)),
	})
	
	
	return video
}

func (video *Video) Crop(width int64, height int64, startingPositions ...int64) (modifiedVideo *Video) {
	var x, y int64 = 0, 0
	for index, value := range startingPositions {
		if index == 0 {x = value}
		if index == 1  {y = value}
	} 
	video.args.addArg("-filter_complex", 
		subArg{
			Key: "crop",
			Value: fmt.Sprintf(`crop=%d:%d:%d:%d`, width, height, x, y),
		})
		
	return modifiedVideo
}

func (video *Video) CropOutTop(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", 
		subArg{
			Key: "crop",
			Value: fmt.Sprintf(`crop=in_w:in_h-%d:0:out_h`, pixels),
		})
		
	return video
}

func (video *Video) CropOutBottom(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", 
		subArg{
			Key: "crop",
			Value: fmt.Sprintf(`crop=in_w:in_h-%d:0:0`, pixels),
		})
	return video
}

func (video *Video) CropOutLeft(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", 
		subArg{
			Key: "crop",
			Value: 	fmt.Sprintf(`crop=in_w-%d:in_h:%d:0`, pixels, pixels),
		})
	return video
}

func (video *Video) CropOutRight(pixels int64) (modifiedVideo *Video) {
	video.args.addArg("-filter_complex", 
		subArg{
			Key: "crop",
			Value: fmt.Sprintf(`crop=in_w-%d:in_h:0:0`, pixels),
		})
	return video
}

func (video *Video) Blur(intensity int16) (modifiedVideo *Video) {
	if (intensity < 0 || intensity > 50) {
		Logger.Warn("Blur intensity should be between 0 and 50")
		return &Video{}
	}
 	video.args.addArg("-filter_complex", 
		subArg{
			Key: "boxblur",
			Value: fmt.Sprintf(`boxblur=%d`, intensity),
		})
	return video
}

func (video *Video) NewAspectRatio(aspectRatio float32) (modifiedVideo *Video) {
	video.args.addArg("-aspect", 
		subArg{
			Key: "aspect",
			Value: fmt.Sprintf(`%f`,aspectRatio),
		})
	return video
}

// func (video *Video) NewAspectRatioPadAuto(aspectRatio float32) (modifiedVideo *Video) {
// 	return video
// }

func (video *Video) ChangeVolume(multiplier float64) (modifiedVideo *Video) {
	video.args.addArg("-filter:a", 
		subArg{
			Key: "volume",
			Value: fmt.Sprintf(`volume=%f`, multiplier),
		})
	return  video
}

func (video *Video) MuteAudio() (modifiedVideo *Video) {
	video.args.addArg("-filter:a", 
		subArg{
			Key: "volume",
			Value: `volume=0`,
		})

	return video
}

func (video *Video) Saturate(multiplier float64) (modifiedVideo *Video) {
	video.args.addArg("-vf", 
		subArg{
			Key: "saturation",
			Value: fmt.Sprintf("eq=saturation=%f", multiplier),
		})
	return video
}

/*
	If there exists a file at the specified outputPath, the file will be overwritten.
*/
func (video Video) Render(outputPath string, videoEncoding string) (outputVideo Video) {
	removeIfExists(outputPath)

	if videoEncoding == "" {videoEncoding = VIDEO_ENCODINGS.Best}

	g := GetRenderGraph(VideoGraph)
	renderStages := g.ProduceOrdering(video.args, &video)
	
	if len(renderStages) == 0 {
		Logger.Errorf("No effects applied. Aborting render.\n")
		return video
	}

	fmt.Printf("\nALL STAGES %+v\n\n", renderStages)
	startRender(&renderStages, video, outputPath)
	outputVideo , err := LoadVideo(outputPath)

	if err != nil {
		return video
	}
	return outputVideo
}