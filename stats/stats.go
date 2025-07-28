package stats

import (
	"bytes"
	"git.sophuwu.com/manweb/CFG"
	"git.sophuwu.com/manweb/logs"
	"github.com/asdine/storm/v3"
	"go.etcd.io/bbolt"
	"html/template"
	"path/filepath"
	"strings"
)

var DB *storm.DB

func OpenDB() {
	if !CFG.EnableStats || CFG.StatisticDB == "" {
		return
	}
	var err error
	DB, err = storm.Open(CFG.StatisticDB, storm.BoltOptions(0660, &bbolt.Options{
		Timeout: 1000 * 1000 * 1000, // 1 second
	}))
	logs.CheckFatal("failed to open statistics database", err)
}

func CloseDB() {
	if DB == nil {
		return
	}
	err := DB.Close()
	logs.CheckFatal("failed to close statistics database", err)
	DB = nil
}

type Stat struct {
	Query string `storm:"id,unique"` // The query string
	Count int
}

func count(query string) {
	s := &Stat{Query: query}
	err := DB.One("Query", query, s)
	if err != nil {
		s = &Stat{
			Query: query,
			Count: 1,
		}
		DB.Save(s)
		return
	}
	s.Count++
	DB.Update(s)
}

func Count(query string) {
	if !CFG.EnableStats || DB == nil {
		return
	}
	go count(query)
}

type Special struct {
	Page  string `storm:"id,unique"` // The page name
	Count int
}

func countSpecial(page string) {
	s := &Special{Page: page}
	err := DB.One("Page", page, s)
	if err != nil {
		s = &Special{
			Page:  page,
			Count: 1,
		}
		DB.Save(s)
		return
	}
	s.Count++
	DB.Update(s)
}

func SpecialCount(page string) {
	if !CFG.EnableStats || DB == nil {
		return
	}
	go countSpecial(page)
}

func GetStats() ([]*Stat, error) {
	var stats []*Stat
	err := DB.Select().OrderBy("Count").Reverse().Find(&stats)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

type SectionCount struct {
	Section string
	Count   int
	Unique  int
}

func addSection(sec *map[string]*SectionCount, name string, count int) {
	if name == "" {
		return
	}
	name = filepath.Ext(name)
	name = strings.TrimPrefix(name, ".")
	secName, ok := (*sec)[name]
	if !ok {
		(*sec)[name] = &SectionCount{
			Section: name,
			Count:   count,
			Unique:  1,
		}
		return
	}
	secName.Count += count
	secName.Unique++
	(*sec)[name] = secName
}

type html struct {
	TotalLoads    int
	TotalPages    int
	UniquePages   int
	Searches      int
	Errors        int
	TotalSections []SectionCount
	AllPages      []Special
	MaxLen        int
}

var T *template.Template

func Html() string {
	if !CFG.EnableStats || DB == nil {
		return ""
	}
	stats, err := GetStats()
	if err != nil || len(stats) == 0 {
		return ""
	}
	var ht html
	var Specialc Special
	_ = DB.One("Page", "Search", &Specialc)
	ht.Searches = Specialc.Count
	_ = DB.One("Page", "Error", &Specialc)
	ht.Errors = Specialc.Count

	sec := make(map[string]*SectionCount)
	Maxlen := 0
	for _, s := range stats {
		addSection(&sec, s.Query, s.Count)
		ht.AllPages = append(ht.AllPages, Special{Page: s.Query, Count: s.Count})
		if len(s.Query) > Maxlen {
			Maxlen = len(s.Query)
		}
	}
	for _, v := range sec {
		ht.TotalSections = append(ht.TotalSections, *v)
		ht.TotalPages += v.Count
		ht.UniquePages += v.Unique
	}
	for i := range ht.TotalSections {
		for j := i + 1; j < len(ht.TotalSections); j++ {
			if ht.TotalSections[i].Count > ht.TotalSections[j].Count {
				ht.TotalSections[i], ht.TotalSections[j] = ht.TotalSections[j], ht.TotalSections[i]
			}
		}
	}
	ht.TotalLoads = ht.TotalPages + ht.UniquePages + ht.Searches + ht.Errors
	var b bytes.Buffer
	_ = T.ExecuteTemplate(&b, "stats", &ht)
	return b.String()
}
