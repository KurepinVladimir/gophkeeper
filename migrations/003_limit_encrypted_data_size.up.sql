ALTER TABLE secrets
ADD CONSTRAINT secrets_encrypted_data_size_check
CHECK (octet_length(encrypted_data) <= 1048576);
