CREATE SCHEMA restore;
SET search_path TO restore;
-- begin table creation

DO $$ BEGIN
	CREATE TYPE platform_enum AS enum('test_harness', 'discord');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
	CREATE TYPE url_job_state AS enum('queued', 'claimed', 'retry', 'pending', 'error', 'success');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

CREATE TABLE url_jobs (
	internal_id SERIAL PRIMARY KEY,
	state url_job_state NOT NULL DEFAULT('queued'),
	platform platform_enum,
	url text NOT NULL,
	save_outlinks bool DEFAULT(false),

	claim_proc_id text NULL,  -- used to detect hung claims
	ia_job_id     uuid NULL,
	ia_archive_ts varchar(15) NULL,  -- used to construct /web/ links. only present for 'success'

	submitted_at timestamptz DEFAULT(current_timestamp),
	terminated_at timestamptz NULL, -- error, success are terminal states
	was_already_archived bool DEFAULT(false)
);

INSERT INTO url_jobs(state, platform, url) VALUES ('queued', 'test_harness', 'https://example.com/');


ALTER TYPE platform_enum ADD VALUE IF NOT EXISTS 'telegram';

ALTER TABLE url_jobs ALTER COLUMN platform TYPE platform_enum USING platform::text::platform_enum;

