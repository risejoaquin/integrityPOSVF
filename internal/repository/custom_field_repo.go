package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
)

type CustomFieldRepository struct {
	Pool *pgxpool.Pool
}

func (r *CustomFieldRepository) GetDefinitions(ctx context.Context, entity string) ([]model.CustomFieldDefinition, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, entity, field_name, field_type, options, display_order FROM custom_field_definitions WHERE entity = $1 ORDER BY display_order ASC`, entity)
	if err != nil {
		return nil, fmt.Errorf("error listing custom field definitions: %w", err)
	}
	defer rows.Close()

	var defs []model.CustomFieldDefinition
	for rows.Next() {
		var d model.CustomFieldDefinition
		if err := rows.Scan(&d.ID, &d.Entity, &d.FieldName, &d.FieldType, &d.Options, &d.DisplayOrder); err != nil {
			return nil, fmt.Errorf("error scanning custom field definition: %w", err)
		}
		defs = append(defs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return defs, nil
}

func (r *CustomFieldRepository) CreateDefinition(ctx context.Context, d *model.CustomFieldDefinition) error {
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO custom_field_definitions (entity, field_name, field_type, options, display_order) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		d.Entity, d.FieldName, d.FieldType, d.Options, d.DisplayOrder).Scan(&d.ID)
	if err != nil {
		return fmt.Errorf("error creating custom field definition: %w", err)
	}
	return nil
}

func (r *CustomFieldRepository) UpdateDefinition(ctx context.Context, d *model.CustomFieldDefinition) error {
	cmd, err := r.Pool.Exec(ctx,
		`UPDATE custom_field_definitions SET entity=$1, field_name=$2, field_type=$3, options=$4, display_order=$5 WHERE id=$6`,
		d.Entity, d.FieldName, d.FieldType, d.Options, d.DisplayOrder, d.ID)
	if err != nil {
		return fmt.Errorf("error updating custom field definition: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("custom field definition with id %s not found", d.ID)
	}
	return nil
}

func (r *CustomFieldRepository) DeleteDefinition(ctx context.Context, id string) error {
	cmd, err := r.Pool.Exec(ctx, `DELETE FROM custom_field_definitions WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("error deleting custom field definition: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("custom field definition with id %s not found", id)
	}
	return nil
}
