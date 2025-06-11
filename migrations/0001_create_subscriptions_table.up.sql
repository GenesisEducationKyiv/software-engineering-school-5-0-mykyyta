CREATE TABLE subscriptions (
                               id TEXT PRIMARY KEY,
                               email TEXT NOT NULL UNIQUE,
                               city TEXT NOT NULL,
                               frequency TEXT NOT NULL,
                               is_confirmed BOOLEAN DEFAULT FALSE,
                               is_unsubscribed BOOLEAN DEFAULT FALSE,
                               token TEXT NOT NULL,
                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);