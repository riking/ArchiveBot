
CREATE TYPE platform_enum AS enum('test_harness', 'discord');

CREATE TABLE url_jobs (
	internal_id SERIAL PRIMARY KEY,
	state enum('queued', 'claimed', 'retry', 'pending', 'error', 'success') NOT NULL DEFAULT('queued'),
	platform platform_enum,
	url text NOT NULL,
	save_outlinks bool DEFAULT(false),

	claim_proc_id text NULL,  -- used to detect hung claims
	ia_job_id     text NULL,
	ia_archive_ts text(15) NULL,  -- used to construct /web/ links. only present for 'success'

	submitted_at timestamp with zone DEFAULT(current_timestamp),
	terminated_at timestamp with zone NULL, -- error, success are terminal states
	was_already_archived bool DEFAULT(false),
)

INSERT INTO url_jobs(state, platform, url) VALUES ('queued', 'test_harness', 'https://example.com/');


BEGIN;
ALTER TYPE platform_enum ADD VALUE 'telegram';
ALTER TABLE url_jobs ALTER COLUMN platform TYPE platform_enum USING platform::text::platform_enum;
COMMIT;

