CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    transaction_date DATE NOT NULL,
    currency TEXT NOT NULL,
    amount FLOAT NOT NULL,
    category TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    confirm BOOLEAN NOT NULL DEFAULT false
);

-- Feature: asset-tracking Table for storing asset tracking data

CREATE TABLE IF NOT EXISTS assets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    institution_name TEXT NOT NULL,
    institution_type TEXT NOT NULL,
    asset_name TEXT NOT NULL,
    current_value FLOAT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'INR',
    last_updated DATETIME NOT NULL DEFAULT (datetime('now')),
    description TEXT NOT NULL DEFAULT '',
    confirm BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS asset_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id INTEGER NOT NULL,
    value_date DATETIME NOT NULL DEFAULT (datetime('now')),
    value FLOAT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'INR',
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE
);
