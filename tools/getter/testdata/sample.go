package testdata

type Struct struct{}

type Enum int

const (
	Zero Enum = iota
	One
	Two
	Three
)

type Embedded struct {
	ad int
	ae string
}

type Sample struct {
	Embedded

	a uint8
	b uint16
	c uint32
	d uint64

	e int8
	f int16
	g int32
	h int64

	i float32
	j float64

	k complex64
	l complex128

	m byte
	n rune

	o uint
	p int
	q uintptr

	r string

	s [32]byte
	t []int

	u *Struct
	v Struct

	w func()

	x interface{}

	y map[int]int

	z chan int

	// TODO Make this work
	//aa struct {
	//	a int
	//}
	//ab *struct{
	//	a int
	//}

	ac Enum

	extends *Sample
	//parent *Sample
}

func test() {
	s := Sample{}
	s.a = 0
	s.b = 0
	s.c = 0
	s.d = 0
	s.e = 0
	s.f = 0
	s.g = 0
	s.h = 0
	s.i = 0
	s.j = 0
	s.k = 0
	s.l = 0
	s.m = 0
	s.n = 0
	s.o = 0
	s.p = 0

}
