create table if not exists public.users
(
    id       serial      not null
        unique
        constraint users_pk
            primary key,
    username varchar(50) not null
        constraint users_pk_2
            unique,
    password text        not null,
    email    text        not null
);

create table if not exists public.models
(
    id      serial      not null
        constraint models_pk
            primary key,
    name    varchar(50) not null,
    user_id int         not null,
    constraint fk_user foreign key (user_id) references public.users (id)
);

create table if not exists public.versions
(
    id       serial not null
        constraint versions_pk
            primary key,
    number   int    not null,
    model_id int    not null,
    unique (number, model_id),
    constraint fk_model foreign key (model_id) references public.models (id) on delete cascade
);

create table if not exists public.messages
(
    id         serial      not null
        constraint messages_pk
            primary key,
    user_id    int         not null
        constraint fk_user
            references public.users (id),
    model_id   int         not null
        constraint fk_model
            references public.models (id),
    version_id   int         not null
        constraint fk_version
            references public.versions (id),
    input1 bytea not null,
    input2 bytea not null,
    results     TEXT[] not null,
    created_at timestamptz not null
);
