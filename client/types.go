package client

var SelectedAudioFormat string

type AudioType struct {
	Mime      string
	Extension string
}

var AudioFormatMap = map[string]AudioType{
	"aac":  {"mp4a", "m4a"},
	"opus": {"opus", "webm"},
	"wav":  {"wav", "wav"},
	"mp3":  {"mp3", "mp3"},
}
