-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create links table
CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    link_id VARCHAR(10) UNIQUE NOT NULL,
    business_unit VARCHAR(10) NOT NULL,
    network VARCHAR(50),
    total_cap INTEGER DEFAULT 0,
    current_hits INTEGER DEFAULT 0,
    backup_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create targets table
CREATE TABLE IF NOT EXISTS targets (
    id SERIAL PRIMARY KEY,
    link_id INTEGER REFERENCES links(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    weight INTEGER DEFAULT 100,
    cap INTEGER DEFAULT 0,
    current_hits INTEGER DEFAULT 0,
    countries TEXT DEFAULT '[]',
    param_mapping TEXT DEFAULT '{}',
    static_params TEXT DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create access_logs table
CREATE TABLE IF NOT EXISTS access_logs (
    id SERIAL PRIMARY KEY,
    link_id INTEGER REFERENCES links(id),
    target_id INTEGER REFERENCES targets(id),
    ip VARCHAR(45),
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(2),
    client_ip VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create templates table
CREATE TABLE IF NOT EXISTS templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    business_unit VARCHAR(10),
    network VARCHAR(50),
    total_cap INTEGER DEFAULT 0,
    backup_url TEXT,
    config JSONB,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create user_links table for access control
CREATE TABLE IF NOT EXISTS user_links (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    link_id INTEGER REFERENCES links(id) ON DELETE CASCADE,
    can_edit BOOLEAN DEFAULT false,
    can_delete BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, link_id)
);

-- Create indexes
CREATE INDEX idx_links_link_id ON links(link_id);
CREATE INDEX idx_links_business_unit ON links(business_unit);
CREATE INDEX idx_targets_link_id ON targets(link_id);
CREATE INDEX idx_access_logs_link_id ON access_logs(link_id);
CREATE INDEX idx_access_logs_ip ON access_logs(ip);
CREATE INDEX idx_access_logs_created_at ON access_logs(created_at);

-- Insert default admin user (password: admin123)
INSERT INTO users (username, email, password, role) 
VALUES ('admin', 'admin@example.com', '$2a$10$YourHashedPasswordHere', 'admin')
ON CONFLICT (username) DO NOTHING;