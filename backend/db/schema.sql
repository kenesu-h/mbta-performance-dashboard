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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: dwell; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.dwell (
    stop_id character varying(255) NOT NULL,
    route_id character varying(255) NOT NULL,
    direction boolean NOT NULL,
    arr_dt timestamp without time zone NOT NULL,
    dep_dt timestamp without time zone NOT NULL,
    dwell_time_sec integer NOT NULL
);


--
-- Name: headway; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.headway (
    stop_id character varying(255) NOT NULL,
    route_id character varying(255) NOT NULL,
    prev_route_id character varying(255) NOT NULL,
    direction boolean NOT NULL,
    current_dep_dt timestamp without time zone NOT NULL,
    previous_dep_dt timestamp without time zone NOT NULL,
    headway_time_sec integer NOT NULL,
    benchmark_headway_time_sec integer NOT NULL
);


--
-- Name: last_dwell_cache_datetime; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.last_dwell_cache_datetime (
    stop_id character varying(255) NOT NULL,
    route_id character varying(255) NOT NULL,
    value timestamp without time zone NOT NULL
);


--
-- Name: last_headway_cache_datetime; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.last_headway_cache_datetime (
    stop_id character varying(255) NOT NULL,
    route_id character varying(255) NOT NULL,
    value timestamp without time zone NOT NULL
);


--
-- Name: last_travel_time_cache_datetime; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.last_travel_time_cache_datetime (
    from_stop_id character varying(255) NOT NULL,
    to_stop_id character varying(255) NOT NULL,
    route_id character varying(255) NOT NULL,
    value timestamp without time zone NOT NULL
);


--
-- Name: route; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.route (
    id character varying(255) NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(128) NOT NULL
);


--
-- Name: shape; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.shape (
    id character varying(255) NOT NULL,
    route_id character varying(255) NOT NULL,
    polyline character varying NOT NULL
);


--
-- Name: stop; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stop (
    id character varying(255) NOT NULL,
    route_id character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL
);


--
-- Name: travel_time; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.travel_time (
    from_stop_id character varying(255) NOT NULL,
    to_stop_id character varying(255) NOT NULL,
    route_id character varying(255) NOT NULL,
    direction boolean NOT NULL,
    dep_dt timestamp without time zone NOT NULL,
    arr_dt timestamp without time zone NOT NULL,
    travel_time_sec integer NOT NULL,
    benchmark_travel_time_sec integer NOT NULL
);


--
-- Name: route route_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.route
    ADD CONSTRAINT route_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: shape shape_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shape
    ADD CONSTRAINT shape_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20230906195458');
