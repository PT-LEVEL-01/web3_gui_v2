package engine

import (
	sjson "encoding/json"
	"fmt"
	"reflect"
	"web3_gui/utils"
)

type ParamValid struct {
	Must        bool           //是否必要参数
	Name        string         //参数名称
	ValueKind   reflect.Kind   //值的类型
	IsSlice     bool           //本参数是否数组
	CustomValid CustomValidFun //自定义验证方法
	Desc        string         //参数说明
	NewKey      string         //验证转换参数后的key
}

func NewParamValid_Mast_Panic(name string, valueKind reflect.Kind, desc string) ParamValid {
	return NewParamValid_UnsupportedTypePanic(true, name, valueKind, false, nil, "", desc)
}

func NewParamValid_Nomast_Panic(name string, valueKind reflect.Kind, desc string) ParamValid {
	return NewParamValid_UnsupportedTypePanic(false, name, valueKind, false, nil, "", desc)
}

func NewParamValid_Mast_CustomFun_Panic(name string, valueKind reflect.Kind, cf CustomValidFun, newkey, desc string) ParamValid {
	return NewParamValid_UnsupportedTypePanic(true, name, valueKind, false, cf, newkey, desc)
}

func NewParamValid_Nomast_CustomFun_Panic(name string, valueKind reflect.Kind, cf CustomValidFun, newkey, desc string) ParamValid {
	return NewParamValid_UnsupportedTypePanic(false, name, valueKind, false, cf, newkey, desc)
}

func NewParamValid_MastSlice_Panic(name string, valueKind reflect.Kind, desc string) ParamValid {
	return NewParamValid_UnsupportedTypePanic(true, name, valueKind, true, nil, "", desc)
}

func NewParamValid_NomastSlice_Panic(name string, valueKind reflect.Kind, desc string) ParamValid {
	return NewParamValid_UnsupportedTypePanic(false, name, valueKind, true, nil, "", desc)
}

func NewParamValid_MastSlice_CustomFun_Panic(name string, valueKind reflect.Kind, cf CustomValidFun, newkey, desc string) ParamValid {
	return NewParamValid_UnsupportedTypePanic(true, name, valueKind, true, cf, newkey, desc)
}

func NewParamValid_NomastSlice_CustomFun_Panic(name string, valueKind reflect.Kind, cf CustomValidFun, newkey, desc string) ParamValid {
	return NewParamValid_UnsupportedTypePanic(false, name, valueKind, true, cf, newkey, desc)
}

/*
创建参数类型，不支持的类型会panic
*/
func NewParamValid_UnsupportedTypePanic(must bool, name string, valueKind reflect.Kind, isSlice bool, cf CustomValidFun,
	newkey, desc string) ParamValid {
	if valueKind != reflect.String && valueKind != reflect.Int64 && valueKind != reflect.Uint64 && valueKind !=
		reflect.Float64 && valueKind != reflect.Bool {
		resultStr := fmt.Sprintf("Unsupported types:%d,want type:%d %d %d %d %d", valueKind, reflect.String,
			reflect.Int64, reflect.Uint64, reflect.Float64, reflect.Bool)
		panic(resultStr)
	}
	pv := ParamValid{
		Must:        must,      //是否必要参数
		Name:        name,      //参数名称
		ValueKind:   valueKind, //值的类型
		IsSlice:     isSlice,   //
		CustomValid: cf,        //自定义验证方法
		Desc:        desc,      //参数说明
		NewKey:      newkey,    //
	}
	return pv
}

/*
验证rpc传入的参数
json值只支持string、int64、float64这三种。
*/
func validParam(params *map[string]interface{}, paramValid *[]ParamValid) ([]reflect.Value, utils.ERROR) {
	paramValues := make([]reflect.Value, 0, len(*paramValid))
	if paramValid == nil {
		return nil, utils.NewErrorSuccess()
	}
	//判断必要参数
	ERR := CheckMastParam(params, paramValid)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//填充可选参数
	FullNomastParam(params, paramValid)
	//判断参数类型是否正确
	ERR = CheckParamType(params, paramValid)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//验证自定义参数
	ERR = CustomValid(params, paramValid)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//组装参数
	paramValues = append(paramValues, reflect.ValueOf(params))
	for _, one := range *paramValid {
		valueItr := (*params)[one.Name]
		if one.IsSlice {
			temp := valueItr.([]interface{})
			newSlices := BuildValueByKind(temp, one.ValueKind)
			paramValues = append(paramValues, newSlices)
		} else {
			paramValues = append(paramValues, reflect.ValueOf(valueItr))
		}
	}
	return paramValues, utils.NewErrorSuccess()
}

/*
验证必要参数
*/
func CheckMastParam(params *map[string]interface{}, paramValid *[]ParamValid) utils.ERROR {
	for _, one := range *paramValid {
		_, exist := (*params)[one.Name]
		if !exist {
			if one.Must {
				//参数名称未找到
				return utils.NewErrorBus(ERROR_code_rpc_param_not_found, one.Name)
			}
		}
	}
	return utils.NewErrorSuccess()
}

/*
填充可选参数
可选参数初始化
*/
func FullNomastParam(params *map[string]interface{}, paramValid *[]ParamValid) {
	for _, one := range *paramValid {
		if !one.Must {
			_, exist := (*params)[one.Name]
			if exist {
				continue
			}
			//可选参数初始化
			if one.IsSlice {
				switch one.ValueKind {
				case reflect.String:
					(*params)[one.Name] = []string{}
				case reflect.Int64:
					(*params)[one.Name] = []int64{}
				case reflect.Uint64:
					(*params)[one.Name] = []uint64{}
				case reflect.Float64:
					(*params)[one.Name] = []float64{}
				case reflect.Bool:
					(*params)[one.Name] = []bool{}
				}
			} else {
				switch one.ValueKind {
				case reflect.String:
					(*params)[one.Name] = ""
				case reflect.Int64:
					(*params)[one.Name] = int64(0)
				case reflect.Uint64:
					(*params)[one.Name] = uint64(0)
				case reflect.Float64:
					(*params)[one.Name] = float64(0)
				case reflect.Bool:
					(*params)[one.Name] = false
				}
			}
		}
	}
}

func CheckParamType(params *map[string]interface{}, paramValid *[]ParamValid) utils.ERROR {
	for _, one := range *paramValid {
		valueItr, _ := (*params)[one.Name]
		typeOf := reflect.TypeOf(valueItr)
		if one.IsSlice {
			if typeOf.Kind() != reflect.Slice {
				utils.Log.Error().Uint("类型错误", uint(typeOf.Kind())).Send()
				return utils.NewErrorBus(ERROR_code_rpc_param_type_fail, one.Name)
			}
			slice, ok := valueItr.([]interface{})
			if !ok {
				utils.Log.Error().Uint("类型错误", uint(typeOf.Elem().Kind())).Send()
				return utils.NewErrorBus(ERROR_code_rpc_param_type_fail, one.Name)
			}
			//整理类型
			for i, oneValue := range slice {
				itr, ERR := ConverNeedType(oneValue, one.ValueKind)
				if ERR.CheckFail() {
					return ERR
				}
				if itr != nil {
					slice[i] = itr
				}
			}
			//刷新类型
			(*params)[one.Name] = slice
			//验证类型
			for _, oneValue := range slice {
				if !TryConverType(oneValue, one.ValueKind) {
					utils.Log.Error().Uint("类型错误", uint(typeOf.Elem().Kind())).Send()
					return utils.NewErrorBus(ERROR_code_rpc_param_type_fail, one.Name)
				}
			}
		} else {
			//utils.Log.Info().Str("key", one.Name).Uint("kind", uint(typeOf.Kind())).Str("pkgPath",
			//	typeOf.PkgPath()).Str("name", typeOf.Name()).Send()
			//如果是整型，字符串方式表示
			//json处理数值类型存在精度问题，所以数值类型全部转成字符串类型
			itr, ERR := ConverNeedType(valueItr, one.ValueKind)
			if ERR.CheckFail() {
				return ERR
			}
			if itr != nil {
				//utils.Log.Info().Uint("刷新类型", uint(reflect.TypeOf(itr).Kind())).Send()
				(*params)[one.Name] = itr
				//刷新下类型
				valueItr = (*params)[one.Name]
				typeOf = reflect.TypeOf(valueItr)
			}
			//判断参数值类型是否正确
			if typeOf.Kind() != one.ValueKind {
				//参数值的类型错误
				utils.Log.Error().Uint("类型错误", uint(typeOf.Kind())).Send()
				return utils.NewErrorBus(ERROR_code_rpc_param_type_fail, one.Name)
			}
		}
	}
	return utils.NewErrorSuccess()
}

func CustomValid(params *map[string]interface{}, paramValid *[]ParamValid) utils.ERROR {
	for _, one := range *paramValid {
		if one.CustomValid == nil {
			continue
		}
		valueItr, _ := (*params)[one.Name]
		zero := false
		if one.IsSlice {
			valuesItr := valueItr.([]interface{})
			if len(valuesItr) == 0 {
				zero = true
			}
		} else {
			switch one.ValueKind {
			case reflect.Int64:
				value := valueItr.(int64)
				if value == 0 {
					zero = true
				}
			case reflect.Uint64:
				value := valueItr.(uint64)
				if value == 0 {
					zero = true
				}
			case reflect.Float64:
				value := valueItr.(float64)
				if value == 0 {
					zero = true
				}
			case reflect.String:
				value := valueItr.(string)
				if value == "" {
					zero = true
				}
			case reflect.Bool:
				value := valueItr.(bool)
				if value == false {
					zero = true
				}
			}
		}

		//非必要参数，并且是零值的时候，不做验证
		if one.Must || !zero {
			newValue, ERR := one.CustomValid(valueItr)
			if ERR.CheckFail() {
				//错误信息叠加
				ERR.Msg = one.Name + ":" + ERR.Msg
				return ERR
			}
			if one.NewKey != "" {
				(*params)[one.NewKey] = newValue
			}
		}
	}
	return utils.NewErrorSuccess()
}

func (this *ParamValid) ConverVO() ParamValidVO {
	pvvo := ParamValidVO{
		Must:      this.Must,
		Name:      this.Name,
		ValueKind: this.ValueKind,
		IsSlice:   this.IsSlice,
		Desc:      this.Desc, //参数说明
	}
	return pvvo
}

type ParamValidVO struct {
	Must      bool         //是否必要参数
	Name      string       //参数名称
	ValueKind reflect.Kind //值的类型
	IsSlice   bool         //本参数是否数组
	Desc      string       //参数说明
}

/*
尝试类型转换
@return    bool    是否成功。true=成功;false=失败;
*/
func TryConverType(itr interface{}, kind reflect.Kind) bool {
	switch kind {
	case reflect.Int64:
		_, ok := itr.(int64)
		return ok
	case reflect.Uint64:
		_, ok := itr.(uint64)
		return ok
	case reflect.Float64:
		_, ok := itr.(float64)
		return ok
	case reflect.Float32:
		_, ok := itr.(float32)
		return ok
	case reflect.String:
		_, ok := itr.(string)
		return ok
	case reflect.Bool:
		_, ok := itr.(bool)
		return ok
	}
	return true
}

/*
转换为需要的类型
*/
func ConverNeedType(valueItr interface{}, wantKind reflect.Kind) (interface{}, utils.ERROR) {
	typeOf := reflect.TypeOf(valueItr)
	//utils.Log.Info().Uint("kind", uint(typeOf.Kind())).Str("pkgPath",
	//	typeOf.PkgPath()).Str("name", typeOf.Name()).Bool("类型", typeOf.Kind() == reflect.String).
	//	Bool("PkgPath", typeOf.PkgPath() == "encoding/json").
	//	Bool("Name", typeOf.Name() == "Number").Send()
	var result interface{}
	if typeOf.Kind() == reflect.String && typeOf.PkgPath() == "encoding/json" && typeOf.Name() == "Number" {
		value, ok := valueItr.(sjson.Number)
		if !ok {
			//参数值的类型错误
			utils.Log.Error().Uint("类型错误", uint(typeOf.Kind())).Send()
			return nil, utils.NewErrorBus(ERROR_code_rpc_param_type_fail, "")
		}
		switch wantKind {
		case reflect.Int64:
			valueInt64, err := value.Int64()
			if err != nil {
				utils.Log.Error().Str("error", err.Error()).Send()
				return nil, utils.NewErrorSysSelf(err)
			}
			result = valueInt64
		case reflect.Uint64:
			valueInt64, err := value.Int64()
			if err != nil {
				utils.Log.Error().Str("error", err.Error()).Send()
				return nil, utils.NewErrorSysSelf(err)
			}
			result = uint64(valueInt64)
		case reflect.Float64:
			valueFloat64, err := value.Float64()
			if err != nil {
				utils.Log.Error().Str("error", err.Error()).Send()
				return nil, utils.NewErrorSysSelf(err)
			}
			result = valueFloat64
		case reflect.String:
			valueString := value.String()
			result = valueString
		}
		//utils.Log.Info().Interface("result", result).Send()
	} else {

	}
	return result, utils.NewErrorSuccess()
}

func BuildValueByKind(itr []interface{}, kind reflect.Kind) reflect.Value {
	switch kind {
	case reflect.Int64:
		slice := make([]int64, 0, len(itr))
		for _, v := range itr {
			slice = append(slice, v.(int64))
		}
		return reflect.ValueOf(slice)
	case reflect.Uint64:
		slice := make([]uint64, 0, len(itr))
		for _, v := range itr {
			slice = append(slice, v.(uint64))
		}
		return reflect.ValueOf(slice)
	case reflect.Float64:
		slice := make([]float64, 0, len(itr))
		for _, v := range itr {
			slice = append(slice, v.(float64))
		}
		return reflect.ValueOf(slice)
	case reflect.String:
		slice := make([]string, 0, len(itr))
		for _, v := range itr {
			slice = append(slice, v.(string))
		}
		return reflect.ValueOf(slice)
	case reflect.Bool:
		slice := make([]bool, 0, len(itr))
		for _, v := range itr {
			slice = append(slice, v.(bool))
		}
		return reflect.ValueOf(slice)
	}
	return reflect.ValueOf(itr)
}
