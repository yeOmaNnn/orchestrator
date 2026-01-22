CREATE TABLE steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    task_id UUID NOT NULL,

    agent TEXT NOT NULL,
    input JSONB NOT NULL,
    output JSONB,

    status TEXT NOT NULL,
    retry_count INT NOT NULL DEFAULT 0,

    depends_on UUID[] NOT NULL DEFAULT '{}',

    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT fk_steps_task
        FOREIGN KEY (task_id)
        REFERENCES tasks(id)
        ON DELETE CASCADE
);
