package engine

/*
	关闭连接回调接口
*/
type CloseCallback func(ss Session)

/*
	服务器有新连接，触发的回调方法
*/
type ServerNewConnCallback func(ss Session, params interface{}) error

/*
	客户端有新连接，触发的回调方法
*/
type ClientNewConnCallback func(ss Session, params interface{}) error
