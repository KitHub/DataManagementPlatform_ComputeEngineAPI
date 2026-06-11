Create Database PackageDB;

Create table package (
    id int primary key auto_increment,
    origin_id varchar(255) not null default "",
    display_name varchar(255) not null default "",
    comment text not null default "",
    platform varchar(255) not null default "",
    bucket_name varchar(255) not null default "",
    key_name varchar(255) not null default "",
    register_time datetime default current_timestamp,
    create_time datetime default current_timestamp,
    update_time datetime default current_timestamp on update current_timestamp,
    uk_origin_id unique (origin_id)
);

