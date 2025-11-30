CREATE TABLE IF NOT EXISTS departments (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(127) NOT NULL,
    budget DECIMAL(15, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    employee_id VARCHAR(63) UNIQUE NOT NULL,
    name VARCHAR(63) NOT NULL,
    email VARCHAR(127) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    `role` VARCHAR(31) NOT NULL,
    project_id VARCHAR(36),
    department_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(127) NOT NULL,
    description TEXT,
    budget DECIMAL(15, 2),
    start_date DATE,
    end_date DATE,
    department_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS expenses (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    description TEXT,
    status VARCHAR(31) NOT NULL,
    purpose VARCHAR(127),
    approved_by VARCHAR(36),
    approved_at TIMESTAMP,
    reviewed_by VARCHAR(36),
    reviewed_at TIMESTAMP,
    is_reconciled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS bills (
    id VARCHAR(36) PRIMARY KEY,
    expense_id VARCHAR(36) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    description TEXT,
    attachment_url VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (expense_id) REFERENCES expenses(id)
);

CREATE TABLE IF NOT EXISTS advances (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    purpose VARCHAR(127),
    description TEXT,
    status VARCHAR(31) NOT NULL,
    reconciled_expense_id VARCHAR(36),
    approved_by VARCHAR(36),
    approved_at TIMESTAMP,
    reviewed_by VARCHAR(36),
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

ALTER TABLE users
ADD FOREIGN KEY (project_id) REFERENCES projects(id),
ADD FOREIGN KEY (department_id) REFERENCES departments(id);

ALTER TABLE projects
ADD FOREIGN KEY (department_id) REFERENCES departments(id);

ALTER TABLE expenses
ADD FOREIGN KEY (approved_by) REFERENCES users(id),
ADD FOREIGN KEY (reviewed_by) REFERENCES users(id);

ALTER TABLE advances
ADD FOREIGN KEY (reconciled_expense_id) REFERENCES expenses(id),
ADD FOREIGN KEY (approved_by) REFERENCES users(id),
ADD FOREIGN KEY (reviewed_by) REFERENCES users(id);
