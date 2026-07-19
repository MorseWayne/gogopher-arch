CREATE TABLE alerts (
    tenant_id text NOT NULL,
    id text NOT NULL,
    message text NOT NULL,
    acknowledged boolean NOT NULL DEFAULT false,
    PRIMARY KEY (tenant_id, id)
);
