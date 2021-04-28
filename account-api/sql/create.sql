CREATE TABLE Users (
	email VARCHAR(255) UNIQUE NOT NULL,
	password VARCHAR(255) NOT NULL,
	firstname VARCHAR (255) NOT NULL, 
	lastname VARCHAR (255) NOT NULL,
	ip VARCHAR (255) ,
	user_agent VARCHAR (255),
	role VARCHAR (255)
);

CREATE TABLE Wallets (
      address VARCHAR(255) UNIQUE NOT NULL,
      email VARCHAR(255) NOT NULL, 
      currency VARCHAR(50) NOT NULL,
      balance NUMERIC(32,6) NOT NULL
);

CREATE TABLE Auths (
    email VARCHAR(255) UNIQUE NOT NULL,
    date TIMESTAMP,
    attempt INT, 
    active TIMESTAMP,
    jwt VARCHAR(255)
);

CREATE TABLE Mail (
    email VARCHAR(255) UNIQUE NOT NULL,
    confirmed BOOLEAN,
    link VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE Transaction (
    amount FLOAT,
    commission FLOAT,
    date TIMESTAMP,
    currency VARCHAR(50),
    sender VARCHAR(255),
    recipient VARCHAR(255)
);