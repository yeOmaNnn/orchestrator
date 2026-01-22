CREATE INDEX idx_tasks_status
    ON tasks(status);

CREATE INDEX idx_steps_task_id
    ON steps(task_id);

CREATE INDEX idx_steps_status
    ON steps(status);
