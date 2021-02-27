package transformer

import (
	"errors"
	"os"
	"path/filepath"
)

type WorkflowProcessor interface {
	Process() error
}

func NewWorkflow(config *Config, workflow *Workflow) (WorkflowProcessor, error) {
	switch workflow.Type {
	case "Copy":
		return &CopyProcessor{config: config, wf: workflow}, nil
	default:
		return nil, errors.New("Unknown workflow type " + workflow.Type)
	}
}

type CopyProcessor struct {
	config *Config
	wf     *Workflow
}

func (self *CopyProcessor) ProcessTransforms(ctx *TransformContext) error {
	for _, transform := range self.wf.Transforms {
		transform, err := NewTransform(
			self.config, self.wf, transform)
		if err != nil {
			return err
		}

		keep_going, err := transform.Process(ctx)
		if err != nil {
			return err
		}

		if !keep_going {
			break
		}
	}
	return nil
}

func (self *CopyProcessor) Process() error {
	if self.wf.Disabled {
		return nil
	}

	files := []*FileInfo{}
	if len(files) == 0 {
		err := filepath.Walk(self.wf.SrcRepository,
			func(path string, info os.FileInfo, err error) error {
				files = append(files, &FileInfo{
					base: info, path: path,
				})
				return nil
			})
		if err != nil {
			return err
		}
	}

	for _, info := range files {
		ctx, err := NewTransformContext(self.config, self.wf, info.path, info.base)
		if err != nil {
			return err
		}
		err = self.ProcessTransforms(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
