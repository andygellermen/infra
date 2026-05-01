# Datei: `docs/06-sqlite-ddl.md`

## 1. Ziel

Dieses DDL ist ein erster, bewusst pragmatischer Startpunkt für das Mini-ERP. Es ist MVP-nah, aber bereits auf Storno, Anzahlungen, E-Rechnung und Google-Sheets-Sync vorbereitet.

## 2. Konventionen

- IDs als `TEXT` mit UUID/ULID.
- Geldbeträge als `INTEGER` in Minor Units, z. B. Cent.
- Steuersätze als Basispunkte oder Dezimal-Text; im Entwurf `INTEGER` in basis points, z. B. 1900 = 19,00 %.
- Zeitwerte als ISO-8601 `TEXT`.
- JSON-Snapshots als `TEXT`.
- Fremdschlüssel aktiviert.

```sql
PRAGMA foreign_keys = ON;
```

## 3. Identity & Access

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT,
    role_key TEXT NOT NULL DEFAULT 'reader',
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE roles (
    role_key TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE role_permissions (
    id TEXT PRIMARY KEY,
    role_key TEXT NOT NULL,
    permission_key TEXT NOT NULL,
    allowed INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(role_key, permission_key),
    FOREIGN KEY(role_key) REFERENCES roles(role_key)
);

CREATE TABLE magic_link_tokens (
    id TEXT PRIMARY KEY,
    user_email TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    purpose TEXT NOT NULL DEFAULT 'login',
    expires_at TEXT NOT NULL,
    used_at TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    session_hash TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    revoked_at TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id)
);
```

## 4. Settings

```sql
CREATE TABLE settings_sync_runs (
    id TEXT PRIMARY KEY,
    source_spreadsheet_id TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    status TEXT NOT NULL,
    error_message TEXT,
    created_by TEXT
);

CREATE TABLE app_settings (
    id TEXT PRIMARY KEY,
    setting_group TEXT NOT NULL,
    setting_key TEXT NOT NULL,
    value TEXT NOT NULL,
    value_type TEXT NOT NULL,
    description TEXT,
    version TEXT,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    active INTEGER NOT NULL DEFAULT 1,
    source_sheet TEXT,
    source_row INTEGER,
    sync_run_id TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(setting_group, setting_key, version),
    FOREIGN KEY(sync_run_id) REFERENCES settings_sync_runs(id)
);

CREATE TABLE company_profiles (
    id TEXT PRIMARY KEY,
    profile_key TEXT NOT NULL UNIQUE,
    company_name TEXT NOT NULL,
    street TEXT,
    postal_code TEXT,
    city TEXT,
    country_code TEXT NOT NULL DEFAULT 'DE',
    tax_number TEXT,
    vat_id TEXT,
    email TEXT,
    phone TEXT,
    website TEXT,
    iban TEXT,
    bic TEXT,
    bank_name TEXT,
    legal_footer TEXT,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE number_ranges (
    id TEXT PRIMARY KEY,
    range_key TEXT NOT NULL UNIQUE,
    document_type TEXT NOT NULL,
    prefix TEXT NOT NULL,
    pattern TEXT NOT NULL,
    padding INTEGER NOT NULL DEFAULT 5,
    reset_policy TEXT NOT NULL DEFAULT 'yearly',
    current_year INTEGER,
    next_number INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE tax_rates (
    id TEXT PRIMARY KEY,
    tax_rate_key TEXT NOT NULL UNIQUE,
    country_code TEXT NOT NULL DEFAULT 'DE',
    tax_name TEXT NOT NULL,
    rate_basis_points INTEGER NOT NULL,
    category_code TEXT,
    description TEXT,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE payment_terms (
    id TEXT PRIMARY KEY,
    payment_term_key TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    due_days INTEGER NOT NULL DEFAULT 14,
    deposit_required INTEGER NOT NULL DEFAULT 0,
    deposit_mode TEXT NOT NULL DEFAULT 'none',
    deposit_value_basis_points INTEGER,
    deposit_value_minor INTEGER,
    final_due_days INTEGER,
    text_block TEXT,
    active INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

## 5. Customers

```sql
CREATE TABLE customers (
    id TEXT PRIMARY KEY,
    customer_number TEXT UNIQUE,
    customer_type TEXT NOT NULL DEFAULT 'business',
    display_name TEXT NOT NULL,
    company_name TEXT,
    first_name TEXT,
    last_name TEXT,
    email TEXT,
    phone TEXT,
    vat_id TEXT,
    tax_number TEXT,
    notes TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE customer_contacts (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    role TEXT,
    first_name TEXT,
    last_name TEXT,
    email TEXT,
    phone TEXT,
    is_primary INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(customer_id) REFERENCES customers(id)
);

CREATE TABLE customer_addresses (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    address_type TEXT NOT NULL DEFAULT 'billing',
    name_line TEXT,
    street TEXT,
    postal_code TEXT,
    city TEXT,
    country_code TEXT NOT NULL DEFAULT 'DE',
    is_primary INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(customer_id) REFERENCES customers(id)
);
```

## 6. Catalog / Select-Box-Staging

```sql
CREATE TABLE catalog_categories (
    id TEXT PRIMARY KEY,
    category_key TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    source_spreadsheet_id TEXT,
    source_sheet_name TEXT,
    source_row INTEGER,
    synced_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE manufacturers (
    id TEXT PRIMARY KEY,
    manufacturer_key TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    website TEXT,
    active INTEGER NOT NULL DEFAULT 1,
    source_sheet_name TEXT,
    source_row INTEGER,
    synced_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE product_groups (
    id TEXT PRIMARY KEY,
    group_key TEXT NOT NULL UNIQUE,
    category_id TEXT NOT NULL,
    manufacturer_id TEXT,
    label TEXT NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    source_sheet_name TEXT,
    source_row INTEGER,
    synced_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(category_id) REFERENCES catalog_categories(id),
    FOREIGN KEY(manufacturer_id) REFERENCES manufacturers(id)
);

CREATE TABLE products (
    id TEXT PRIMARY KEY,
    sku TEXT NOT NULL UNIQUE,
    category_id TEXT NOT NULL,
    manufacturer_id TEXT,
    group_id TEXT,
    product_name TEXT NOT NULL,
    description TEXT,
    unit_code TEXT NOT NULL DEFAULT 'C62',
    price_net_minor INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'EUR',
    tax_rate_key TEXT NOT NULL DEFAULT 'vat_19',
    stock_quantity INTEGER,
    availability_status TEXT NOT NULL DEFAULT 'available',
    order_required INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    source_sheet_name TEXT,
    source_row INTEGER,
    synced_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(category_id) REFERENCES catalog_categories(id),
    FOREIGN KEY(manufacturer_id) REFERENCES manufacturers(id),
    FOREIGN KEY(group_id) REFERENCES product_groups(id)
);

CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_manufacturer ON products(manufacturer_id);
CREATE INDEX idx_products_group ON products(group_id);
```

## 7. Documents

```sql
CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    document_type TEXT NOT NULL,
    document_number TEXT UNIQUE,
    customer_id TEXT NOT NULL,
    predecessor_document_id TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    version_no INTEGER NOT NULL DEFAULT 1,
    currency TEXT NOT NULL DEFAULT 'EUR',
    company_profile_id TEXT,
    payment_term_id TEXT,
    legal_text_key TEXT,
    billing_address_snapshot TEXT,
    delivery_address_snapshot TEXT,
    customer_snapshot TEXT,
    issue_date TEXT,
    service_date TEXT,
    service_period_start TEXT,
    service_period_end TEXT,
    due_date TEXT,
    total_net_minor INTEGER NOT NULL DEFAULT 0,
    total_tax_minor INTEGER NOT NULL DEFAULT 0,
    total_gross_minor INTEGER NOT NULL DEFAULT 0,
    prepaid_minor INTEGER NOT NULL DEFAULT 0,
    payable_minor INTEGER NOT NULL DEFAULT 0,
    e_invoice_required INTEGER NOT NULL DEFAULT 0,
    e_invoice_profile_key TEXT,
    finalized_at TEXT,
    sent_at TEXT,
    cancelled_at TEXT,
    created_by TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(customer_id) REFERENCES customers(id),
    FOREIGN KEY(predecessor_document_id) REFERENCES documents(id),
    FOREIGN KEY(company_profile_id) REFERENCES company_profiles(id),
    FOREIGN KEY(payment_term_id) REFERENCES payment_terms(id)
);

CREATE TABLE document_items (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    position_no INTEGER NOT NULL,
    product_id TEXT,
    sku_snapshot TEXT,
    product_name_snapshot TEXT NOT NULL,
    description_snapshot TEXT,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_code TEXT NOT NULL DEFAULT 'C62',
    unit_price_net_minor INTEGER NOT NULL DEFAULT 0,
    discount_minor INTEGER NOT NULL DEFAULT 0,
    surcharge_minor INTEGER NOT NULL DEFAULT 0,
    tax_rate_key TEXT NOT NULL DEFAULT 'vat_19',
    tax_rate_basis_points INTEGER NOT NULL DEFAULT 1900,
    line_net_minor INTEGER NOT NULL DEFAULT 0,
    line_tax_minor INTEGER NOT NULL DEFAULT 0,
    line_gross_minor INTEGER NOT NULL DEFAULT 0,
    service_date TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(document_id) REFERENCES documents(id),
    FOREIGN KEY(product_id) REFERENCES products(id)
);

CREATE TABLE document_references (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    reference_type TEXT NOT NULL,
    referenced_document_id TEXT,
    external_reference TEXT,
    description TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(document_id) REFERENCES documents(id),
    FOREIGN KEY(referenced_document_id) REFERENCES documents(id)
);

CREATE TABLE document_status_events (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    old_status TEXT,
    new_status TEXT NOT NULL,
    reason TEXT,
    metadata_json TEXT,
    created_by TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(document_id) REFERENCES documents(id)
);

CREATE INDEX idx_documents_customer ON documents(customer_id);
CREATE INDEX idx_documents_type_status ON documents(document_type, status);
CREATE INDEX idx_document_items_document ON document_items(document_id);
```

## 8. Payments

```sql
CREATE TABLE payment_requests (
    id TEXT PRIMARY KEY,
    document_id TEXT,
    request_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    amount_minor INTEGER NOT NULL,
    currency TEXT NOT NULL DEFAULT 'EUR',
    due_date TEXT,
    description TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(document_id) REFERENCES documents(id)
);

CREATE TABLE payments (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    payment_date TEXT NOT NULL,
    amount_minor INTEGER NOT NULL,
    currency TEXT NOT NULL DEFAULT 'EUR',
    payment_method TEXT NOT NULL DEFAULT 'bank_transfer',
    reference TEXT,
    bank_booking_text TEXT,
    status TEXT NOT NULL DEFAULT 'received',
    created_by TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(customer_id) REFERENCES customers(id)
);

CREATE TABLE payment_allocations (
    id TEXT PRIMARY KEY,
    payment_id TEXT NOT NULL,
    document_id TEXT,
    payment_request_id TEXT,
    allocation_type TEXT NOT NULL,
    amount_minor INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    created_by TEXT,
    FOREIGN KEY(payment_id) REFERENCES payments(id),
    FOREIGN KEY(document_id) REFERENCES documents(id),
    FOREIGN KEY(payment_request_id) REFERENCES payment_requests(id)
);

CREATE TABLE refunds (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    source_payment_id TEXT,
    source_document_id TEXT,
    amount_minor INTEGER NOT NULL,
    currency TEXT NOT NULL DEFAULT 'EUR',
    status TEXT NOT NULL DEFAULT 'pending',
    reason TEXT,
    refunded_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(customer_id) REFERENCES customers(id),
    FOREIGN KEY(source_payment_id) REFERENCES payments(id),
    FOREIGN KEY(source_document_id) REFERENCES documents(id)
);

CREATE INDEX idx_payment_allocations_payment ON payment_allocations(payment_id);
CREATE INDEX idx_payment_allocations_document ON payment_allocations(document_id);
```

## 9. Cancellation & Correction

```sql
CREATE TABLE cancellation_policies (
    id TEXT PRIMARY KEY,
    policy_key TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    applies_to TEXT NOT NULL,
    customer_type TEXT NOT NULL DEFAULT 'any',
    product_category_key TEXT,
    time_reference TEXT NOT NULL DEFAULT 'service_start',
    threshold_from_hours INTEGER NOT NULL,
    threshold_to_hours INTEGER NOT NULL,
    fee_mode TEXT NOT NULL DEFAULT 'none',
    fee_value_basis_points INTEGER,
    fee_value_minor INTEGER,
    fee_basis TEXT NOT NULL DEFAULT 'order_gross',
    min_fee_minor INTEGER NOT NULL DEFAULT 0,
    max_fee_minor INTEGER,
    tax_rate_key TEXT,
    requires_manual_review INTEGER NOT NULL DEFAULT 0,
    active INTEGER NOT NULL DEFAULT 1,
    valid_from TEXT NOT NULL,
    valid_to TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE cancellation_events (
    id TEXT PRIMARY KEY,
    target_document_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'requested',
    cancellation_reason TEXT,
    policy_id TEXT,
    requested_at TEXT NOT NULL,
    evaluated_at TEXT,
    approved_at TEXT,
    completed_at TEXT,
    fee_amount_minor INTEGER NOT NULL DEFAULT 0,
    fee_tax_minor INTEGER NOT NULL DEFAULT 0,
    fee_gross_minor INTEGER NOT NULL DEFAULT 0,
    decision_json TEXT,
    created_by TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY(target_document_id) REFERENCES documents(id),
    FOREIGN KEY(policy_id) REFERENCES cancellation_policies(id)
);

CREATE TABLE correction_documents (
    id TEXT PRIMARY KEY,
    correction_document_id TEXT NOT NULL,
    original_document_id TEXT NOT NULL,
    correction_type TEXT NOT NULL,
    reason TEXT,
    created_at TEXT NOT NULL,
    created_by TEXT,
    FOREIGN KEY(correction_document_id) REFERENCES documents(id),
    FOREIGN KEY(original_document_id) REFERENCES documents(id)
);
```

## 10. Output, Files, E-Invoice

```sql
CREATE TABLE generated_files (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    file_type TEXT NOT NULL,
    file_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    local_path TEXT,
    google_drive_file_id TEXT,
    google_doc_id TEXT,
    sha256_hash TEXT,
    created_at TEXT NOT NULL,
    created_by TEXT,
    FOREIGN KEY(document_id) REFERENCES documents(id)
);

CREATE TABLE mail_dispatches (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    template_key TEXT,
    to_email TEXT NOT NULL,
    cc_email TEXT,
    bcc_email TEXT,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    sent_at TEXT,
    error_message TEXT,
    created_at TEXT NOT NULL,
    created_by TEXT,
    FOREIGN KEY(document_id) REFERENCES documents(id)
);

CREATE TABLE e_invoice_exports (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    format TEXT NOT NULL,
    profile_key TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    xml_file_id TEXT,
    validation_status TEXT,
    validation_report_file_id TEXT,
    error_count INTEGER NOT NULL DEFAULT 0,
    warning_count INTEGER NOT NULL DEFAULT 0,
    generated_at TEXT,
    validated_at TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(document_id) REFERENCES documents(id),
    FOREIGN KEY(xml_file_id) REFERENCES generated_files(id),
    FOREIGN KEY(validation_report_file_id) REFERENCES generated_files(id)
);
```

## 11. Sync

```sql
CREATE TABLE sheet_sources (
    id TEXT PRIMARY KEY,
    source_key TEXT NOT NULL UNIQUE,
    spreadsheet_id TEXT NOT NULL,
    sheet_name TEXT NOT NULL,
    purpose TEXT NOT NULL,
    active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE sheet_sync_runs (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    status TEXT NOT NULL,
    rows_read INTEGER NOT NULL DEFAULT 0,
    rows_created INTEGER NOT NULL DEFAULT 0,
    rows_updated INTEGER NOT NULL DEFAULT 0,
    rows_failed INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    FOREIGN KEY(source_id) REFERENCES sheet_sources(id)
);

CREATE TABLE sheet_sync_errors (
    id TEXT PRIMARY KEY,
    sync_run_id TEXT NOT NULL,
    sheet_name TEXT,
    row_no INTEGER,
    field_name TEXT,
    error_code TEXT,
    error_message TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY(sync_run_id) REFERENCES sheet_sync_runs(id)
);
```

## 12. Audit

```sql
CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    actor_user_id TEXT,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    action TEXT NOT NULL,
    old_value_json TEXT,
    new_value_json TEXT,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY(actor_user_id) REFERENCES users(id)
);

CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_created_at ON audit_logs(created_at);
```

## 13. Offene DDL-Entscheidungen

| Thema | Empfehlung |
|---|---|
| Migration Tool | `goose`, `atlas` oder eigenes kleines Migrationspaket |
| IDs | ULID, weil sortierbar und logfreundlich |
| Geldbeträge | immer Minor Units, keine Float-Werte |
| JSON | für Snapshots erlaubt, aber Kernfelder relational halten |
| E-Rechnung | XML-Generator später als separates Package |
| Audit | für MVP generisch, später fachlich feiner |

---
