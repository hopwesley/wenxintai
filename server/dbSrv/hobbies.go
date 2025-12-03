package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type TestPlan struct {
	PlanKey     string         `json:"key"`           // plan_key
	Name        string         `json:"name"`          // name
	Price       float64        `json:"price"`         // price (NUMERIC -> float64)
	Description string         `json:"desc"`          // description
	Tag         sql.NullString `json:"tag,omitempty"` // tag，可能为 NULL
}

func (pdb *psDatabase) ListHobbies(ctx context.Context) ([]string, error) {
	const q = `
		SELECT name
		FROM app.hobbies
		WHERE enabled = TRUE
		ORDER BY display_order NULLS LAST, name
	`
	rows, err := pdb.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (pdb *psDatabase) ListTestPlans(ctx context.Context) ([]TestPlan, error) {
	const q = `
		SELECT 
		    plan_key,
		    name,
		    price,
		    description,
		    tag
		FROM app.test_plans
		ORDER BY price ASC, plan_key
	`

	rows, err := pdb.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TestPlan
	for rows.Next() {
		var p TestPlan
		if err := rows.Scan(
			&p.PlanKey,
			&p.Name,
			&p.Price,
			&p.Description,
			&p.Tag,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (pdb *psDatabase) PlanByKey(ctx context.Context, key string) (*TestPlan, error) {
	const q = `
        SELECT 
            plan_key,
            name,
            price,
            description,
            tag
        FROM app.test_plans
        WHERE plan_key = $1
    `

	var p TestPlan

	row := pdb.db.QueryRowContext(ctx, q, key)
	if err := row.Scan(
		&p.PlanKey,
		&p.Name,
		&p.Price,
		&p.Description,
		&p.Tag,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("plan %s not found", key)
		}
		return nil, err
	}

	return &p, nil
}
