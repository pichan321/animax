package animax

import (
	"fmt"
)

type Video struct {
	FileName string
	Fps float64
	Width int64
	Height int64
	Duration int64
	Volume int32
	OutputWidth int64
	OutputHeight int64
	OutputDuration int64
	OutputVolume int32
	Format string
	Args map[string][]string
	IsMuted bool
}

type TrimSection struct {
	StartTime int64
	EndTime int64
}

func PrintHello() {
	fmt.Println("Hello")
}

func (Video *Video) Load(filePath string) {
	
}

func (video *Video) Resize(width int64, height int64) {

}

func (video *Video) ResizeByWidth(width int64) {

}

func (video *Video) ResizeByHeight(height int64) {

}

func (video *Video) Trim() {

}

func (video *Video) MultipleTrim() {

}

func (video *Video) Crop(width int64, height int) {

}

func (video *Video) CropTop() {

}

func (video *Video) CropBottom() {

}

func (video *Video) CropLeft() {

}

func (video *Video) CropRight() {

}

func (video *Video) Skipper(skipDuration int64, skipInterval int64) {

}


func (video *Video) NewAspectRatio() {
	
}

func (video *Video) ChangeVideoVolume() {

}

func (video *Video) MuteAudio() {
	
}

func AddOverlayBackground() {

}

func ConcatenateVideos(videos []Video) {

}

func (video Video) Render() {

}
