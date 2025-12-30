-- Drop trigger
DROP TRIGGER IF EXISTS trigger_create_user_balance;

-- Drop tables in reverse order (due to foreign key dependencies)
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS balances;
DROP TABLE IF EXISTS users;
