package squadron

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/foomo/squadron/util"
)

const (
	TagMap    = "!!map"
	TagString = "!!str"
)

type Build struct {
	Image      string   `yaml:"image,omitempty"`
	Tag        string   `yaml:"tag,omitempty"`
	Context    string   `yaml:"context,omitempty"`
	Dockerfile string   `yaml:"dockerfile,omitempty"`
	Args       []string `yaml:"args,omitempty"`
	Secrets    []string `yaml:"secrets,omitempty"`
	Labels     []string `yaml:"labels,omitempty"`
	CacheFrom  []string `yaml:"cache_from,omitempty"`
	Network    string   `yaml:"network,omitempty"`
	Target     string   `yaml:"target,omitempty"`
	ShmSize    string   `yaml:"shm_size,omitempty"`
	ExtraHosts []string `yaml:"extra_hosts,omitempty"`
	Isolation  string   `yaml:"isolation,omitempty"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Build ...
func (b *Build) Build(ctx context.Context) error {
	logrus.Infof("running docker build for %q", b.Context)
	_, err := util.NewDockerCommand().Build(b.Context).
		Arg("-t", fmt.Sprintf("%s:%s", b.Image, b.Tag)).
		Arg("--file", b.Dockerfile).
		ListArg("--build-arg", b.Args).
		ListArg("--secrets", b.Secrets).
		ListArg("--label", b.Labels).
		ListArg("--cache-from", b.CacheFrom).
		Arg("--network", b.Network).
		Arg("--target", b.Target).
		Arg("--shm-size", b.ShmSize).
		ListArg("--add-host", b.ExtraHosts).
		Arg("--isolation", b.Isolation).Run(ctx)
	return err
}

// Push ...
func (b *Build) Push(ctx context.Context) error {
	logrus.Infof("running docker push for %s:%s", b.Image, b.Tag)
	_, err := util.NewDockerCommand().Push(ctx, b.Image, b.Tag)
	return err
}

// UnmarshalYAML ...
func (b *Build) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == TagMap {
		type wrapper Build
		return value.Decode((*wrapper)(b))
	}
	if value.Tag == TagString {
		var vString string
		if err := value.Decode(&vString); err != nil {
			return err
		}
		b.Context = vString
	}
	return fmt.Errorf("unsupported node tag type for %T: %q", b, value.Tag)
}
