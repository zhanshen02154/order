package repository

import "github.com/jinzhu/gorm"

type Paginator[T any] struct {
	Page        int64
	PageSize    int64
	Total       int64
	Pages       int64
	Data        []T
}

// Paginate
// Paginate[T any]
//  @Description: 分页查询
//  @param page
//  @return func(db *gorm.DB) *gorm.DB
//
func Paginate[T any](page *Paginator[T]) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page.Page <= 0 {
			page.Page = 0
		}
		switch {
		case page.PageSize > 100:
			page.PageSize = 100
		case page.PageSize <= 0:
			page.PageSize = 10
		}

		page.Pages = page.Total / page.PageSize
		if page.Total % page.PageSize != 0 {
			page.Pages++
		}
		p := page.Page
		if page.Page > page.Pages {
			p = page.Pages
		}
		size := page.PageSize
		offset := int((p - 1) * size)
		return db.Offset(offset).Limit(int(size))
	}
}

//
// FindPagedList
//  @Description: 分页查询
//  @receiver page
//  @param db
//  @param join
//  @return err
//
func (page *Paginator[T]) FindPagedList(db *gorm.DB) (err error) {
	err = nil
	err = db.Scopes(Paginate(page)).Find(&page.Data).Offset(-1).Limit(-1).Count(page.Total).Error
	return
}


