--
-- PostgreSQL database dump
--

-- Dumped from database version 11.5
-- Dumped by pg_dump version 11.5

-- Started on 2019-10-21 21:51:20 MSK

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_with_oids = false;

--
-- TOC entry 196 (class 1259 OID 24706)
-- Name: ic_chats_photos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ic_chats_photos (
    chat_id bigint NOT NULL,
    photo_id integer NOT NULL
);


--
-- TOC entry 199 (class 1259 OID 24727)
-- Name: ic_page; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ic_page (
    id integer NOT NULL,
    pins_link character varying(255) NOT NULL,
    cursor character varying(1024) NOT NULL
);


--
-- TOC entry 200 (class 1259 OID 24733)
-- Name: ic_page_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ic_page_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 197 (class 1259 OID 24709)
-- Name: ic_photos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ic_photos (
    id integer NOT NULL,
    url character varying(2048)
);


--
-- TOC entry 198 (class 1259 OID 24715)
-- Name: ic_photos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ic_photos_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 3146 (class 0 OID 0)
-- Dependencies: 198
-- Name: ic_photos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ic_photos_id_seq OWNED BY public.ic_photos.id;


--
-- TOC entry 3014 (class 2604 OID 24717)
-- Name: ic_photos id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ic_photos ALTER COLUMN id SET DEFAULT nextval('public.ic_photos_id_seq'::regclass);


--
-- TOC entry 3016 (class 2606 OID 24719)
-- Name: ic_chats_photos chat_photo_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ic_chats_photos
    ADD CONSTRAINT chat_photo_pkey PRIMARY KEY (chat_id, photo_id);


--
-- TOC entry 3018 (class 2606 OID 24721)
-- Name: ic_photos ic_photos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ic_photos
    ADD CONSTRAINT ic_photos_pkey PRIMARY KEY (id);


--
-- TOC entry 3019 (class 2606 OID 24722)
-- Name: ic_chats_photos ic_chats_photos_id_photo_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ic_chats_photos
    ADD CONSTRAINT ic_chats_photos_id_photo_fkey FOREIGN KEY (photo_id) REFERENCES public.ic_photos(id) ON UPDATE CASCADE;


-- Completed on 2019-10-21 21:51:20 MSK

--
-- PostgreSQL database dump complete
--

