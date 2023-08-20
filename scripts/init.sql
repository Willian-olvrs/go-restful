CREATE DATABASE gorestful;
create extension if not exists pg_trgm;

CREATE TABLE pessoa (
		id uuid PRIMARY KEY,
		apelido varchar(32) UNIQUE NOT NULL,
		nome varchar(100) NOT NULL,
		nascimento date NOT NULL,
		stack text,
		search_p text NOT NULL
	);

CREATE INDEX index_pessoa_search ON pessoa USING gin (search_p gin_trgm_ops);
--CREATE INDEX index_pessoa_apelido ON pessoa USING gin (apelido gin_trgm_ops;
--CREATE INDEX index_pessoa_nome ON pessoa USING gin (nome gin_trgm_ops);
--CREATE INDEX index_pessoa_apelido ON pessoa USING GIN (to_tsvector('english', apelido));
--CREATE INDEX index_pessoa_nome ON pessoa USING GIN (to_tsvector('english', nome));

/*
CREATE TABLE ling (
		id SERIAL PRIMARY KEY,
		ling varchar(32) NOT NULL
	);
	
CREATE INDEX index_ling_ling ON ling USING gin (to_tsvector('english', ling));

CREATE TABLE stack (
		id_pessoa uuid NOT NULL,
		id_ling integer NOT NULL,
		PRIMARY KEY (id_pessoa, id_ling),
		CONSTRAINT fk_pessoa FOREIGN KEY (id_pessoa) REFERENCES pessoa(id),
		CONSTRAINT fk_ling FOREIGN KEY (id_ling) REFERENCES ling(id)
	);

INSERT INTO pessoa (id, apelido, nome, nascimento) VALUES ('f7379ae8-8f9b-4cd5-8221-51efe19e721b', 'josé', 'José Roberto', '2000-10-01');
INSERT INTO pessoa (id, apelido, nome, nascimento) VALUES ('5ce4668c-4710-4cfb-ae5f-38988d6d49cb', 'ana', 'Ana Barbosa', '1985-09-23');
INSERT INTO ling(ling) VALUES ('C#');
INSERT INTO ling(ling) VALUES ('Node');
INSERT INTO ling(ling) VALUES ('Oracle');
INSERT INTO ling(ling) VALUES ('Postgres');

INSERT INTO stack(id_pessoa, id_ling) VALUES ('f7379ae8-8f9b-4cd5-8221-51efe19e721b', 1);
INSERT INTO stack(id_pessoa, id_ling) VALUES ('f7379ae8-8f9b-4cd5-8221-51efe19e721b', 2);
INSERT INTO stack(id_pessoa, id_ling) VALUES ('f7379ae8-8f9b-4cd5-8221-51efe19e721b', 3);
INSERT INTO stack(id_pessoa, id_ling) VALUES ('5ce4668c-4710-4cfb-ae5f-38988d6d49cb', 2);
INSERT INTO stack(id_pessoa, id_ling) VALUES ('5ce4668c-4710-4cfb-ae5f-38988d6d49cb', 4);

select (pessoa.id,apelido,nome,nascimento,ling) 
	from pessoa left join 
		(select * from stack right join (select * from ling where ling='C#' OR ling='Node') as ling_select ON id_ling=ling_select.id) as stack_select
	on pessoa.id=stack_select.id_pessoa
	where apelido='josé' OR nome like '%Roberto%';*/
