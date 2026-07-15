PRAGMA foreign_keys = ON;


CREATE TABLE IF NOT EXISTS firmware (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    size INTEGER NOT NULL,
    version TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    path TEXT NOT NULL,
    is_current INTEGER NOT NULL DEFAULT 0 CHECK(is_current IN (0,1)),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_firmware_name_version
ON firmware(name, version);

CREATE UNIQUE INDEX IF NOT EXISTS idx_firmware_current
ON firmware(is_current)
WHERE is_current = 1;


CREATE TABLE IF NOT EXISTS flasher (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    os TEXT NOT NULL ,
    version TEXT NOT NULL,
    size INTEGER NOT NULL, 
    sha256 TEXT NOT NULL,
    path TEXT NOT NULL,
    is_current INTEGER NOT NULL DEFAULT 0 CHECK(is_current IN (0,1)),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
 
CREATE UNIQUE INDEX IF NOT EXISTS idx_flasher_name_os_version
ON flasher(name, os, version);

CREATE INDEX IF NOT EXISTS idx_flasher_os
ON flasher(os);

CREATE UNIQUE INDEX IF NOT EXISTS idx_flasher_os_current
ON flasher(os)
WHERE is_current = 1;