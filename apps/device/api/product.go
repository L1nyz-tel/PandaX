package api

// ==========================================================================
// 生成日期：2023-06-30 08:57:48 +0800 CST
// 生成路径: apps/device/api/products.go
// 生成人：panda
// ==========================================================================
import (
	"fmt"
	"pandax/kit/biz"
	"pandax/kit/model"
	"pandax/kit/restfulx"
	"pandax/pkg/global"
	model2 "pandax/pkg/global/model"
	"strings"

	"pandax/apps/device/entity"
	"pandax/apps/device/services"
	ruleService "pandax/apps/rule/services"
)

type ProductApi struct {
	ProductApp  services.ProductModel
	DeviceApp   services.DeviceModel
	TemplateApp services.ProductTemplateModel
	OtaAPP      services.ProductOtaModel
	RuleApp     ruleService.RuleChainModel
}

// GetProductList Product列表数据
func (p *ProductApi) GetProductList(rc *restfulx.ReqCtx) {
	data := entity.Product{}
	pageNum := restfulx.QueryInt(rc, "pageNum", 1)
	pageSize := restfulx.QueryInt(rc, "pageSize", 10)
	data.Status = restfulx.QueryParam(rc, "status")
	data.ProductCategoryId = restfulx.QueryParam(rc, "productCategoryId")
	data.ProtocolName = restfulx.QueryParam(rc, "protocolName")
	data.DeviceType = restfulx.QueryParam(rc, "deviceType")
	data.Name = restfulx.QueryParam(rc, "name")

	list, total, err := p.ProductApp.FindListPage(pageNum, pageSize, data)
	biz.ErrIsNil(err, "查询产品列表失败")
	rc.ResData = model.ResultPage{
		Total:    total,
		PageNum:  int64(pageNum),
		PageSize: int64(pageSize),
		Data:     list,
	}
}

func (p *ProductApi) GetProductListAll(rc *restfulx.ReqCtx) {
	data := entity.Product{}
	data.Status = restfulx.QueryParam(rc, "status")
	data.ProductCategoryId = restfulx.QueryParam(rc, "productCategoryId")
	data.ProtocolName = restfulx.QueryParam(rc, "protocolName")
	data.DeviceType = restfulx.QueryParam(rc, "deviceType")
	data.Name = restfulx.QueryParam(rc, "name")
	list, err := p.ProductApp.FindList(data)
	biz.ErrIsNil(err, "查询产品列表失败")
	rc.ResData = list
}

// GetProductTsl Template列表数据
func (p *ProductApi) GetProductTsl(rc *restfulx.ReqCtx) {
	data := entity.ProductTemplate{}
	data.Pid = restfulx.PathParam(rc, "id")
	templates, err := p.TemplateApp.FindList(data)
	biz.ErrIsNil(err, "查询产品模板列表失败")
	attributes := make([]map[string]interface{}, 0)
	telemetry := make([]map[string]interface{}, 0)
	commands := make([]map[string]interface{}, 0)
	for _, template := range *templates {
		tslData := map[string]interface{}{
			"name":       template.Name,
			"identifier": template.Key,
			"type":       template.Type,
			"define":     template.Define,
		}
		if template.Classify == global.TslAttributesType {
			attributes = append(attributes, tslData)
		}
		if template.Classify == global.TslTelemetryType {
			telemetry = append(telemetry, tslData)
		}
		if template.Classify == global.TslCommandsType {
			commands = append(commands, tslData)
		}
	}

	rc.ResData = map[string]interface{}{
		"attributes": attributes,
		"telemetry":  telemetry,
		"commands":   commands,
	}
}

// GetProduct 获取Product
func (p *ProductApi) GetProduct(rc *restfulx.ReqCtx) {
	id := restfulx.PathParam(rc, "id")
	one, err := p.ProductApp.FindOne(id)
	biz.ErrIsNil(err, "查询失败")
	rc.ResData = one
}

// InsertProduct 添加Product
func (p *ProductApi) InsertProduct(rc *restfulx.ReqCtx) {
	var data entity.Product
	restfulx.BindJsonAndValid(rc, &data)
	data.Id = model2.GenerateID()
	data.Owner = rc.LoginAccount.UserName
	data.OrgId = rc.LoginAccount.OrganizationId
	// 如果未设置规则链，默认为主链
	if data.RuleChainId == "" {
		root, err := p.RuleApp.FindOneByRoot()
		biz.ErrIsNil(err, "规则链查询错误")
		data.RuleChainId = root.Id
	}

	p.ProductApp.Insert(data)
}

// UpdateProduct 修改Product
func (p *ProductApi) UpdateProduct(rc *restfulx.ReqCtx) {
	var data entity.Product
	restfulx.BindQuery(rc, &data)

	p.ProductApp.Update(data)
}

func (p *ProductApi) UpdateProductStatue(rc *restfulx.ReqCtx) {
	var data entity.Product
	restfulx.BindJsonAndValid(rc, &data)

	p.ProductApp.Update(data)
}

// DeleteProduct 删除Product
func (p *ProductApi) DeleteProduct(rc *restfulx.ReqCtx) {
	id := restfulx.PathParam(rc, "id")
	ids := strings.Split(id, ",")
	for _, id := range ids {
		list, _ := p.DeviceApp.FindList(entity.Device{Pid: id})
		biz.IsTrue(!(list != nil && len(*list) > 0), fmt.Sprintf("产品ID%s下存在设备，不能被删除", id))
	}
	// 删除产品
	p.ProductApp.Delete(ids)
	for _, id := range ids {
		// 删除所有模型，固件
		p.TemplateApp.Delete([]string{id})
		p.OtaAPP.Delete([]string{id})
	}
}
