-- Refresh tokens são opacos: o banco guarda só o SHA-256 do token, nunca o
-- valor. Cada login abre uma "família"; cada /refresh gira o token dentro
-- dela. Reuso de um token girado fora da janela de tolerância revoga a
-- família inteira.
CREATE TABLE refresh_tokens (
    id          UUID        PRIMARY KEY,
    family_id   UUID        NOT NULL,
    user_id     UUID        NOT NULL,
    token_hash  TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    rotated_at  TIMESTAMPTZ,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX refresh_tokens_family_id_idx ON refresh_tokens (family_id);
