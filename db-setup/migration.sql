BEGIN;

CREATE TABLE IF NOT EXISTS public."user" (
    id BIGSERIAL PRIMARY KEY,
    username TEXT,
    email TEXT,
    phone TEXT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    middle_name TEXT,
    password TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    owner_id BIGINT,
    photo_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_user_username ON public."user"(username) WHERE username IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ux_user_email ON public."user"(email) WHERE email IS NOT NULL;

CREATE TABLE IF NOT EXISTS public.role (
    id BIGINT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS public.user_roles (
    user_id BIGINT NOT NULL REFERENCES public."user"(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES public.role(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS ix_user_roles_role_id ON public.user_roles(role_id);

CREATE TABLE IF NOT EXISTS public.language (
    id INT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.status (
    id INT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS public.real_estate_type (
    id INT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS public.real_estate_purpose (
    id INT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS public.country (
    id BIGSERIAL PRIMARY KEY,
    name_ru TEXT,
    name_en TEXT,
    name_tr TEXT,
    name_de TEXT,
    code TEXT,
    phone_code TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.region (
    id BIGSERIAL PRIMARY KEY,
    name_ru TEXT,
    name_en TEXT,
    name_tr TEXT,
    name_de TEXT,
    country_id BIGINT REFERENCES public.country(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.district (
    id BIGSERIAL PRIMARY KEY,
    name_ru TEXT,
    name_en TEXT,
    name_tr TEXT,
    name_de TEXT,
    region_id BIGINT REFERENCES public.region(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.amenities (
    id BIGSERIAL PRIMARY KEY,
    name_ru TEXT,
    name_en TEXT,
    name_de TEXT,
    name_tr TEXT,
    icon TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.files (
    id BIGSERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    user_id BIGINT REFERENCES public."user"(id),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.real_estate (
    id BIGSERIAL PRIMARY KEY,
    price NUMERIC(12,2) NOT NULL,
    area NUMERIC(12,2) NOT NULL,
    rooms TEXT,
    region_id BIGINT REFERENCES public.region(id),
    "type" INT REFERENCES public.real_estate_type(id),
    purpose_id INT REFERENCES public.real_estate_purpose(id),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    completion_date DATE,
    manager_id BIGINT REFERENCES public."user"(id),
    "location" TEXT,
    address TEXT,
    apartment TEXT,
    floor INT,
    total_floors INT,
    parking_available BOOLEAN,
    built_year INT,
    price_per_square_meter NUMERIC(12,2),
    has_balcony BOOLEAN,
    distance_to_sea DOUBLE PRECISION,
    district_id BIGINT REFERENCES public.district(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.real_estate_url (
    real_estate_id BIGINT NOT NULL REFERENCES public.real_estate(id) ON DELETE CASCADE,
    files_id BIGINT NOT NULL REFERENCES public.files(id) ON DELETE CASCADE,
    order_num INT NOT NULL DEFAULT 1,
    PRIMARY KEY (real_estate_id, files_id)
);

CREATE INDEX IF NOT EXISTS ix_real_estate_url_order ON public.real_estate_url(real_estate_id, order_num);

CREATE TABLE IF NOT EXISTS public.real_estate_history (
    id BIGSERIAL PRIMARY KEY,
    real_estate_id BIGINT NOT NULL REFERENCES public.real_estate(id) ON DELETE CASCADE,
    status_id INT NOT NULL REFERENCES public.status(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    manager_id BIGINT REFERENCES public."user"(id)
);

CREATE INDEX IF NOT EXISTS ix_real_estate_history_estate_id ON public.real_estate_history(real_estate_id, id DESC);

CREATE TABLE IF NOT EXISTS public.real_estate_translations (
    id BIGSERIAL PRIMARY KEY,
    real_estate_id BIGINT NOT NULL REFERENCES public.real_estate(id) ON DELETE CASCADE,
    description TEXT,
    "name" TEXT,
    lang_id INT NOT NULL REFERENCES public.language(id),
    summary_title TEXT,
    terms_and_conditions TEXT,
    UNIQUE(real_estate_id, lang_id)
);

CREATE TABLE IF NOT EXISTS public.real_estate_amenities (
    real_estate_id BIGINT NOT NULL REFERENCES public.real_estate(id) ON DELETE CASCADE,
    amenity_id BIGINT NOT NULL REFERENCES public.amenities(id) ON DELETE CASCADE,
    PRIMARY KEY (real_estate_id, amenity_id)
);

CREATE TABLE IF NOT EXISTS public.collection (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    url TEXT,
    manager_id BIGINT REFERENCES public."user"(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    photo_url TEXT,
    is_temporary BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_collection_url ON public.collection(url) WHERE url IS NOT NULL;

CREATE TABLE IF NOT EXISTS public.collection_item (
    collection_id BIGINT NOT NULL REFERENCES public.collection(id) ON DELETE CASCADE,
    real_estate_id BIGINT NOT NULL REFERENCES public.real_estate(id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, real_estate_id)
);

CREATE TABLE IF NOT EXISTS public.collection_history (
    id BIGSERIAL PRIMARY KEY,
    collection_id BIGINT NOT NULL REFERENCES public.collection(id) ON DELETE CASCADE,
    status_id INT NOT NULL REFERENCES public.status(id),
    manager_id BIGINT REFERENCES public."user"(id),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_collection_history_collection_id ON public.collection_history(collection_id, id DESC);

CREATE TABLE IF NOT EXISTS public.collection_translations (
    collection_id BIGINT NOT NULL REFERENCES public.collection(id) ON DELETE CASCADE,
    lang_id INT NOT NULL REFERENCES public.language(id),
    "name" TEXT NOT NULL,
    description TEXT,
    PRIMARY KEY (collection_id, lang_id)
);

CREATE TABLE IF NOT EXISTS public.page (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.page_translations (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES public.page(id) ON DELETE CASCADE,
    language_id INT NOT NULL REFERENCES public.language(id),
    title TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.block_type (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.block (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES public.page(id) ON DELETE CASCADE,
    block_type_id BIGINT REFERENCES public.block_type(id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.block_content (
    id BIGSERIAL PRIMARY KEY,
    block_id BIGINT NOT NULL REFERENCES public.block(id) ON DELETE CASCADE,
    image_url TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.block_content_translations (
    id BIGSERIAL PRIMARY KEY,
    block_content_id BIGINT NOT NULL REFERENCES public.block_content(id) ON DELETE CASCADE,
    language_id INT NOT NULL REFERENCES public.language(id),
    title TEXT,
    body TEXT,
    UNIQUE(block_content_id, language_id)
);

CREATE TABLE IF NOT EXISTS public.exchange_rates (
    currency_code TEXT PRIMARY KEY,
    rate DOUBLE PRECISION NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now()
);

INSERT INTO public.language (id, code) VALUES
    (1, 'ru'),
    (2, 'en'),
    (3, 'de'),
    (4, 'tr')
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.role (id, code) VALUES
    (1, 'admin'),
    (2, 'manager')
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.status (id, name) VALUES
    (1, 'created'),
    (2, 'available'),
    (3, 'sold'),
    (4, 'archived')
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.real_estate_type (id, name) VALUES
    (1, 'apartment'),
    (2, 'commercial'),
    (3, 'land'),
    (4, 'house'),
    (5, 'secondary_house'),
    (6, 'townhouses'),
    (7, 'villas')
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.real_estate_purpose (id, name) VALUES
    (1, 'sale'),
    (2, 'rent')
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.block_type (id, code, name) VALUES
    (1, 'text', 'Text'),
    (2, 'image', 'Image'),
    (3, 'video', 'Video'),
    (4, 'link', 'Link'),
    (5, 'html', 'HTML')
ON CONFLICT (id) DO NOTHING;

COMMIT;
