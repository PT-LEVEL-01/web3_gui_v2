package bind

import "web3_gui/chain/evm/abi"

type tmplData struct {
	Package   string                   // Name of the package to place the generated file in
	Contracts map[string]*tmplContract // List of contracts to generate into this file
	Libraries map[string]string        // Map the bytecode's link pattern to the library name
	Structs   map[string]*tmplStruct   // Contract struct type definitions
	HasEvent  bool
}

// tmplContract contains the data needed to generate an individual contract binding.
type tmplContract struct {
	Type     string // Type name of the main contract binding
	InputABI string // JSON ABI used as the input to generate the binding from
	InputBin string // Optional EVM bytecode used to generate deploy code from
	//暂时用不到
	//FuncSigs    map[string]string      // Optional map: string signature -> 4-byte signature
	Constructor abi.Method             // Contract constructor for deploy parametrization
	Calls       map[string]*tmplMethod // Contract calls that only read state data
	Transacts   map[string]*tmplMethod // Contract calls that write state data
	Fallback    *tmplMethod            // Additional special fallback function
	Receive     *tmplMethod            // Additional special receive function
	Events      map[string]*tmplEvent  // Contract events accessors
	Libraries   map[string]string      // Same as tmplData, but filtered to only keep what the contract needs
	Library     bool                   // Indicator whether the contract is a library
}

// tmplMethod is a wrapper around an abi.Method that contains a few preprocessed
// and cached data fields.
type tmplMethod struct {
	Original   abi.Method // Original method as parsed by the abi package
	Normalized abi.Method // Normalized version of the parsed method (capitalized names, non-anonymous args/returns)
	Structured bool       // Whether the returns should be accumulated into a struct
}

// tmplEvent is a wrapper around an abi.Event that contains a few preprocessed
// and cached data fields.
type tmplEvent struct {
	Original   abi.Event // Original event as parsed by the abi package
	Normalized abi.Event // Normalized version of the parsed fields
}

// tmplField is a wrapper around a struct field with binding language
// struct type definition and relative filed name.
type tmplField struct {
	Type    string   // Field type representation depends on target binding language
	Name    string   // Field name converted from the raw user-defined field name
	SolKind abi.Type // Raw abi type information
}

// tmplStruct is a wrapper around an abi.tuple and contains an auto-generated
// struct name.
type tmplStruct struct {
	Name   string       // Auto-generated struct name(before solidity v0.5.11) or raw name.
	Fields []*tmplField // Struct fields definition depends on the binding language.
}

var tmplSource = map[string]string{
	"go": tmplSourceGo,
}

const tmplSourceGo = `
// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package {{.Package}}
{{if .HasEvent}}
import (
	"math/big"
	"strings"
	"errors"
	"time"
	"context"
	"web3_gui/libp2parea/adapter/engine"
	"github.com/google/uuid"
	"github.com/golang/protobuf/proto"
	"web3_gui/chain/protos/go_protos"
	"google.golang.org/grpc"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/abi/bind"
	"web3_gui/chain/evm/common"
	
)
{{else}}
import (
	"math/big"
	"strings"
	"errors"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/abi/bind"
	"web3_gui/chain/evm/common"
	
)
{{end}}


// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
)

{{$structs := .Structs}}
{{range $structs}}
	// {{.Name}} is an auto generated low-level Go binding around an user-defined struct.
	type {{.Name}} struct {
	{{range $field := .Fields}}
	{{$field.Name}} {{$field.Type}}{{end}}
	}
{{end}}

{{range $contract := .Contracts}}
	// {{.Type}}MetaData contains all meta data concerning the {{.Type}} contract.
	var {{.Type}}MetaData = &bind.MetaData{
		ABI: "{{.InputABI}}",
		{{if .InputBin -}}
		Bin: "0x{{.InputBin}}",
		{{end}}
	}
	// {{.Type}}ABI is the input ABI used to generate the binding from.
	// Deprecated: Use {{.Type}}MetaData.ABI instead.
	var {{.Type}}ABI = {{.Type}}MetaData.ABI

	{{if .InputBin}}
		// {{.Type}}Bin is the compiled bytecode used for deploying new contracts.
		// Deprecated: Use {{.Type}}MetaData.Bin instead.
		var {{.Type}}Bin = {{.Type}}MetaData.Bin

		func (_{{$contract.Type}} *{{$contract.Type}})Deploy{{.Type}}(opts *bind.TransactionOpts {{range .Constructor.Inputs}}, {{.Name}} {{bindtype .Type $structs}}{{end}}) (string,string,error) {
		  {{range $pattern, $name := .Libraries}}
			{{decapitalise $name}}Addr, _, _, _ := Deploy{{capitalise $name}}(auth, backend)
			{{$contract.Type}}Bin = strings.ReplaceAll({{$contract.Type}}Bin, "__${{$pattern}}$__", {{decapitalise $name}}Addr.String()[2:])
		  {{end}}
			return _{{$contract.Type}}.contract.DeployContract(opts,{{.Type}}Bin {{range .Constructor.Inputs}}, {{.Name}}{{end}})
		}
	{{end}}
	type {{.Type}} struct {
	  contract *bind.BoundContract
	}
	// New{{.Type}} creates a new instance of {{.Type}}, bound to a specific deployed contract.
	func New{{.Type}}(api string) (*{{.Type}}, error) {
	  contract, err := bind{{.Type}}(api)
	  if err != nil {
	    return nil, err
	  }
	  return &{{.Type}}{contract: contract}, nil
	}
	// bind{{.Type}} binds a generic wrapper to an already deployed contract.
	func bind{{.Type}}(api string) (*bind.BoundContract, error) {
	  parsed, err := abi.JSON(strings.NewReader({{.Type}}ABI))
	  if err != nil {
	    return nil, err
	  }
	  return bind.NewBoundContract(api, parsed), nil
	}
	{{range .Calls}}
		// {{.Normalized.Name}} is a free data retrieval call binding the contract method 0x{{printf "%x" .Original.ID}}.
		//
		// Solidity: {{.Original.String}}
		func (_{{$contract.Type}} *{{$contract.Type}}) {{.Normalized.Name}}(opts *bind.TransactionOpts {{range .Normalized.Inputs}}, {{.Name}} {{bindtype .Type $structs}} {{end}}) ({{if .Structured}}struct{ {{range .Normalized.Outputs}}{{.Name}} {{bindtype .Type $structs}};{{end}} },{{else}}{{range .Normalized.Outputs}}{{bindtype .Type $structs}},{{end}}{{end}} error) {
			var out []interface{}
			err := _{{$contract.Type}}.contract.Call(opts, &out, "{{.Original.Name}}" {{range .Normalized.Inputs}}, {{.Name}}{{end}})
			{{if .Structured}}
			outstruct := new(struct{ {{range .Normalized.Outputs}} {{.Name}} {{bindtype .Type $structs}}; {{end}} })
			if err != nil {
				return *outstruct, err
			}
			{{range $i, $t := .Normalized.Outputs}} 
			outstruct.{{.Name}} = *abi.ConvertType(out[{{$i}}], new({{bindtype .Type $structs}})).(*{{bindtype .Type $structs}}){{end}}

			return *outstruct, err
			{{else}}
			if err != nil {
				return {{range $i, $_ := .Normalized.Outputs}}*new({{bindtype .Type $structs}}), {{end}} err
			}
			{{range $i, $t := .Normalized.Outputs}}
			out{{$i}} := *abi.ConvertType(out[{{$i}}], new({{bindtype .Type $structs}})).(*{{bindtype .Type $structs}}){{end}}
			
			return {{range $i, $t := .Normalized.Outputs}}out{{$i}}, {{end}} err
			{{end}}
		}
	{{end}}

	{{range .Transacts}}
		// {{.Normalized.Name}} is a paid mutator sendRequest binding the contract method 0x{{printf "%x" .Original.ID}}.
		//
		// Solidity: {{.Original.String}}
		func (_{{$contract.Type}} *{{$contract.Type}}) {{.Normalized.Name}}(opts *bind.TransactionOpts {{range .Normalized.Inputs}}, {{.Name}} {{bindtype .Type $structs}} {{end}}) (string,error) {
			return _{{$contract.Type}}.contract.Transfer(opts, "{{.Original.Name}}" {{range .Normalized.Inputs}}, {{.Name}}{{end}})
		}
	{{end}}
	
	{{if .Events}}
	//合约事件部分
	type {{$contract.Type}}Subscribe struct {
		client go_protos.SubscriberClient
		abi abi.ABI
	}
	func New{{$contract.Type}}Subscribe(ctx context.Context,endpoins string) (*{{$contract.Type}}Subscribe,error) {
		conn,err:=grpc.DialContext(ctx,endpoins,grpc.WithInsecure())
		if err !=nil{
			return nil, err
		}
		client := go_protos.NewSubscriberClient(conn)
		parsed, err := abi.JSON(strings.NewReader({{.Type}}ABI))
		if err != nil {
			return nil, err
		}
		return &{{$contract.Type}}Subscribe{
			client: client,
			abi:    parsed,
		},nil
	}
	func (_{{$contract.Type}} *{{$contract.Type}}Subscribe)UnpackLog(out interface{}, event string, contractEventInfo *go_protos.ContractEventInfo) error {
		if contractEventInfo.Topic != common.Bytes2Hex(_{{$contract.Type}}.abi.Events[event].ID.Bytes()) {
			return errors.New("事件签名不符")
		}
		eventData := contractEventInfo.EventData
		logData := eventData[len(eventData)-1]
		//if log.Topics[0] != _{{$contract.Type}}.abi.Events[event].ID {
		//	return fmt.Errorf("event signature mismatch")
		//}
		if logData != "" {
			if err := _{{$contract.Type}}.abi.UnpackIntoInterface(out, event, common.Hex2Bytes(logData)); err != nil {
				return err
			}
		}
		var indexed abi.Arguments
		for _, arg := range _{{$contract.Type}}.abi.Events[event].Inputs {
			if arg.Indexed {
				indexed = append(indexed, arg)
			}
		}
		//去除掉第一个主题
		topics := []common.Hash{}
		for i := 0; i < len(eventData)-1; i++ {
			topics = append(topics, common.BytesToHash(common.Hex2Bytes(eventData[i])))
		}
		return abi.ParseTopics(out, indexed, topics)
	}
	//含有event事件
	type EventQuery struct {
		StartBlock   int64
		EndBlock     int64
		ContractAddr string
		Topic        string
	}
	//生成请求体
	func createPayload(query *EventQuery) ([]byte, error) {
		payload := &go_protos.SubscribeContractEventPayload{
			Topic:           query.Topic,
			ContractAddress: query.ContractAddr,
			StartBlock:      query.StartBlock,
			EndBlock:        query.EndBlock,
		}
		bt, err := proto.Marshal(payload)
		if err != nil {
			return nil, err
		}
		return bt, nil
	}
	func createReq(txId string, reqType go_protos.ReqType, payload []byte) *go_protos.SubscriberReq {
		header := &go_protos.ReqHeader{ReqType: reqType, TxId: txId, Timestamp: config.TimeNow().Unix(), ExpTime: 0}
		req := &go_protos.SubscriberReq{
			Header:  header,
			Payload: payload,
		}
		return req
	}
	{{end}}
	{{range .Events}}
	// {{$contract.Type}}{{.Normalized.Name}} represents a {{.Normalized.Name}} event raised by the {{$contract.Type}} contract.
		type LogEvent{{$contract.Type}}{{.Normalized.Name}} struct { {{range .Normalized.Inputs}}
			{{capitalise .Name}} {{if .Indexed}}{{bindtopictype .Type $structs}}{{else}}{{bindtype .Type $structs}}{{end}}; {{end}}
		}
	func (_{{$contract.Type}} *{{$contract.Type}}Subscribe) SubscribeEvent{{.Normalized.Name}}(ctx context.Context,query *EventQuery) (<-chan LogEvent{{$contract.Type}}{{.Normalized.Name}},error){
		//执行订阅
		payload,_:=createPayload(query)
		txid:=uuid.New().String()
		req:=createReq(txid,go_protos.ReqType_SUBSCRIBE_CONTRACT_EVENT_INFO, payload)
		resp, err :=_{{$contract.Type}}.client.EventSub(ctx,req)
		if err !=nil{
			return nil,err
		}
		c:=make(chan LogEvent{{$contract.Type}}{{.Normalized.Name}})
		go func() {
			defer close(c)
			for{
				select {
					case <-ctx.Done():
						return
				default:
					result, err := resp.Recv()
					if err !=nil{
						engine.Log.Error("")
						return
					}
					//解析数据
					events := &go_protos.ContractEventInfoList{}
					err = proto.Unmarshal(result.Data, events)
					if err != nil {
						engine.Log.Error("事件解析失败")
						return
					}
					for _, event := range events.ContractEvents {
						//
						log:=_{{$contract.Type}}.UnPack{{.Normalized.Name}}Log(event)
						c <- log
					}
					continue
				}
			}
		}()
		return c,nil
	}

	//解析日志
	func(_{{$contract.Type}} *{{$contract.Type}}Subscribe) UnPack{{.Normalized.Name}}Log(contractEventInfo *go_protos.ContractEventInfo) LogEvent{{$contract.Type}}{{.Normalized.Name}} {
		
		var log LogEvent{{$contract.Type}}{{.Normalized.Name}}
		_{{$contract.Type}}.UnpackLog(&log,"{{.Original.Name}}",contractEventInfo)
		return log
	}


	{{end}}
	{{if .Fallback}} 
		// Fallback is a paid mutator sendRequest binding the contract fallback function.
		//
		// Solidity: {{.Fallback.Original.String}}
		func (_{{$contract.Type}} *{{$contract.Type}}) Fallback(opts *bind.TransactionOpts, calldata []byte) (string,error) {
			return _{{$contract.Type}}.contract.RawTransact(opts, calldata)
		}

		func (_{{$contract.Type}} *{{$contract.Type}}) StaticFallback(opts *bind.TransactionOpts, calldata []byte) (string,error) {
			return _{{$contract.Type}}.contract.RawTransactStatic(opts, calldata)
		}
	{{end}}

	{{if .Receive}} 
		// Receive is a paid mutator sendRequest binding the contract receive function.
		//
		// Solidity: {{.Receive.Original.String}}
		func (_{{$contract.Type}} *{{$contract.Type}}) Receive(opts *bind.TransactionOpts){
			_{{$contract.Type}}.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
		}
	{{end}}

	func (_{{$contract.Type}} *{{$contract.Type}}) PreCall(opts *bind.TransactionOpts, calldata []byte) (uint64,error) {
		return _{{$contract.Type}}.contract.PreCall(opts, calldata)
	}
{{end}}
`
