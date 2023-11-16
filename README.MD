# Animax 

 ![Animax](https://i.ibb.co/ccKy1PK/logo-circle.jpg)

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
	audio.Render("output.mp4")
```

### Contact:

 
Emails:
- pichsereyvattanchan@gmail.com

**Pichsereyvattana Chan**
