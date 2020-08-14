package lib

import (
	"errors"
	"github.com/LimeHD/limehd-syslog-server/constants"
	"io/ioutil"
	"os"
	"strings"
)

type (
	Template struct {
		configurationMap map[string]int
	}

	TemplateConfig struct {
		config string
	}
)

func (t *Template) Load(config TemplateConfig) error {
	if !t.exist(config.config) {
		return errors.New("Файл конфигурации не найден")
	}

	if content, err := t.read(config.config); err == nil {
		t.configurationMap = t.parse(content)
	}

	return errors.New("Не удалось загрузить шаблон")
}

func (t Template) Value(key string) int {
	if pos, ok := t.configurationMap[t.key(key)]; ok {
		return pos
	}

	return -1
}

func (t Template) key(key string) string {
	return "$" + key
}

func (t Template) exist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (t Template) read(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (t Template) parse(raw string) map[string]int {
	items := strings.Split(raw, constants.LOG_DELIM)
	tmp := make(map[string]int, len(items))

	for pos, item := range items {
		tmp[item] = pos
	}

	return tmp
}
