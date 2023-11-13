CREATE TABLE public.urls (
	short_url varchar NOT NULL,
	full_url varchar NOT NULL,
	CONSTRAINT urls_pk PRIMARY KEY (short_url)
);



insert into public.urls (short_url, full_url)
values ('ehefjwk', 'google.com');


SELECT full_url FROM urls WHERE short_url = 'fdjfkdjf' LIMIT 1