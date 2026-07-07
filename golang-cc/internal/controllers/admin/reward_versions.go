package admin

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rafa/golang-cc/internal/domain"
)

func (c *Controller) PublishComponentVersion(ctx *gin.Context) {
	var request struct {
		RewardUnitID       string  `json:"reward_unit_id"`
		Name               string  `json:"name"`
		Description        string  `json:"description"`
		EffectType         string  `json:"effect_type"`
		RewardValue        string  `json:"reward_value"`
		EffectiveFrom      string  `json:"effective_from"`
		EffectiveTo        *string `json:"effective_to"`
		AnnouncedAt        *string `json:"announced_at"`
		ChangeReason       string  `json:"change_reason"`
		DisplayChangeUntil *string `json:"display_change_until"`
	}
	if ctx.ShouldBindJSON(&request) != nil {
		failure(ctx, 400, "validation_failed", "版本資料無效")
		return
	}
	effectiveFrom, err := time.Parse(time.RFC3339, request.EffectiveFrom)
	if err != nil {
		failure(ctx, 400, "validation_failed", "生效時間格式錯誤")
		return
	}
	parseOptional := func(value *string) (*time.Time, error) {
		if value == nil || *value == "" {
			return nil, nil
		}
		parsed, err := time.Parse(time.RFC3339, *value)
		return &parsed, err
	}
	effectiveTo, err := parseOptional(request.EffectiveTo)
	if err != nil {
		failure(ctx, 400, "validation_failed", "失效時間格式錯誤")
		return
	}
	announcedAt, err := parseOptional(request.AnnouncedAt)
	if err != nil {
		failure(ctx, 400, "validation_failed", "公告時間格式錯誤")
		return
	}
	displayUntil, err := parseOptional(request.DisplayChangeUntil)
	if err != nil {
		failure(ctx, 400, "validation_failed", "提示期限格式錯誤")
		return
	}
	id, err := c.service.PublishComponentVersion(ctx, ctx.Param("componentId"), domain.RewardComponentVersionInput{
		RewardUnitID: request.RewardUnitID, Name: request.Name, Description: request.Description,
		EffectType: request.EffectType, RewardValue: request.RewardValue, EffectiveFrom: effectiveFrom, EffectiveTo: effectiveTo,
		AnnouncedAt: announcedAt, ChangeReason: request.ChangeReason, DisplayChangeUntil: displayUntil,
	})
	if isNotFound(err) {
		failure(ctx, 404, "not_found", "找不到回饋項目")
		return
	}
	if err != nil {
		failure(ctx, 409, "version_conflict", "版本期間重疊或資料無效")
		return
	}
	data(ctx, 201, gin.H{"id": id})
}

func (c *Controller) PublishConditionVersion(ctx *gin.Context) {
	var request struct {
		Operator      string          `json:"operator"`
		Configuration json.RawMessage `json:"configuration_json"`
		Description   string          `json:"description"`
		EffectiveFrom string          `json:"effective_from"`
		EffectiveTo   *string         `json:"effective_to"`
	}
	if ctx.ShouldBindJSON(&request) != nil {
		failure(ctx, 400, "validation_failed", "條件版本資料無效")
		return
	}
	effectiveFrom, err := time.Parse(time.RFC3339, request.EffectiveFrom)
	if err != nil {
		failure(ctx, 400, "validation_failed", "生效時間格式錯誤")
		return
	}
	var effectiveTo *time.Time
	if request.EffectiveTo != nil && *request.EffectiveTo != "" {
		value, err := time.Parse(time.RFC3339, *request.EffectiveTo)
		if err != nil {
			failure(ctx, 400, "validation_failed", "失效時間格式錯誤")
			return
		}
		effectiveTo = &value
	}
	id, err := c.service.PublishConditionVersion(ctx, ctx.Param("conditionId"), domain.RewardConditionVersionInput{
		Operator: request.Operator, Configuration: request.Configuration, Description: request.Description,
		EffectiveFrom: effectiveFrom, EffectiveTo: effectiveTo,
	})
	versionResponse(ctx, id, err, "找不到回饋條件")
}

func (c *Controller) PublishCapVersion(ctx *gin.Context) {
	var request struct {
		CapType       string  `json:"cap_type"`
		LimitValue    string  `json:"limit_value"`
		LimitFormula  string  `json:"limit_formula"`
		RewardUnitID  *string `json:"reward_unit_id"`
		PeriodType    string  `json:"period_type"`
		EffectiveFrom string  `json:"effective_from"`
		EffectiveTo   *string `json:"effective_to"`
	}
	if ctx.ShouldBindJSON(&request) != nil {
		failure(ctx, 400, "validation_failed", "上限版本資料無效")
		return
	}
	effectiveFrom, err := time.Parse(time.RFC3339, request.EffectiveFrom)
	if err != nil {
		failure(ctx, 400, "validation_failed", "生效時間格式錯誤")
		return
	}
	var effectiveTo *time.Time
	if request.EffectiveTo != nil && *request.EffectiveTo != "" {
		value, err := time.Parse(time.RFC3339, *request.EffectiveTo)
		if err != nil {
			failure(ctx, 400, "validation_failed", "失效時間格式錯誤")
			return
		}
		effectiveTo = &value
	}
	id, err := c.service.PublishCapVersion(ctx, ctx.Param("capId"), domain.RewardCapVersionInput{
		CapType: request.CapType, LimitValue: request.LimitValue, LimitFormula: request.LimitFormula, RewardUnitID: request.RewardUnitID,
		PeriodType: request.PeriodType, EffectiveFrom: effectiveFrom, EffectiveTo: effectiveTo,
	})
	versionResponse(ctx, id, err, "找不到回饋上限")
}

func versionResponse(ctx *gin.Context, id string, err error, notFoundMessage string) {
	if isNotFound(err) {
		failure(ctx, 404, "not_found", notFoundMessage)
		return
	}
	if err != nil {
		failure(ctx, 409, "version_conflict", "版本期間重疊或資料無效")
		return
	}
	data(ctx, 201, gin.H{"id": id})
}
