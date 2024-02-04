SELECT 'CREATE DATABASE main'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'main');

SELECT 'CREATE DATABASE main_test'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'main_test');
