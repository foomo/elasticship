package util

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/foomo/config-bob/builder"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func RelativePath(path, basePath string) string {
	return strings.Replace(path, basePath+"/", "", -1)
}

func join(value interface{}, sep string) (string, error) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return "", fmt.Errorf("Value must be a slice or array")
	}
	var l []string
	for i := 0; i < v.Len(); i++ {
		l = append(l, fmt.Sprint(v.Index(i)))
	}
	return strings.Join(l, sep), nil
}

func joinMap(value interface{}, mapSep, sep string) (string, error) {
	if mapSep == sep {
		return "", fmt.Errorf("Cannot use same separator %q for map entries and items", sep)
	}
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Map {
		return "", fmt.Errorf("Value must be a map")
	}
	var l []string
	for _, key := range v.MapKeys() {
		l = append(l, fmt.Sprintf("%v%v%v", key, mapSep, v.MapIndex(key)))
	}
	return strings.Join(l, sep), nil
}

func ExecuteTemplate(file string, templateVars interface{}) ([]byte, error) {
	tplFuncs := builder.TemplateFuncs
	tplFuncs["join"] = join
	tplFuncs["joinMap"] = joinMap

	templateBytes, errRead := ioutil.ReadFile(file)
	if errRead != nil {
		return nil, errRead
	}
	tpl, err := template.New("squadron").Option("missingkey=error").Funcs(tplFuncs).Parse(string(templateBytes))
	if err != nil {
		return nil, err
	}
	out := bytes.NewBuffer([]byte{})
	if err := tpl.Option("missingkey=error").Funcs(tplFuncs).Execute(out, templateVars); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

type Cmd struct {
	l *logrus.Entry
	// cmd           *exec.Cmd
	command       []string
	cwd           string
	env           []string
	stdin         io.Reader
	stdoutWriters []io.Writer
	stderrWriters []io.Writer
	wait          bool
	t             time.Duration
	preStartFunc  func() error
	postStartFunc func() error
	postEndFunc   func() error
}

func NewCommand(l *logrus.Entry, name string) *Cmd {
	return &Cmd{
		l:       l,
		command: []string{name},
		wait:    true,
		env:     os.Environ(),
	}
}

func (c Cmd) Base() *Cmd {
	c.command = []string{c.command[0]}
	return &c
}

func (c Cmd) Command() []string {
	return c.command
}

func (c *Cmd) Args(args ...string) *Cmd {
	c.command = append(c.command, args...)
	return c
}

func (c *Cmd) Cwd(path string) *Cmd {
	c.cwd = path
	return c
}

func (c *Cmd) Env(env ...string) *Cmd {
	c.env = append(c.env, env...)
	return c
}

func (c *Cmd) Stdin(r io.Reader) *Cmd {
	c.stdin = r
	return c
}

func (c *Cmd) Stdout(w io.Writer) *Cmd {
	if w == nil {
		w, _ = os.Open(os.DevNull)
	}
	c.stdoutWriters = append(c.stdoutWriters, w)
	return c
}

func (c *Cmd) Stderr(w io.Writer) *Cmd {
	if w == nil {
		w, _ = os.Open(os.DevNull)
	}
	c.stderrWriters = append(c.stderrWriters, w)
	return c
}

func (c *Cmd) Timeout(t time.Duration) *Cmd {
	c.t = t
	return c
}

func (c *Cmd) NoWait() *Cmd {
	c.wait = false
	return c
}

func (c *Cmd) PreStart(f func() error) *Cmd {
	c.preStartFunc = f
	return c
}

func (c *Cmd) PostStart(f func() error) *Cmd {
	c.postStartFunc = f
	return c
}

func (c *Cmd) PostEnd(f func() error) *Cmd {
	c.postEndFunc = f
	return c
}

func (c *Cmd) Run() (string, error) {
	cmd := exec.Command("/bin/bash", "-e", "-c", strings.Join(c.command, " "))
	cmd.Env = append(os.Environ(), c.env...)
	if c.cwd != "" {
		cmd.Dir = c.cwd
	}
	c.l.Tracef("executing %q", cmd.String())

	combinedBuf := new(bytes.Buffer)
	traceWriter := c.l.WriterLevel(logrus.TraceLevel)
	warnWriter := c.l.WriterLevel(logrus.WarnLevel)

	c.stdoutWriters = append(c.stdoutWriters, combinedBuf, traceWriter)
	c.stderrWriters = append(c.stderrWriters, combinedBuf, warnWriter)
	cmd.Stdout = io.MultiWriter(c.stdoutWriters...)
	cmd.Stderr = io.MultiWriter(c.stderrWriters...)

	if c.preStartFunc != nil {
		if err := c.preStartFunc(); err != nil {
			return "", err
		}
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if c.postStartFunc != nil {
		if err := c.postStartFunc(); err != nil {
			return "", err
		}
	}

	if c.wait {
		if c.t != 0 {
			timer := time.AfterFunc(c.t, func() {
				cmd.Process.Kill()
			})
			defer timer.Stop()
		}

		if err := cmd.Wait(); err != nil {
			if c.t == 0 {
				return "", err
			}
		}
		if c.postEndFunc != nil {
			if err := c.postEndFunc(); err != nil {
				return "", err
			}
		}
	}

	return combinedBuf.String(), nil
}

func GenerateYaml(path string, data interface{}) error {
	out, marshalErr := yaml.Marshal(data)
	if marshalErr != nil {
		return marshalErr
	}
	file, crateErr := os.Create(path)
	if crateErr != nil {
		return crateErr
	}
	defer file.Close()
	_, writeErr := file.Write(out)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func IsYaml(file string) bool {
	return StringInSlice(filepath.Ext(file), []string{".yml", ".yaml"})
}

func IsJson(file string) bool {
	return filepath.Ext(file) == ".json"
}

func ValidatePath(wd string, p *string) error {
	if !filepath.IsAbs(*p) {
		*p = path.Join(wd, *p)
	}
	absPath, err := filepath.Abs(*p)
	if err != nil {
		return err
	}
	_, err = os.Stat(absPath)
	if err != nil {
		return err
	}
	*p = absPath
	return nil
}
