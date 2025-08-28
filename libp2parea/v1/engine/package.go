package engine

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/quic-go/quic-go"
	"io"
	"net"
	"web3_gui/libp2parea/v1/engine/protobuf/go_protobuf"
	"web3_gui/utils"
)

const (
	max_size = 1024 * 1024 * 1024 //一个包内容最大容量
)

// var Netid uint32 = 2 //网络版本id，避免冲突

type GetPacket func(cache *[]byte, index *uint32) (packet *Packet, n int)
type GetPacketBytes func(msgID, opt, errcode uint32, cryKey []byte, data *[]byte) *[]byte

type Packet struct {
	MsgID    uint64 //
	Size     uint64 //数据包长度，包含头部4字节
	Data     []byte
	Dataplus []byte //未加密部分数据分开
	Session  Session
}

/*
系统默认的消息接收并转化为Packet的方法
一个packet包括包头和包体，保证在接收到包头后两秒钟内接收到包体，否则线程会一直阻塞
因此，引入了超时机制
*/
func RecvPackage(conn net.Conn) (*Packet, error) {
	// defer PrintPanicStack()
	packet := new(Packet)

	packetSizeBs, err := ReadFull(8, conn)
	if err != nil {
		return nil, err
	}

	// wantCount := 8
	// //先读包长度
	// packetSizeBs := make([]byte, wantCount)
	// count := 0
	// for {
	// 	n, err := io.ReadFull(conn, packetSizeBs[count:])
	// 	if err != nil && err.Error() != io.EOF.Error() {
	// 		// utils.Log.Error().Msgf("recv remote packet size error:%s", err.Error())
	// 		return nil, err
	// 	}
	// 	count += n
	// 	if count == wantCount {
	// 		break
	// 	}
	// }
	//解析包长度
	packet.Size = binary.LittleEndian.Uint64(packetSizeBs[:8])

	if packet.Size > max_size {
		//包头错误 包长度大于最大值
		return nil, errors.New("packet size too big")
	}

	//再读包内容
	packetBodyBs, err := ReadFull(int(packet.Size), conn)
	if err != nil {
		return nil, err
	}
	// packetBodyBs := make([]byte, packet.Size)
	// n, err := io.ReadFull(conn, packetBodyBs)
	// if err != nil {
	// 	// fmt.Println("接收对方node错误 44444", err)
	// 	if strings.Contains(err.Error(), "use of closed network connection") {
	// 	} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
	// 	} else {
	// 		utils.Log.Error().Msgf("recv remote packet error:%s", err.Error())
	// 	}
	// 	return nil, err
	// }
	// // utils.Log.Info().Msgf("recv remote packet size:%d", n)
	// if uint64(n) != packet.Size {
	// 	return nil, errors.New("recv remote packet size error")
	// }

	packetProto := go_protobuf.Packet{}
	err = proto.Unmarshal(packetBodyBs, &packetProto)
	if err != nil {
		utils.Log.Error().Msgf("recv packet proto unmarshal error:%s", err.Error())
		return nil, err
	}
	packet.MsgID = packetProto.MsgID
	packet.Data = packetProto.Data
	packet.Dataplus = packetProto.Dataplus
	return packet, nil
}

func MarshalPacket(msgID uint64, data, dataplus *[]byte) (*[]byte, error) {
	packetProto := go_protobuf.Packet{
		MsgID: msgID,
		// Data:     *data,
		// Dataplus: *dataplus,
	}
	if data != nil {
		packetProto.Data = *data
	}
	if dataplus != nil {
		packetProto.Dataplus = *dataplus
	}
	bs, err := packetProto.Marshal()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	totalSize := uint64(len(bs))
	binary.Write(buf, binary.LittleEndian, totalSize)
	// utils.Log.Info().Msgf("打包头大小 %d 字节 %s", totalSize, hex.EncodeToString(buf.Bytes()))
	buf.Write(bs)
	bs = buf.Bytes()
	//	NLog.Info(LOG_file, "MarshalPacket size:%d;data长度:%d;dataplus长度:%d", len(bs), len(dataBuf.Bytes()), dataplusSize)
	return &bs, nil
}

///*
//	系统默认的消息接收并转化为Packet的方法
//	一个packet包括包头和包体，保证在接收到包头后两秒钟内接收到包体，否则线程会一直阻塞
//	因此，引入了超时机制
//*/
//func RecvPackage(conn net.Conn, packet *Packet) error {
//	// fmt.Println("packet   11111", *index, (*cache))
//	defer PrintPanicStack()
//	if len(packet.temp) < 16 {
//		cache := make([]byte, 16)
//		n, err := conn.Read(cache)
//		//	fmt.Println(n, err != nil)
//		if err != nil {
//			return err
//		}
//		packet.temp = append(packet.temp, cache[:n]...)
//	}

//	packet.Size = binary.LittleEndian.Uint64(packet.temp[:8])
//	packet.MsgID = binary.LittleEndian.Uint64(packet.temp[8:16])

//	if packet.Size < 16 {
//		return errors.New("包头错误")
//	}

//	for uint64(len(packet.temp)) < packet.Size {
//		cache := make([]byte, packet.Size-16)
//		n, err := conn.Read(cache)
//		if err != nil {
//			utils.Log.Debug().Msgf("err %v %d %d", err, n, uint64(n))
//			return err
//		}
//		packet.temp = append(packet.temp, cache[:n]...)
//	}
//	packet.Data = packet.temp[16:packet.Size]

//	//	packet.temp = make([]byte, uint64(len(packet.temp))-packet.Size)
//	//	if uint64(len(packet.temp))-packet.Size != 0 {
//	//		copy(packet.temp, packet.temp[packet.Size:])
//	//	}

//	oldtemp := packet.temp
//	packet.temp = make([]byte, uint64(len(oldtemp))-packet.Size)
//	//	fmt.Println("packet  ", uint64(len(oldtemp)), packet.Size)
//	if uint64(len(oldtemp))-packet.Size > 0 {
//		copy(packet.temp, oldtemp[packet.Size:])
//	}

//	return nil
//}

//func MarshalPacket(msgID uint64, data, dataplus *[]byte) *[]byte {
//	//	newCryKey := RandKey128()
//	if data == nil || len(*data) <= 0 {
//		buf := bytes.NewBuffer([]byte{})
//		binary.Write(buf, binary.LittleEndian, uint64(16))
//		binary.Write(buf, binary.LittleEndian, msgID)
//		bs := buf.Bytes()
//		return &bs
//	}

//	bodyBytes := *data
//	buf := bytes.NewBuffer([]byte{})
//	binary.Write(buf, binary.LittleEndian, uint64(len(bodyBytes)+16))
//	binary.Write(buf, binary.LittleEndian, msgID)
//	buf.Write(bodyBytes)
//	bs := buf.Bytes()
//	return &bs
//}

// func cry(in []byte) []byte {
// 	i := 0
// 	tmpBuf := make([]byte, 128)
// 	for i < len(in) {
// 		if i+1 < len(in) {
// 			tmpBuf[i] = in[i+1]
// 			tmpBuf[i+1] = in[i]
// 		} else {
// 			tmpBuf[i] = in[i]
// 		}
// 		i += 2
// 	}
// 	out := make([]byte, len(in))
// 	for i := 0; i < len(in); i++ {
// 		out[i] = tmpBuf[i] & 0x01
// 		out[i] = tmpBuf[i] & 0x0f
// 		out[i] <<= 4
// 		out[i] |= ((tmpBuf[i] & 0xf0) >> 4)
// 	}
// 	return out
// }

func ReadFull(length int, conn net.Conn) ([]byte, error) {
	//先读包长度
	packetSizeBs := make([]byte, length)
	_, err := io.ReadFull(conn, packetSizeBs)
	if err != nil {
		// utils.Log.Error().Msgf("recv remote packet size error:%s", err.Error())
		return nil, err
	}
	return packetSizeBs, nil
}

// func ReadFull(length int, conn net.Conn) ([]byte, error) {
// 	wantCount := length
// 	//先读包长度
// 	packetSizeBs := make([]byte, wantCount)
// 	count := 0
// 	for {
// 		n, err := io.ReadFull(conn, packetSizeBs[count:])
// 		if err != nil && err.Error() != io.EOF.Error() {
// 			// utils.Log.Error().Msgf("recv remote packet size error:%s", err.Error())
// 			return nil, err
// 		}
// 		count += n
// 		if count == wantCount {
// 			break
// 		}
// 		utils.Log.Info().Msgf("not read full:%d want:%d", n, wantCount)
// 		time.Sleep(time.Millisecond * 50)
// 	}
// 	return packetSizeBs, nil
// }

func cry(in []byte) []byte {
	i := 0
	tmpBuf := make([]byte, 128)
	for i < len(in) {
		if i+1 < len(in) {
			tmpBuf[i] = in[i+1]
			tmpBuf[i+1] = in[i]
		} else {
			tmpBuf[i] = in[i]
		}
		i += 2
	}
	out := make([]byte, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = tmpBuf[i] & 0x01
		out[i] = tmpBuf[i] & 0x0f
		out[i] <<= 4
		out[i] |= ((tmpBuf[i] & 0xf0) >> 4)
	}
	return out
}

/*
系统默认的消息接收并转化为Packet的方法
一个packet包括包头和包体，保证在接收到包头后两秒钟内接收到包体，否则线程会一直阻塞
因此，引入了超时机制
*/
func RecvQuicPackage(stream quic.Stream) (*Packet, error) {
	// defer PrintPanicStack()
	packet := new(Packet)

	packetSizeBs, err := ReadQuicFull(8, stream)
	if err != nil {
		return nil, err
	}

	//解析包长度
	packet.Size = binary.LittleEndian.Uint64(packetSizeBs[:8])

	if packet.Size > max_size {
		//包头错误 包长度大于最大值
		return nil, errors.New("packet size too big")
	}

	//再读包内容
	packetBodyBs, err := ReadQuicFull(int(packet.Size), stream)
	if err != nil {
		return nil, err
	}

	packetProto := go_protobuf.Packet{}
	err = proto.Unmarshal(packetBodyBs, &packetProto)
	if err != nil {
		utils.Log.Error().Msgf("recv packet proto unmarshal error:%s", err.Error())
		return nil, err
	}
	packet.MsgID = packetProto.MsgID
	packet.Data = packetProto.Data
	packet.Dataplus = packetProto.Dataplus
	return packet, nil
}

func ReadQuicFull(length int, conn quic.Stream) ([]byte, error) {
	//先读包长度
	packetSizeBs := make([]byte, length)
	_, err := io.ReadFull(conn, packetSizeBs)
	if err != nil {
		// utils.Log.Error().Msgf("recv remote packet size error:%s", err.Error())
		return nil, err
	}
	return packetSizeBs, nil
}
