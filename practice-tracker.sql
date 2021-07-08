--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner:
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner:
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: users; Type: TABLE; Schema: public; Owner: hasmik
--

CREATE TABLE public.users (
    id public.citext DEFAULT ('us-'::text || encode(public.gen_random_bytes(6), 'hex'::text)) UNIQUE NOT NULL,
    email public.citext NOT NULL,
    user_type text NOT NULL DEFAULT 'admin',
    activated_at timestamp without time zone,
    email_verified boolean DEFAULT false,
    password text,
    first_name text,
    last_name text,
    auth_method text,
    reset_password_token text,
    email_verify_token text,
    reset_password_token_expires timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);


CREATE TABLE public.groups (
    id public.citext DEFAULT ('ug-'::text || encode(public.gen_random_bytes(6), 'hex'::text)) UNIQUE NOT NULL,
    user_id public.citext NOT NULL,
    group_name text,
    descript text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.group_users (
    group_id public.citext NOT NULL,
    user_id public.citext NOT NULL,
    permission text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.tickets (
    id public.citext DEFAULT ('tick-'::text || encode(public.gen_random_bytes(6), 'hex'::text)) UNIQUE NOT NULL,
    user_id public.citext NOT NULL,
    topic text,
    repo text,
    status_info text,
    summary text,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.group_tickets (
    group_id public.citext NOT NULL,
    ticket_id public.citext NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.languages (
    id public.citext DEFAULT ('lang-'::text || encode(public.gen_random_bytes(6), 'hex'::text)) UNIQUE NOT NULL,
    lang_name text
);

CREATE TABLE public.ticket_lang (
    ticket_id public.citext NOT NULL,
    lang_id public.citext NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.technologies (
    id public.citext DEFAULT ('tech-'::text || encode(public.gen_random_bytes(6), 'hex'::text)) UNIQUE NOT NULL,
    tech_name text
);

CREATE TABLE public.ticket_tech (
    ticket_id public.citext NOT NULL,
    tech_id public.citext NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.yt_channels (
    id public.citext DEFAULT ('ytch-'::text || encode(public.gen_random_bytes(6), 'hex'::text)) UNIQUE NOT NULL,
    ytch_name text
);

CREATE TABLE public.ticket_yt_channels (
    ticket_id public.citext NOT NULL,
    yt_channels_id public.citext NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.resources (
    id public.citext DEFAULT ('res-'::text || encode(public.gen_random_bytes(6), 'hex'::text)) UNIQUE NOT NULL,
    rscs_name text
);

CREATE TABLE public.ticket_resources (
    ticket_id public.citext NOT NULL,
    resources_id public.citext NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.sources (
    id public.citext DEFAULT ('src-'::text || encode(public.gen_random_bytes(6), 'hex'::text)) UNIQUE NOT NULL,
    src_name text
);

CREATE TABLE public.ticket_sources (
    ticket_id public.citext NOT NULL,
    sources_id public.citext NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);

CREATE TABLE public.docs (
    id public.citext DEFAULT ('doc-'::text || encode(public.gen_random_bytes(6), 'hex'::text)) UNIQUE NOT NULL,
    docs_name text
);

CREATE TABLE public.ticket_docs (
    ticket_id public.citext NOT NULL,
    docs_id public.citext NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);
--
-- public.users table
--
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users ADD CONSTRAINT users_email_key UNIQUE (email);
--
-- public.groups table
--
ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX group_name_constraint ON public.groups USING btree ("user_id", "group_name") WHERE ("deleted_at" IS NULL);
--
-- public.tickets table
--
ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_pkey PRIMARY KEY (id);
--
-- public.groups_users table
--
CREATE UNIQUE INDEX group_users_constraint ON public.group_users USING btree ("user_id", "group_id") WHERE ("deleted_at" IS NULL);

ALTER TABLE ONLY public.group_users
    ADD CONSTRAINT group_users_user_id_fkey FOREIGN KEY ("user_id") REFERENCES public.users(id);

ALTER TABLE ONLY public.group_users
    ADD CONSTRAINT group_users_group_id_fkey FOREIGN KEY ("group_id") REFERENCES public.groups(id);
--
-- public.groups_tickets table
--
CREATE UNIQUE INDEX group_tickets_constraint ON public.group_tickets USING btree ("ticket_id", "group_id") WHERE ("deleted_at" IS NULL);

ALTER TABLE ONLY public.group_tickets
    ADD CONSTRAINT group_tickets_group_id_fkey FOREIGN KEY ("group_id") REFERENCES public.groups(id);

ALTER TABLE ONLY public.group_tickets
    ADD CONSTRAINT group_tickets_ticket_id_fkey FOREIGN KEY ("ticket_id") REFERENCES public.tickets(id);

-- public.ticket_lang table

CREATE UNIQUE INDEX ticket_lang_constraint ON public.ticket_lang USING btree ("ticket_id", "lang_id") WHERE ("deleted_at" IS NULL);

ALTER TABLE ONLY public.ticket_lang
    ADD CONSTRAINT ticket_lang_ticket_id_fkey FOREIGN KEY ("ticket_id") REFERENCES public.tickets(id);

ALTER TABLE ONLY public.ticket_lang
    ADD CONSTRAINT ticket_lang_language_id_fkey FOREIGN KEY ("lang_id") REFERENCES public.languages(id);
--
-- public.ticket_tech table
--
CREATE UNIQUE INDEX ticket_tech_constraint ON public.ticket_tech USING btree ("ticket_id", "tech_id") WHERE ("deleted_at" IS NULL);

ALTER TABLE ONLY public.ticket_tech
    ADD CONSTRAINT ticket_tech_ticket_id_fkey FOREIGN KEY ("ticket_id") REFERENCES public.tickets(id);

ALTER TABLE ONLY public.ticket_tech
    ADD CONSTRAINT ticket_tech_id_fkey FOREIGN KEY ("tech_id") REFERENCES public.technologies(id);
--
-- public.ticket_yt_channels table
--
CREATE UNIQUE INDEX ticket_yt_channels_constraint ON public.ticket_yt_channels USING btree ("ticket_id", "yt_channels_id") WHERE ("deleted_at" IS NULL);

ALTER TABLE ONLY public.ticket_yt_channels
    ADD CONSTRAINT ticket_yt_channels_ticket_id_fkey FOREIGN KEY ("ticket_id") REFERENCES public.tickets(id);

ALTER TABLE ONLY public.ticket_yt_channels
    ADD CONSTRAINT ticket_yt_channels_id_fkey FOREIGN KEY ("yt_channels_id") REFERENCES public.technologies(id);
--
-- public.ticket_resources table
--
CREATE UNIQUE INDEX ticket_resources_constraint ON public.ticket_resources USING btree ("ticket_id", "resources_id") WHERE ("deleted_at" IS NULL);

ALTER TABLE ONLY public.ticket_resources
    ADD CONSTRAINT ticket_resources_ticket_id_fkey FOREIGN KEY ("ticket_id") REFERENCES public.tickets(id);

ALTER TABLE ONLY public.ticket_resources
    ADD CONSTRAINT ticket_resources_id_fkey FOREIGN KEY ("resources_id") REFERENCES public.technologies(id);
--
-- public.ticket_sources table
--
CREATE UNIQUE INDEX ticket_sources_constraint ON public.ticket_sources USING btree ("ticket_id", "sources_id") WHERE ("deleted_at" IS NULL);

ALTER TABLE ONLY public.ticket_sources
    ADD CONSTRAINT ticket_sources_ticket_id_fkey FOREIGN KEY ("ticket_id") REFERENCES public.tickets(id);

ALTER TABLE ONLY public.ticket_sources
    ADD CONSTRAINT ticket_sources_sources_id_fkey FOREIGN KEY ("sources_id") REFERENCES public.technologies(id);
--
-- public.ticket_docs table
--
CREATE UNIQUE INDEX ticket_docs_constraint ON public.ticket_docs USING btree ("ticket_id", "docs_id") WHERE ("deleted_at" IS NULL);

ALTER TABLE ONLY public.ticket_docs
    ADD CONSTRAINT ticket_docs_ticket_id_fkey FOREIGN KEY ("ticket_id") REFERENCES public.tickets(id);

ALTER TABLE ONLY public.ticket_docs
    ADD CONSTRAINT ticket_docs_docs_id_fkey FOREIGN KEY ("docs_id") REFERENCES public.technologies(id);
