CREATE TABLE instance_events (
    timestamp DateTime,
    project_id UUID,
    uuid UUID,
    type LowCardinality(String),
    state Enum('active' = 1,
               'building' = 2,
               'paused' = 3,
               'suspended' = 4,
               'stopped' = 5,
               'rescued' = 6,
               'resized' = 7,
               'soft_deleted' = 8,
               'deleted' = 9,
               'error' = 10,
               'shelved' = 11,
               'shelved_offloaded' = 12),
    image UUID
) ENGINE = MergeTree()
ORDER BY (project_id, type, state, image, timestamp);
