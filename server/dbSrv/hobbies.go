package dbSrv

import (
	"context"
	"database/sql"
)

type TestPlan struct {
	PlanKey     string         // plan_key
	Name        string         // name
	Price       float64        // price (NUMERIC -> float64)
	Description string         // description
	Tag         sql.NullString // tag，可能为 NULL
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
