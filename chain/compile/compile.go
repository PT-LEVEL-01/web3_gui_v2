package compile

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/keystore/adapter/crypto"
)

type compileDate struct {
	AbiJson string
}
type InputData struct {
	Method  string `json:"method"` //abi名字  setrecord
	ReqData []struct {
		Parameter string `json:"parameter"` //参数值 0111  880
		Index     int64  `json:"index"`     //参数的位置  1   4
		Type      string `json:"type"`      //参数的位置  1   4
	} `json:"req_data"`
	ReqAddress []struct {
		Address string `json:"address"` //地址  1xxx  2xxx
		Index   int64  `json:"index"`   //参数的位置2  3
	} `json:"req_address"` //json
}

func Compile(ContractName string, filePath string, Env string) (string, error) {
	os.RemoveAll(ContractName)
	//var result compileDate
	//reqVersion := Env[len(Env)-4:]
	//if reqVersion == ".exe" {
	//	reqVersion = Env[len(Env)-8 : len(Env)-4]
	//}
	//version, err := GetContractCodeVersion(filePath)
	//if strings.ReplaceAll(version, ".", "") != reqVersion || err != nil {
	//	return compileDate{}, errors.New(version + "version mismatch") //版本不对
	//}
	fmt.Println("filePath:", filePath)
	fmt.Println("Env:", Env)
	//编译源代码
	e := exec.Command(Env, "--combined-json", "abi,bin", "-o", ContractName, filePath)
	var stderr, stdout bytes.Buffer
	e.Stderr = &stderr
	e.Stdout = &stdout
	errs := e.Run()
	fmt.Println("exec:", errors.New(stderr.String()))
	if errs != nil {
		return "", errors.New(stderr.String()) //编译错误
	}
	//获取abi json
	result := GetContractAbiJson(ContractName + "/combined.json")
	//删除生成的abi文件
	err := os.RemoveAll(ContractName)
	if err != nil {
		//删除文件失败
		return "", err
	}
	//删除源代码文件
	//err = os.RemoveAll(filePath)
	//if err != nil {
	//	return compileDate{}, err
	//}
	fmt.Println("err:", err)
	return result, nil
}

// CompileInput 编译合约输出
func CompileInput(Abi string, inputData string) (string, error) {
	var reqData InputData
	var paramsData []interface{}
	abiObj, _ := abi.JSON(strings.NewReader(Abi))
	var err error
	var values uint64
	var c int
	var hexBs []byte

	err = json.Unmarshal([]byte(inputData), &reqData)
	if err != nil {
		return "", err
	}
	if len(reqData.ReqData) < 0 {
		return "", errors.New("reqData Cannot be empty")
	}
	params := make(map[int64]interface{}) //针对地址处理
	for i, _ := range reqData.ReqData {
		switch reqData.ReqData[i].Type {
		case "uint256":
			values, err = strconv.ParseUint(reqData.ReqData[i].Parameter, 10, 64)
			params[reqData.ReqData[i].Index] = big.NewInt(int64(values))
			break
		case "int256":
			values, err = strconv.ParseUint(reqData.ReqData[i].Parameter, 10, 64)
			params[reqData.ReqData[i].Index] = big.NewInt(int64(values))
			break
		case "uint64":
			values, err = strconv.ParseUint(reqData.ReqData[i].Parameter, 10, 64)
			params[reqData.ReqData[i].Index] = big.NewInt(int64(values))
			break
		case "int64":
			values, err = strconv.ParseUint(reqData.ReqData[i].Parameter, 10, 64)
			params[reqData.ReqData[i].Index] = big.NewInt(int64(values))
			break
		case "uint32":
			values, err = strconv.ParseUint(reqData.ReqData[i].Parameter, 10, 32)
			params[reqData.ReqData[i].Index] = big.NewInt(int64(values))
			break
		case "int32":
			values, err = strconv.ParseUint(reqData.ReqData[i].Parameter, 10, 32)
			params[reqData.ReqData[i].Index] = big.NewInt(int64(values))
			break
		case "uint8":
			values, err = strconv.ParseUint(reqData.ReqData[i].Parameter, 10, 8)
			params[reqData.ReqData[i].Index] = uint8(values)
			break
		case "string":
			params[reqData.ReqData[i].Index] = reqData.ReqData[i].Parameter
			break
		case "bool":
			params[reqData.ReqData[i].Index] = false
			if reqData.ReqData[i].Parameter == "true" {
				params[reqData.ReqData[i].Index] = true
			}
			break
		case "bytes32":
			c, err = strconv.Atoi(reqData.ReqData[i].Parameter)
			cd := strconv.FormatInt(int64(c), 16)
			cc := fmt.Sprintf("%064s", cd)
			hexBs, err = hex.DecodeString(cc)
			params[reqData.ReqData[i].Index] = NameHash2(hexBs)
			break
		case "bytes":
			params[reqData.ReqData[i].Index] = []byte(reqData.ReqData[i].Parameter)
			break
		case "bytes32[]":
			c, err = strconv.Atoi(reqData.ReqData[i].Parameter)
			cd := strconv.FormatInt(int64(c), 16)
			cc := fmt.Sprintf("%064s", cd)
			a := make([]common.Hash, 0)
			s := []byte{}
			strbs := []byte(cc)
			for k := range strbs {
				s = append(s, cc[k])
				if (k+1)%64 == 0 {
					v, _ := hex.DecodeString(string(s))
					a = append(a, NameHash1(v))
					s = []byte{}
				}
			}
			params[reqData.ReqData[i].Index] = a
			break
		case "tuple[]":

			// 解析 JSON 数据
			var tuples [][]float64
			err = json.Unmarshal([]byte(reqData.ReqData[i].Parameter), &tuples)
			if err == nil {
				var req [][]*big.Int
				for _, arr := range tuples {
					var tuple []*big.Int
					for _, element := range arr {
						tuple = append(tuple, big.NewInt(int64(element)))
					}
					req = append(req, tuple)
				}
				params[reqData.ReqData[i].Index] = req
			}

			break
		default:
			params[reqData.ReqData[i].Index] = reqData.ReqData[i].Parameter
			break
		}
		if err != nil {
			return "", err
		}
	}
	//如果存在需要处理的地址 那就格式化地址
	if len(reqData.ReqAddress) > 0 {
		for i, _ := range reqData.ReqAddress {
			address := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(reqData.ReqAddress[i].Address)))
			params[reqData.ReqAddress[i].Index] = address
		}
	}
	var keys []int64
	for k := range params {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	fmt.Println("params", params)
	fmt.Println("keys", keys)
	for _, v := range keys {
		fmt.Println(params[v])
		paramsData = append(paramsData, params[v])
	}

	fmt.Println("编译前Method数据:", reqData.Method)
	fmt.Println("params:", params)
	//fmt.Println("编译前abi参数:", paramsData[1])
	input, err := abiObj.Pack(reqData.Method, paramsData...)
	fmt.Println("编译后数据:", input)
	fmt.Println("编译错误:", err)
	fmt.Println("编译后数据:", common.Bytes2Hex(input))
	if err != nil {
		return "", err
	}
	return common.Bytes2Hex(input), nil
}

func bytesToUint8Arr(data []byte) []uint8 {
	uint8Array := make([]uint8, len(data))
	for i, b := range data {
		uint8Array[i] = uint8(b)
	}
	return uint8Array
}

// GetContractCodeVersion 获取合约源代码版本号
func GetContractCodeVersion(file string) (string, error) {
	i, err := os.Open(file)
	if err != nil {
		return "", errors.New("file does not exist")
	}
	defer i.Close()
	h := bufio.NewReader(i)
	r := make([]rune, 0)
	for {
		l, err := h.ReadString('\n')

		for k := 0; k < len(l); k++ {
			if l[k] == '^' {
				kk := k + 1
				for {
					if l[kk] == ';' {
						break
					}
					r = append(r, rune(l[kk]))
					kk++
				}
				if len(r) != 0 {
					break
				}
			}
		}
		if len(r) != 0 {
			break
		}
		if err == io.EOF {
			return "", err
		}
	}
	return string(r), nil
}

// GetContractAbiJson 获取ABI
func GetContractAbiJson(file string) string {
	i, err := os.Open(file)
	fmt.Println("GetContractAbiJson:", err)
	if err != nil {
		return ""
	}
	defer i.Close()
	h := bufio.NewReader(i)
	r := make([]rune, 0)
	for {
		l, _ := h.ReadString('\n')
		return l
	}
	return string(r)
}

type OutData struct {
	OutputData []*ReqData `json:"output_data"`
}
type ReqData struct {
	Name   any `json:"name"`
	Type   any `json:"type"`
	Value  any `json:"value"`
	Method any `json:"method"`
}

//解译evm返回的comment

func Analysis(Abi string, method string, comment string) (string, error) {
	resolverAbi, err := abi.JSON(strings.NewReader(Abi))
	if err != nil {
		return "", err
	}
	comments, err := hex.DecodeString(comment)
	if err != nil {
		return "", err
	}
	out, err := resolverAbi.Unpack(method, comments)
	if err != nil {
		return "", err
	}
	Req := make([]*ReqData, 0, len(resolverAbi.Methods[method].Outputs))
	for i := 0; i < len(resolverAbi.Methods[method].Outputs); i++ {
		data := &ReqData{
			Name:   resolverAbi.Methods[method].Outputs[i].Name,
			Type:   resolverAbi.Methods[method].Outputs[i].Type.String(),
			Value:  out[i],
			Method: method,
		}
		Req = append(Req, data)
	}
	reqJson, err := json.Marshal(Req)
	if err != nil {
		return "", err
	}
	return string(reqJson), nil
}
func NameHash1(inputName []byte) common.Hash {
	node := common.Hash{}
	copy(node[:], inputName)
	return node
}
func NameHash2(inputName []byte) common.Hash {
	fmt.Println("dddd ", len(inputName), inputName)
	node := common.Hash{}

	for i := len(inputName) - 1; i >= 0; i-- {
		node[i] = inputName[i]
	}

	return node
}
