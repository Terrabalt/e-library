CREATE TABLE user_account (
	email varchar(255) NOT NULL,
	password varchar(60),
	g_id text,
	activated BOOLEAN NOT NULL DEFAULT 'false',
	name varchar(255) NOT NULL,
	CONSTRAINT user_account_pk PRIMARY KEY (email)
);

CREATE TABLE user_devices (
	user_id varchar(255) NOT NULL,
	verifier uuid NOT NULL,
	expires_in timestamptz NOT NULL,
	CONSTRAINT user_devices_pk PRIMARY KEY (user_id, verifier)
);

ALTER TABLE user_devices ADD CONSTRAINT user_devices_fk0 FOREIGN KEY (user_id) REFERENCES user_account(email) ON DELETE CASCADE;
