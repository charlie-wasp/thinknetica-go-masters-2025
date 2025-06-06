-- +goose Up
CREATE TABLE beers (
    id serial primary key,
    name VARCHAR(100) NOT NULL,
    style VARCHAR(50),
    brewery VARCHAR(100)
);

CREATE TABLE users (
    id serial primary key,
    fullname text
);

CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    beer_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    rating INTEGER,

    CONSTRAINT fk_user
        FOREIGN KEY(user_id) 
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_beer
        FOREIGN KEY(beer_id) 
        REFERENCES beers(id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE reviews;
DROP TABLE users;
DROP TABLE beers;
