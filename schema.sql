DROP DATABASE IF EXISTS ovto CASCADE;
CREATE DATABASE IF NOT EXISTS ovto;
SET DATABASE = ovto;

CREATE TABLE IF NOT EXISTS users
(
    id              SERIAL  NOT NULL PRIMARY KEY,
    email           VARCHAR NOT NULL UNIQUE,
    fullname        VARCHAR(50) NOT NULL UNIQUE,
    avatar          VARCHAR,
    address         VARCHAR(255),
    phone           VARCHAR
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS credentials
(
    id              SERIAL  NOT NULL PRIMARY KEY,
    user_id         INT NOT NULL REFERENCES users,
    password        VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS foodprovider
(
    id              SERIAL  NOT NULL PRIMARY KEY,
    email           VARCHAR NOT NULL UNIQUE,
    fullname        VARCHAR(50) NOT NULL UNIQUE,
    phone           VARCHAR NOT NULL,
    password        VARCHAR NOT NULL
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS restuarant
(
    id              SERIAL  NOT NULL PRIMARY KEY,
    name            VARCHAR(50) NOT NULL UNIQUE,
    owner_id        INT NOT NULL REFERENCES foodprovider,
    about           VARCHAR(50) NOT NULL UNIQUE,
    location        VARCHAR NOT NULL,
    city            VARCHAR NOT NULL,
    area            VARCHAR NOT NULL,
    country         VARCHAR NOT NULL,
    phone           VARCHAR NOT NULL,
    opening_time    VARCHAR NOT NULL,
    closing_time    VARCHAR NOT NULL,
    ambassador_code VARCHAR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- INSERT INTO users (id, email, fullname)
-- VALUES (1, 'jon@example.org', 'jon snow'),
--        (2, 'jane@example.org', 'night king');