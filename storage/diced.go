package storage

import (
	"bytes"
	"encoding/hex"
	"golang.org/x/crypto/sha3"
	"io"
	"os"
	"path/filepath"
	"time"
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/utils"
)

/*
把文件切片
@filePath     string    要切片的文件路径
@chunkPath    string    切片存放文件夹路径
@return    *model.FileIndex    文件索引
@return    []byte              文件原始hash，作为加密key
@return    error               错误
*/
func Diced(key, iv []byte, filePath, chunkDirPath string) (*model.FileIndex, utils.ERROR) {
	//filehash, err := utils.FileSHA3_256(filePath)
	//if err != nil {
	//	return nil, err
	//}
	err := utils.CheckCreateDir(chunkDirPath)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}

	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		// fmt.Println("111", err)
		return nil, utils.NewErrorSysSelf(err)
	}
	fileinfo, err := f.Stat()

	chunkCount := fileinfo.Size() / config.Chunk_size
	if (fileinfo.Size() % config.Chunk_size) > 0 {
		chunkCount = chunkCount + 1
	}
	fileindex := model.FileIndex{
		//ID: filehash, //文件加密后的hash值
		//UserID         []byte   //用户地址
		Version:      config.FileChunk_version,  //版本号
		Name:         []string{fileinfo.Name()}, //文件名称
		FileSize:     uint64(fileinfo.Size()),   //文件总大小
		ChunkCount:   uint32(chunkCount),        //分片总量
		ChunkOneSize: config.Chunk_size,         //每一个分片大小
		//Chunks         [][]byte //每一个分片ID
		PermissionType: []uint8{config.PermissionType_self}, //权限类型 0=仅自己可访问;1=仅自己授权者可访问;2=所有人可访问;
		EncryptionType: config.EncryptionType_AES_ctr,       //加密类型
		Time:           []int64{time.Now().Unix()},          //
	}
	//按照块大小分割
	bs := make([]byte, config.Chunk_size)
	for i := 0; i < int(fileinfo.Size()/config.Chunk_size); i++ {
		// fmt.Println("4444444444444")
		f.Seek(int64(i*config.Chunk_size), 0)
		_, err := f.Read(bs)
		if err != nil {
			utils.Log.Info().Msgf("切片报错:%s", err.Error())
			return nil, utils.NewErrorSysSelf(err)
		}
		chunkHash, ERR := saveFileChunk(key, iv, chunkDirPath, bs, fileindex.EncryptionType)
		if !ERR.CheckSuccess() {
			// fmt.Println("444", err)
			return nil, ERR
		}
		fileindex.Chunks = append(fileindex.Chunks, chunkHash)
	}
	//有余就再把余下的文件做成一个块
	if fileinfo.Size()%config.Chunk_size > 0 {
		f.Seek(fileinfo.Size()/config.Chunk_size*config.Chunk_size, 0)
		n, err := f.Read(bs)
		if err != nil {
			// fmt.Println("555", err)
			return nil, utils.NewErrorSysSelf(err)
		}
		bs = bs[:n]
		chunkHash, ERR := saveFileChunk(key, iv, chunkDirPath, bs, fileindex.EncryptionType)
		if !ERR.CheckSuccess() {
			// fmt.Println("444", err)
			return nil, ERR
		}
		fileindex.Chunks = append(fileindex.Chunks, chunkHash)
	}
	fileindex.Hash = utils.BuildMerkleRoot(fileindex.Chunks)
	fileindex.ChunkOffsetIndex = make([]uint64, len(fileindex.Chunks))
	fileindex.PullIDs = make([][]byte, len(fileindex.Chunks))
	//utils.Log.Info().Msgf("文件切片:%+v", fileindex.Chunks)
	//utils.Log.Info().Msgf("加密文件Hash:%+v", fileindex.ID)
	return &fileindex, utils.NewErrorSuccess()
}

/*
写一个文件块到磁盘
@return    string    新文件名称（文件hash值）
@return    error     返回错误
*/
func saveFileChunk(key, iv []byte, chunkPath string, bs []byte, encryptionType uint32) ([]byte, utils.ERROR) {
	var err error
	switch encryptionType {
	case config.EncryptionType_AES_ctr:
		bs, err = utils.AesCTR_Encrypt(key, iv, bs)
	default:
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_encry_type_Not_Supported, "") //config.ERROR_encry_type_Not_Supported
	}
	if err != nil {
		return nil, utils.NewErrorSysSelf(err) // err
	}

	buf := bytes.NewBuffer(bs)
	hash_sha3 := sha3.New256()
	_, err = io.Copy(hash_sha3, buf)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err) // err
	}
	chunkhash := hash_sha3.Sum(nil)
	chunkhashStr := hex.EncodeToString(chunkhash)

	//把文件块写到目标文件夹
	file, err := os.OpenFile(filepath.Join(chunkPath, chunkhashStr), os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		file.Close()
		return nil, utils.NewErrorSysSelf(err) // err
	}
	_, err = file.Write(bs)
	if err != nil {
		file.Close()
		return nil, utils.NewErrorSysSelf(err) // err
	}
	file.Close()
	return chunkhash, utils.NewErrorSuccess()
}

/*
从传输的参数中采集key和iv
把key填充到block size大小
长度太短的，前面补零
@return    []byte    key
@return    []byte    iv
*/
func splitKeyAndIv(key []byte) ([]byte, []byte) {
	if len(key) >= 32 {
		return key[:16], key[16:]
	}
	length := len(key)
	bs := make([]byte, 32)
	copy(bs[32-length:], key)
	return bs[:16], bs[16:]
}

/*
将文件切片合并成完整的文件
@fileIndex    *model.FileIndex    要组装的文件
@chunkPath    string              切片文件夹
@filePath     string              新文件存放路径
*/
func MergeFile(key, iv []byte, fileIndex *model.FileIndex, chunkPath, filePath string) error {
	dirName, _ := filepath.Split(filePath)
	err := utils.CheckCreateDir(dirName)
	if err != nil {
		return err
	}
	//先创建一个文件
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	//设置文件大小，先占用磁盘空间
	if err := f.Truncate(int64(fileIndex.FileSize)); err != nil {
		return err
	}
	index := int64(0)
	for _, one := range fileIndex.Chunks {
		bs, err := os.ReadFile(filepath.Join(chunkPath, hex.EncodeToString(one)))
		if err != nil {
			return err
		}
		//解密
		bs, err = utils.AesCTR_Decrypt(key, iv, bs)
		if err != nil {
			return err
		}
		_, err = f.Seek(index, 0)
		if err != nil {
			return err
		}
		_, err = f.Write(bs)
		if err != nil {
			return err
		}
		index += int64(len(bs))
	}
	return nil
}
