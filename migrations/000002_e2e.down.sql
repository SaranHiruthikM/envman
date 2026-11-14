DROP TABLE IF EXISTS project_keys;

ALTER TABLE users
    DROP COLUMN IF EXISTS kdf_parallelism,
    DROP COLUMN IF EXISTS kdf_memory,
    DROP COLUMN IF EXISTS kdf_time,
    DROP COLUMN IF EXISTS kdf_salt,
    DROP COLUMN IF EXISTS encrypted_private_key_nonce,
    DROP COLUMN IF EXISTS encrypted_private_key,
    DROP COLUMN IF EXISTS public_key;
