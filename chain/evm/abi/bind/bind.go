package bind

import (
	"bytes"
	"fmt"
	"go/format"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"web3_gui/chain/evm/abi"
)

var methodNormalizer = map[string]func(string) string{
	"go": abi.ToCamelCase,
}

// bindStructType is a set of type binders that convert Solidity tuple types to some supported
// programming language struct definition.
var bindStructType = map[string]func(kind abi.Type, structs map[string]*tmplStruct) string{
	"go": bindStructTypeGo,
}
var capitalise = abi.ToCamelCase

// bindStructTypeGo converts a Solidity tuple type to a Go one and records the mapping
// in the given map.
// Notably, this function will resolve and record nested struct recursively.
func bindStructTypeGo(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.TupleTy:
		// We compose a raw struct name and a canonical parameter expression
		// together here. The reason is before solidity v0.5.11, kind.TupleRawName
		// is empty, so we use canonical parameter expression to distinguish
		// different struct definition. From the consideration of backward
		// compatibility, we concat these two together so that if kind.TupleRawName
		// is not empty, it can have unique id.
		id := kind.TupleRawName + kind.String()
		if s, exist := structs[id]; exist {
			return s.Name
		}
		var (
			names  = make(map[string]bool)
			fields []*tmplField
		)
		for i, elem := range kind.TupleElems {
			name := capitalise(kind.TupleRawNames[i])
			name = abi.ResolveNameConflict(name, func(s string) bool { return names[s] })
			names[name] = true
			fields = append(fields, &tmplField{Type: bindStructTypeGo(*elem, structs), Name: name, SolKind: *elem})
		}
		name := kind.TupleRawName
		if name == "" {
			name = fmt.Sprintf("Struct%d", len(structs))
		}
		name = capitalise(name)

		structs[id] = &tmplStruct{
			Name:   name,
			Fields: fields,
		}
		return name
	case abi.ArrayTy:
		return fmt.Sprintf("[%d]", kind.Size) + bindStructTypeGo(*kind.Elem, structs)
	case abi.SliceTy:
		return "[]" + bindStructTypeGo(*kind.Elem, structs)
	default:
		return bindBasicTypeGo(kind)
	}
}

// bindBasicTypeGo converts basic solidity types(except array, slice and tuple) to Go ones.
func bindBasicTypeGo(kind abi.Type) string {
	switch kind.T {
	case abi.AddressTy:
		return "common.Address"
	case abi.IntTy, abi.UintTy:
		parts := regexp.MustCompile(`(u)?int([0-9]*)`).FindStringSubmatch(kind.String())
		switch parts[2] {
		case "8", "16", "32", "64":
			return fmt.Sprintf("%sint%s", parts[1], parts[2])
		}
		return "*big.Int"
	case abi.FixedBytesTy:
		return fmt.Sprintf("[%d]byte", kind.Size)
	case abi.BytesTy:
		return "[]byte"
	case abi.FunctionTy:
		return "[24]byte"
	default:
		// string, bool types
		return kind.String()
	}
}
func isKeyWord(arg string) bool {
	switch arg {
	case "break":
	case "case":
	case "chan":
	case "const":
	case "continue":
	case "default":
	case "defer":
	case "else":
	case "fallthrough":
	case "for":
	case "func":
	case "go":
	case "goto":
	case "if":
	case "import":
	case "interface":
	case "iota":
	case "map":
	case "make":
	case "new":
	case "package":
	case "range":
	case "return":
	case "select":
	case "struct":
	case "switch":
	case "type":
	case "var":
	default:
		return false
	}

	return true
}
func structured(args abi.Arguments) bool {
	if len(args) < 2 {
		return false
	}
	exists := make(map[string]bool)
	for _, out := range args {
		// If the name is anonymous, we can't organize into a struct
		if out.Name == "" {
			return false
		}
		// If the field name is empty when normalized or collides (var, Var, _var, _Var),
		// we can't organize into a struct
		field := capitalise(out.Name)
		if field == "" || exists[field] {
			return false
		}
		exists[field] = true
	}
	return true
}
func Bind(types, abis, bins []string, pkg string) (string, error) {
	var (
		//合约
		contracts = make(map[string]*tmplContract)
		//结构体
		structs = make(map[string]*tmplStruct)
	)
	hasEvent := false
	for i := 0; i < len(types); i++ {
		//获取abi操作对象
		evmABI, err := abi.JSON(strings.NewReader(abis[i]))
		if err != nil {
			return "", err
		}
		// Strip any whitespace from the JSON ABI
		strippedABI := strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, abis[i])
		var (
			//只读类型函数
			calls = make(map[string]*tmplMethod)
			//产生状态变化，需要通过交易调用的函数
			transacts = make(map[string]*tmplMethod)
			//合约中的事件
			events = make(map[string]*tmplEvent)
			//fallback函数，最多有一个
			fallback *tmplMethod
			//receive函数，最多有一个
			receive *tmplMethod
			//作用暂时未知
			callIdentifiers     = make(map[string]bool)
			transactIdentifiers = make(map[string]bool)
			eventIdentifiers    = make(map[string]bool)
		)
		//遍历接口构造函数的输入
		for _, input := range evmABI.Constructor.Inputs {
			if hasStruct(input.Type) {
				bindStructType["go"](input.Type, structs)
			}
		}
		//遍历接口的方法
		for _, original := range evmABI.Methods {
			//normailed 统一格式
			normalized := original
			//统一代码风格
			normalizedName := methodNormalizer["go"](original.Name)
			var identifiers = callIdentifiers
			//判断是否是只读类型函数
			if !original.IsConstant() {
				identifiers = transactIdentifiers
			}
			if identifiers[normalizedName] {
				return "", fmt.Errorf("duplicated identifier \"%s\"(normalized \"%s\"), use --alias for renaming", original.Name, normalizedName)
			}
			identifiers[normalizedName] = true
			//修改normalized函数的属性为统一风格后的
			normalized.Name = normalizedName
			normalized.Inputs = make([]abi.Argument, len(original.Inputs))
			copy(normalized.Inputs, original.Inputs)
			for j, input := range normalized.Inputs {
				//如果参数名是空，或者是关键词，那么默认给参数名设置成arg1....
				if input.Name == "" || isKeyWord(input.Name) {
					normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
				}
				if hasStruct(input.Type) {
					bindStructType["go"](input.Type, structs)
				}
			}
			normalized.Outputs = make([]abi.Argument, len(original.Outputs))
			copy(normalized.Outputs, original.Outputs)
			//遍历函数的输出
			for j, output := range normalized.Outputs {
				if output.Name != "" {
					normalized.Outputs[j].Name = capitalise(output.Name)
				}
				if hasStruct(output.Type) {
					bindStructType["go"](output.Type, structs)
				}
			}
			//将函数进行归类
			if original.IsConstant() {
				calls[original.Name] = &tmplMethod{Original: original, Normalized: normalized, Structured: structured(original.Outputs)}
			} else {
				transacts[original.Name] = &tmplMethod{Original: original, Normalized: normalized, Structured: structured(original.Outputs)}
			}
		}
		//遍历事件
		for _, original := range evmABI.Events {
			//如果事件是用anonymous声明的，则忽略
			if original.Anonymous {
				continue
			}
			normalized := original
			normalizedName := methodNormalizer["go"](original.Name)
			if eventIdentifiers[normalizedName] {
				return "", fmt.Errorf("duplicated identifier \"%s\"(normalized \"%s\"), use --alias for renaming", original.Name, normalizedName)
			}
			eventIdentifiers[normalizedName] = true
			normalized.Name = normalizedName
			used := make(map[string]bool)
			normalized.Inputs = make([]abi.Argument, len(original.Inputs))
			copy(normalized.Inputs, original.Inputs)
			for j, input := range normalized.Inputs {
				if input.Name == "" || isKeyWord(input.Name) {
					normalized.Inputs[j].Name = fmt.Sprintf("arg%d", j)
				}
				// Event is a bit special, we need to define event struct in binding,
				// ensure there is no camel-case-style name conflict.
				for index := 0; ; index++ {
					if !used[capitalise(normalized.Inputs[j].Name)] {
						used[capitalise(normalized.Inputs[j].Name)] = true
						break
					}
					normalized.Inputs[j].Name = fmt.Sprintf("%s%d", normalized.Inputs[j].Name, index)
				}
				if hasStruct(input.Type) {
					bindStructType["go"](input.Type, structs)
				}
			}
			// Append the event to the accumulator list
			events[original.Name] = &tmplEvent{Original: original, Normalized: normalized}
		}
		//如果合约中有fallback函数
		if evmABI.HasFallback() {
			fallback = &tmplMethod{Original: evmABI.Fallback}
		}
		//如果合约中有receive函数
		if evmABI.HasReceive() {
			receive = &tmplMethod{Original: evmABI.Receive}
		}
		if len(events) > 0 {
			hasEvent = true
		}
		contracts[types[i]] = &tmplContract{
			Type:        capitalise(types[i]),
			InputABI:    strings.ReplaceAll(strippedABI, "\"", "\\\""),
			InputBin:    strings.TrimPrefix(strings.TrimSpace(bins[i]), "0x"),
			Constructor: evmABI.Constructor,
			Calls:       calls,
			Transacts:   transacts,
			Fallback:    fallback,
			Receive:     receive,
			Events:      events,
			Libraries:   make(map[string]string),
		}

	}
	data := &tmplData{
		Package:   pkg,
		Contracts: contracts,
		Structs:   structs,
		HasEvent:  hasEvent,
	}
	buffer := new(bytes.Buffer)
	funcs := map[string]interface{}{
		"bindtype":      bindType["go"],
		"bindtopictype": bindTopicType["go"],
		"namedtype":     namedType["go"],
		"capitalise":    capitalise,
		"decapitalise":  decapitalise,
	}
	tmpl := template.Must(template.New("").Funcs(funcs).Parse(tmplSource["go"]))
	if err := tmpl.Execute(buffer, data); err != nil {
		return "", err
	}
	// For Go bindings pass the code through gofmt to clean it up
	code, err := format.Source(buffer.Bytes())
	if err != nil {
		//return "", fmt.Errorf("%v\n%s", err, buffer)
		return "", fmt.Errorf("%v", err)
	}
	return string(code), nil
	//if lang == LangGo {
	//	code, err := format.Source(buffer.Bytes())
	//	if err != nil {
	//		return "", fmt.Errorf("%v\n%s", err, buffer)
	//	}
	//	return string(code), nil
	//}
	//// For all others just return as is for now
	//return buffer.String(), nil
	//return "", nil
}
func decapitalise(input string) string {
	if len(input) == 0 {
		return input
	}

	goForm := abi.ToCamelCase(input)
	return strings.ToLower(goForm[:1]) + goForm[1:]
}

var namedType = map[string]func(string, abi.Type) string{
	"go": func(string, abi.Type) string { panic("this shouldn't be needed") },
	//LangJava: namedTypeJava,
}
var bindTopicType = map[string]func(kind abi.Type, structs map[string]*tmplStruct) string{
	"go": bindTopicTypeGo,
	//LangJava: bindTopicTypeJava,
}

// bindTopicTypeGo converts a Solidity topic type to a Go one. It is almost the same
// functionality as for simple types, but dynamic types get converted to hashes.
func bindTopicTypeGo(kind abi.Type, structs map[string]*tmplStruct) string {
	bound := bindTypeGo(kind, structs)

	// todo(rjl493456442) according solidity documentation, indexed event
	// parameters that are not value types i.e. arrays and structs are not
	// stored directly but instead a keccak256-hash of an encoding is stored.
	//
	// We only convert stringS and bytes to hash, still need to deal with
	// array(both fixed-size and dynamic-size) and struct.
	if bound == "string" || bound == "[]byte" {
		bound = "common.Hash"
	}
	return bound
}

var bindType = map[string]func(kind abi.Type, structs map[string]*tmplStruct) string{
	"go": bindTypeGo,
}

func bindTypeGo(kind abi.Type, structs map[string]*tmplStruct) string {
	switch kind.T {
	case abi.TupleTy:
		return structs[kind.TupleRawName+kind.String()].Name
	case abi.ArrayTy:
		return fmt.Sprintf("[%d]", kind.Size) + bindTypeGo(*kind.Elem, structs)
	case abi.SliceTy:
		return "[]" + bindTypeGo(*kind.Elem, structs)
	default:
		return bindBasicTypeGo(kind)
	}
}

// 判断是否是结构体，结构体切片，结构体数组
func hasStruct(t abi.Type) bool {
	switch t.T {
	case abi.SliceTy:
		return hasStruct(*t.Elem)
	case abi.ArrayTy:
		return hasStruct(*t.Elem)
	case abi.TupleTy:
		return true
	default:
		return false
	}
}
