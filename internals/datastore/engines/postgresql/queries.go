package postgresql

const (
	createExtensionQuery = `
		CREATE EXTENSION IF NOT EXISTS pg_trgm;
	`
	createTableQuery = `
		CREATE TABLE IF NOT EXISTS redix_data_v5 (
			_id 		BIGSERIAL PRIMARY KEY,
			_expires_at BIGINT,
			_key 		TEXT,
			_value 		JSONB
		);
	`
	createSetTableQuery = `
		CREATE TABLE IF NOT EXISTS redis_set (
			key TEXT NOT NULL,
			member TEXT NOT NULL,
			expires_at BIGINT DEFAULT 0,
			PRIMARY KEY (key, member)
		);
	`
	createUniqueIndexQuery = `
		CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_redix_data_v5_key ON redix_data_v5 (_key);
	`
	createTrgmIndexQuery = `
		CREATE INDEX IF NOT EXISTS trgm_idx_redix_data_v5_key ON redix_data_v5 USING GIN(_key gin_trgm_ops);
	`
	createExpiresAtIndexQuery = `
		CREATE INDEX IF NOT EXISTS idx_redix_data_v5_expires_at ON redix_data_v5 (_expires_at);
	`
	createSetExpiresAtIndexQuery = `
		CREATE INDEX IF NOT EXISTS idx_redis_set_expires_at ON redis_set (expires_at);
	`
	createSetKeyIndexQuery = `
		CREATE INDEX IF NOT EXISTS idx_redis_set_key ON redis_set (key);
	`

	deleteExpiredKeysQuery = `
		DELETE FROM redix_data_v5 WHERE _expires_at != 0 and _expires_at <= $1
	`

	deleteAllKeysQuery = `
		DELETE FROM redix_data_v5
	`

	deleteAllSetsQuery = `
		DELETE FROM redis_set
	`

	deleteMatchingKeysQuery = `
		DELETE FROM redix_data_v5 WHERE _key LIKE $1
	`

	deleteMatchingSetsQuery = `
		DELETE FROM redis_set WHERE key LIKE $1
	`

	insertQuery = `
		INSERT INTO redix_data_v5(_key, _value, _expires_at) VALUES($1, $2, $3)
	`

	onConflictDoNothing = `
		"ON CONFLICT (_key) DO NOTHING"
	`

	incrementInsertQuery = `
		INSERT INTO redix_data_v5(_key, _value, _expires_at) VALUES($1, $2, $3) ON CONFLICT (_key) DO UPDATE SET _value = (EXCLUDED._value::text::float + redix_data_v5._value::text::float)::text::jsonb
	`

	appendInsertQuery = `
		INSERT INTO redix_data_v5(_key, _value, _expires_at) VALUES($1, $2, $3) ON CONFLICT (_key) DO UPDATE SET _value = (redix_data_v5._value::text || EXCLUDED._value::text)::jsonb
	`

	updateQuery = `
		UPDATE redix_data_v5 SET _value = $2::jsonb, _expires_at = $3::bigint WHERE _key = $1
	`

	expiresAtUpdateQuery = `
		, _expires_at = $3::bigint
	`

	returningQuery = `
		RETURNING _value, _expires_at
	`

	selectQuery = `
		SELECT _value, _expires_at FROM redix_data_v5 WHERE _key = $1
	`

	deleteQuery = `
		DELETE FROM redix_data_v5 WHERE _key = $1
	`

	selectWhereQuery = `
		SELECT _key, _value, _expires_at FROM redix_data_v5 WHERE _key LIKE $1 ORDER BY _id ASC
	`

	selectNotifyQuery = `
		SELECT pg_notify($1, $2)
	`

	listenQuery = `
		LISTEN %s
	`

	// Set-related queries
	deleteExpiredSetMembersQuery = `
		DELETE FROM redis_set WHERE expires_at != 0 and expires_at <= $1
	`

	insertSetMemberQuery = `
		INSERT INTO redis_set(key, member, expires_at) VALUES($1, $2, $3) ON CONFLICT (key, member) DO NOTHING
	`

	selectSetMembersQuery = `
		SELECT member FROM redis_set WHERE key = $1 AND (expires_at = 0 OR expires_at > $2)
	`

	countSetMembersQuery = `
		SELECT COUNT(*) FROM redis_set WHERE key = $1 AND (expires_at = 0 OR expires_at > $2)
	`

	updateSetExpirationQuery = `
		UPDATE redis_set SET expires_at = $2 WHERE key = $1
	`

	deleteSetQuery = `
		DELETE FROM redis_set WHERE key = $1
	`

	checkSetExistsQuery = `
		SELECT COUNT(*) FROM redis_set WHERE key = $1 AND (expires_at = 0 OR expires_at > $2) LIMIT 1
	`
)
