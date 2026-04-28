--
-- PostgreSQL database dump
--

\restrict MEWDCBhaKYTHbu6C9on52BgkQi5rOaSrMPgFRdpmv9ni18SHHklPahc16TKsvmZ

-- Dumped from database version 18.1 (Homebrew)
-- Dumped by pg_dump version 18.1 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET timezone = 'Asia/Shanghai';

--
-- Name: vector; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;


--
-- Name: EXTENSION vector; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION vector IS 'vector data type and ivfflat and hnsw access methods';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: conversation_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.conversation_users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    conversation_id bigint NOT NULL,
    unread_count bigint DEFAULT 0,
    is_pinned boolean DEFAULT false,
    remark text,
    last_message_id bigint,
    is_banned boolean DEFAULT false
);


--
-- Name: conversations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.conversations (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    type smallint,
    owner_id bigint DEFAULT 0,
    group_name varchar(64) DEFAULT NULL
);


--
-- Name: files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.files (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    file_name character varying(255) NOT NULL,
    file_type character varying(50) NOT NULL,
    file_url character varying(255) NOT NULL,
    file_size bigint NOT NULL,
    hash_value character(64) NOT NULL
);


--
-- Name: file_vectors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.file_vectors (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    file_id bigint NOT NULL,
    vector public.vector(1024) NOT NULL,
    number integer NOT NULL
);


--
-- Name: friendship_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.friendship_requests (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    sender_id bigint NOT NULL,
    receiver_id bigint NOT NULL,
    verification_message character varying(128),
    status character varying(16) NOT NULL,
    sender_name character varying(64) NOT NULL,
    
    CONSTRAINT chk_friendship_requests_status CHECK (((status)::text = ANY (ARRAY[('pending'::character varying)::text, ('accepted'::character varying)::text, ('canceled'::character varying)::text])))
);


--
-- Name: friendships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.friendships (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    friend_id bigint NOT NULL,
    friend_remark character varying(64) NOT NULL
);


--
-- Name: message_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.message_users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint,
    message_id bigint NOT NULL,
    is_starred boolean DEFAULT false,
    is_deleted boolean DEFAULT false
);


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.messages (
    sender_id bigint,
    conversation_id bigint,
    status smallint DEFAULT 0,
    id bigint NOT NULL,
    file_id bigint DEFAULT 0,
    content varchar(1024) NOT NULL,
    created_at timestamp with time zone,
    deleted_at timestamp with time zone,
    updated_at timestamp with time zone
);

--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    name character varying(64) NOT NULL,
    password character varying(72) NOT NULL,
    uid character varying(20) NOT NULL,
    region character varying(32),
    phone_number character varying(20) NOT NULL,
    signature character varying(128),
    gender character varying(12),
    head character varying(255),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    
    CONSTRAINT chk_users_gender CHECK (((gender)::text = ANY (ARRAY[('male'::character varying)::text, ('female'::character varying)::text, (''::character varying)::text])))
);


--
-- Name: user_files; Type: TABLE; Schema: public; Owner: -
--
CREATE TABLE public.user_files (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    file_id bigint NOT NULL
);

--
-- Name: user_files user_files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_files
    ADD CONSTRAINT user_files_pkey PRIMARY KEY (id);

--
-- Name: conversation_users conversation_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversation_users
    ADD CONSTRAINT conversation_users_pkey PRIMARY KEY (id);


--
-- Name: conversations conversations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_pkey PRIMARY KEY (id);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);


--
-- Name: file_vectors file_vectors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.file_vectors
    ADD CONSTRAINT file_vectors_pkey PRIMARY KEY (id);


--
-- Name: friendship_requests friendship_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.friendship_requests
    ADD CONSTRAINT friendship_requests_pkey PRIMARY KEY (id);


--
-- Name: friendships friendships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.friendships
    ADD CONSTRAINT friendships_pkey PRIMARY KEY (id);


--
-- Name: message_users message_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.message_users
    ADD CONSTRAINT message_users_pkey PRIMARY KEY (id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);

--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: conversation_users_conversation_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX conversation_users_conversation_id_idx ON public.conversation_users USING btree (conversation_id);


--
-- Name: conversation_users_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX conversation_users_user_id_idx ON public.conversation_users USING btree (user_id);


--
-- Name: idx_conv_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_conv_user 
    ON public.conversation_users USING btree (user_id, conversation_id) WHERE deleted_at IS NULL;


--
-- Name: idx_conversation_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_conversation_owner ON public.conversations USING btree (owner_id);


--
-- Name: idx_conversation_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_conversation_type ON public.conversations USING btree (type);

--
-- Name: idx_friendship; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_friendship 
    ON public.friendships USING btree (user_id, friend_id) WHERE deleted_at IS NULL;


--
-- Name: idx_friendship_request; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_friendship_request 
    ON public.friendship_requests USING btree (sender_id, receiver_id) WHERE deleted_at IS NULL AND status = 'pending';


--
-- Name: idx_friendship_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_friendship_user_id ON public.friendships USING btree (user_id);


--
-- Name: idx_message_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_message_user 
    ON public.message_users USING btree (user_id, message_id) WHERE deleted_at IS NULL;


--
-- Name: idx_messages_conversation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_messages_conversation_id ON public.messages USING btree (conversation_id);


--
-- Name: idx_messages_sender_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_messages_sender_id ON public.messages USING btree (sender_id);


--
-- Name: idx_messages_file_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_messages_file_id ON public.messages USING btree (file_id);

--
-- Name: idx_files_hash_value; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_files_hash_value
    ON public.files USING btree (hash_value) WHERE deleted_at IS NULL;

--
-- Name: idx_file_vectors_file_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_file_vectors_file_id
    ON public.file_vectors USING btree (file_id);

--
-- Name: idx_file_vectors_file_id_number; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_file_vectors_file_id_number
    ON public.file_vectors USING btree (file_id, number) WHERE deleted_at IS NULL;

--
-- Name: idx_receiver; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_receiver ON public.friendship_requests USING btree (receiver_id);

--
-- Name: idx_users_phone_number; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_phone_number 
    ON public.users USING btree (phone_number) WHERE deleted_at IS NULL;


--
-- Name: idx_users_uid; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_uid 
    ON public.users USING btree (uid) WHERE deleted_at IS NULL;

--
-- Name: idx_user_files_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_files_user ON public.user_files USING btree (user_id);

--
-- Name: idx_user_file_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_user_file_unique
    ON public.user_files USING btree (user_id, file_id) WHERE deleted_at IS NULL;

--
-- PostgreSQL database dump complete
--

\unrestrict MEWDCBhaKYTHbu6C9on52BgkQi5rOaSrMPgFRdpmv9ni18SHHklPahc16TKsvmZ
