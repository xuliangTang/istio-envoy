package utils

import (
	"cuelang.org/go/cue"
	_ "embed"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/tidwall/gjson"
	"istio-envoy/mygateway/utils/helpers"
	"log"
)

const (
	xdsCueTpl = "mygateway/tpls/xds.cue"
)

type TplObj interface {
	proto.Message
}

type TplGenerator[T TplObj] struct {
	cuecv cue.Value
}

// go:embed xds.cue
// var xdstpl []byte

var xdsCV *cue.Value

// NewTplGenerator 生成xds模板的cue对象
func NewTplGenerator[T TplObj]() *TplGenerator[T] {
	if xdsCV == nil {
		// cc := cuecontext.New()
		// cv := cc.CompileBytes(xdstpl)
		// 如果cue使用了import，不能直接读取文件，需要使用下面这种方式获取
		cv := helpers.MustLoadFileInstance(xdsCueTpl)
		if cv.Err() != nil {
			log.Fatalln(cv.Err())
		}

		// 填充控制面系统配置
		cv = cv.FillPath(cue.ParsePath("sysconfig"), SysConfig)
		if cv.Err() != nil {
			log.Fatalln(cv.Err())
		}

		xdsCV = &cv
	}
	return &TplGenerator[T]{
		cuecv: *xdsCV,
	}
}

var ObjNotFound = fmt.Errorf("没有找到对应的对象")

// GetOutputs 获取渲染结果的数组对象
func (t *TplGenerator[T]) GetOutputs(input interface{}, objName string, f func() T) ([]T, error) {
	// 填充input
	inputCv := t.cuecv.Context().Encode(input)
	filledCv := t.cuecv.FillPath(cue.ParsePath("input"), inputCv)
	if filledCv.Err() != nil {
		return nil, filledCv.Err()
	}

	// 解析output
	b, err := filledCv.LookupPath(cue.ParsePath("output")).MarshalJSON()
	if err != nil {
		return nil, err
	}

	// 获取指定对象
	getObj := gjson.Get(string(b), objName)
	if !getObj.Exists() {
		return nil, ObjNotFound
	}

	// array情况有些复杂 目前没有找到直接反序列化切片的方法的方法
	var ret []T
	for _, r := range getObj.Array() {
		obj := f()
		err = jsonpb.UnmarshalString(r.String(), obj)
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj)
	}

	return ret, nil
}

// GetOutput 获取渲染结果的单个对象
func (t *TplGenerator[T]) GetOutput(input interface{}, objName string, obj T) error {
	// 填充input
	inputCv := t.cuecv.Context().Encode(input)
	filledCv := t.cuecv.FillPath(cue.ParsePath("input"), inputCv)
	if filledCv.Err() != nil {
		return filledCv.Err()
	}

	// 解析output
	b, err := filledCv.LookupPath(cue.ParsePath("output")).MarshalJSON()
	if err != nil {
		return err
	}

	// 获取指定对象
	getObj := gjson.Get(string(b), objName)
	if !getObj.Exists() {
		return ObjNotFound
	}

	return jsonpb.UnmarshalString(getObj.String(), obj)
}
