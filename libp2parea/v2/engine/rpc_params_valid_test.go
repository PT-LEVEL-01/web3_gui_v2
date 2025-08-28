package engine

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"web3_gui/utils"
)

func TestParamsValid(t *testing.T) {
	testCases := []struct {
		input   map[string]interface{}
		valid   ParamValid
		ERRCode uint64
	}{
		{map[string]interface{}{}, NewParamValid_UnsupportedTypePanic(true, "age", reflect.String, false,
			nil, "", "年龄，必填项"), ERROR_code_rpc_param_not_found},
		{map[string]interface{}{"age": "18"}, NewParamValid_UnsupportedTypePanic(true, "age", reflect.String, false,
			nil, "", "年龄，字符串格式"), utils.ERROR_CODE_success},
		{map[string]interface{}{"age": 18}, NewParamValid_UnsupportedTypePanic(true, "age", reflect.Int64, false,
			nil, "", "年龄"), utils.ERROR_CODE_success},
		{map[string]interface{}{"age": 19}, NewParamValid_UnsupportedTypePanic(true, "age", reflect.Int64, false,
			ValidAge18, "", "年龄"), utils.ERROR_CODE_system_error_self},
	}

	for i, tc := range testCases {
		//参数要用json处理下
		bs, err := json.Marshal(tc.input)
		if err != nil {
			t.Errorf("json marshal error:%s", err.Error())
		}
		//解析为map
		params := make(map[string]interface{})
		decoder := json.NewDecoder(bytes.NewBuffer(bs))
		decoder.UseNumber()
		err = decoder.Decode(&params)
		if err != nil {
			t.Errorf("json decode error:%s", err.Error())
		}

		valids := make([]ParamValid, 0, 1)
		valids = append(valids, tc.valid)
		_, ERR := validParam(&params, &valids)
		//utils.Log.Info().Uint64("err code", ERR.Code).Str("err msg", ERR.Msg).Send()
		if ERR.Code != tc.ERRCode {
			t.Errorf("index:%d validParam(%v,%v) = %v, expected %v", i, tc.input, tc.valid, ERR.String(), tc.ERRCode)
		}
	}
}

func ValidAge18(ageItr interface{}) (any, utils.ERROR) {
	age := ageItr.(int64)
	if age != 18 {
		return nil, utils.NewErrorSysSelf(errors.New("age must 18"))
	}
	return age, utils.NewErrorSuccess()
}
