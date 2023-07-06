-- Creating persongdb table
create table car (
	id uuid,
	productionyear INTEGER,
	isrunning BOOLEAN,
	brand VARCHAR(30),
	primary key (id)
);
-- Creating users table
create table users (
	id uuid,
	username VARCHAR(30),
	password VARCHAR,
	refreshToken VARCHAR,
	primary key (id)
);
