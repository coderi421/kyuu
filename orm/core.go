package orm

import (
	"github.com/coderi421/kyuu/orm/internal/valuer"
	"github.com/coderi421/kyuu/orm/model"
)

type core struct {
	dialect    Dialect
	r          model.Registry // 存储数据库表和 struct 映射关系的实例
	valCreator valuer.Creator // 与DB交互映射的实现
	mdls       []Middleware
}
