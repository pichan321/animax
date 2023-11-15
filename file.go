package animax

type File interface {
	GetType() string
	GetFilename() string
	GetFilePath() string
	GetExtension() string
}

const (
	video = "video"
	audio = "audio"
)