package transformer

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type TransformContext struct {
	original_path string
	relative_path string
	output_path   string
	content       string

	file_info os.FileInfo

	// If a transform uses regex captures the next transform will
	// find possible captures here.
	LastRegex    *regexp.Regexp
	LastCaptures []string
}

func NewTransformContext(config *Config,
	wf *Workflow, path string, info os.FileInfo) (*TransformContext, error) {
	relpath, err := filepath.Rel(wf.SrcRepository, path)
	if err != nil {
		return nil, err
	}

	dest_path := filepath.Join(wf.DestRepository, relpath)
	return &TransformContext{
		original_path: path,
		relative_path: relpath,
		output_path:   dest_path,
		file_info:     info,
	}, nil
}

type TransformProcessor interface {
	Process(ctx *TransformContext) (bool, error)
}

func NewTransform(config *Config, wf *Workflow, t *Transform) (TransformProcessor, error) {
	switch t.Type {
	case "Read":
		return &ReadTransform{}, nil

	case "Write":
		return &WriteTransform{config, wf, t}, nil

	case "Select":
		return &SelectorTransform{config, t}, nil

	case "ModTime":
		return &ModTimeTransform{config, wf}, nil

	case "Rename":
		return &RenameTransform{config, wf, t}, nil

	case "Symlink":
		return &SymlinkTransform{config, wf}, nil

	case "Ignore":
		return &IgnoreTransform{config}, nil
	default:
		return nil, errors.New("Unknown transform " + t.Type)
	}
}

type IgnoreTransform struct {
	config *Config
}

func (self IgnoreTransform) Process(ctx *TransformContext) (bool, error) {
	NewLogger(self.config).Printf("Ignoring %v\n", ctx.original_path)
	return true, nil
}

type ReadTransform struct{}

func (self ReadTransform) Process(ctx *TransformContext) (bool, error) {
	// We can not read directories.
	if !ctx.file_info.Mode().IsRegular() {
		return false, nil
	}

	data, err := ioutil.ReadFile(ctx.original_path)
	if err != nil {
		return false, err
	}
	ctx.content = string(data)

	// Once we read the file, we need to keep going.
	return true, nil
}

type RenameTransform struct {
	config *Config
	wf     *Workflow
	t      *Transform
}

func (self RenameTransform) Process(ctx *TransformContext) (bool, error) {
	dest_path := ctx.output_path
	if self.t.RewriteDest != "" {
		dest_path = ctx.LastRegex.ReplaceAllString(ctx.relative_path, self.t.RewriteDest)
		dest_path = filepath.Join(self.wf.DestRepository, dest_path)
		ctx.output_path = dest_path
	}

	return true, nil
}

type WriteTransform struct {
	config *Config
	wf     *Workflow
	t      *Transform
}

func (self WriteTransform) Process(ctx *TransformContext) (bool, error) {
	dest_path := ctx.output_path

	NewLogger(self.config).Printf("Will write %v bytes to %v\n",
		len(ctx.content), dest_path)

	err := self.config.EnsureDirectory(dest_path)
	if err != nil {
		return false, err
	}

	fd, err := os.OpenFile(dest_path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return false, err
	}
	defer fd.Close()

	_, err = fd.Write([]byte(ctx.content))
	if err != nil {
		return false, err
	}

	return false, nil
}

type SymlinkTransform struct {
	config *Config
	wf     *Workflow
}

func (self SymlinkTransform) Process(ctx *TransformContext) (bool, error) {
	NewLogger(self.config).Printf("Will create a symlink from %v to %v\n",
		ctx.original_path, ctx.output_path)
	os.Symlink(ctx.original_path, ctx.output_path)
	return false, nil
}

type ModTimeTransform struct {
	config *Config
	wf     *Workflow
}

func (self ModTimeTransform) Process(ctx *TransformContext) (bool, error) {
	// Unable to stat the file is not an error - the file may not exist yet.
	stat, err := os.Lstat(ctx.output_path)
	if err != nil {
		return true, nil
	}

	// If the src file is before the dest file, we can skip it.
	if ctx.file_info.ModTime().Before(stat.ModTime()) {
		return false, nil
	}
	return true, nil
}

type SelectorTransform struct {
	config *Config
	t      *Transform
}

func matchAny(config *Config, ctx *TransformContext, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		inc, err := config.RegexpCompile(pattern)
		if err != nil {
			return false, err
		}
		ctx.LastRegex = inc
		ctx.LastCaptures = inc.FindStringSubmatch(ctx.relative_path)
		if len(ctx.LastCaptures) > 0 {
			return true, nil
		}
	}

	return false, nil
}

func (self SelectorTransform) Process(ctx *TransformContext) (bool, error) {
	if len(self.t.IncludeRegex) > 0 {
		included, err := matchAny(self.config, ctx, self.t.IncludeRegex)
		if err != nil {
			return false, err
		}

		if !included {
			return false, nil
		}
	}

	if len(self.t.ExcludeRegex) > 0 {
		excluded, err := matchAny(self.config, ctx, self.t.ExcludeRegex)
		if err != nil {
			return false, err
		}

		if excluded {
			return false, nil
		}
	}

	// If we get here, proceed with the next transform
	return true, nil
}
