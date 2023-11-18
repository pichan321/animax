# Animax 

### Description
 Animax is a Go (Golang) public module created in hope of making video/audio editing on Golang become a much easier task for everyone. Unlike any other binders/wrappers, animax uses graph + rules in order to maximize effiency in video editing especially for clipping videos/audios by choosing when to re-encode or not depending on the effects you use.
  
### Installations

Animax heavily relies on FFMPEG. To run Animax you would first have to install FFMPEG and add it to your environment path. 
1. Instructions on how to install FFMPEG can be found on this link: [How to Install FFmpeg ](https://gist.github.com/barbietunnie/47a3de3de3274956617ce092a3bc03a1)
2. go get "github.com/pichan321/animax"

### Documentation

>https://pkg.go.dev/github.com/pichan321/animax

### Usage

  #### Video

 ##### Load Video

```go
	video, err := animax.LoadVideo("shin.mp4")
	if err != nil {
		panic(err)
	}
```
##### Apply effects
Effects can be applied as many as you want on a particular video. Trim effects will always be prioritized and executed first regardless of when it is applied (saving CPU and time).

```go
	video.Trim(200, 300).MuteAudio().Saturate(1.5)
```
##### Render
Render method will actually perform render on the video based on all the effects you have chained. Render takes in an output path and video encoding. 

The below two lines are equivalent.
```go
	video.Render("output.mp4", animax.VIDEO_ENCODINGS.Best)
```

```go
	video.Render("output.mp4", "")
```

![Video Render Graph](https://i.ibb.co/8rfdWsQ/Untitled-2023-11-18-0002.png)

##### What happens if trim effect is used only once

A video with only a single trim effect applied will always be re-encoded.

```go
	func main() {
		video, err := animax.LoadVideo("shin.mp4")
		if err != nil {
			panic(err)
		}
		video.Trim(0, 500)
		video.Render("output.mp4", "")
	}
```

![Video Render Graph](https://i.ibb.co/Y0c8Yfy/Reencode.png)

##### What happens if trim effect is used multiple times on a video but not continiously

```go
	func main() {
		video, err := animax.LoadVideo("shin.mp4")
		if err != nil {
			panic(err)
		}
		video.Trim(0, 500).Saturate(1.5).Trim(100, 350).MuteAudio().Trim(0, 150).CropOutTop(100)
		video.Render("output.mp4", "")
	}
```

Render process graph from above code:

![Video Render Graph](https://i.ibb.co/K50zWg8/Complex.png)

### Audio

#### Load Audio

```go
	audio, err := animax.LoadAudio("audio.mp3")
	if err != nil {
		panic(err)
	}
```

#### Apply effects
The same as video, you can apply and chain as many effects as you want, but trim effects will always be prioritized and executed first for efficiency.
  
```go
	audio.Trim(100, 200).Nightcore()
```

#### Render

```go
	audio.Render("output.mp3")
```

![Audio Render Graph](https://i.ibb.co/pdbgdwb/Audio-Render.png)

### Contact:
 
Emails:
- pichsereyvattanchan@gmail.com

**Pichsereyvattana Chan**
