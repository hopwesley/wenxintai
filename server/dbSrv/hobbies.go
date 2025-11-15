package dbSrv

import (
	"context"
)

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
