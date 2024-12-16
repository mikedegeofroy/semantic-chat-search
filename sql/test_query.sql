SELECT
  *
FROM
  "public"."blog_contents_embeddings_store"
LIMIT 10;

select COUNT(*) from messages_contents_embeddings_store;

SELECT * FROM ai.vectorizer_status;