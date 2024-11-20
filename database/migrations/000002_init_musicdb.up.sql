CREATE TABLE songs (
    id BIGSERIAL PRIMARY KEY,
    "group" VARCHAR(255) NOT NULL,
    song VARCHAR(255) NOT NULL,
    release_date DATE NOT NULL,
    text TEXT,
    link VARCHAR(1000)
);

CREATE INDEX idx_song_group ON songs ("group", song);
