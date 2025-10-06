CREATE TABLE
  environments (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE,
    sort_order integer NOT NULL DEFAULT 0
  );

CREATE TABLE
  applications (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE,
    sort_order integer NOT NULL DEFAULT 0
  );

CREATE TABLE
  instances (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE,
    environment_id integer NOT NULL REFERENCES environments (id) ON DELETE CASCADE,
    application_id integer NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    UNIQUE (environment_id, application_id)
  );

CREATE TABLE
  deployments (
    id UUID PRIMARY KEY,
    instance_id integer NOT NULL REFERENCES instances (id) ON DELETE CASCADE,
    version text NOT NULL,
    deployed_at timestamptz NOT NULL
  );