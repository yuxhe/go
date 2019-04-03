package mongoutil

import (
	"bufio"
	"io"
	"os"
	"strings"
)

/*
 * Config
 */
type Config struct {
	Mymap map[string]string
}

type Db struct {
	//URL        string `mongodb://luyoutcore:luyoutcore123@192.168.1.13:21001,192.168.1.14:21001/luyoutcore`
	URL      string "datebase_luyoutcore"
	DataBase string "luyoutcore" //连接的数据库
}

func (c *Config) InitConfig(path string) {
	c.Mymap = make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		s := strings.TrimSpace(string(b))
		if strings.HasPrefix(s, "#") {
			continue
		}

		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}

		frist := strings.TrimSpace(s[:index])
		if len(frist) == 0 {
			continue
		}
		second := strings.TrimSpace(s[index+1:])
		if len(second) == 0 {
			continue
		}

		key := frist
		c.Mymap[key] = strings.TrimSpace(second)
	}
}

func (c Config) Get(key string) string {
	v, found := c.Mymap[key]
	if !found {
		return ""
	}
	return v
}
