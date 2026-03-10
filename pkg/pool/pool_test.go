package pool

import (
	"testing"

	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

func TestPool_GetPut(t *testing.T) {
	p := New(func() *model.ResettableStruct {
		return &model.ResettableStruct{}
	})

	// Get создаёт новый объект
	obj1 := p.Get()
	if obj1 == nil {
		t.Fatal("Get() returned nil")
	}
	obj1.IntVal = 42
	obj1.StrVal = "test"

	// Put возвращает объект в пул
	p.Put(obj1)

	// Get возвращает тот же объект (после Reset)
	obj2 := p.Get()
	if obj2 != obj1 {
		t.Error("expected same object from pool")
	}
	if obj2.IntVal != 0 || obj2.StrVal != "" {
		t.Errorf("Reset() was not called: IntVal=%d, StrVal=%q", obj2.IntVal, obj2.StrVal)
	}
}
