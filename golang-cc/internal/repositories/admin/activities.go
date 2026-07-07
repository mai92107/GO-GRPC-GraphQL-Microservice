package admin

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rafa/golang-cc/internal/domain"
	service "github.com/rafa/golang-cc/internal/services/admin"
	"gorm.io/gorm"
)

type activityModel struct {
	ID            string `gorm:"column:id;primaryKey"`
	BankID        string `gorm:"column:bank_id"`
	CardProductID string `gorm:"column:card_product_id"`
	Title         string `gorm:"column:title"`
	Description   string `gorm:"column:description"`
	SourceURL     string `gorm:"column:source_url"`
	EffectiveFrom string `gorm:"column:effective_from"`
	EffectiveTo   string `gorm:"column:effective_to"`
	IsActive      bool   `gorm:"column:is_active"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Groups        []activityGroupModel `gorm:"foreignKey:ActivityID"`
}

func (activityModel) TableName() string { return "reward.activities" }

type activityGroupModel struct {
	ID           string `gorm:"column:id;primaryKey"`
	ActivityID   string `gorm:"column:activity_id"`
	Name         string `gorm:"column:name"`
	Description  string `gorm:"column:description"`
	DisplayOrder int    `gorm:"column:display_order"`
	IsActive     bool   `gorm:"column:is_active"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (activityGroupModel) TableName() string { return "reward.activity_groups" }

type activityComponentModel struct {
	ID            string `gorm:"column:id;primaryKey"`
	Name          string `gorm:"column:name"`
	Description   string `gorm:"column:description"`
	Layer         int    `gorm:"column:layer"`
	StackGroup    string `gorm:"column:stack_group"`
	StackMode     string `gorm:"column:stack_mode"`
	Priority      int    `gorm:"column:priority"`
	EffectiveFrom string `gorm:"column:effective_from"`
	EffectiveTo   string `gorm:"column:effective_to"`
	IsActive      bool   `gorm:"column:is_active"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Requirements  []activityRequirementModel `gorm:"foreignKey:RewardComponentID"`
	Benefits      []activityBenefitModel     `gorm:"foreignKey:RewardComponentID"`
}

func (activityComponentModel) TableName() string { return "reward.activity_components" }

type activityComponentGroupModel struct {
	RewardComponentID string `gorm:"column:reward_component_id;primaryKey"`
	RewardGroupID     string `gorm:"column:reward_group_id;primaryKey"`
}

func (activityComponentGroupModel) TableName() string { return "reward.activity_component_groups" }

type activityRequirementModel struct {
	ID                string          `gorm:"column:id;primaryKey"`
	RewardComponentID string          `gorm:"column:reward_component_id"`
	RequirementType   string          `gorm:"column:requirement_type"`
	Operator          string          `gorm:"column:operator"`
	Configuration     json.RawMessage `gorm:"column:configuration_json;type:jsonb"`
	Description       string          `gorm:"column:description"`
	IsActive          bool            `gorm:"column:is_active"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (activityRequirementModel) TableName() string { return "reward.activity_requirements" }

type activityBenefitModel struct {
	ID                string  `gorm:"column:id;primaryKey"`
	RewardComponentID string  `gorm:"column:reward_component_id"`
	BenefitType       string  `gorm:"column:benefit_type"`
	Value             string  `gorm:"column:value"`
	RewardUnitID      string  `gorm:"column:reward_unit_id"`
	CapAmount         *string `gorm:"column:cap_amount"`
	CapFormula        *string `gorm:"column:cap_formula"`
	CapPeriod         *string `gorm:"column:cap_period"`
	Description       string  `gorm:"column:description"`
	IsActive          bool    `gorm:"column:is_active"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (activityBenefitModel) TableName() string { return "reward.activity_benefits" }

func (r *Repository) ListActivities(ctx context.Context, filters service.ActivityFilters) ([]domain.ActivitySummary, error) {
	query := r.db.WithContext(ctx).
		Table("reward.activities AS a").
		Select(`a.id,a.bank_id,b.name AS bank_name,a.card_product_id,cp.name AS card_name,a.title,a.description,a.source_url,
			a.effective_from::text AS effective_from,a.effective_to::text AS effective_to,a.is_active,a.created_at,a.updated_at,
			COUNT(DISTINCT g.id) AS group_count,COUNT(DISTINCT c.id) AS component_count`).
		Joins("JOIN banks AS b ON b.id=a.bank_id").
		Joins("JOIN catalog.card_products AS cp ON cp.id=a.card_product_id").
		Joins("LEFT JOIN reward.activity_groups AS g ON g.activity_id=a.id").
		Joins("LEFT JOIN reward.activity_component_groups AS cg ON cg.reward_group_id=g.id").
		Joins("LEFT JOIN reward.activity_components AS c ON c.id=cg.reward_component_id").
		Group("a.id,b.name,cp.name")
	if filters.BankID != "" {
		query = query.Where("a.bank_id = ?", filters.BankID)
	}
	if filters.CardProductID != "" {
		query = query.Where("a.card_product_id = ?", filters.CardProductID)
	}
	if filters.IsActive != nil {
		query = query.Where("a.is_active = ?", *filters.IsActive)
	}
	var out []domain.ActivitySummary
	err := query.Order("a.effective_from DESC,a.title").Scan(&out).Error
	return out, err
}

func (r *Repository) GetActivity(ctx context.Context, id string) (domain.ActivityFlow, error) {
	var model activityModel
	err := r.db.WithContext(ctx).
		Preload("Groups", func(db *gorm.DB) *gorm.DB { return db.Order("display_order,name") }).
		Where("id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ActivityFlow{}, domain.ErrNotFound
		}
		return domain.ActivityFlow{}, err
	}
	flow := flowFromModel(model)
	if err := r.loadFlowComponents(ctx, &flow); err != nil {
		return domain.ActivityFlow{}, err
	}
	var names struct{ BankName, CardName string }
	if err := r.db.WithContext(ctx).
		Table("banks AS b").
		Select("b.name AS bank_name,cp.name AS card_name").
		Joins("JOIN catalog.card_products AS cp ON cp.id=?", model.CardProductID).
		Where("b.id=?", model.BankID).
		Scan(&names).Error; err != nil {
		return domain.ActivityFlow{}, err
	}
	flow.Activity.BankName = names.BankName
	flow.Activity.CardName = names.CardName
	return flow, nil
}

func (r *Repository) CreateActivity(ctx context.Context, flow domain.ActivityFlow) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := modelFromFlow(flow)
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		return saveFlowChildren(tx, flow)
	})
}

func (r *Repository) UpdateActivity(ctx context.Context, flow domain.ActivityFlow) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exists bool
		if err := tx.Model(&activityModel{}).Select("count(*) > 0").Where("id = ?", flow.Activity.ID).Find(&exists).Error; err != nil {
			return err
		}
		if !exists {
			return domain.ErrNotFound
		}
		model := modelFromFlow(flow)
		if err := tx.Model(&activityModel{}).Where("id = ?", model.ID).Updates(map[string]any{
			"bank_id": model.BankID, "card_product_id": model.CardProductID, "title": model.Title,
			"description": model.Description, "source_url": model.SourceURL, "effective_from": model.EffectiveFrom,
			"effective_to": model.EffectiveTo, "is_active": model.IsActive, "updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}
		groupIDs, componentIDs, requirementIDs, benefitIDs := collectFlowIDs(flow)
		if err := deactivateMissing(tx, flow.Activity.ID, groupIDs, componentIDs, requirementIDs, benefitIDs); err != nil {
			return err
		}
		if err := clearCurrentComponentLinks(tx, flow.Activity.ID, componentIDs); err != nil {
			return err
		}
		for _, group := range model.Groups {
			if err := tx.Save(&group).Error; err != nil {
				return err
			}
		}
		return saveFlowChildren(tx, flow)
	})
}

func (r *Repository) SetActivityStatus(ctx context.Context, id string, active bool) error {
	result := r.db.WithContext(ctx).Model(&activityModel{}).Where("id = ?", id).Updates(map[string]any{"is_active": active, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteActivity(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&activityModel{}).Where("id = ?", id).Updates(map[string]any{"is_active": false, "updated_at": time.Now()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		if err := tx.Exec(`UPDATE reward.activity_groups g SET is_active=false,updated_at=now() WHERE g.activity_id=?`, id).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM reward.activity_component_groups cg USING reward.activity_groups g WHERE cg.reward_group_id=g.id AND g.activity_id=?`, id).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM reward.activity_components c WHERE NOT EXISTS (SELECT 1 FROM reward.activity_component_groups cg WHERE cg.reward_component_id=c.id)`).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) ListRequirementTypes(ctx context.Context) ([]domain.RequirementTypeOption, error) {
	out := []domain.RequirementTypeOption{}
	err := r.db.WithContext(ctx).Raw(`SELECT id AS code,name,value_key,value_source FROM reward.requirement_types WHERE is_active ORDER BY display_order,name`).Scan(&out).Error
	return out, err
}

func (r *Repository) ActivityRequirementOptions(ctx context.Context, filters service.RequirementOptionFilters) (domain.RequirementOptionSet, error) {
	out := domain.RequirementOptionSet{
		Operators:  []domain.CodeNameOption{{Code: "IN", Name: "包含任一"}, {Code: "NOT_IN", Name: "不包含"}, {Code: "EQ", Name: "等於"}, {Code: "GTE", Name: "大於等於"}, {Code: "LTE", Name: "小於等於"}, {Code: "BETWEEN", Name: "介於"}},
		Channels:   []domain.CodeNameOption{{Code: "ONLINE", Name: "線上"}, {Code: "PHYSICAL", Name: "實體"}},
		Currencies: []domain.CodeNameOption{{Code: "TWD", Name: "新台幣"}},
	}

	switch filters.RequirementType {
	case "PAYMENT_METHOD":
		if err := r.scanCodeNames(ctx, "SELECT id::text AS code,name FROM payment_methods WHERE is_active ORDER BY name", &out.PaymentMethods); err != nil {
			return out, err
		}
	case "MERCHANT":
		if err := r.scanCodeNames(ctx, "SELECT id::text AS code,name FROM merchants WHERE is_active ORDER BY name", &out.Merchants); err != nil {
			return out, err
		}
	case "MERCHANT_CATEGORY", "CONSUMPTION_CATEGORY":
		if err := r.scanCodeNames(ctx, "SELECT id::text AS code,name FROM categories WHERE is_active ORDER BY name", &out.Categories); err != nil {
			return out, err
		}
	case "CARD_NETWORK":
		if filters.CardProductID == "" {
			return out, nil
		}
		if err := r.db.WithContext(ctx).Raw(`SELECT trim(value) AS code,trim(value) AS name FROM catalog.card_products cp CROSS JOIN LATERAL regexp_split_to_table(COALESCE(cp.networks,''), ',') value WHERE cp.id = ? AND trim(value)<>'' ORDER BY name`, filters.CardProductID).Scan(&out.CardNetworks).Error; err != nil {
			return out, err
		}
	case "CARD_PRODUCT":
		if err := r.scanCodeNames(ctx, "SELECT id::text AS code,name FROM catalog.card_products WHERE is_active ORDER BY name", &out.CardProducts); err != nil {
			return out, err
		}
	case "CARD_PLAN":
		if filters.CardProductID == "" {
			return out, nil
		}
		if err := r.db.WithContext(ctx).Table("catalog.card_plans AS p").
			Select("p.id,p.card_product_id,p.plan_type,pv.name").
			Joins("JOIN LATERAL (SELECT name FROM catalog.card_plan_versions v WHERE v.card_plan_id=p.id ORDER BY v.effective_from DESC LIMIT 1) pv ON true").
			Where("p.is_active AND p.card_product_id = ? AND p.plan_type = ?", filters.CardProductID, "selectable").
			Order("pv.name").
			Scan(&out.CardPlans).Error; err != nil {
			return out, err
		}
	case "ACCOUNT_TIER":
		if filters.CardProductID == "" {
			return out, nil
		}
		if err := r.db.WithContext(ctx).Raw(`
			SELECT DISTINCT trim(value) AS code,trim(value) AS name
			FROM catalog.card_products cp
			CROSS JOIN LATERAL regexp_split_to_table(COALESCE(cp.qualified_type,''), ',') value
			WHERE cp.id = ? AND trim(value)<>''
			ORDER BY name`, filters.CardProductID).Scan(&out.AccountTiers).Error; err != nil {
			return out, err
		}
	case "REGION":
		if err := r.scanCodeNames(ctx, "SELECT id AS code,name FROM catalog.regions WHERE is_active ORDER BY name", &out.Regions); err != nil {
			return out, err
		}
	case "USER_QUALIFICATION":
		if err := r.scanCodeNames(ctx, "SELECT id AS code,name FROM catalog.user_qualifications WHERE is_active ORDER BY name", &out.UserQualifications); err != nil {
			return out, err
		}
	}
	return out, nil
}

func (r *Repository) scanCodeNames(ctx context.Context, query string, out *[]domain.CodeNameOption) error {
	return r.db.WithContext(ctx).Raw(query).Scan(out).Error
}
func modelFromFlow(flow domain.ActivityFlow) activityModel {
	model := activityModel{ID: flow.Activity.ID, BankID: flow.Activity.BankID, CardProductID: flow.Activity.CardProductID, Title: flow.Activity.Title, Description: flow.Activity.Description, SourceURL: flow.Activity.SourceURL, EffectiveFrom: flow.Activity.EffectiveFrom, EffectiveTo: flow.Activity.EffectiveTo, IsActive: flow.Activity.IsActive}
	for _, group := range flow.RewardGroups {
		groupModel := activityGroupModel{ID: group.ID, ActivityID: flow.Activity.ID, Name: group.Name, Description: group.Description, DisplayOrder: group.DisplayOrder, IsActive: group.IsActive}
		model.Groups = append(model.Groups, groupModel)
	}
	return model
}

func flowFromModel(model activityModel) domain.ActivityFlow {
	flow := domain.ActivityFlow{Activity: domain.ActivitySummary{ID: model.ID, BankID: model.BankID, CardProductID: model.CardProductID, Title: model.Title, Description: model.Description, SourceURL: model.SourceURL, EffectiveFrom: model.EffectiveFrom, EffectiveTo: model.EffectiveTo, IsActive: model.IsActive, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}}
	for _, group := range model.Groups {
		outGroup := domain.RewardGroup{ID: group.ID, ActivityID: group.ActivityID, Name: group.Name, Description: group.Description, DisplayOrder: group.DisplayOrder, IsActive: group.IsActive}
		flow.RewardGroups = append(flow.RewardGroups, outGroup)
	}
	return flow
}

type activityComponentLinkRow struct {
	RewardGroupID string `gorm:"column:reward_group_id"`
	ID            string `gorm:"column:id"`
	Name          string `gorm:"column:name"`
	Description   string `gorm:"column:description"`
	Layer         int    `gorm:"column:layer"`
	StackGroup    string `gorm:"column:stack_group"`
	StackMode     string `gorm:"column:stack_mode"`
	Priority      int    `gorm:"column:priority"`
	EffectiveFrom string `gorm:"column:effective_from"`
	EffectiveTo   string `gorm:"column:effective_to"`
	IsActive      bool   `gorm:"column:is_active"`
}

func (r *Repository) loadFlowComponents(ctx context.Context, flow *domain.ActivityFlow) error {
	if len(flow.RewardGroups) == 0 {
		return nil
	}
	rows := []activityComponentLinkRow{}
	if err := r.db.WithContext(ctx).
		Table("reward.activity_component_groups AS cg").
		Select(`cg.reward_group_id,c.id,c.name,c.description,c.layer,c.stack_group,c.stack_mode,c.priority,
			c.effective_from::text AS effective_from,c.effective_to::text AS effective_to,c.is_active`).
		Joins("JOIN reward.activity_components AS c ON c.id=cg.reward_component_id").
		Joins("JOIN reward.activity_groups AS g ON g.id=cg.reward_group_id").
		Where("g.activity_id = ?", flow.Activity.ID).
		Order("g.display_order,g.name,c.layer,c.priority,c.name").
		Scan(&rows).Error; err != nil {
		return err
	}
	componentIDs := []string{}
	components := map[string]domain.RewardComponent{}
	for _, row := range rows {
		component := components[row.ID]
		if component.ID == "" {
			component = domain.RewardComponent{
				ID: row.ID, Name: row.Name, Description: row.Description, Layer: row.Layer, StackGroup: row.StackGroup,
				StackMode: row.StackMode, Priority: row.Priority, EffectiveFrom: row.EffectiveFrom, EffectiveTo: row.EffectiveTo, IsActive: row.IsActive,
			}
			componentIDs = append(componentIDs, row.ID)
		}
		component.RewardGroupIDs = appendUnique(component.RewardGroupIDs, row.RewardGroupID)
		components[row.ID] = component
	}
	if len(componentIDs) > 0 {
		requirements := []activityRequirementModel{}
		if err := r.db.WithContext(ctx).Where("reward_component_id IN ?", componentIDs).Order("requirement_type,id").Find(&requirements).Error; err != nil {
			return err
		}
		for _, requirement := range requirements {
			component := components[requirement.RewardComponentID]
			component.Requirements = append(component.Requirements, domain.RewardRequirement{ID: requirement.ID, RewardComponentID: requirement.RewardComponentID, RequirementType: requirement.RequirementType, Operator: requirement.Operator, Configuration: requirement.Configuration, Description: requirement.Description, IsActive: requirement.IsActive})
			components[requirement.RewardComponentID] = component
		}
		benefits := []activityBenefitModel{}
		if err := r.db.WithContext(ctx).Where("reward_component_id IN ?", componentIDs).Order("benefit_type,id").Find(&benefits).Error; err != nil {
			return err
		}
		for _, benefit := range benefits {
			component := components[benefit.RewardComponentID]
			component.Benefits = append(component.Benefits, domain.RewardBenefit{ID: benefit.ID, RewardComponentID: benefit.RewardComponentID, BenefitType: benefit.BenefitType, Value: benefit.Value, RewardUnitID: benefit.RewardUnitID, CapAmount: benefit.CapAmount, CapFormula: benefit.CapFormula, CapPeriod: benefit.CapPeriod, Description: benefit.Description, IsActive: benefit.IsActive})
			components[benefit.RewardComponentID] = component
		}
	}
	for index := range flow.RewardGroups {
		for _, row := range rows {
			if row.RewardGroupID == flow.RewardGroups[index].ID {
				flow.RewardGroups[index].Components = append(flow.RewardGroups[index].Components, components[row.ID])
			}
		}
	}
	return nil
}

func saveFlowChildren(tx *gorm.DB, flow domain.ActivityFlow) error {
	components := uniqueFlowComponents(flow)
	for _, component := range components {
		componentModel := activityComponentModel{ID: component.ID, Name: component.Name, Description: component.Description, Layer: component.Layer, StackGroup: component.StackGroup, StackMode: component.StackMode, Priority: component.Priority, EffectiveFrom: component.EffectiveFrom, EffectiveTo: component.EffectiveTo, IsActive: component.IsActive}
		if err := tx.Save(&componentModel).Error; err != nil {
			return err
		}
		for _, groupID := range component.RewardGroupIDs {
			if err := tx.Save(&activityComponentGroupModel{RewardComponentID: component.ID, RewardGroupID: groupID}).Error; err != nil {
				return err
			}
		}
		for _, requirement := range component.Requirements {
			if err := tx.Save(&activityRequirementModel{ID: requirement.ID, RewardComponentID: component.ID, RequirementType: requirement.RequirementType, Operator: requirement.Operator, Configuration: requirement.Configuration, Description: requirement.Description, IsActive: requirement.IsActive}).Error; err != nil {
				return err
			}
		}
		for _, benefit := range component.Benefits {
			if err := tx.Save(&activityBenefitModel{ID: benefit.ID, RewardComponentID: component.ID, BenefitType: benefit.BenefitType, Value: benefit.Value, RewardUnitID: benefit.RewardUnitID, CapAmount: benefit.CapAmount, CapFormula: benefit.CapFormula, CapPeriod: benefit.CapPeriod, Description: benefit.Description, IsActive: benefit.IsActive}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func clearCurrentComponentLinks(tx *gorm.DB, activityID string, componentIDs []string) error {
	if len(componentIDs) == 0 {
		return nil
	}
	return tx.Exec("DELETE FROM reward.activity_component_groups cg USING reward.activity_groups g WHERE cg.reward_group_id=g.id AND g.activity_id=? AND cg.reward_component_id IN ?", activityID, componentIDs).Error
}

func uniqueFlowComponents(flow domain.ActivityFlow) []domain.RewardComponent {
	components := []domain.RewardComponent{}
	seen := map[string]bool{}
	for _, group := range flow.RewardGroups {
		for _, component := range group.Components {
			if seen[component.ID] {
				continue
			}
			seen[component.ID] = true
			components = append(components, component)
		}
	}
	return components
}

func collectFlowIDs(flow domain.ActivityFlow) ([]string, []string, []string, []string) {
	groups, components, requirements, benefits := []string{}, []string{}, []string{}, []string{}
	seenComponents := map[string]bool{}
	for _, group := range flow.RewardGroups {
		groups = append(groups, group.ID)
		for _, component := range group.Components {
			if seenComponents[component.ID] {
				continue
			}
			seenComponents[component.ID] = true
			components = append(components, component.ID)
			for _, requirement := range component.Requirements {
				requirements = append(requirements, requirement.ID)
			}
			for _, benefit := range component.Benefits {
				benefits = append(benefits, benefit.ID)
			}
		}
	}
	return groups, components, requirements, benefits
}

func deactivateMissing(tx *gorm.DB, activityID string, groupIDs, componentIDs, requirementIDs, benefitIDs []string) error {
	if err := tx.Exec("DELETE FROM reward.activity_groups WHERE activity_id=? AND id NOT IN ?", activityID, groupIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM reward.activity_component_groups cg USING reward.activity_groups g WHERE cg.reward_group_id=g.id AND g.activity_id=? AND (cg.reward_group_id NOT IN ? OR cg.reward_component_id NOT IN ?)", activityID, groupIDs, componentIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM reward.activity_component_groups cg USING reward.activity_groups g WHERE cg.reward_group_id=g.id AND g.activity_id=? AND NOT EXISTS (SELECT 1 FROM reward.activity_components c WHERE c.id=cg.reward_component_id)", activityID).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM reward.activity_components c WHERE c.id NOT IN ? AND NOT EXISTS (SELECT 1 FROM reward.activity_component_groups cg WHERE cg.reward_component_id=c.id)", componentIDs).Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM reward.activity_requirements r WHERE r.reward_component_id IN ? AND r.id NOT IN ?", componentIDs, requirementIDs).Error; err != nil {
		return err
	}
	return tx.Exec("DELETE FROM reward.activity_benefits b WHERE b.reward_component_id IN ? AND b.id NOT IN ?", componentIDs, benefitIDs).Error
}

func appendUnique(values []string, next string) []string {
	for _, value := range values {
		if value == next {
			return values
		}
	}
	return append(values, next)
}
