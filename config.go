package transformer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Velocidex/yaml/v2"
	"github.com/pkg/errors"
)

type FileInfo struct {
	base os.FileInfo
	path string
}

type Transform struct {
	Type         string   `json:"type"`
	IncludeRegex []string `json:"include_regex"`
	ExcludeRegex []string `json:"exclude_regex"`
	RewriteDest  string   `json:"rewrite_dest"`
}

type Workflow struct {
	Type        string       `json:"type"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Disabled    bool         `json:"disabled"`
	Transforms  []*Transform `json:"transforms"`

	SrcRepository  string `json:"src_repository"`
	DestRepository string `json:"dest_repository"`
}

type Config struct {
	Verbose   bool        `json:"verbose"`
	Workflows []*Workflow `json:"workflows"`

	// A global cache of created directories.
	dircache map[string]bool

	files       []*FileInfo
	regex_cache map[string]*regexp.Regexp
}

func (self *Config) RegexpCompile(pattern string) (*regexp.Regexp, error) {
	r, pres := self.regex_cache[pattern]
	if pres {
		return r, nil
	}

	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	self.regex_cache[pattern] = r
	return r, nil
}

func (self *Config) EnsureDirectory(dest_path string) error {
	// Ensure the directories exist but cache them for faster
	// access.
	dirname := filepath.Dir(dest_path)
	_, pres := self.dircache[dirname]
	if !pres {
		err := os.MkdirAll(dirname, 0700)
		if err != nil {
			return err
		}
		self.dircache[dirname] = true
	}
	return nil
}

func LoadConfig(filename string) (*Config, error) {
	result := &Config{}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = yaml.UnmarshalStrict(data, result)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result.dircache = make(map[string]bool)
	result.regex_cache = make(map[string]*regexp.Regexp)
	return result, nil
}
