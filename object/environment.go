package object

func NewEnclosedEnvironment(outer *Environment) *Environment {
	// 블록문을 만나면 환경을 새로 만들어주고
	env := NewEnvironment()
	// 그 환경에 상위 환경을 저장함 (포인터 참조이기 때문에 성능상 문제는 없을듯?)
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	// 찾고 있는 name이 현재 환경에 존재하지 않고 상위 환경이 있는 경우 거슬러 올라가서 탐색 진행
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
