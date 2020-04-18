CREATE DATABASE IF NOT EXISTS `rds`;

CREATE USER IF NOT EXISTS 'write_user'@'%' IDENTIFIED BY '';
GRANT INSERT ON *.* TO 'write_user'@'%';

CREATE USER IF NOT EXISTS 'read_user'@'%' IDENTIFIED BY '';
GRANT SELECT ON *.* TO 'read_user'@'%';

CREATE USER IF NOT EXISTS 'delete_user'@'%' IDENTIFIED BY '';
GRANT DELETE ON *.* TO 'read_user'@'%';

CREATE TABLE IF NOT EXISTS redirects
(
    token      varchar(20)  not null primary key,
    url        varchar(255) not null,
    user_id    varchar(50)  not null,
    created_at datetime     not null,
    constraint Redirects_token_uindex
        unique (token)
);

alter table redirects
    add primary key (token);

CREATE TABLE IF NOT EXISTS stats
(
    id         int          not null
        primary key,
    useragent  varchar(255) null,
    ip         varchar(40)  not null,
    city       varchar(100) null,
    created_at datetime     not null,
    token      varchar(100) null
);
