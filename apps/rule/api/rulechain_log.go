package api

import (
	"pandax/apps/rule/entity"
	"pandax/apps/rule/services"
	"pandax/kit/model"
	"pandax/kit/restfulx"
	"pandax/pkg/rule_engine/nodes"
)

type RuleChainMsgLogApi struct {
	RuleChainMsgLogApp services.RuleChainMsgLogModel
}

func (r *RuleChainMsgLogApi) GetNodeLabels(rc *restfulx.ReqCtx) {
	rc.ResData = nodes.GetCategory()
}

// GetRuleChainMsgLogList 列表数据
func (p *RuleChainMsgLogApi) GetRuleChainMsgLogList(rc *restfulx.ReqCtx) {
	data := entity.RuleChainMsgLog{}
	pageNum := restfulx.QueryInt(rc, "pageNum", 1)
	pageSize := restfulx.QueryInt(rc, "pageSize", 10)
	data.DeviceName = restfulx.QueryParam(rc, "deviceName")
	data.MsgType = restfulx.QueryParam(rc, "msgType")

	data.RoleId = rc.LoginAccount.RoleId
	data.Owner = rc.LoginAccount.UserName

	list, total := p.RuleChainMsgLogApp.FindListPage(pageNum, pageSize, data)

	rc.ResData = model.ResultPage{
		Total:    total,
		PageNum:  int64(pageNum),
		PageSize: int64(pageSize),
		Data:     list,
	}
}

// DeleteRuleChainMsgLog 删除规则链
func (p *RuleChainMsgLogApi) DeleteRuleChainMsgLog(rc *restfulx.ReqCtx) {
	data := entity.RuleChainMsgLog{}
	data.DeviceName = restfulx.QueryParam(rc, "deviceName")
	data.MsgType = restfulx.QueryParam(rc, "msgType")
	p.RuleChainMsgLogApp.Delete(data)
}
