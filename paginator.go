package pag

import (
	"fmt"
	"math"
	"strings"

	"github.com/mmuflih/golib/filter"
	"gorm.io/gorm"
)

/**
 * Created by Muhammad Muflih Kholidin
 * at 2020-09-29 00:37:55
 * https://github.com/mmuflih
 * muflic.24@gmail.com
 **/

type Config struct {
	DB      *gorm.DB
	Page    int
	Size    int
	OrderBy []string
	GroupBy []string
	Filters filter.Where
	ShowSQL bool
}

type PaginatorSvc struct {
	Data      interface{} `json:"data"`
	Page      int         `json:"page"`
	Size      int         `json:"size"`
	Total     int64       `json:"total"`
	PageCount int         `json:"page_count"`
}

type Paginator struct {
	Data     interface{} `json:"data"`
	Paginate Paginate    `json:"paginate"`
}

type Paginate struct {
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	NextPage   *int  `json:"next_page"`
	PrevPage   *int  `json:"prev_page"`
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
	db = p.Filters.GenerateCondition(db)
	db.Model(ds).Count(&count)
	db.Scopes(gormPaginate(p.Page, p.Size)).Find(ds)

	pageCount := math.Floor(float64(count / int64(p.Size)))
	if count%int64(p.Size) > 0 {
		pageCount++
	}
	result.Data = ds
	result.Paginate = buildPaginator(p, count)

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

	query += p.Filters.GenerateConditionRaw()
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
	queries := strings.Split(strings.ToLower(query), "from")
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
	result.Paginate = buildPaginator(p, count)

	return &result
}

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

func buildPaginator(p *Config, c int64) Paginate {
	var prevPage, nextPage *int
	var page int = p.Page
	var size int = p.Size
	totalPages := int(math.Ceil(float64(c) / float64(size)))

	if page > 1 {
		np := page - 1
		prevPage = &np
	}
	if page == totalPages {
	} else {
		np := page + 1
		nextPage = &np
	}
	return Paginate{
		Total:      c,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
		NextPage:   nextPage,
		PrevPage:   prevPage,
	}
}
