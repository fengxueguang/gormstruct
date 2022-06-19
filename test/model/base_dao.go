package model

/**
 * @Author: prince.lee <leeprince@foxmail.com>
 * @Date:   2022-06-20 00:31:38
 * @Desc:   DAO 的基本方法
 */

import (
	"context"
	"gorm.io/gorm"
	"time"
)

const (
	// 分页器时，默认允许最大的记录数
	MaxLimit = 1000
)

// 初始化 gorm 实例的其他字段
type _BaseDAO struct {
	*gorm.DB
	ctx              context.Context
	cancel           context.CancelFunc
	timeout          time.Duration
	columns          []string
	isDefaultColumns bool
}

// 设置超时
func (obj *_BaseDAO) SetTimeOut(timeout time.Duration) {
	obj.ctx, obj.cancel = context.WithTimeout(context.Background(), timeout)
	obj.timeout = timeout
}

// 设置上下文
func (obj *_BaseDAO) SetCtx(c context.Context) {
	if c != nil {
		obj.ctx = c
	}
}

// 获取上下文
func (obj *_BaseDAO) GetCtx() context.Context {
	return obj.ctx
}

// 取消上下文
func (obj *_BaseDAO) Cancel(c context.Context) {
	obj.cancel()
}

// 获取 DB 实例
func (obj *_BaseDAO) GetDB() *gorm.DB {
	return obj.DB
}

// 更新 DB 实例
func (obj *_BaseDAO) UpdateDB(db *gorm.DB) {
	obj.DB = db
}

// 重置 gorm
func (obj *_BaseDAO) New() {
	obj.UpdateDB(obj.NewDB())
}

// 重置 gorm 会话
func (obj *_BaseDAO) NewDB() *gorm.DB {
	return obj.GetDB().Session(&gorm.Session{NewDB: true, Context: obj.ctx})
}

// 设置上下文获取 *grom.DB
func (obj *_BaseDAO) WithContext() (db *gorm.DB) {
	return obj.GetDB().WithContext(obj.ctx)
}

// 设置 sql 语句是否默认选择表的所有字段：没有通过WithSelect指定字段时，是否默认选择表的所有字段。更新/统计（count）语句时设置为false。
func (obj *_BaseDAO) setIsDefaultColumns(b bool) {
	obj.isDefaultColumns = b
}

// 查询指定字段
func (obj *_BaseDAO) WithSelect(query interface{}, args ...interface{}) Option {
	return selectOptionFunc(func(o *options) {
		o.selectField = queryArg{
			query: query,
			arg:   args,
		}
	})
}

// Or 查询：将所有 Withxxx 的 options.query 作为 Or 的查询条件
func (obj *_BaseDAO) WithOrOption(opts ...Option) Option {
	optionOr := initOption(opts...)

	return queryOrOptionFunc(func(o *options) {
		if len(optionOr.query) > 0 {
			o.queryOr = make(map[string]interface{}, len(optionOr.query))
			for key, value := range optionOr.query {
				o.queryOr[key] = value
			}
		}
	})
}

// 设置 offset、limit 作为 option 条件支持分页
func (obj *_BaseDAO) WithPage(offset int, limit int) Option {
	return pageOptionFunc(func(o *options) {
		o.page.offset = offset
		o.page.limit = limit
	})
}

// 设置 offset、limit 作为 option 条件支持分页
func (obj *_BaseDAO) WithOrderBy(orderBy string) Option {
	return orderByOptionFunc(func(o *options) {
		o.orderBy = orderBy
	})
}

// 分组
func (obj *_BaseDAO) WithGroupBy(groupBy string) Option {
	return groupByOptionFunc(func(o *options) {
		o.groupBy = groupBy
	})
}

// 分组后筛选
func (obj *_BaseDAO) WithHaving(query interface{}, args ...interface{}) Option {
	return havingByOptionFunc(func(o *options) {
		o.having = queryArg{
			query: query,
			arg:   args,
		}
	})
}

// 执行 sql 前的准备
func (obj *_BaseDAO) prepare(opts ...Option) (tx *gorm.DB) {
	options := initOption(opts...)

	tx = obj.WithContext().
		Scopes(obj.selectField(&options)).
		Where(options.query).
		Or(options.queryOr).
		Scopes(obj.paginate(&options)).
		Scopes(obj.orderBy(&options)).
		Scopes(obj.groupBy(&options)).
		Scopes(obj.having(&options))
	return
}

// 选择字段
func (obj *_BaseDAO) selectField(opt *options) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if opt.selectField.query != nil && opt.selectField.query != "" {
			if opt.selectField.arg != nil && len(opt.selectField.arg) > 0 {
				return db.Select(opt.selectField.query, opt.selectField.arg)
			}
			return db.Select(opt.selectField.query)
		} else if obj.isDefaultColumns {
			return db.Select(obj.columns)
		}
		return db
	}
}

// 排序
func (obj *_BaseDAO) orderBy(opt *options) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if opt.orderBy != "" {
			return db.Order(opt.orderBy)
		}
		return db
	}
}

// 分组
func (obj *_BaseDAO) groupBy(opt *options) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if opt.groupBy != "" {
			return db.Group(opt.groupBy)
		}
		return db
	}
}

// 分组后筛选
func (obj *_BaseDAO) having(opt *options) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if opt.having.query != nil && opt.having.query != "" {
			if opt.having.arg != nil && len(opt.having.arg) > 0 {
				return db.Having(opt.having.query, opt.having.arg)
			}
			return db.Having(opt.having.query)
		}
		return db
	}
}

// 分页器
func (obj *_BaseDAO) paginate(opt *options) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if opt.page.limit <= 0 {
			return db
		}
		if opt.page.limit > MaxLimit {
			opt.page.limit = MaxLimit
		}
		return db.Offset(opt.page.offset).Limit(opt.page.limit)
	}
}

// 函数选项模式的参数字段
type options struct {
	selectField queryArg
	query       map[string]interface{}
	queryOr     map[string]interface{}
	page        paging
	orderBy     string
	groupBy     string
	having      queryArg
}

// 分页数据结构
type queryArg struct {
	query interface{}
	arg   []interface{}
}

// 分页数据结构
type paging struct {
	offset int
	limit  int
}

// 函数选项模式接口
type Option interface {
	apply(*options)
}

// options.query 实现 Option 接口
type selectOptionFunc func(*options)

func (f selectOptionFunc) apply(o *options) {
	f(o)
}

// options.query 实现 Option 接口
type queryOptionFunc func(*options)

func (f queryOptionFunc) apply(o *options) {
	f(o)
}

// options.query 实现 Option 接口
type queryOrOptionFunc func(*options)

func (f queryOrOptionFunc) apply(o *options) {
	f(o)
}

// options.page 实现 Option 接口
type pageOptionFunc func(*options)

func (f pageOptionFunc) apply(o *options) {
	f(o)
}

// options.update 实现 Option 接口
type orderByOptionFunc func(*options)

func (f orderByOptionFunc) apply(o *options) {
	f(o)
}

// options.update 实现 Option 接口
type groupByOptionFunc func(*options)

func (f groupByOptionFunc) apply(o *options) {
	f(o)
}

// options.update 实现 Option 接口
type havingByOptionFunc func(*options)

func (f havingByOptionFunc) apply(o *options) {
	f(o)
}

// 初始化 options
func initOption(opts ...Option) options {
	opt := options{
		selectField: queryArg{
			query: nil,
			arg:   nil,
		},
		query:   make(map[string]interface{}, len(opts)),
		queryOr: make(map[string]interface{}, len(opts)),
		page: paging{
			offset: 0,
			limit:  0,
		},
		orderBy: "",
		groupBy: "",
	}
	for _, o := range opts {
		o.apply(&opt)
	}
	return opt
}
