package leveldb

import "errors"

type DataType byte

const (
	NoneType byte = iota
	KVType
	HashType
	HSizeType
)
const (
	// key最大size
	MaxKeySize int = 1024

	// field最大大小
	MaxHashFieldSize int = 1024

	// value最大大小
	MaxValueSize int = 1024 * 1024 * 1024
)

var (
	errKeySize       = errors.New("invalid key size")
	errValueSize     = errors.New("invalid value size")
	errHashFieldSize = errors.New("invalid hash field size")
	ErrNotFound      = errors.New("leveldb: not found")
)

const (
	KB int = 1024
	MB int = KB * 1024
	GB int = MB * 1024
)
