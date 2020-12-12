package paginator

import (
	"fmt"
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
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
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

	result.Data = ds
	result.Paginate = Paginate{
		p.Page,
		p.Size,
		count,
	}

	return &result
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
			where += " where " + filter.Field + " " + filter.Operation + " " + filter.Value.(string)
			continue
		}
		where += "	and " + filter.Field + " " + filter.Operation + " " + filter.Value.(string)
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
		order += " group by "
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
	p.DB.Raw("select count(*) FROM " + queries[1]).Scan(&count)

	result.Data = ds
	result.Paginate = Paginate{
		p.Page,
		p.Size,
		count,
	}

	return &result
}
