-- Create database if not exists
CREATE DATABASE smart_redirect_dev;

-- Connect to the database
\c smart_redirect_dev;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create links table
CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    link_id VARCHAR(10) UNIQUE NOT NULL,
    business_unit VARCHAR(10) NOT NULL,
    network VARCHAR(50) NOT NULL,
    total_cap INTEGER DEFAULT 0,
    current_hits INTEGER DEFAULT 0,
    backup_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create targets table
CREATE TABLE IF NOT EXISTS targets (
    id SERIAL PRIMARY KEY,
    link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    weight INTEGER DEFAULT 100 CHECK (weight >= 0 AND weight <= 100),
    cap INTEGER DEFAULT 0,
    current_hits INTEGER DEFAULT 0,
    countries JSONB,
    param_mapping JSONB,
    static_params JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create access_logs table
CREATE TABLE IF NOT EXISTS access_logs (
    id SERIAL PRIMARY KEY,
    link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    target_id INTEGER REFERENCES targets(id) ON DELETE SET NULL,
    ip_address INET NOT NULL,
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(2),
    region VARCHAR(100),
    city VARCHAR(100),
    params JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create link_permissions table
CREATE TABLE IF NOT EXISTS link_permissions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    link_id INTEGER NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    permission VARCHAR(20) DEFAULT 'read' CHECK (permission IN ('read', 'write', 'admin')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, link_id)
);

-- Create link_templates table
CREATE TABLE IF NOT EXISTS link_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    business_unit VARCHAR(10) NOT NULL,
    network VARCHAR(50) NOT NULL,
    total_cap INTEGER DEFAULT 0,
    backup_url TEXT,
    template_targets JSONB NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_links_link_id ON links(link_id);
CREATE INDEX IF NOT EXISTS idx_links_business_unit_network ON links(business_unit, network);
CREATE INDEX IF NOT EXISTS idx_targets_link_id ON targets(link_id);
CREATE INDEX IF NOT EXISTS idx_targets_is_active ON targets(is_active);
CREATE INDEX IF NOT EXISTS idx_access_logs_link_id ON access_logs(link_id);
CREATE INDEX IF NOT EXISTS idx_access_logs_created_at ON access_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_access_logs_ip_address ON access_logs(ip_address);

-- Insert default admin user (password: admin123)
INSERT INTO users (username, email, password, role) VALUES 
('admin', 'admin@smartredirect.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin')
ON CONFLICT (username) DO NOTHING;

-- Insert sample data for testing
INSERT INTO links (link_id, business_unit, network, total_cap, backup_url) VALUES 
('abc123', 'bu01', 'mi', 1000, 'https://backup.example.com'),
('def456', 'bu02', 'google', 2000, 'https://backup2.example.com')
ON CONFLICT (link_id) DO NOTHING;

-- Insert sample targets
INSERT INTO targets (link_id, url, weight, cap, countries, param_mapping, static_params) VALUES 
(1, 'https://target1.example.com', 70, 500, '["US","CA"]', '{"kw":"q"}', '{"ref":"test"}'),
(1, 'https://target2.example.com', 30, 300, '["UK","DE"]', '{}', '{}'),
(2, 'https://target3.example.com', 50, 800, '["US"]', '{"keyword":"search"}', '{"source":"redirect"}'),
(2, 'https://target4.example.com', 50, 800, '["ALL"]', '{}', '{}')
ON CONFLICT DO NOTHING;