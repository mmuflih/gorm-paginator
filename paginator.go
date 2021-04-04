package paginator

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

/**
 * Created by Muhammad Muflih Kholidin
 * at 2020-09-29 00:37:55
 * https://github.com/mmuflih
 * muflic.24@gmail.com
 **/

func gormPaginate(page, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}

		if size <= 0 {
			size = 100
		}

		offset := (page - 1) * size
		return db.Offset(offset).Limit(size)
	}
}

type Config struct {
	DB      *gorm.DB
	Page    int
	Size    int
	OrderBy []string
	GroupBy []string
	Filters []Filter
	ShowSQL bool
}

type Filter struct {
	Field     string
	Operation string
	Value     interface{}
}

type Paginator struct {
	Data     interface{} `json:"data"`
	Paginate Paginate    `json:"paginate"`
}

type Paginate struct {
	Page      int   `json:"page"`
	Size      int   `json:"size"`
	Total     int64 `json:"total"`
	PageCount int   `json:"page_count"`
}

func Make(p *Config, ds interface{}) *Paginator {
	db := p.DB

	if p.ShowSQL {
		db = db.Debug()
	}

	if len(p.OrderBy) > 0 {
		for _, o := range p.OrderBy {
			db = db.Order(o)
		}
	}

	var result Paginator
	var count int64

	for _, filter := range p.Filters {
		if filter.Operation == "raw" {
			db.Where(filter.Value)
			continue
		}
		if filter.Value == nil {
			db.Where(filter.Field + " is null")
			continue
		} else {
			db.Where(filter.Field+" "+filter.Operation+" ?", filter.Value)
			continue
		}
	}

	db.Model(ds).Count(&count)
	db.Scopes(gormPaginate(p.Page, p.Size)).Find(ds)

	pageCount := math.Floor(float64(count / int64(p.Size)))
	if count%int64(p.Size) > 0 {
		pageCount++
	}
	result.Data = ds
	result.Paginate = Paginate{
		p.Page,
		p.Size,
		count,
		int(pageCount),
	}

	return &result
}

func (f Filter) generateFilterRaw() string {
	if f.Operation == "raw" {
		return f.Value.(string)
	}
	if f.Value == nil {
		return f.Field + " is null"
	} else {
		return f.Field + " " + f.Operation + " " + f.getValue()
	}
}

func (f Filter) getValue() string {
	v := reflect.ValueOf(f.Value)
	switch v.Type().Name() {
	case "int":
		return strconv.Itoa(f.Value.(int))
	case "string":
		return f.Value.(string)
	}
	return ""
}

func MakeRaw(query string, p *Config, ds interface{}) *Paginator {
	var result Paginator
	var count int64

	if p.Page <= 0 {
		p.Page = 1
	}

	if p.Size <= 0 {
		p.Size = 100
	}

	var where string

	for id, filter := range p.Filters {
		if id == 0 {
			where += " where " + filter.generateFilterRaw()
			continue
		}
		where += "	and " + filter.generateFilterRaw()
	}
	query += where
	limitOffset := fmt.Sprintf(" limit %d offset %d ", p.Size, (p.Page-1)*p.Size)

	order := ""
	if len(p.OrderBy) > 0 {
		order += " order by "
		for _, o := range p.OrderBy {
			order += o + ","
		}
		order = order[:len(order)-1]
	}

	group := ""
	if len(p.GroupBy) > 0 {
		group += " group by "
		for _, g := range p.GroupBy {
			group += g + ","
		}
		group = group[:len(group)-1]
	}

	err := p.DB.Raw(query + group + order + limitOffset).Scan(ds).Error
	if err != nil {
		fmt.Println("ERROR Paginator RAW", err)
	}
	queries := strings.Split(query, "from")
	nextStatement := ""
	for k, query := range queries {
		if k == 0 {
			continue
		}
		if k == 1 {
			nextStatement += query
			continue
		}
		nextStatement += " from " + query
	}
	nQuery := "select count(*) FROM " + nextStatement
	p.DB.Raw(nQuery).Scan(&count)

	pageCount := math.Floor(float64(count / int64(p.Size)))
	if count%int64(p.Size) > 0 {
		pageCount++
	}
	result.Data = ds
	result.Paginate = Paginate{
		p.Page,
		p.Size,
		count,
		int(pageCount),
	}

	return &result
}
