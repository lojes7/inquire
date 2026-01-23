--
-- PostgreSQL database dump
--

\restrict 7If4iLKASUsmIrqkae3mSelVe4UixeM0YZtACJnrEYee9oCadfdeK5axuEfHHmQ

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
-- Name: conversation_friends; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.conversation_friends (
    id bigint NOT NULL,
    created_at timestamp without time zone,
    deleted_at timestamp without time zone,
    updated_at timestamp without time zone,
    conversation_id bigint NOT NULL,
    user_id bigint NOT NULL,
    friend_id bigint NOT NULL
);


--
-- Name: conversation_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.conversation_users (
    id bigint NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    user_id bigint,
    conversation_id bigint,
    unread_count bigint DEFAULT 0,
    is_pinned boolean DEFAULT false,
    remark text,
    last_message_id bigint
);


--
-- Name: conversations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.conversations (
    id bigint NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    deleted_at timestamp without time zone,
    type smallint
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
    file_content text,
    content_vector public.vector(1536),
    message_id bigint NOT NULL
);


--
-- Name: files_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.files_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.files_id_seq OWNED BY public.files.id;


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
    CONSTRAINT chk_friendship_requests_status CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'accepted'::character varying, 'rejected'::character varying, 'canceled'::character varying])::text[])))
);


--
-- Name: friendship_requests_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.friendship_requests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: friendship_requests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.friendship_requests_id_seq OWNED BY public.friendship_requests.id;


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
-- Name: friendships_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.friendships_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: friendships_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.friendships_id_seq OWNED BY public.friendships.id;


--
-- Name: message_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.message_users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint,
    message_id bigint,
    is_starred boolean DEFAULT false
);


--
-- Name: message_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.message_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: message_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.message_users_id_seq OWNED BY public.message_users.id;


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.messages (
    sender_id bigint,
    conversation_id bigint,
    status smallint DEFAULT 0,
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.messages_id_seq OWNED BY public.messages.id;


--
-- Name: texts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.texts (
    text character varying(1024) CONSTRAINT texts_content_not_null NOT NULL,
    message_id bigint NOT NULL
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
    deleted_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT chk_users_gender CHECK (((gender)::text = ANY ((ARRAY['male'::character varying, 'female'::character varying, ''::character varying])::text[])))
);


--
-- Name: conversation_friends conversation_friends_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversation_friends
    ADD CONSTRAINT conversation_friends_pkey PRIMARY KEY (id);


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
-- Name: conversation_friends_user_id_friend_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX conversation_friends_user_id_friend_id_idx ON public.conversation_friends USING btree (user_id, friend_id);


--
-- Name: conversation_users_conversation_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX conversation_users_conversation_id_idx ON public.conversation_users USING btree (conversation_id);


--
-- Name: idx_conv_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_conv_user ON public.conversation_users USING btree (user_id, conversation_id);


--
-- Name: idx_conversation_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_conversation_type ON public.conversations USING btree (type);


--
-- Name: idx_conversation_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_conversation_users_deleted_at ON public.conversation_users USING btree (deleted_at);


--
-- Name: idx_file_msg; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_file_msg ON public.files USING btree (message_id);


--
-- Name: idx_files_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_files_deleted_at ON public.files USING btree (deleted_at);


--
-- Name: idx_friendship_requests_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_friendship_requests_deleted_at ON public.friendship_requests USING btree (deleted_at);


--
-- Name: idx_friendships_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_friendships_deleted_at ON public.friendships USING btree (deleted_at);


--
-- Name: idx_friendships_u_f; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_friendships_u_f ON public.friendships USING btree (user_id, friend_id, deleted_at);


--
-- Name: idx_message_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_message_user ON public.message_users USING btree (user_id, message_id);


--
-- Name: idx_message_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_message_users_deleted_at ON public.message_users USING btree (deleted_at);


--
-- Name: idx_messages_conversation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_messages_conversation_id ON public.messages USING btree (conversation_id);


--
-- Name: idx_messages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_messages_deleted_at ON public.messages USING btree (deleted_at);


--
-- Name: idx_messages_sender_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_messages_sender_id ON public.messages USING btree (sender_id);


--
-- Name: idx_requests_sender_receiver; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_requests_sender_receiver ON public.friendship_requests USING btree (sender_id, receiver_id, deleted_at);


--
-- Name: idx_text_msg; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_text_msg ON public.texts USING btree (message_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_name ON public.users USING btree (name);


--
-- Name: idx_users_phone_number; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_phone_number ON public.users USING btree (phone_number);


--
-- Name: idx_users_uid; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_uid ON public.users USING btree (uid);


--
-- PostgreSQL database dump complete
--

\unrestrict 7If4iLKASUsmIrqkae3mSelVe4UixeM0YZtACJnrEYee9oCadfdeK5axuEfHHmQ

