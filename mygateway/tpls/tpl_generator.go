package tpls

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	_ "embed"
	"fmt"
	_ "github.com/envoyproxy/go-control-plane/pkg/cache/v3" // 要加这个引入，否则下面jsonpb.unmarshal会报错
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/tidwall/gjson"
	"log"
)

type TplObj interface {
	proto.Message
}

type TplGenerator[T TplObj] struct {
	cuecv cue.Value
}

//go:embed xds_test.cue
var xdstpl []byte

// NewTplGenerator 生成xds模板的cue对象
func NewTplGenerator[T TplObj]() *TplGenerator[T] {
	cc := cuecontext.New()
	cv := cc.CompileBytes(xdstpl)
	if cv.Err() != nil {
		log.Fatalln(cv.Err())
	}
	return &TplGenerator[T]{
		cuecv: cv,
	}
}

var ObjNotFound = fmt.Errorf("没有找到对应的对象")

// GetOutput 获取渲染结果的指定对象
func (t *TplGenerator[T]) GetOutput(input, objName string, isarray bool, obj T) error {
	// 填充input
	inputCv := t.cuecv.Context().CompileString(input)
	filledCv := t.cuecv.FillPath(cue.ParsePath("input"), inputCv)
	if filledCv.Err() != nil {
		return filledCv.Err()
	}

	// 解析output
	b, err := t.cuecv.LookupPath(cue.ParsePath("output")).MarshalJSON()
	if err != nil {
		return err
	}

	// 获取指定对象
	getObj := gjson.Get(string(b), objName)
	if !getObj.Exists() {
		return ObjNotFound
	}

	if !isarray {
		err = jsonpb.UnmarshalString(getObj.String(), obj)
		if err != nil {
			return err
		}
	}
	return nil
}
