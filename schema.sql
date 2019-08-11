-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DROP DATABASE IF EXISTS ovto CASCADE;
CREATE DATABASE IF NOT EXISTS ovto;
SET DATABASE = ovto;

CREATE TABLE IF NOT EXISTS users
(
    id              SERIAL  NOT NULL PRIMARY KEY,
    email           VARCHAR NOT NULL UNIQUE,
    fullname        VARCHAR(150) NOT NULL,
    avatar          VARCHAR,
    address         VARCHAR(255),
    phone           VARCHAR,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS credentials
(
    id              SERIAL  NOT NULL PRIMARY KEY,
    user_id         INT NOT NULL REFERENCES users,
    password        VARCHAR NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS foodprovider
(
    id              SERIAL  NOT NULL PRIMARY KEY,
    email           VARCHAR NOT NULL UNIQUE,
    fullname        VARCHAR(50) NOT NULL,
    phone           VARCHAR NOT NULL UNIQUE,
    password        VARCHAR NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ambassador
(
    id              SERIAL  NOT NULL PRIMARY KEY,
    email           VARCHAR NOT NULL UNIQUE,
    fullname        VARCHAR(50) NOT NULL,
    phone           VARCHAR NOT NULL UNIQUE,
    bkash           VARCHAR,
    rocket          VARCHAR,
    password        VARCHAR NOT NULL,
    facebook        VARCHAR NOT NULL,
    city            VARCHAR NOT NULL,
    area            VARCHAR NOT NULL,
    address         VARCHAR NOT NULL,
    referral_code   varchar NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS restaurant
(
    id              UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    title           VARCHAR(50) NOT NULL UNIQUE,
    owner_id        INT NOT NULL REFERENCES foodprovider,
    avatar          VARCHAR,
    cover           VARCHAR,
    about           VARCHAR,
    location        VARCHAR NOT NULL,
    city            VARCHAR NOT NULL,
    area            VARCHAR NOT NULL,
    country         VARCHAR NOT NULL,
    phone           VARCHAR NOT NULL,
    opening_time    VARCHAR NOT NULL,
    closing_time    VARCHAR NOT NULL,
    ambassador_code VARCHAR,
    vat_reg_no      VARCHAR,
    rating          DECIMAL(1,1) DEFAULT 0.0 CHECK (rating >= 0),
    active          BOOLEAN NOT NULL DEFAULT true,
    close_status    BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS permission
(
    id              INT,
    restaurant_id   UUID,
    restaurant      VARCHAR NOT NULL,
    role            INT NOT NULL,

    PRIMARY KEY (id, restaurant_id)
);

CREATE TABLE IF NOT EXISTS restaurant_gallery
(
    id              SERIAL NOT NULL PRIMARY KEY,
    restaurant_id   UUID NOT NULL REFERENCES restaurant,
    image           VARCHAR NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()

    PRIMARY KEY (id, restaurant_id)
);

CREATE TABLE IF NOT EXISTS category
(
    id              SERIAL NOT NULL PRIMARY KEY,
    name            VARCHAR(25) NOT NULL,
    availability    BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()

    PRIMARY KEY (id, restaurant_id)
);

CREATE TABLE IF NOT EXISTS item
(
    id              SERIAL NOT NULL PRIMARY KEY,
    restaurant_id   UUID NOT NULL REFERENCES restaurant,
    category_id     INT NOT NULL REFERENCES category,
    name            VARCHAR(25) NOT NULL,
    description     VARCHAR(255) NOT NULL,
    price           DECIMAL(1,1) NOT NULL CHECK (price >= 0),
    availability    BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()

    PRIMARY KEY (id, restaurant_id)
);

-- INSERT INTO users (id, email, fullname)
-- VALUES (1, 'jon@example.org', 'jon snow'),
--        (2, 'jane@example.org', 'night king');
