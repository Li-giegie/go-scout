package go_scout

import (
	"io/fs"
	"time"
)

var (
	SkipDir = fs.SkipDir
	SkipAll = fs.SkipAll
)

type Config struct {
	Paths         []string
	Sleep         time.Duration
	EnableHex     bool
	EventChanSize uint
}

type EventValue struct {
	Type     EventType
	Error    error
	FileInfo *FileInfo
}

type fileInfoX struct {
	info     *FileInfo
	updateAt int64
}

type ErrHash struct {
	path string
	err  error
}

func (h *ErrHash) Error() string {
	return "calculate Hash file " + h.path + " err: " + h.err.Error()
}

type EventType uint8

func (e EventType) String() string {
	switch e {
	case EventType_Init:
		return "INIT"
	case EventType_Create:
		return "CREATE"
	case EventType_Change:
		return "CHANGE"
	case EventType_Remove:
		return "REMOVE"
	case EventType_Error:
		return "ERROR"
	default:
		return "invalid event type"
	}
}

const (
	EventType_Init EventType = iota
	EventType_Create
	EventType_Change
	EventType_Remove
	EventType_Error
)

type SkipType uint8

const (
	SkipType_NoSkip SkipType = iota
	SkipType_File
	SkipType_Dir
	SkipType_All
)
