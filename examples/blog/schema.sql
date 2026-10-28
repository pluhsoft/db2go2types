-- Schema of the example. Generate the package with:
--   go test . -run TestExample -update
DROP SCHEMA IF EXISTS blog CASCADE;
CREATE SCHEMA blog;

CREATE TYPE blog.post_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE blog.label_color AS ENUM ('#A6D2FF', '#F87659', '#BCF1A5');

CREATE TABLE blog.authors (
    id         serial PRIMARY KEY,
    name       text NOT NULL,
    email      varchar(255),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE blog.posts (
    id           serial PRIMARY KEY,
    author_id    integer NOT NULL REFERENCES blog.authors (id) ON DELETE CASCADE,
    title        text NOT NULL,
    status       blog.post_status NOT NULL DEFAULT 'draft',
    labels       blog.label_color[] NOT NULL DEFAULT '{}',
    tags         text[],
    "order"      integer NOT NULL DEFAULT 0,
    views        bigint NOT NULL DEFAULT 0,
    rating       double precision,
    meta         jsonb,
    published_at timestamp
);

CREATE TABLE blog.post_likes (
    post_id   integer NOT NULL REFERENCES blog.posts (id) ON DELETE CASCADE,
    author_id integer NOT NULL REFERENCES blog.authors (id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, author_id)
);

CREATE VIEW blog.published_posts AS
    SELECT * FROM blog.posts WHERE status = 'published';
