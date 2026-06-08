Create Database PackageDB;

Create table package (
    id int primary key auto_increment,
    name varchar(255) not null,
    description text,
    url varchar(255),
    platform varchar(255),
    register_time datetime default current_timestamp,
    create_time datetime default current_timestamp,
    update_time datetime default current_timestamp on update current_timestamp,
    uk_name unique (name)
);