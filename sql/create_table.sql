CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP,
    first_name TEXT,
    last_name TEXT,
    username TEXT,
    content TEXT
);

select * from messages;
