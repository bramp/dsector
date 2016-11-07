// TODO This file is full of panics, because I don't know the correct way to deal with errors under Lua
package ufwb

import (
	"encoding/binary"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/layeh/gopher-luar"
	"github.com/yuin/gopher-lua"
	"io"
)

const luaInit = `` // TODO Put any Lua init script here.

type luaMapper Decoder

func (m *luaMapper) GetCurrentLogSrc() *luaLogger {
	return (*luaLogger)(nil)
}

func (m *luaMapper) GetCurrentResults() *luaResults {
	return (*luaResults)(m)
}

// TODO Move this somewhere else
// TODO Do we already have this somewhere?
func (e Endian) ByteOrder() binary.ByteOrder {
	switch e {
	case LittleEndian:
		return binary.LittleEndian
	case BigEndian:
		return binary.BigEndian
	}
	panic(fmt.Sprintf("unknown endian %q", e))
}

func FromByteOrder(bo binary.ByteOrder) Endian {
	switch bo {
	case binary.LittleEndian:
		return LittleEndian
	case binary.BigEndian:
		return BigEndian
	}
	return UnknownEndian
}

func (m *luaMapper) GetDynamicEndianness() Endian {
	d := (*Decoder)(m)
	return FromByteOrder(d.dynamicEndian)
}

func (m *luaMapper) SetDynamicEndianness(endian Endian) {
	d := (*Decoder)(m)
	d.dynamicEndian = endian.ByteOrder()
}

type luaResults Decoder

func (r *luaResults) GetLastResult() *luaResult {
	d := (*Decoder)(r)
	v, err := d.prev()
	if err != nil {
		panic(err)
	}
	return &luaResult{
		decoder: d,
		value:   v,
	}
}

func (r *luaResults) GetResultByName(name string) *luaResult {
	d := (*Decoder)(r)
	v, err := d.prevByName(name)
	if err != nil {
		panic(err)
	}
	return &luaResult{
		decoder: d,
		value:   v,
	}
}

type luaLogger struct{}

func (l *luaLogger) LogMessage(module string, messageId int, severity logrus.Level, message string) {
	LogAtLevel(severity, module, messageId, message)
}

func (l *luaLogger) LogMessageForced(module string, messageId int, severity logrus.Level, message string) {
	LogAtLevel(severity, module, messageId, message)
}

func (l *luaLogger) LogMessageHighlight(module string, messageId int, severity logrus.Level, message string) {
	LogAtLevel(severity, module, messageId, message)
}

type luaResult struct {
	decoder *Decoder `luar:"-"`
	value   *Value   `luar:"-"`
}

func (l *luaResult) GetValue() *luaValue {
	return &luaValue{
		file:  l.decoder.f,
		value: l.value,
	}
}

type luaValue struct {
	file  io.ReaderAt `luar:"-"`
	value *Value      `luar:"-"`
}

func (l *luaValue) GetName() string {
	return l.value.Element.Name()
}

func (l *luaValue) GetType() string {
	panic("TODO")
}

// NumberValue
func (l *luaValue) GetUnsignedNumber() uint64 {
	n := l.value.Element.(*Number)
	i, err := n.Uint(l.file, l.value)
	if err != nil {
		panic(err)
	}
	return i
}

func (l *luaValue) getSignedNumber() int64 {
	n := l.value.Element.(*Number)
	i, err := n.Int(l.file, l.value)
	if err != nil {
		panic(err)
	}
	return i
}

func registerSynalysisType(L *lua.LState) {
	synalysis := L.NewTable()
	L.SetGlobal("synalysis", synalysis)

	L.SetField(synalysis, "ENDIAN_BIG", lua.LNumber(BigEndian))
	L.SetField(synalysis, "ENDIAN_LITTLE", lua.LNumber(LittleEndian))

	L.SetField(synalysis, "SEVERITY_FATAL", lua.LNumber(logrus.FatalLevel))
	L.SetField(synalysis, "SEVERITY_ERROR", lua.LNumber(logrus.ErrorLevel))
	L.SetField(synalysis, "SEVERITY_WARN", lua.LNumber(logrus.WarnLevel))
	L.SetField(synalysis, "SEVERITY_INFO", lua.LNumber(logrus.InfoLevel))
	L.SetField(synalysis, "SEVERITY_DEBUG", lua.LNumber(logrus.DebugLevel))
}

func registerCurrentMapperType(L *lua.LState, decoder *Decoder) {
	L.SetGlobal("currentMapper", luar.New(L, (*luaMapper)(decoder)))
}

func registerDebug(L *lua.LState, decoder *Decoder) {
	if decoder.debugFunc != nil {
		L.SetGlobal("debug", luar.New(L, decoder.debugFunc))
	}
}

// registerPackages registers the built in Lua packages that we deem safe
// TODO Check if this is actually safe!!!
// TODO Harden Lua so it can't read file system, etc
func registerPackages(L *lua.LState) {
	// TODO Remove all of this, as no built in should be used.
	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		//{lua.LoadLibName, lua.OpenPackage},
		{lua.BaseLibName, lua.OpenBase},
		//{lua.TabLibName, lua.OpenTable},
	} {
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n))
		if err != nil {
			panic(err)
		}
	}
}

func luaState(d *Decoder) (*lua.LState, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	registerPackages(L)

	registerSynalysisType(L)
	registerCurrentMapperType(L, d)
	registerDebug(L, d)

	err := L.DoString(luaInit)
	if err != nil {
		// This is a bug in the library, not in the grammar!
		return nil, fmt.Errorf("lua init error: %s", err)
	}

	return L, nil
}

func (s *Script) RunLua(d *Decoder) error {
	// TODO Move the state into the Decoder
	// TODO We shouldn't be init Lua every time, but for development it is ok
	L, err := luaState(d)
	if err != nil {
		return err
	}
	defer L.Close()

	err = L.DoString(s.Text())
	if err != nil {
		return fmt.Errorf("lua error: %s", err)
	}

	return nil
}
