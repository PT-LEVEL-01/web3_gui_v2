package file_transfer

type UploadFile interface {
	GetFileSize() (string, uint64)                              //获取文件名称和大小
	Read(filename string, offset, length int64) ([]byte, error) //
}

type DownloadFile interface {
	GetFileSize() uint64                                     //获取文件大小
	Writer(filename string, data []byte, offset int64) error //
}
