CREATE TABLE IF NOT EXISTS user_account (
	email varchar(255) NOT NULL,
	password varchar(60),
	g_id TEXT,
	activated BOOLEAN NOT NULL DEFAULT 'false',
	activation_token uuid,
	expires_in timestamptz,
	name varchar(255) NOT NULL,
	CONSTRAINT user_account_pk PRIMARY KEY (email)
);

CREATE TABLE IF NOT EXISTS user_session (
	user_id varchar(255) NOT NULL,
	session_token uuid NOT NULL,
	expires_in timestamptz NOT NULL,
	CONSTRAINT user_session_pk PRIMARY KEY (user_id, session_token),
	CONSTRAINT user_session_fk_user_id FOREIGN KEY (user_id) REFERENCES user_account(email)
);

CREATE TABLE IF NOT EXISTS book (
	id uuid NOT NULL,
	title TEXT NOT NULL,
	author TEXT NOT NULL,
	cover_image TEXT NOT NULL,
	summary TEXT NOT NULL,
	readers_count integer NOT NULL DEFAULT '0',
	is_new bool NOT NULL DEFAULT 'true',
	is_popular bool NOT NULL DEFAULT 'false',
	CONSTRAINT book_pk PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS fav_book (
	user_id varchar(255) NOT NULL,
	book_id uuid NOT NULL,
	CONSTRAINT fav_book_pk PRIMARY KEY (user_id, book_id),
	CONSTRAINT fav_book_fk_user_id FOREIGN KEY (user_id) REFERENCES user_account(email),
	CONSTRAINT fav_book_fk_book_id FOREIGN KEY (book_id) REFERENCES book(id)
);

CREATE TABLE IF NOT EXISTS rate_book (
	user_id varchar(255) NOT NULL,
	book_id uuid NOT NULL,
	rating int8 NOT NULL,
	CONSTRAINT rate_book_pk PRIMARY KEY (user_id, book_id),
	CONSTRAINT rate_book_fk_user_id FOREIGN KEY (user_id) REFERENCES user_account(email),
	CONSTRAINT rate_book_fk_book_id FOREIGN KEY (book_id) REFERENCES book(id)
);

CREATE OR REPLACE VIEW rating_avg AS 
	SELECT 
		book.id, 
		(
			COALESCE(
				avg(rate_book.rating), 
				(0):: numeric
			)
		):: numeric(10, 2) AS rating 
	FROM 
		book 
		LEFT JOIN rate_book ON book.id = rate_book.book_id 
	GROUP BY 
		book.id;
