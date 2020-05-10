package service

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	"go-shipment-api/models"
	"go-shipment-api/pkg/global"
)

// 获取casbin策略管理器
func Casbin() (*casbin.Enforcer, error) {
	// 初始化数据库适配器, 添加自定义表前缀
	a, err := gormadapter.NewAdapterByDB(global.Mysql)
	if err != nil {
		return nil, err
	}
	// 读取配置文件
	config, err := global.ConfBox.Find(global.Conf.Casbin.ModelPath)
	cabinModel := model.NewModel()
	// 从字符串中加载casbin配置
	err = cabinModel.LoadModelFromText(string(config))
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(cabinModel, a)
	if err != nil {
		return nil, err
	}
	// 加载策略
	err = e.LoadPolicy()
	return e, err
}

// 创建一条casbin规则, 按角色
func CreateRoleCasbin(c models.SysRoleCasbin) (bool, error) {
	e, _ := Casbin()
	return e.AddPolicy(c.Keyword, c.Path, c.Method)
}

// 批量创建多条casbin规则, 按角色
func BatchCreateRoleCasbins(cs []models.SysRoleCasbin) (bool, error) {
	e, _ := Casbin()
	// 按角色构建
	rules := make([][]string, 0)
	for _, c := range cs {
		rules = append(rules, []string{
			c.Keyword,
			c.Path,
			c.Method,
		})
	}
	return e.AddPolicies(rules)
}

// 删除一条casbin规则, 按角色
func DeleteRoleCasbin(c models.SysRoleCasbin) (bool, error) {
	e, _ := Casbin()
	return e.RemovePolicy(c.Keyword, c.Path, c.Method)
}

// 批量删除多条casbin规则, 按角色
func BatchDeleteRoleCasbins(cs []models.SysRoleCasbin) (bool, error) {
	e, _ := Casbin()
	// 按角色构建
	rules := make([][]string, 0)
	for _, c := range cs {
		rules = append(rules, []string{
			c.Keyword,
			c.Path,
			c.Method,
		})
	}
	return e.RemovePolicies(rules)
}

// 根据权限编号读取casbin规则
func GetCasbinListByRoleId(roleId uint) ([]models.SysCasbin, error) {
	casbins := make([]models.SysCasbin, 0)
	var role models.SysRole
	err := global.Mysql.Where("id = ?", roleId).First(&role).Error
	if err != nil {
		return casbins, err
	}
	e, _ := Casbin()
	// 查询符合字段v0=role.Keyword所有casbin规则
	list := e.GetFilteredPolicy(0, role.Keyword)
	for _, v := range list {
		casbins = append(casbins, models.SysCasbin{
			PType: "p",
			V0:    v[0],
			V1:    v[1],
			V2:    v[2],
		})
	}
	return casbins, nil
}