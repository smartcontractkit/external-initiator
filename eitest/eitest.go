package eitest

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

type closeable interface {
	Close() error
}

func MustClose(toClose closeable) {
	Must(toClose.Close())
}
