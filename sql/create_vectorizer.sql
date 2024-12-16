SELECT ai.create_vectorizer(
     'blog'::regclass,
     destination => 'blog_contents_embeddings',
     embedding => ai.embedding_ollama('nomic-embed-text', 768),
     chunking => ai.chunking_recursive_character_text_splitter('contents')
);

SELECT ai.create_vectorizer(
     'messages'::regclass,
     destination => 'messages_contents_embeddings',
     embedding => ai.embedding_ollama('nomic-embed-text', 768),
     chunking => ai.chunking_recursive_character_text_splitter('content')
);

SELECT ai.create_vectorizer(
     'messages'::regclass,
     destination => 'messages_contents_embeddings',
     embedding => ai.embedding_ollama('nomic-embed-text', 768),
     chunking => ai.chunking_recursive_character_text_splitter('content')
);

SELECT 
    group_id, 
    string_agg(
        '[' || timestamp::text || ' - ' || username || ']: ' || content, 
        E'\n'
    ) AS combined_content
FROM (
    SELECT 
        *,
        SUM(CASE WHEN lag_username IS DISTINCT FROM username OR time_diff > INTERVAL '5 minutes' THEN 1 ELSE 0 END)
        OVER (ORDER BY timestamp ASC) AS group_id
    FROM (
        SELECT 
            *,
            LAG(username) OVER (ORDER BY timestamp ASC) AS lag_username,
            timestamp - LAG(timestamp) OVER (ORDER BY timestamp ASC) AS time_diff
        FROM messages
    ) AS subquery
) AS grouped_messages
GROUP BY group_id;


CREATE MATERIALIZED VIEW combined_messages_view AS
SELECT 
    group_id, 
    string_agg(
        '[' || timestamp::text || ' - ' || username || ']: ' || content, 
        E'\n'
    ) AS combined_content
FROM (
    SELECT 
        *,
        SUM(CASE WHEN lag_username IS DISTINCT FROM username OR time_diff > INTERVAL '5 minutes' THEN 1 ELSE 0 END)
        OVER (ORDER BY timestamp ASC) AS group_id
    FROM (
        SELECT 
            *,
            LAG(username) OVER (ORDER BY timestamp ASC) AS lag_username,
            timestamp - LAG(timestamp) OVER (ORDER BY timestamp ASC) AS time_diff
        FROM messages
    ) AS subquery
) AS grouped_messages
GROUP BY group_id;

CREATE TABLE combined_messages_table AS
SELECT 
    group_id, 
    string_agg(
        '[' || timestamp::text || ' - ' || username || ']: ' || content, 
        E'\n'
    ) AS combined_content
FROM (
    SELECT 
        *,
        SUM(CASE WHEN lag_username IS DISTINCT FROM username OR time_diff > INTERVAL '5 minutes' THEN 1 ELSE 0 END)
        OVER (ORDER BY timestamp ASC) AS group_id
    FROM (
        SELECT 
            *,
            LAG(username) OVER (ORDER BY timestamp ASC) AS lag_username,
            timestamp - LAG(timestamp) OVER (ORDER BY timestamp ASC) AS time_diff
        FROM messages
    ) AS subquery
) AS grouped_messages
GROUP BY group_id;

ALTER TABLE combined_messages_table ADD PRIMARY KEY (group_id);

select * from combined_messages_table;

SELECT ai.create_vectorizer(
    'combined_messages_table'::regclass,
    destination => 'messages_contents_embeddings_combined',
    embedding => ai.embedding_ollama('nomic-embed-text', 768),
    chunking => ai.chunking_character_text_splitter('combined_content', 128, 10, E'\n')
);

select COUNT(*) from messages_contents_embeddings_combined;



