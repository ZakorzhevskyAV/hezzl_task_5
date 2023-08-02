CREATE TABLE IF NOT EXISTS PROJECTS(
id integer UNIQUE NOT NULL PRIMARY KEY,
name varchar(40) NOT NULL,
created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS GOODS(
id integer UNIQUE NOT NULL PRIMARY KEY,
project_id integer UNIQUE NOT NULL,
name varchar(40) NOT NULL,
description varchar(40),
priority SERIAL,
removed BOOLEAN DEFAULT FALSE,
created_at TIMESTAMP DEFAULT NOW(),
CONSTRAINT fk_project_id FOREIGN KEY (project_id) REFERENCES PROJECTS(id)
);
INSERT INTO PROJECTS(id, name) VALUES (1, 'First entry');
INSERT INTO GOODS(id, project_id, name, description) VALUES (1, 1, 'good_1', 'First entry');
