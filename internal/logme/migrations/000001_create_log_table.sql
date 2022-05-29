CREATE TABLE IF NOT EXISTS logs (
    `uuid`       UUID DEFAULT generateUUIDv4(),
    `account_id` UInt32,
    `name`       String,
    `dt`         DateTime,
    `content`    String
) engine=MergeTree()
ORDER BY (account_id, dt)