package engine

import (
	"encoding/binary"
	"github.com/rs/zerolog"
	"io"
	"strings"
	"web3_gui/utils"
)

/*
循环接收消息
*/
func loopRecv(reader IOConn, ss Session, router *Router, authFun AuthFunc, log *zerolog.Logger) {
	//log.Info().Str("这里循环接收消息", "").Send()
	//处理客户端主动断开连接的情况
	var ERR utils.ERROR
	for {
		var packet *Packet
		packet, ERR = ReadStream(reader, log)
		if ERR.CheckFail() {
			//utils.Log.Error().Str("ERR", ERR.String()).Send()
			break
		}
		packet.Session = ss
		//utils.Log.Info().Str("这里触发了方法", "").Send()
		if ShowSendAndRecvLog {
			log.Info().Msgf("server conn recv: %d %s <- %s %d", packet.MsgID, ss.GetLocalHost(), ss.GetRemoteHost(), len(packet.Data)+16)
		}
		//这里决定了消息是否异步处理
		go handlerProcess(router, packet, authFun, log)
	}
}

/*
从连接中读取一个包大小的内容
*/
func ReadStream(conn IOConn, log *zerolog.Logger) (*Packet, utils.ERROR) {
	//接收的第一个包，一定是包头
	protocol_num := uint8(0)
	var err error
	bss := make([][]byte, 0)
	totalSize := int(0)
	for {
		//log.Info().Str("读取包长度", conn.LocalAddr().String()+"<-"+conn.RemoteAddr().String()).Send()
		//先读取5字节，4字节为包长度，1字节为协议号
		//bs, err := ReadFull(4+1, conn)
		bs := make([]byte, 4+1)
		_, err = io.ReadFull(conn, bs)
		if err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") &&
				!strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
				log.Error().Err(err).Send()
			}
			//log.Info().Str("读取包长度", conn.LocalAddr().String()+"<-"+conn.RemoteAddr().String()).Send()
			// Log.Error("recv remote packet size error:%s", err.Error())
			return nil, utils.NewErrorSysSelf(err)
		}
		//log.Info().Str("读取包长度", conn.LocalAddr().String()+"<-"+conn.RemoteAddr().String()).Send()
		//utils.Log.Info().Hex("Recv data length", bs).Send()
		//先读4字节包长度
		length := binary.BigEndian.Uint32(bs[:4])
		//判断包大小是否超过限制
		//if length > package_max_size {
		//	return nil, utils.NewErrorBus(ERROR_code_packege_over_max_size, "")
		//}
		totalSize += int(length)
		//再读1字节协议号
		protocol := bs[4]
		//utils.Log.Info().Int("协议号", int(protocol)).Send()
		//判断协议号对不对
		if protocol_num == 0 {
			if protocol == Package_protocol_start {
				protocol_num = Package_protocol_start
			} else if protocol == Package_protocol_single {
				protocol_num = Package_protocol_single
			} else {
				log.Error().Int("协议号错误", int(protocol)).Send()
				return nil, utils.NewErrorBus(ERROR_code_package_protocol_fail, "")
			}
		} else if protocol_num == Package_protocol_start {
			if protocol == Package_protocol_start {
				//上个包未传完，又开始传新的包了，允许将上个包因超时丢弃
				return nil, utils.NewErrorSuccess()
			} else if protocol == Package_protocol_single {
				//上个包未传完，又开始传新的包了，允许将上个包因超时丢弃
				return nil, utils.NewErrorSuccess()
			} else if protocol == Package_protocol_keep {
				protocol_num = Package_protocol_keep
			} else if protocol == Package_protocol_end {
				protocol_num = Package_protocol_end
			} else {
				log.Error().Str("协议号错误", "").Send()
				return nil, utils.NewErrorBus(ERROR_code_package_protocol_fail, "")
			}
		} else if protocol_num == Package_protocol_keep {
			if protocol == Package_protocol_start {
				//上个包未传完，又开始传新的包了，允许将上个包因超时丢弃
				return nil, utils.NewErrorSuccess()
			} else if protocol == Package_protocol_single {
				//上个包未传完，又开始传新的包了，允许将上个包因超时丢弃
				return nil, utils.NewErrorSuccess()
			} else if protocol == Package_protocol_keep {
				protocol_num = Package_protocol_keep
			} else if protocol == Package_protocol_end {
				protocol_num = Package_protocol_end
			} else {
				log.Error().Str("协议号错误", "").Send()
				return nil, utils.NewErrorBus(ERROR_code_package_protocol_fail, "")
			}
		} else {
			//utils.Log.Info().Int("接收包大小", len(bs)).Send()
		}
		//utils.Log.Info().Str("Recv start", conn.LocalAddr().String()+"<-"+conn.RemoteAddr().String()).Send()
		//再读包内容
		bs = make([]byte, length)
		_, err = io.ReadFull(conn, bs)
		if err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "use of closed network connection") {
				log.Error().Err(err).Send()
			}
			return nil, utils.NewErrorSysSelf(err)
		}
		bss = append(bss, bs)
		//log.Info().Str("读取包内容", conn.LocalAddr().String()+"<-"+conn.RemoteAddr().String()).Send()
		//判断传完了
		if protocol == Package_protocol_end || protocol == Package_protocol_single {
			//utils.Log.Info().Int("接收包大小", len(bs)).Send()
			break
		}
	}
	bs := make([]byte, 0, totalSize)
	for _, one := range bss {
		bs = append(bs, one...)
	}
	//utils.Log.Info().Int("接收包大小", len(bs)).Send()
	packet, err := ParsePacket(bs)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Int("接收包大小", len(bs)).Send()
	return packet, utils.NewErrorSuccess()
}

/*
异步执行用户注册的业务方法
*/
func handlerProcess(router *Router, packet *Packet, authFun AuthFunc, log *zerolog.Logger) {
	//log.Info().Msgf("engine执行方法回调:%d", packet.MsgID)
	//消息处理模块报错将不会引起宕机
	defer utils.PrintPanicStack(log)
	if len(packet.replyID) > 0 {
		//log.Warn().Interface("是回复消息", "").Send()
		//回复消息
		ResponseByteKey(Wait_major_engine_msg, packet.sendID, &packet.Data)
		return
	}
	//普通消息
	handler, ok := router.GetHandler(packet.MsgID)
	if !ok {
		log.Warn().Uint64("server The message is not registered", packet.MsgID).Send()
		return
	}
	//权限验证
	if authFun != nil {
		ERR := authFun(packet)
		if ERR.CheckFail() {
			log.Warn().Str("ERR", ERR.String()).Send()
			return
		}
	}
	handler(packet)
}

func ReadFull(length int, conn IOConn) ([]byte, error) {
	//先读包长度
	packetSizeBs := make([]byte, length)
	_, err := io.ReadFull(conn, packetSizeBs)
	if err != nil {
		// utils.Log.Error().Msgf("recv remote packet size error:%s", err.Error())
		return nil, err
	}
	return packetSizeBs, nil
}

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
