package frame

import (
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/robertkrimen/otto"
)

const (
	ExtJS = ".js"
)

const (
	RegisterFunc = "register"
)

type Frame struct {
	plugins   []*Plugin
	pluginDir string
}

func NewFrame(pluginDir string) *Frame {
}

func (frame *Frame) loadPlugins() error {
	files, err := ioutil.ReadDir(frame.pluginDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && path.Ext(file.Name() == ExtJS) {
			err = frame.loadPlugin(file.Name())
			if err != nil {
				// TODO warning
			}
		}
	}
}

func (frame *Frame) loadPlugin(file string) error {
	pluginCnt, err := readPlugin(frame.pluginDir, file)
	if err != nil {
		return err
	}
	vm := otto.New()
	_, err := vm.Run(pluginCnt)
	if err != nil {
		return err
	}
	register, err := vm.Call(RegisterFunc)
	if err != nil {
		return err
	}
	ph, _ := otto.ToValue(nil)
	hook, err := register.Call(ph)
	if err != nil {
		return err
	}
}

func readPlugin(dir, file string) ([]byte, error) {
	filepath := filepath.Join(dir, file)
	return ioutil.ReadFile(filepath)
}
