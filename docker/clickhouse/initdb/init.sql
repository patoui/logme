CREATE DATABASE IF NOT EXISTS logs;
CREATE TABLE IF NOT EXISTS logs.logs (
    `uuid` UUID DEFAULT generateUUIDv4(),
    `account_id` UInt32,
    `name` String,
    `content` String,
    `dt` DateTime,
    `recorded_at` DateTime
) engine = MergeTree()
ORDER BY (account_id, dt);

CREATE DATABASE IF NOT EXISTS logs_test;
CREATE TABLE IF NOT EXISTS logs_test.logs (
    `uuid` UUID DEFAULT generateUUIDv4(),
    `account_id` UInt32,
    `name` String,
    `content` String,
    `dt` DateTime,
    `recorded_at` DateTime
) engine = MergeTree()
ORDER BY (account_id, dt);