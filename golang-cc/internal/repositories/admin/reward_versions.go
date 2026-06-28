package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (r *Repository) PublishComponentVersion(ctx context.Context, componentID string, input domain.RewardComponentVersionInput) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT true FROM reward.components WHERE id=$1 FOR UPDATE`, componentID).Scan(&exists); errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	} else if err != nil {
		return "", err
	}
	var previousID *string
	err = tx.QueryRow(ctx, `SELECT id::text FROM reward.component_versions
		WHERE reward_component_id=$1 AND effective_from < $2
		ORDER BY effective_from DESC LIMIT 1 FOR UPDATE`, componentID, input.EffectiveFrom).Scan(&previousID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	var overlap bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM reward.component_versions WHERE reward_component_id=$1
		AND tstzrange(effective_from,effective_to,'[)') &&
		    tstzrange($2,$3::timestamptz,'[)')
		AND ($4::uuid IS NULL OR id<>$4)
	)`, componentID, input.EffectiveFrom, input.EffectiveTo, previousID).Scan(&overlap); err != nil {
		return "", err
	}
	if overlap {
		return "", domain.ErrInvalidInput
	}
	if previousID != nil {
		if _, err := tx.Exec(ctx, `UPDATE reward.component_versions SET effective_to=$2
			WHERE id=$1 AND (effective_to IS NULL OR effective_to>$2)`, *previousID, input.EffectiveFrom); err != nil {
			return "", err
		}
	}
	id := secure.UUID()
	if _, err := tx.Exec(ctx, `INSERT INTO reward.component_versions(
		id,reward_component_id,reward_unit_id,name,description,effect_type,reward_value,effective_from,effective_to,
		announced_at,published_at,supersedes_version_id,change_reason,display_change_until)
		VALUES ($1,$2,$3,$4,$5,$6,$7::numeric,$8,$9,$10,now(),$11,$12,$13)`,
		id, componentID, input.RewardUnitID, strings.TrimSpace(input.Name), strings.TrimSpace(input.Description),
		input.EffectType, input.RewardValue, input.EffectiveFrom, input.EffectiveTo, input.AnnouncedAt,
		previousID, strings.TrimSpace(input.ChangeReason), input.DisplayChangeUntil); err != nil {
		return "", err
	}
	payload, _ := json.Marshal(map[string]any{"component_id": componentID, "version_id": id, "effective_from": input.EffectiveFrom})
	if _, err := tx.Exec(ctx, `INSERT INTO integration.outbox_events(id,aggregate_type,aggregate_id,event_type,payload_json)
		VALUES ($1,'reward_component',$2,'RewardComponentVersionPublished',$3)`, secure.UUID(), componentID, payload); err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}

func (r *Repository) PublishConditionVersion(ctx context.Context, conditionID string, input domain.RewardConditionVersionInput) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT true FROM reward.conditions WHERE id=$1 FOR UPDATE`, conditionID).Scan(&exists); errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	} else if err != nil {
		return "", err
	}
	previousID, err := closePreviousVersion(ctx, tx, "reward.condition_versions", "reward_condition_id", conditionID, input.EffectiveFrom, input.EffectiveTo)
	if err != nil {
		return "", err
	}
	id := secure.UUID()
	if _, err := tx.Exec(ctx, `INSERT INTO reward.condition_versions(
		id,reward_condition_id,operator,configuration_json,description,effective_from,effective_to,published_at,supersedes_version_id)
		VALUES($1,$2,$3,$4,$5,$6,$7,now(),$8)`, id, conditionID, input.Operator, input.Configuration,
		strings.TrimSpace(input.Description), input.EffectiveFrom, input.EffectiveTo, previousID); err != nil {
		return "", err
	}
	if err := insertCatalogOutbox(ctx, tx, "reward_condition", conditionID, "RewardConditionVersionPublished", id, input.EffectiveFrom); err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}

func (r *Repository) PublishCapVersion(ctx context.Context, capID string, input domain.RewardCapVersionInput) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT true FROM reward.caps WHERE id=$1 FOR UPDATE`, capID).Scan(&exists); errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	} else if err != nil {
		return "", err
	}
	previousID, err := closePreviousVersion(ctx, tx, "reward.cap_versions", "reward_cap_id", capID, input.EffectiveFrom, input.EffectiveTo)
	if err != nil {
		return "", err
	}
	id := secure.UUID()
	if _, err := tx.Exec(ctx, `INSERT INTO reward.cap_versions(
		id,reward_cap_id,cap_type,limit_value,reward_unit_id,period_type,effective_from,effective_to,supersedes_version_id)
		VALUES($1,$2,$3,$4::numeric,$5,$6,$7,$8,$9)`, id, capID, input.CapType, input.LimitValue,
		input.RewardUnitID, input.PeriodType, input.EffectiveFrom, input.EffectiveTo, previousID); err != nil {
		return "", err
	}
	if err := insertCatalogOutbox(ctx, tx, "reward_cap", capID, "RewardCapVersionPublished", id, input.EffectiveFrom); err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}

func closePreviousVersion(ctx context.Context, tx pgx.Tx, table, ownerColumn, ownerID string, effectiveFrom time.Time, effectiveTo *time.Time) (*string, error) {
	query := fmt.Sprintf(`SELECT id::text FROM %s WHERE %s=$1 AND effective_from < $2 ORDER BY effective_from DESC LIMIT 1 FOR UPDATE`, table, ownerColumn)
	var previousID *string
	if err := tx.QueryRow(ctx, query, ownerID, effectiveFrom).Scan(&previousID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	overlapQuery := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s=$1
		AND tstzrange(effective_from,effective_to,'[)') && tstzrange($2,$3::timestamptz,'[)')
		AND ($4::uuid IS NULL OR id<>$4))`, table, ownerColumn)
	var overlap bool
	if err := tx.QueryRow(ctx, overlapQuery, ownerID, effectiveFrom, effectiveTo, previousID).Scan(&overlap); err != nil {
		return nil, err
	}
	if overlap {
		return nil, domain.ErrInvalidInput
	}
	if previousID != nil {
		updateQuery := fmt.Sprintf(`UPDATE %s SET effective_to=$2 WHERE id=$1 AND (effective_to IS NULL OR effective_to>$2)`, table)
		if _, err := tx.Exec(ctx, updateQuery, *previousID, effectiveFrom); err != nil {
			return nil, err
		}
	}
	return previousID, nil
}

func insertCatalogOutbox(ctx context.Context, tx pgx.Tx, aggregateType, aggregateID, eventType, versionID string, effectiveFrom time.Time) error {
	payload, _ := json.Marshal(map[string]any{"aggregate_id": aggregateID, "version_id": versionID, "effective_from": effectiveFrom})
	_, err := tx.Exec(ctx, `INSERT INTO integration.outbox_events(id,aggregate_type,aggregate_id,event_type,payload_json)
		VALUES($1,$2,$3,$4,$5)`, secure.UUID(), aggregateType, aggregateID, eventType, payload)
	return err
}
