package utils

import (
	"github.com/helloh2o/lucky/log"
	"testing"
)

func TestNewAuthValidator(t *testing.T) {
	// &与，|或，^异或
	auValidator := NewAuthValidator()
	opRead := 1 << 1   //000010 2
	opWrite := 1 << 2  //0000100 4
	opFix := 1 << 3    //0001000 8
	opUpdate := 1 << 4 //0010000 16
	opDel := 1 << 5    //0100000 32
	key := "ay"
	auValidator.AddAuthData(key, 48)
	log.Debug("opRead:%v, val:%d", auValidator.Validate(key, opRead), opRead)
	log.Debug("opWrite:%v, val:%d", auValidator.Validate(key, opWrite), opWrite)
	log.Debug("opFix:%v, val:%d", auValidator.Validate(key, opFix), opFix)
	log.Debug("opUpdate:%v, val:%d", auValidator.Validate(key, opUpdate), opUpdate)
	log.Debug("opDel:%v, val:%d", auValidator.Validate(key, opDel), opDel)
}
