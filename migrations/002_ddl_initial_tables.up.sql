-- location_types table
CREATE TABLE IF NOT EXISTS location.location_types (
    id      INTEGER     PRIMARY KEY,
    type    VARCHAR     NOT NULL
);

-- suppliers table
CREATE TABLE IF NOT EXISTS location.suppliers (
   id       INTEGER     PRIMARY KEY,
   name     VARCHAR     NOT NULL
);

-- sub_location_types table
CREATE TABLE IF NOT EXISTS location.sub_location_types (
    id      INTEGER     PRIMARY KEY,
    type    VARCHAR     NOT NULL
);

-- locations
CREATE TABLE IF NOT EXISTS location.locations (
    id                      UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                    CITEXT          NOT NULL,
    location_type_id        INTEGER         NOT NULL REFERENCES location.location_types (id),
    supplier_id             INTEGER         NOT NULL REFERENCES location.suppliers (id),
    active                  BOOL            NOT NULL DEFAULT TRUE,
    created_at              timestamptz     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              timestamptz     DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS locations_name ON location.locations USING btree (name);

-- location_information
CREATE TABLE IF NOT EXISTS location.location_information (
    id                      UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    location_id             UUID            NOT NULL REFERENCES location.locations (id),
    address                 CITEXT          NOT NULL,
    city                    CITEXT          NOT NULL,
    state                   CITEXT          NOT NULL,
    zipcode                 CITEXT          NOT NULL,
    contact_person          CITEXT          NULL,
    phone_number            CITEXT          NULL,
    email                   CITEXT          NULL,
    latitude                NUMERIC         NOT NULL,
    longitude               NUMERIC         NOT NULL
);

-- sub_locations
CREATE TABLE IF NOT EXISTS location.sub_locations (
    id                      UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                    CITEXT          NOT NULL,
    location_id             UUID            NOT NULL REFERENCES location.locations (id),
    sub_location_type_id    INTEGER         NOT NULL REFERENCES location.sub_location_types (id),
    active                  BOOL            NOT NULL DEFAULT TRUE,
    created_at              timestamptz     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              timestamptz     DEFAULT NULL
);
