CREATE TABLE users (
    id              UUID PRIMARY KEY,
    email           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    telegram_chat_id BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE subscriptions (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    keyword     TEXT NOT NULL,
    area_id     TEXT NOT NULL DEFAULT '',
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_active ON subscriptions(is_active) WHERE is_active = true;

CREATE TABLE vacancies (
    id              UUID PRIMARY KEY,
    hh_vacancy_id   TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    employer_name   TEXT NOT NULL DEFAULT '',
    url             TEXT NOT NULL,
    area_name       TEXT NOT NULL DEFAULT '',
    salary_from     INTEGER,
    salary_to       INTEGER,
    salary_currency TEXT NOT NULL DEFAULT '',
    published_at    TIMESTAMPTZ
);

CREATE TABLE seen_vacancies (
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hh_vacancy_id TEXT NOT NULL REFERENCES vacancies(hh_vacancy_id) ON DELETE CASCADE,
    sent_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, hh_vacancy_id)
);
