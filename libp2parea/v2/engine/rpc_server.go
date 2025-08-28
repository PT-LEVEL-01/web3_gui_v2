package engine

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"web3_gui/utils"
)

type RPCServer struct {
	engine         *Engine             //
	tcpAddr        *net.TCPAddr        //
	httpListen     *HttpListen         //
	closed         *atomic.Bool        //是否已经关闭
	rpcHandlerLock *sync.RWMutex       //
	rpcHandler     map[string]*RpcInfo //rpc回调方法
	rpcUserLock    *sync.RWMutex       //
	rpcUser        map[string]string   //
	contextRoot    context.Context     //
	canceRoot      context.CancelFunc  //
}

func NewRPCServer(engine *Engine) *RPCServer {
	closed := new(atomic.Bool)
	closed.Store(true)
	rpcServer := &RPCServer{
		engine:         engine,
		closed:         closed,
		rpcUserLock:    new(sync.RWMutex),
		rpcUser:        make(map[string]string),
		rpcHandlerLock: new(sync.RWMutex),
		rpcHandler:     make(map[string]*RpcInfo),
	}
	rpcServer.rpcUser[RPC_username] = RPC_password
	rpcServer.contextRoot, rpcServer.canceRoot = context.WithCancel(engine.contextRoot)
	ERR := rpcServer.RegisterRpcHandler(0, RPC_method_rpclist, rpcServer.rpcList, "")
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	ERR = rpcServer.RegisterRpcHandler(1, RPC_method_errors_desc, rpcServer.errorsDesc, "获取系统中所有错误编号及说明")
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	return rpcServer
}

/*
注册一个rpc接口
*/
func (this *RPCServer) RegisterRpcHandler(sortNumber int, method string, handler any, desc string, pvs ...ParamValid) utils.ERROR {
	//utils.Log.Info().Str("注册一个方法", method).Send()
	this.rpcHandlerLock.Lock()
	defer this.rpcHandlerLock.Unlock()
	_, exist := this.rpcHandler[method]
	if exist {
		return utils.NewErrorBus(ERROR_code_rpc_method_repeat, method)
	}
	future := true
	if handler != nil {
		t := reflect.TypeOf(handler)
		//必须是一个方法
		if t.Kind() != reflect.Func {
			return utils.NewErrorBus(ERROR_code_rpc_method_fail, method)
		}
		//检查传入参数
		ERR := checkParamsIn(t, pvs...)
		if ERR.CheckFail() {
			return ERR
		}
		//检查返回参数
		ERR = checkParamsOut(t)
		if ERR.CheckFail() {
			return ERR
		}
		future = false
	}
	fi := RpcInfo{
		MethodName: method,                   //
		Handler:    reflect.ValueOf(handler), //rpc回调方法
		Desc:       desc,                     //rpc接口描述
		ParamValid: pvs,                      //rpc参数验证
		Index:      sortNumber,               //编号，排序用
		Future:     future,                   //
	}
	//utils.Log.Info().Str("注册一个方法", method).Send()
	this.rpcHandler[method] = &fi
	return utils.NewErrorSuccess()
}

func checkParamsIn(funType reflect.Type, pvs ...ParamValid) utils.ERROR {
	//utils.Log.Info().Int("方法参数总量", 1).Send()
	numParams := funType.NumIn()
	//对比参数数量是否一致
	if numParams > 0 && numParams != len(pvs)+1 {
		return utils.NewErrorBus(ERROR_code_rpc_method_param_total_fail, "")
	}

	//if numParams == 0 {
	//	return utils.NewErrorSuccess()
	//}
	//utils.Log.Info().Int("方法参数总量", numParams).Send()
	// 遍历并打印每个参数的类型
	for i := 0; i < numParams; i++ {
		if i == 0 {
			//第一个参数
			paramOne := funType.In(0)
			//utils.Log.Info().Int("第0个参数类型", int(paramOne.Kind())).Int("类型", int(paramOne.Elem().Kind())).
			//	Bool("1", paramOne.Kind() != reflect.Pointer).Bool("2", paramOne.Elem().Kind() != reflect.Map).Send()
			if paramOne.Kind() != reflect.Pointer || paramOne.Elem().Kind() != reflect.Map {
				return utils.NewErrorBus(ERROR_code_rpc_method_param_type_fail, fmt.Sprintf("第%d个参数，类型应该是：*map[string]interface{}", i))
			}
			continue
		}
		paramType := funType.In(i) // 获取第i个参数的类型
		//utils.Log.Info().Int("对比类型1", int(paramType.Kind())).Int("对比类型2", int(pvs[i-1].ValueKind)).Send()
		//对比参数类型是否一致
		pvsOne := pvs[i-1]
		if pvsOne.IsSlice {
			if paramType.Kind() != reflect.Slice || paramType.Elem().Kind() != pvsOne.ValueKind {
				return utils.NewErrorBus(ERROR_code_rpc_method_param_type_fail, fmt.Sprintf("第%d个参数", i))
			}
		} else {
			if paramType.Kind() != pvsOne.ValueKind {
				return utils.NewErrorBus(ERROR_code_rpc_method_param_type_fail, fmt.Sprintf("第%d个参数", i))
			}
		}
	}
	return utils.NewErrorSuccess()
}

func checkParamsOut(funType reflect.Type) utils.ERROR {
	if funType.NumOut() != 1 {
		return utils.NewErrorBus(ERROR_code_rpc_method_result_total_fail, "")
	}
	outType := funType.Out(0)
	if outType.Kind() != reflect.Struct {
		//utils.Log.Info().Uint("返回参数类型", uint(outType.Kind())).Send()
		return utils.NewErrorBus(ERROR_code_rpc_method_result_type_fail, "")
	}
	//对比返回类型中的field是否一致
	pv := reflect.TypeOf(PostResult{})
	if pv.NumField() != outType.NumField() {
		return utils.NewErrorBus(ERROR_code_rpc_method_result_type_fail, "")
	}
	for i := range pv.NumField() {
		if outType.Field(i).Name != pv.Field(i).Name || outType.Field(i).Type.Kind() != pv.Field(i).Type.Kind() {
			return utils.NewErrorBus(ERROR_code_rpc_method_result_type_fail, "")
		}
	}
	return utils.NewErrorSuccess()
}

/*
查询一个rpc接口
@return    reflect.Value    这个方法
@return    *[]ParamValid    方法的传入参数
@return    bool             是否找到这个方法
*/
func (this *RPCServer) findRpcHandler(method string) (*RpcInfo, bool) {
	this.rpcHandlerLock.RLock()
	defer this.rpcHandlerLock.RUnlock()
	rpcInfo, exist := this.rpcHandler[method]
	return rpcInfo, exist
}

/*
接口列表
*/
func (this *RPCServer) rpcGetMethod(params map[string]interface{}) (string, utils.ERROR) {
	//utils.Log.Info().Interface("输入参数", params).Send()
	//methodItr, ok := params["method"]
	//utils.Log.Info().Interface("输入参数", params).Bool("method", ok).Send()
	methodItr, ok := params[RPC_method]
	if !ok {
		//utils.Log.Info().Str("RPC 调用开始", "11111111111111").Send()
		ERR := utils.NewErrorBus(ERROR_code_rpc_method_not_found, "")
		return "", ERR
	}
	methodName, ok := methodItr.(string)
	if !ok {
		//utils.Log.Info().Str("RPC 调用开始", "11111111111111").Send()
		ERR := utils.NewErrorBus(ERROR_code_rpc_method_type_fail, "")
		return methodName, ERR
	}
	return methodName, utils.NewErrorSuccess()
}

/*
检查用户名密码
*/
func (this *RPCServer) checkUserPassword(params map[string]interface{}, method string) utils.ERROR {
	return utils.NewErrorSuccess()
	need := false //默认需要验证用户名密码
	//检查用户名
	usernameItr, ok := params[RPC_username]
	if !ok {
		utils.Log.Info().Str("验证用户名密码", "").Send()
		//判断是否不需要用户名密码
		if this.checkRpcUser("", "") {
			utils.Log.Info().Str("验证用户名密码", "").Send()
			need = false
		} else {
			//查看接口列表不需要用户名密码
			if method == RPC_method_rpclist {
				return utils.NewErrorSuccess()
			}
			utils.Log.Info().Str("验证用户名密码", "").Send()
			//utils.Log.Info().Str("RPC 调用开始", "11111111111111").Send()
			ERR := utils.NewErrorBus(ERROR_code_rpc_user_fail, "")
			return ERR
		}
	}
	if need {
		utils.Log.Info().Str("验证用户名密码", "").Send()
		username, ok := usernameItr.(string)
		if !ok {
			return utils.NewErrorBus(ERROR_code_rpc_user_fail, "")
		}
		//检查密码
		pwdItr, ok := params[RPC_password]
		if !ok {
			return utils.NewErrorBus(ERROR_code_rpc_user_fail, "")
		}
		pwd, ok := pwdItr.(string)
		if !ok {
			return utils.NewErrorBus(ERROR_code_rpc_user_fail, "")

		}
		//判断rpc用户名密码
		if !this.checkRpcUser(username, pwd) {
			return utils.NewErrorBus(ERROR_code_rpc_user_fail, "")

		}
	}
	return utils.NewErrorSuccess()
}

/*
执行对应方法
*/
func (this *RPCServer) proessMethodHandler(method string, params map[string]interface{}) (map[string]interface{}, utils.ERROR) {
	//utils.Log.Info().Str("RPC 调用开始", "11111111111111").Send()
	rpcInfo, ok := this.findRpcHandler(method)
	if !ok {
		//utils.Log.Info().Str("RPC 调用开始", "11111111111111").Send()
		return nil, utils.NewErrorBus(ERROR_code_rpc_method_not_found, "")
	}
	if rpcInfo.Future {
		return nil, utils.NewErrorBus(ERROR_code_rpc_method_is_future, "")
	}

	//utils.Log.Info().Str("RPC 调用开始", "11111111111111").Send()
	//验证传入参数
	paramsValue, ERR := validParam(&params, &rpcInfo.ParamValid)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//调用方法
	fOuts := rpcInfo.Handler.Call(paramsValue)
	prValue := fOuts[0]
	//
	m := make(map[string]interface{})
	pv := reflect.TypeOf(PostResult{})
	for i := range pv.NumField() {
		valueItr := prValue.Field(i).Interface()
		m[pv.Field(i).Name] = valueItr
	}
	//r := fnInfo(params)
	return m, utils.NewErrorSuccess()
}

/*
接口列表
*/
func (this *RPCServer) rpcList(params *map[string]interface{}) PostResult {
	pr := NewPostResult()
	this.rpcHandlerLock.Lock()
	defer this.rpcHandlerLock.Unlock()
	list := make([]RpcInfoVO, 0, len(this.rpcHandler))
	//utils.Log.Info().Int("方法总数", len(this.rpcHandler)).Send()
	for methodName, rpcInfoOne := range this.rpcHandler {
		//utils.Log.Info().Str("方法", methodName).Send()
		if methodName == RPC_method_rpclist {
			continue
		}
		voOne := rpcInfoOne.ConverVO()
		list = append(list, *voOne)
	}
	//排序
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Index < list[j].Index
	})
	pr.Data["list"] = list
	return *pr
}

/*
错误编号描述
*/
func (this *RPCServer) errorsDesc(params *map[string]interface{}) PostResult {
	pr := NewPostResult()
	this.rpcHandlerLock.Lock()
	defer this.rpcHandlerLock.Unlock()
	eds := utils.GetErrorsList()
	pr.Data["list"] = eds
	return *pr
}

/*
添加一个用户
*/
func (this *RPCServer) AddRpcUser(user, password string) utils.ERROR {
	this.rpcUserLock.Lock()
	defer this.rpcUserLock.Unlock()
	if _, ok := this.rpcUser[user]; ok {
		return utils.NewErrorBus(ERROR_code_rpc_user_exist, "")
	}
	this.rpcUser[user] = password
	return utils.NewErrorSuccess()
}

/*
添加一个用户
*/
func (this *RPCServer) UpdateRpcUser(user, password string) utils.ERROR {
	this.rpcUserLock.Lock()
	defer this.rpcUserLock.Unlock()
	if user == RPC_username {
		RPC_password = password
	}
	if _, ok := this.rpcUser[user]; !ok {
		return utils.NewErrorBus(ERROR_code_rpc_user_not_found, "")
	}
	this.rpcUser[user] = password
	return utils.NewErrorSuccess()
}

/*
删除一个用户
*/
func (this *RPCServer) DelRpcUser(user string) utils.ERROR {
	this.rpcUserLock.Lock()
	defer this.rpcUserLock.Unlock()
	if _, ok := this.rpcUser[user]; !ok {
		return utils.NewErrorBus(ERROR_code_rpc_user_not_found, "")
	}
	delete(this.rpcUser, user)
	return utils.NewErrorSuccess()
}

/*
验证rpc用户名密码
*/
func (this *RPCServer) checkRpcUser(user, pwd string) bool {
	this.rpcUserLock.RLock()
	defer this.rpcUserLock.RUnlock()
	//当没有添加rpc用户的情况下，所有人都可以连接
	if len(this.rpcUser) == 0 {
		return true
	}
	password, ok := this.rpcUser[user]
	if !ok {
		return false
	}
	if password == pwd {
		return true
	}
	return false
}

/*
销毁，断开连接，关闭监听
@param	areaName	[]byte	区域名
*/
func (this *RPCServer) Destroy() {
	this.closed.Store(true)
	//this.httpListen.SetRPCHandler(nil)
}

type RpcInfo struct {
	MethodName string        //
	Handler    reflect.Value //rpc回调方法
	Desc       string        //rpc接口描述
	ParamValid []ParamValid  //rpc参数验证
	Index      int           //编号，排序用
	Future     bool          //将来待实现
}

func (this *RpcInfo) ConverVO() *RpcInfoVO {
	vo := RpcInfoVO{
		MethodName: this.MethodName,
		Desc:       this.Desc,
		Params:     make([]ParamValidVO, 0, len(this.ParamValid)),
		Index:      this.Index,
	}
	for _, one := range this.ParamValid {
		vo.Params = append(vo.Params, one.ConverVO())
	}
	return &vo
}

type RpcInfoVO struct {
	MethodName string         //方法名称
	Desc       string         //接口描述
	Params     []ParamValidVO //参数
	Index      int            //编号，排序用
}

//type FuncInfo struct {
//	Fn     reflect.Value
//	Params []FuncParam
//}
//
//type FuncParam struct {
//}
