package helpers

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"errors"
	"fmt"
)

// MustLoadFileInstance 根据文件生成cue实例
func MustLoadFileInstance(filepath string) cue.Value {
	cv, err := LoadFileInstance(filepath)
	if err != nil {
		panic(err)
	}

	return cv
}

func LoadFileInstance(filepath string) (cue.Value, error) {
	insts := load.Instances([]string{filepath}, nil)
	if len(insts) != 1 {
		return cue.Value{}, errors.New(fmt.Sprintf("load instance error:%s", filepath))
	}

	cc := cuecontext.New()
	cv := cc.BuildInstance(insts[0])
	if cv.Err() != nil {
		return cue.Value{}, cv.Err()
	}

	return cv, nil
}
