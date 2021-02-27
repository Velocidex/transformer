package transformer

import "time"

type Transformer struct {
	config *Config
}

func NewTransformer(config *Config) *Transformer {
	return &Transformer{config}
}

func (self *Transformer) Transform() error {
	now := time.Now()

	for _, workflow := range self.config.Workflows {
		wf, err := NewWorkflow(self.config, workflow)
		if err != nil {
			return err
		}

		err = wf.Process()
		if err != nil {
			return err
		}
	}

	NewLogger(self.config).Printf("Completed in %v\n", time.Now().Sub(now))
	return nil
}

func (self *Transformer) Watch(period int, name string) error {

	for {
		now := time.Now()

		for _, workflow := range self.config.Workflows {
			if workflow.Name != name {
				continue
			}

			wf, err := NewWorkflow(self.config, workflow)
			if err != nil {
				return err
			}

			err = wf.Process()
			if err != nil {
				return err
			}
		}

		NewLogger(self.config).Printf("Completed in %v\n", time.Now().Sub(now))

		time.Sleep(time.Duration(period) * time.Second)
	}
	return nil
}
