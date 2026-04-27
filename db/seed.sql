-- GoGopher Arch: Database & Transactions tasks (Day 6-7) seed data
-- Run: psql -h localhost -U user -d gogopher -f db/seed.sql

-- Employees table for Day 6: database connection & queries
CREATE TABLE IF NOT EXISTS employees (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    department VARCHAR(100) NOT NULL,
    salary INT NOT NULL CHECK (salary > 0),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

DELETE FROM employees;

INSERT INTO employees (name, department, salary) VALUES
    ('Ming', 'Engineering', 120000),
    ('Yan', 'Engineering', 115000),
    ('Li', 'Product', 105000),
    ('Wang', 'Design', 95000),
    ('Zhang', 'Engineering', 130000);

-- Accounts table for Day 7: transactions
CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    holder VARCHAR(100) NOT NULL,
    balance INT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    updated_at TIMESTAMP DEFAULT NOW()
);

DELETE FROM accounts;

INSERT INTO accounts (holder, balance) VALUES
    ('Ming', 5000),
    ('Yan', 3000);
