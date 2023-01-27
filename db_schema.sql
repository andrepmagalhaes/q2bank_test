CREATE TYPE "user_type" AS ENUM (
  'person',
  'store'
);

CREATE TYPE "transaction_type" AS ENUM (
  'credit',
  'debit'
);

CREATE TABLE "Users" (
  "id" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "email" varchar(255) UNIQUE NOT NULL,
  "password" varchar(255) NOT NULL,
  "cpf" varchar(11) UNIQUE NOT NULL,
  "name" varchar(255) NOT NULL,
  "customer_type" user_type NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "is_deleted" boolean DEFAULT false
);

CREATE TABLE "Transactions" (
  "id" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "payer" int NOT NULL,
  "payee" int NOT NULL,
  "transaction_type" transaction_type NOT NULL,
  "amount" bigint NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "is_valid" boolean DEFAULT true
);

ALTER TABLE "Transactions" ADD FOREIGN KEY ("payer") REFERENCES "Users" ("id");

ALTER TABLE "Transactions" ADD FOREIGN KEY ("payee") REFERENCES "Users" ("id");
