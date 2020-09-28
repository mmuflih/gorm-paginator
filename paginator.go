package paginator

import (
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
		if page == 0 {
			page = 1
		}

		if size <= 0 {
			size = 10
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
	ShowSQL bool
}

type Paginator struct {
	Data     interface{} `json:"data"`
	Paginate Paginate    `json:"paginate"`
}

type Paginate struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

func Make(p *Config, gormDS interface{}) *Paginator {
	db := p.DB

	if p.ShowSQL {
		db = db.Debug()
	}
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Size == 0 {
		p.Size = 10
	}
	if len(p.OrderBy) > 0 {
		for _, o := range p.OrderBy {
			db = db.Order(o)
		}
	}

	var result Paginator
	var count int64

	db.Model(gormDS).Count(&count)

	db.Scopes(gormPaginate(p.Page, p.Size)).Find(gormDS)

	result.Data = gormDS
	result.Paginate = Paginate{
		p.Page,
		p.Size,
		int(count),
	}

	return &result
}
