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
		Template string
	}
)

func NewTemplate(config TemplateConfig) (Template, error) {
	t := Template{}
	err := t.load(config.Template)
	return t, err
}

// создает замыкание для последующего использования,как
// valueOf := t.CreateTemplateParser([]string)
// _ = valueOf("host")
func (t Template) CreateTemplateParser(from []string) func(string) string {
	s := from
	return func(key string) string {
		return t.valueOf(key, s)
	}
}

// получаем позицию значения (индекс) по именованному ключу
// далее пытаемся получить значение из слайса
// т.к. слайс string и не может быть другим повзвращаем пустую строку в случае отсутсвия значения
func (t Template) valueOf(key string, from []string) string {
	pos := t.pos(key)
	// индекс не найден
	if pos == -1 {
		return ""
	}
	// существует ли такой индекс в слайсе
	if len(from) >= pos {
		return from[pos]
	}
	return ""
}

// определяем индекс по названию ключа
func (t Template) pos(key string) int {
	if pos, ok := t.configurationMap[t.key(key)]; ok {
		return pos
	}
	return -1
}

func (t *Template) load(template string) error {
	if !t.exist(template) {
		return errors.New("Файл конфигурации не найден")
	}

	if content, err := t.read(template); err == nil {
		t.configurationMap = t.parse(content)
		return nil
	}

	return errors.New("Не удалось загрузить шаблон")
}

func (t Template) key(key string) string {
	return key
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

// @see template.conf
func (t Template) parse(raw string) map[string]int {
	items := strings.Split(raw, constants.LOG_DELIM)
	tmp := make(map[string]int, len(items))

	// формируем мапу с ключами названиями переменных в качестве значений используем их очередность
	// в дальнейшем они будут индексами
	for pos, item := range items {
		rpl := strings.Replace(item, "$", "", -1)
		key := strings.TrimSpace(rpl)
		tmp[key] = pos
	}

	return tmp
}
