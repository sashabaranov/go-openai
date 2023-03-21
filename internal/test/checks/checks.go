package checks

import (
	"testing"
)

//func NoError(t *testing.T, err error, message ...string) {
//
//	pc, file, lineNo, ok := runtime.Caller(1)
//	if !ok {
//		return
//	}
//	funcName := runtime.FuncForPC(pc).Name()
//	fileName := path.Base(file) // The Base function returns the last element of the path
//
//	loc := fmt.Sprintf("FuncName:%s, file:%s, line:%d ", funcName, fileName, lineNo)
//
//	if err == nil {
//		t.Errorf(loc, message)
//	}
//}

//func NoError(t *testing.T, err error, message ...string) {
//	if err != nil {
//		t.Error(message)
//	}
//}

func NoError(t *testing.T, err error, message ...string) {
	t.Helper()
	if err == nil {
		t.Error(err, message)
	}
}
