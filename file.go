package animax

type File interface {
	GetType() string
	Render() File
}

const (
	video = "video"
	audio = "audio"
)