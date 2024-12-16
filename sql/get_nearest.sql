SELECT
    chunk,
    embedding <=>  ai.ollama_embed('nomic-embed-text', 'good food', host => 'http://ollama:11434') as distance
FROM blog_contents_embeddings
ORDER BY distance;

SELECT
    chunk,
    embedding <=>  ai.ollama_embed('nomic-embed-text', 'когда я ходил на выставку?', host => 'http://ollama:11434') as distance
FROM messages_contents_embeddings
ORDER BY distance;

SELECT
    chunk,
    embedding <=>  ai.ollama_embed('nomic-embed-text', 'когда я ходил на выставку?', host => 'http://ollama:11434') as distance
FROM messages_contents_embeddings_combined
ORDER BY distance;