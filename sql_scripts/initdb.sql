CREATE TABLE IF NOT EXISTS PROJECTS(
id SERIAL PRIMARY KEY,
name varchar(40),
created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS GOODS(
id SERIAL PRIMARY KEY,
project_id integer UNIQUE NOT NULL,
name varchar(40) NOT NULL,
description varchar(40),
priority integer,
removed BOOLEAN NOT NULL DEFAULT FALSE,
created_at TIMESTAMP DEFAULT NOW(),
CONSTRAINT fk_project_id FOREIGN KEY (project_id) REFERENCES PROJECTS(id)
);
CREATE UNIQUE INDEX p_id ON PROJECTS (id);
CREATE UNIQUE INDEX g_id ON GOODS (id);
CREATE UNIQUE INDEX g_p_id ON GOODS (project_id);
CREATE INDEX g_name ON GOODS (name);
INSERT INTO PROJECTS(id, name) VALUES (1, 'First entry');
INSERT INTO GOODS(id, project_id, name, description) VALUES (1, 1, 'good_1', 'First entry');
