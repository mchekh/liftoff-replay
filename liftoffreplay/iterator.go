package liftoffreplay

type Iterator[T any] interface {
	Next() (T, error)
}

type MapIterator[A any, B any] struct {
	src Iterator[A]
	f   func(A) B
}

func NewMapIterator[A any, B any](src Iterator[A], f func(A) B) *MapIterator[A, B] {
	return &MapIterator[A, B]{src: src, f: f}
}

func (it *MapIterator[A, B]) Next() (B, error) {
	a, err := it.src.Next()
	if err != nil {
		var zero B
		return zero, err
	}
	return it.f(a), nil
}
