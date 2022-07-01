package utils

import (
	"github.com/helloh2o/lucky/log"
	"testing"
)

func TestNewAuthValidator(t *testing.T) {
	auValidator := NewAuthValidator()
	opRead := 0x0001
	opWrite := 0x0002
	opFix := 0x0004
	opUpdate := 0x0008
	opDel := 0x0010
	key := "ay"
	auValidator.AddAuthData(key, 31)
	log.Debug("opRead:%v", auValidator.Validate(key, opRead))
	log.Debug("opWrite:%v", auValidator.Validate(key, opWrite))
	log.Debug("opFix:%v", auValidator.Validate(key, opFix))
	log.Debug("opUpdate:%v", auValidator.Validate(key, opUpdate))
	log.Debug("opDel:%v", auValidator.Validate(key, opDel))
}
