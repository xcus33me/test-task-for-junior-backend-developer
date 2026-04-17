CREATE TABLE IF NOT EXISTS schedules (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    recurrence_type TEXT NOT NULL,
    every_n_days INT,
    day_of_month INT,
    specific_dates DATE[],
    start_date DATE NOT NULL,
    end_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_schedules_is_active ON schedules (is_active);

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS schedule_id BIGINT REFERENCES schedules(id) ON DELETE SET NULL;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS scheduled_date DATE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_schedule_date
    ON tasks (schedule_id, scheduled_date)
    WHERE schedule_id IS NOT NULL;
