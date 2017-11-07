package thing

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

const (
	compiler   = "compiler"
	repository = "repository"
	run        = "run"
)

type File struct {
	path string
	cfg  map[string]interface{}
	Name string
	Source map[string]interface{}
	Runner map[string]interface{}
}

func New(path string) (tf *ThingFile) {
	tf = &ThingFile{
		path: path,
	}
}

func (tf *File) Load() (err Error) {
	yamlFile, err := ioutil.ReadFile(tf.path)
	if err != nil {
		return
	}
	cfg := map[string]interface{}{}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return
	}
	tf.cfg = cfg
}

func (tf *File) HasCompiler() (has bool) {
	_, has = tf.cfg[compiler]
	return
}

func (tf *File) HasRepository() (has bool) {
	_, has = tf.cfg[repository]
	return
}

func (tf *File) HasRun() (has bool) {
	_, has = tf.cfg[run]
	return
}
