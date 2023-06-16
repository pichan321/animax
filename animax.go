package animax

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

	logger "github.com/sirupsen/logrus"
)
type Args map[string][]string

type Video struct {
	FileName string
	Fps float64
	Width int64
	Height int64
	Duration int64
	Volume int32
	Format string
	args Args
	IsMuted bool
}

type TrimSection struct {
	StartTime int64
	EndTime int64
	OutputName string
}



func PrintHello() {
	fmt.Println("Hello")
}

func (args Args) addArg(key string, value string) {
    args[key] = append(args[key], value)
}

func pullStats(videoPath string) {
	cmd := exec.Command("ffmpeg", "-i", videoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		
	}
	fmt.Println(string(output))

}


func Load(videoPath string) (video Video, err error) { 
	file, err := os.Stat(videoPath)
	if err != nil {
		logger.Error(fmt.Sprintf(`videoPath: %s does not exist`, videoPath))
		return Video{}, err
	}

	if file.IsDir() {
		logger.Error(fmt.Sprintf(`videoPath: %s is a directory`, videoPath))
		return Video{}, errors.New("videoPath: %s is a directory")
	}

	pullStats(videoPath)

	
	return Video{
		FileName: filepath.Base(videoPath),
		Format: filepath.Ext(videoPath),
		args: make(Args),
	}, nil
}

func (video *Video) Resize(width int64, height int64) (modifiedVideo Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`scale=%d:%d`, width, height))
	return modifiedVideo
}

func (video *Video) ResizeByWidth(width int64) (modifiedVideo Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`scale=%d:%d`, width, -1))
	return modifiedVideo
}

func (video *Video) ResizeByHeight(height int64) (modifiedVideo Video) {
	video.args.addArg("-filter_complex", fmt.Sprintf(`scale=%d:%d`, -1 , height))
	return modifiedVideo
}

func (video *Video) Trim(startTime int64, endTime int64) (modifiedVideo Video){
	if startTime > endTime {
		logger.Error("start time cannot be bigger than end time")
	}
	
	// video.args.addArg("-ss", fmt.Sprint(startTime))
	// video.args.addArg("-to", fmt.Sprint(endTime))
	video.args.addArg("-filter_complex", fmt.Sprintf(`trim=start=%d:end=%d`, startTime, endTime))
	return modifiedVideo
}

func (video *Video) MultipleTrim(concatenateAfter bool, trimSections []TrimSection) {
	for _, v := range trimSections {
		fmt.Print(v)
	}
}

func (video *Video) Crop(width int64, height int) {

}

func (video *Video) CropTop(width int64) {

}

func (video *Video) CropBottom(width int64) {

}

func (video *Video) CropLeft(width int64) {

}

func (video *Video) CropRight(width int64) {

}

func MultiFilter() {
	
}

func (video *Video) Skipper(skipDuration int64, skipInterval int64) {

}


func (video *Video) NewAspectRatio() {
	
}

func (video *Video) ChangeVideoVolume() {

}

func (video *Video) MuteAudio() {
	
}




func (video Video) queryBuilder(outputPath string) []string {
	query := []string{"ffmpeg", "-i", video.FileName}
	for k, v := range video.args {
		query = append(query, k)
		if reflect.TypeOf(v).Name() == "string" {
			// query = append(query, v)
		}
	
	}
	query = append(query, []string{"-c:v", "copy", "-c:a", "copy", outputPath}...)
	// query = append(query, []string{outputPath}...)
	fmt.Println(query)
	return query
}

func (video Video) Render(outputPath string) {
	_, err := os.Stat(outputPath)
	if err == nil {
		os.Remove(outputPath)
	}
	// outputQuery := video.queryBuilder()
	// cmd := exec.Command()
	query := video.queryBuilder(outputPath)
	cmd := exec.Command(query[0], query[1:]...)
	logger.Info("Command to be executed: " + cmd.String())
	ouput, err := cmd.Output()
	if err != nil {
		logger.Error(err.Error())
		return
	}
	logger.Info(string(ouput))
}
