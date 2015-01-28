// sheet.go - music sheet storage for beep
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Sheet struct {
	Id       int
	Name     string
	Dir      string
	Notation string
	file     *os.File
}

func NewSheet(name, dir, notation string) *Sheet {
	s := &Sheet{
		Name:     name,
		Dir:      dir,
		Notation: notation,
	}
	return s
}

func (s *Sheet) Path() string {
	var filename string
	var name = s.Name
	if s.Id > 0 {
		name = fmt.Sprintf("%d-%s", s.Id, s.Name)
	}
	filename = filepath.Join(beepHomeDir(), "sheets", s.Dir, name)
	return filename
}

func (s *Sheet) Save() error {
	dir := filepath.Join(beepHomeDir(), "sheets", s.Dir)
	os.MkdirAll(dir, 0755)
	opt := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	file, err := os.OpenFile(s.Path(), opt, 0644)
	if err != nil {
		return err
	}
	s.file = file
	defer s.file.Close()
	fmt.Fprint(s.file, s.Notation)
	return nil
}

func (s *Sheet) Load() error {
	sheetId := stringNumber(strings.Split(s.Name, "-")[0])
	for _, sheet := range builtinMusic {
		if sheet.Id == sheetId {
			s.Id = sheet.Id
			s.Dir = sheet.Dir
			s.Notation = sheet.Notation
			return nil
		}
	}
	buf, err := ioutil.ReadFile(s.Path())
	if err != nil {
		return err
	}
	s.Notation = string(buf)
	return nil
}

func (s *Sheet) Exists() bool {
	filename := filepath.Join(beepHomeDir(), "sheets", s.Dir, s.Name)
	_, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return true
}

func (s *Sheet) Delete() error {
	err := os.Remove(s.Path())
	if err != nil {
		return err
	}
	return nil
}

func sheetSearch(keyword string) []string {
	var names []string
	keyword = strings.ToLower(keyword)
	root := filepath.Join(beepHomeDir(), "sheets") + string(os.PathSeparator)
	for _, sheet := range builtinMusic {
		name := strings.TrimPrefix(sheet.Path(), root)
		if len(keyword) == 0 || strings.Contains(strings.ToLower(name), keyword) {
			names = append(names, name)
		}
	}
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			name := strings.TrimPrefix(path, root)
			if len(keyword) == 0 || strings.Contains(strings.ToLower(name), keyword) {
				names = append(names, name)
			}
		}
		return nil
	}
	sort.Strings(names)
	filepath.Walk(root, walkFn)
	return names
}

var builtinMusic = []*Sheet{
	&Sheet{
		Id:   1,
		Name: "mozart-k33b-klavierstuck-in-f.txt",
		Dir:  "beep",
		Notation: `VP SA8 SR9
A9HRDE cc DScszs|DEc DQzDE[|cc DScszs|DEc DQz DE[|vv DSvcsc|DEvs ]v|cc DScszs|VN
A3HLDE [n z,    |cHRq HLz, |[n z,    |cHRq HLz,  |sl z,    |]m   pb|z, ]m    |

A9HRDE cz [c|ss DSsz]z|DEs] ps|DSsz][ z][p|DEpDQ[ [ || DERE] DS][p[ |VN
A3HLDE [n ov|]m [n    |  pb ic|  n,   lHRq|HLnc DQ[ || DEcHRq HLvHRw|

A9HRDS ][p[ ][p[|DE] DQp DEi|REc DScszs|cszs |cszs|DEcDQzDE[|REv DSvcsc|DEvs ]v|VN
A3HLDE bHRe HLvw|cHRq   HLic|[n  ]m    |z,   |]m  |zn   z,  |sl  [,    |z. DQp |

A9HRDE REc DScszs|DEcz [c|REs DSsz]z|DEs] ps|DSsz][ z][p|DE[DSitDQrRE|VN
A3HLDE z,  ]m    |[n   ov|]m  [n    |pb   ic|nz     sc  |DQn      [RE|
`,
	},
}
