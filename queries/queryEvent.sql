-- event_base_query.sql

WITH event_data AS (
    SELECT
        ed.id AS event_date_id,
        ed.event_id,
        ed.space_id,
        ed.start,
        ed.end,
        ed.entry_time,
        ed.accessibility_flags,
        ed.visitor_info_flags
    FROM {{schema}}.event_date ed
    {{event-date-conditions}}
    )
SELECT
    e.id AS event_id,
    e.title AS event_title,
    e.teaser_text AS teaser_text,
    array_to_string((regexp_split_to_array(e.description, '\s+'))[1:40], ' ') AS description_preview,
    e.description AS description,

    o.id AS organizer_id,
    o.name AS organizer_name,

    v.id AS venue_id,
    v.name AS venue_name,
    v.street AS venue_street,
    v.house_number AS venue_house_number,
    v.postal_code AS venue_postal_code,
    v.city AS venue_city,
    v.country_code AS venue_country,
    v.state_code AS venue_state,
    ST_AsText(v.wkb_geometry) AS venue_geometry,

    COALESCE(s.id, es.id) AS space_id,
    COALESCE(s.name, es.name) AS space_name,
    COALESCE(s.total_capacity, es.total_capacity) AS space_total_capacity,
    COALESCE(s.seating_capacity, es.seating_capacity) AS space_seating_capacity,
    COALESCE(s.building_level, es.building_level) AS space_building_level,
    COALESCE(s.url, es.url) AS space_url,

    TO_CHAR(ed.start, 'YYYY-MM-DD') AS start_date,
    TO_CHAR(ed.start, 'HH24:MI') AS start_time,
    TO_CHAR(ed.end, 'YYYY-MM-DD') AS end_date,
    TO_CHAR(ed.end, 'HH24:MI') AS end_time,
    TO_CHAR(ed.entry_time, 'HH24:MI') AS entry_time,

    ed.accessibility_flags AS accessibility_flags,
    ed.visitor_info_flags AS visitor_info_flags,

    acc_flags.accessibility_flag_names AS accessibility_flag_names,
    vis_flags.visitor_info_flag_names AS visitor_info_flag_names,

    (img_data.has_main_image) AS has_main_image,
    img_data.img_src_name,

    et_data.event_types,
    gt_data.genre_types

FROM event_data ed
    JOIN {{schema}}.event e ON ed.event_id = e.id
    JOIN {{schema}}.organizer o ON e.organizer_id = o.id
    LEFT JOIN {{schema}}.space s ON ed.space_id = s.id
    LEFT JOIN {{schema}}.space es ON e.space_id = es.id
    JOIN {{schema}}.venue v ON COALESCE(s.venue_id, es.venue_id) = v.id

    LEFT JOIN LATERAL (
        SELECT
        TRUE AS has_main_image,
        img.source_name AS img_src_name
        FROM {{schema}}.event_link_images eli
        JOIN {{schema}}.image img ON eli.image_id = img.id
        WHERE eli.event_id = e.id AND eli.main_image = TRUE
        LIMIT 1
    ) img_data ON true

    LEFT JOIN LATERAL (
        SELECT jsonb_agg(DISTINCT jsonb_build_object('id', elt.event_type_id, 'name', et.name)) AS event_types
        FROM {{schema}}.event_link_types elt
        JOIN {{schema}}.event_type et
        ON et.type_id = elt.event_type_id AND et.iso_639_1 = $1
        WHERE elt.event_id = e.id
    ) et_data ON true

    LEFT JOIN LATERAL (
        SELECT jsonb_agg(DISTINCT jsonb_build_object('id', glt.genre_type_id, 'name', gt.name)) AS genre_types
        FROM {{schema}}.genre_link_types glt
        JOIN {{schema}}.genre_type gt
        ON gt.type_id = glt.genre_type_id AND gt.iso_639_1 = $1
        WHERE glt.event_id = e.id
    ) gt_data ON true

    LEFT JOIN LATERAL (
        SELECT jsonb_agg(name) AS accessibility_flag_names
        FROM {{schema}}.accessibility_flags f
        WHERE (ed.accessibility_flags & (1::BIGINT << f.flag)) = (1::BIGINT << f.flag) AND f.iso_639_1 = $1
    ) acc_flags ON true

    LEFT JOIN LATERAL (
        SELECT jsonb_agg(name) AS visitor_info_flag_names
        FROM {{schema}}.visitor_information_flags f
        WHERE (ed.visitor_info_flags & (1::BIGINT << f.flag)) = (1::BIGINT << f.flag) AND f.iso_639_1 = $1
    ) vis_flags ON true

    {{conditions}}
    {{order}}
    {{limit}}