package engine

import (
	"bytes"
	"encoding/binary"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"time"
	"web3_gui/utils"
)

/*
循环发送消息
*/
func loopSend(sendQueue *SendQueue, conn IOConn, log *zerolog.Logger) {
	csc := NewChunkSizeCalculator()
	//packageSize := 1024 //包起始大小1K
	var isClose bool
	c := sendQueue.GetQueueChan()
	var sp *SendPacket
	var ERR utils.ERROR
	for {
		sp, isClose = <-c
		if !isClose {
			// Log.Warn("out chan is close")
			//log.Error().Str("send queue is close", "").Send()
			return
		}
		//log.Info().Str("这里要发送消息", "111111111").Send()
		ERR = WriteStreamV2(csc, sp, conn, log)
		//log.Info().Str("这里要发送消息", "111111111").Send()
		sendQueue.SetResult(sp.ID, ERR)
		if ERR.CheckFail() {
			log.Error().Str("loopSend fail", ERR.String()).Send()
			return
		} else {
			if ShowSendAndRecvLog {
				log.Info().Msgf("server conn send: %s -> %s %d", conn.LocalAddr(), conn.RemoteAddr(), len(*sp.bs))
			}
		}
	}
}

func WriteStreamV2(csc *ChunkSizeCalculator, packet *SendPacket, conn IOConn, log *zerolog.Logger) utils.ERROR {
	content := *packet.bs
	//utils.Log.Info().Int("这里发送了消息", len(content)).Send()
	//size := startSize
	startIndex := 0
	endIndex := 0
	spend := time.Duration(0) //记录上一轮发送耗时
	fullPacket := false
	var err error
	for {
		chunkSize := csc.GetChunkSize(spend, fullPacket)
		endIndex = startIndex + int(chunkSize)
		fullPacket = false
		//log.Info().Int("startIndex", startIndex).Int("endIndex", endIndex).Send()
		if endIndex > len(content) {
			endIndex = len(content)
		} else {
			fullPacket = true
		}
		t1 := time.Now()
		//判断是否发送超时
		if t1.After(packet.timeout) {
			//log.Error().Str("Send", conn.LocalAddr().String()+"->"+conn.RemoteAddr().String()).
			//	Str("error", err.Error()).Send()
			return utils.NewErrorBus(ERROR_code_timeout_send, "")
		}
		//开始发送
		length := uint32(endIndex - startIndex)
		buf := bytes.NewBuffer(nil)
		//发送4字节包大小
		err = binary.Write(buf, binary.BigEndian, length)
		if err != nil {
			//log.Error().Str("Send", conn.LocalAddr().String()+"->"+conn.RemoteAddr().String()).
			//	Hex("Send data", buf.Bytes()).Str("error", err.Error()).Send()
			return utils.NewErrorSysSelf(err)
		}
		//utils.Log.Info()
		//判断是包头，续传，还是包尾
		protocol_num := Package_protocol_start
		if startIndex == 0 {
			if endIndex == len(content) {
				protocol_num = Package_protocol_single
			} else {
				protocol_num = Package_protocol_start
			}
		} else if endIndex == len(content) {
			protocol_num = Package_protocol_end
		} else {
			protocol_num = Package_protocol_keep
		}
		//再发送1字节协议号
		err = binary.Write(buf, binary.BigEndian, protocol_num)
		if err != nil {
			//log.Error().Str("Send", conn.LocalAddr().String()+"->"+conn.RemoteAddr().String()).
			//	Str("error", err.Error()).Send()
			return utils.NewErrorSysSelf(err)
		}
		//log.Info().Int("startIndex", startIndex).Int("endIndex", endIndex).Send()
		//再发送包内容
		_, err = buf.Write(content[startIndex:endIndex])
		if err != nil {
			//log.Error().Str("Send", conn.LocalAddr().String()+"->"+conn.RemoteAddr().String()).
			//	Str("error", err.Error()).Send()
			return utils.NewErrorSysSelf(err)
		}
		//utils.Log.Info().Hex("Send data", buf.Bytes()).Send()
		_, err := conn.Write(buf.Bytes())
		if err != nil {
			//log.Error().Str("Send", conn.LocalAddr().String()+"->"+conn.RemoteAddr().String()).
			//	Str("error", err.Error()).Send()
			return utils.NewErrorSysSelf(err)
		}
		//log.Info().Str("Send", conn.LocalAddr().String()+"->"+conn.RemoteAddr().String()).Hex("Send data", buf.Bytes()).Int("n", n).Send()
		spend = time.Now().Sub(t1)
		// 检查时间差是否大于10毫秒
		//if time.Now().Sub(t1) >= 10*time.Millisecond {
		//	//fmt.Println("处理时间大于10毫秒")
		//	size = size / 2
		//} else {
		//	if fullPacket {
		//		//fmt.Println("处理时间小于等于10毫秒")
		//		size = size * 2
		//	}
		//}
		if endIndex == len(content) {
			break
		}
		//设置下一轮偏移量
		startIndex = endIndex
		//endIndex = startIndex + size
	}
	return utils.NewErrorSuccess()
}

func WriteStream(startSize int, packet *SendPacket, conn IOConn, csc *ChunkSizeCalculator, log *zerolog.Logger) (int, utils.ERROR) {
	content := *packet.bs
	//utils.Log.Info().Int("这里发送了消息", len(content)).Send()
	size := startSize
	startIndex := 0
	endIndex := startSize
	var err error
	for {
		fullPacket := false
		//log.Info().Int("startIndex", startIndex).Int("endIndex", endIndex).Send()
		if endIndex > len(content) {
			endIndex = len(content)
		} else {
			fullPacket = true
		}
		t1 := time.Now()
		//判断是否发送超时
		if t1.After(packet.timeout) {
			return size, utils.NewErrorBus(ERROR_code_timeout_send, "")
		}
		//开始发送
		length := uint32(endIndex - startIndex)
		buf := bytes.NewBuffer(nil)
		//发送4字节包大小
		err = binary.Write(buf, binary.BigEndian, length)
		if err != nil {
			return size, utils.NewErrorSysSelf(err)
		}
		//utils.Log.Info()
		//判断是包头，续传，还是包尾
		protocol_num := Package_protocol_start
		if startIndex == 0 {
			if endIndex == len(content) {
				protocol_num = Package_protocol_single
			} else {
				protocol_num = Package_protocol_start
			}
		} else if endIndex == len(content) {
			protocol_num = Package_protocol_end
		} else {
			protocol_num = Package_protocol_keep
		}
		//再发送1字节协议号
		err = binary.Write(buf, binary.BigEndian, protocol_num)
		if err != nil {
			return size, utils.NewErrorSysSelf(err)
		}
		//log.Info().Int("startIndex", startIndex).Int("endIndex", endIndex).Send()
		//再发送包内容
		_, err = buf.Write(content[startIndex:endIndex])
		if err != nil {
			return size, utils.NewErrorSysSelf(err)
		}
		//utils.Log.Info().Hex("Send data", buf.Bytes()).Send()
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			return size, utils.NewErrorSysSelf(err)
		}
		//log.Info().Str("Send", conn.LocalAddr().String()+"->"+conn.RemoteAddr().String()).Hex("Send data", buf.Bytes()).Send()
		// 检查时间差是否大于10毫秒
		if time.Now().Sub(t1) >= 10*time.Millisecond {
			//fmt.Println("处理时间大于10毫秒")
			size = size / 2
		} else {
			if fullPacket {
				//fmt.Println("处理时间小于等于10毫秒")
				size = size * 2
			}
		}
		if endIndex == len(content) {
			break
		}
		//计算下一轮偏移量
		startIndex = endIndex
		endIndex = startIndex + size
	}
	return size, utils.NewErrorSuccess()
}

func send(msgID uint64, data *[]byte, timeout time.Duration, queue *SendQueue) (*Packet, utils.ERROR) {
	//utils.Log.Info().Str("send", "").Send()
	p := NewPacket(msgID, data)
	bs, err := p.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return p, queue.AddAndWaitTimeout(bs, timeout)
}

/*
发送消息后等待回复
*/
func sendWait(msgID uint64, data *[]byte, timeout time.Duration, queue *SendQueue, log *zerolog.Logger) (*[]byte, utils.ERROR) {
	//id := ulid.Make().Bytes()
	//log.Info().Hex("发送并等待接口", id).Send()
	p := NewPacket(msgID, data)
	bs, err := p.Proto()
	if err != nil {
		//log.Error().Str("发送失败", err.Error()).Send()
		return nil, utils.NewErrorSysSelf(err)
	}
	RegisterRequestKey(Wait_major_engine_msg, p.sendID)
	defer RemoveRequestKey(Wait_major_engine_msg, p.sendID)
	ERR := queue.AddAndWaitTimeout(bs, timeout)
	if ERR.CheckFail() {
		//log.Info().Str("发送失败", ERR.String()).Send()
		return nil, ERR
	}
	//log.Info().Hex("发送并等待接口", id).Str("超时时间", timeout.String()).Send()
	bs2, ERR := WaitResponseByteKey(Wait_major_engine_msg, p.sendID, timeout)
	if ERR.CheckFail() {
		//log.Info().Str("发送失败", ERR.String()).Send()
		//log.Info().Hex("发送并等待接口", id).Send()
		return nil, ERR
	}
	//log.Info().Hex("发送并等待接口", id).Send()
	//bs2, ok := itr.(*[]byte)
	//if !ok {
	//	return nil, utils.NewErrorBus(ERROR_code_response_type_fail, "")
	//}
	return bs2, utils.NewErrorSuccess()
}
func reply(packet *Packet, data *[]byte, timeout time.Duration, queue *SendQueue) utils.ERROR {
	//utils.Log.Info().Str("reply", "").Send()
	if data != nil {
		packet.Data = *data
	}
	packet.replyID = ulid.Make().Bytes()
	bs, err := packet.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return queue.AddAndWaitTimeout(bs, timeout)
}
