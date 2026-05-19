package internal

import "sync"

func NewMap[T any]() *Map[T] {
	return &Map[T]{datas: map[string]T{}}
}

type Map[T any] struct {
	mutex sync.Mutex
	datas map[string]T
}

func (this *Map[T]) Exist(key string) bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	_, exist := this.datas[key]

	return exist
}

func (this *Map[T]) Get(key string) T {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.datas[key]
}

func (this *Map[T]) GetAll() map[string]T {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.datas
}

func (this *Map[T]) Set(key string, data T) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.datas[key] = data
}

func (this *Map[T]) Remove(key string, finalFunc func(t T)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if _, exist := this.datas[key]; !exist {
		return
	}

	if finalFunc != nil {
		finalFunc(this.datas[key])
	}

	delete(this.datas, key)
}

func (this *Map[T]) RemoveAll(finalFunc func(t T)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for _, data := range this.datas {
		if finalFunc != nil {
			finalFunc(data)
		}
	}

	this.datas = map[string]T{}
}
