// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package postgres

import (
	"context"

	"golang.org/x/discovery/internal"
	"golang.org/x/discovery/internal/database"
	"golang.org/x/discovery/internal/derrors"
)

// InsertAlternativeModulePath inserts the alternative module path into the alternative_module_paths table.
func (db *DB) InsertAlternativeModulePath(ctx context.Context, alternative *internal.AlternativeModulePath) (err error) {
	derrors.Wrap(&err, "DB.InsertAlternativeModulePath(ctx, %v)", alternative)
	_, err = db.db.Exec(ctx, `
		INSERT INTO alternative_module_paths (alternative, canonical)
		VALUES($1, $2) ON CONFLICT DO NOTHING;`,
		alternative.Alternative, alternative.Canonical)
	return err
}

// DeleteAlternatives deletes all modules with the given path.
func (db *DB) DeleteAlternatives(ctx context.Context, alternativePath string) (err error) {
	derrors.Wrap(&err, "DB.DeleteAlternatives(ctx)")

	return db.db.Transact(ctx, func(db *database.DB) error {
		if _, err := db.Exec(ctx,
			`DELETE FROM modules WHERE module_path = $1;`, alternativePath); err != nil {
			return err
		}
		if _, err := db.Exec(ctx,
			`DELETE FROM imports_unique WHERE from_module_path = $1;`, alternativePath); err != nil {
			return err
		}
		if _, err := db.Exec(ctx,
			`UPDATE module_version_states SET status = $1 WHERE module_path = $2;`,
			derrors.ToHTTPStatus(derrors.AlternativeModule), alternativePath); err != nil {
			return err
		}
		return nil
	})
}

// IsAlternativeModulePath reports whether the path is an alternative path for a
// module.
func (db *DB) IsAlternativeModulePath(ctx context.Context, path string) (_ bool, err error) {
	defer derrors.Wrap(&err, "IsAlternativeModulePath(ctx, %q)", path)
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM alternative_module_paths
			WHERE alternative = $1
		);`
	row := db.db.QueryRow(ctx, query, path)
	var isAlternative bool
	err = row.Scan(&isAlternative)
	return isAlternative, err
}
