-- Create activities table
CREATE TABLE IF NOT EXISTS activities (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL DEFAULT 'Sara',
    activity_name VARCHAR(100) NOT NULL,
    duration INTEGER NOT NULL,  -- base unit is a minute
    date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
