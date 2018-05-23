create extension if not exists citext;

create table projects
(
  id     serial primary key,
  mantis text,
  gitlab text unique
);

create table users
(
  id    int primary key unique,
  name  citext,
  email citext not null
);
create unique index unique_user_email
  on users (lower(email));

create table aliases
(
  email citext not null primary key,
  alias citext not null
);
create unique index unique_alias_email
  on aliases (lower(email));

create table issues
(
  commit_hash citext not null,
  issue_id    int    not null,
  email       citext,
  date        timestamp without time zone default (now() :: timestamp at time zone 'utc')
);
