CREATE TABLE IF NOT EXISTS ping(
    id SERIAL,
    user_id BIGINT NOT NULL,
    url TEXT NOT NULL,
    connection_time varchar(6) default '10s',
    ping_time varchar(6) default '300s',
    FOREIGN KEY (user_id)  REFERENCES users (id) ON DELETE CASCADE
);