# Quiz Me - Framework V21: Canonical PS-ID Recursive Structure (Round 2)

**Created**: 2026-04-30
**Purpose**: Close the remaining Q2 decision by selecting the canonical recursive directory structure that will be enforced for all 10 PS-IDs.

---

## Research Snapshot (Evidence-Based)

### Requested Focus Services (interpreting `jose-ca` as `jose-ja`)

- `sm-kms` currently has: `server/businesslogic`, `server/handler`, `server/repository`, `server/repository/migrations`, `server/repository/orm`
- `sm-im` currently has: `server/apis`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`
- `jose-ja` currently has: `server/apis`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`, `server/service`
- `skeleton-template` currently has: `server/apis`, `server/config`, `server/handler`, `server/model`, `server/repository`, `server/repository/migrations`

## Recursive Directory Inventory with Per-Directory CSV File Lists

### Group 1: sm-kms, sm-im, jose-ja, skeleton-template

```text
### sm-kms
- internal/apps/sm-kms | kms.go, kms_cli_test.go, kms_lifecycle_test.go, kms_port_conflict_test.go, kms_usage.go
- internal/apps/sm-kms/client | client_cleanup_test.go, client_oam_mapper.go, client_test.go, client_test_util.go
- internal/apps/sm-kms/e2e | e2e_admin_isolation_test.go, e2e_postgres_mtls_test.go, e2e_test.go, e2e_tls_test.go, testmain_e2e_test.go
- internal/apps/sm-kms/server | contracts_integration_test.go, kms_lifecycle_test.go, kms_port_conflict_test.go, server.go, server_test.go, swagger.go, swagger_test.go, testmain_integration_test.go, testmain_test.go
- internal/apps/sm-kms/server/businesslogic | businesslogic.go, businesslogic_bench_test.go, businesslogic_crud_keyops_test.go, businesslogic_crud_test.go, businesslogic_crypto.go, businesslogic_crypto_roundtrip_test.go, businesslogic_fuzz_test.go, businesslogic_operations_test.go, businesslogic_property_test.go, businesslogic_tenantid_test.go, businesslogic_test.go, elastic_key_status_state_machine.go, elastic_key_status_state_machine_test.go, oam_orm_mapper.go, oam_orm_mapper_conversion_test.go, oam_orm_mapper_query.go, oam_orm_mapper_query_test.go, oam_orm_mapper_sort_test.go, oam_orm_mapper_test.go
- internal/apps/sm-kms/server/handler | handler_methods_test.go, handler_query_params_test.go, handler_response_test.go, handler_test.go, oam_oas_mapper.go, oam_oas_mapper_material.go, oas_handlers.go
- internal/apps/sm-kms/server/repository | migrations.go
- internal/apps/sm-kms/server/repository/migrations | 2001_kms_business_tables.down.sql, 2001_kms_business_tables.up.sql, 2002_add_tenant_id.down.sql, 2002_add_tenant_id.up.sql
- internal/apps/sm-kms/server/repository/orm | business_entities.go, business_entities_builders.go, business_entities_builders_test.go, business_entities_db_errors_test.go, business_entities_error_paths_test.go, business_entities_filter_application_keys_test.go, business_entities_filter_application_test.go, business_entities_filter_pagination_test.go, business_entities_filters.go, business_entities_filters_test.go, business_entities_material_key_test.go, business_entities_mutation_errors_test.go, business_entities_operations.go, business_entities_operations_test.go, business_entities_validation_test.go, business_entities_versioning_filter_test.go, orm_framework_types.go, orm_repository_test.go, orm_repository_test_util.go, orm_transaction_autocommit_test.go, orm_transaction_test.go, orm_transaction_verbose_test.go

### sm-im
- internal/apps/sm-im | im.go, im_cli_commands_test.go, im_cli_url_test.go, im_usage.go
- internal/apps/sm-im/client | concurrent_test.go, message.go, message_errorpaths_test.go, message_test.go, rotation_integration_test.go, service_integration_test.go, testmain_integration_test.go, web_client_integration_test.go
- internal/apps/sm-im/e2e | e2e_registration_test.go, e2e_test.go, e2e_tls_test.go, testmain_e2e_test.go
- internal/apps/sm-im/server | contracts_test.go, http_errors_test.go, http_test.go, im_database_test.go, im_lifecycle_test.go, im_port_conflict_test.go, im_server_lifecycle_test.go, public_server.go, public_server_test.go, registration_test.go, response_body_test.go, server.go, server_test.go, swagger.go, swagger_test.go, testmain_test.go
- internal/apps/sm-im/server/apis | messages.go, messages_dberror_test.go, messages_errorpaths_test.go, messages_receive_test.go, messages_test.go, sessions.go, sessions_test.go
- internal/apps/sm-im/server/config | config.go, config_errorpaths_test.go, config_test.go, config_test_helper.go
- internal/apps/sm-im/server/model | message.go, message_test.go, recipient_message_jwk.go, recipient_message_jwk_test.go
- internal/apps/sm-im/server/repository | concurrent_access_test.go, error_paths_test.go, message_recipient_jwk_repository.go, message_recipient_jwk_repository_test.go, message_repository.go, message_repository_test.go, migrations.go, migrations_test.go, testmain_test.go, user_repository.go, user_repository_adapter.go, user_repository_adapter_test.go, user_repository_test.go
- internal/apps/sm-im/server/repository/migrations | 3001_init.down.sql, 3001_init.up.sql
- internal/apps/sm-im/testing | testmain_helper.go, testmain_helper_test.go

### jose-ja
- internal/apps/jose-ja | ja.go, ja_cli_test.go, ja_lifecycle_test.go, ja_port_conflict_test.go, ja_usage.go, testmain_test.go
- internal/apps/jose-ja/client | client.go, package_test.go
- internal/apps/jose-ja/e2e | e2e_test.go, e2e_tls_test.go, testmain_e2e_test.go
- internal/apps/jose-ja/server | public_server.go, public_server_test.go, server.go, server_integration_test.go, server_test.go, swagger.go, swagger_test.go, testmain_test.go
- internal/apps/jose-ja/server/apis | jwk_handler.go, jwk_handler_elastic_test.go, jwk_handler_errors_test.go, jwk_handler_lifecycle_test.go, jwk_handler_material.go, jwk_handler_material_test.go, jwk_handler_missing_test.go, jwk_handler_test.go, sessions.go, sessions_test.go
- internal/apps/jose-ja/server/config | config.go, config_parse_errors_test.go, config_test.go, config_test_helper.go, config_validation_test.go, coverage
- internal/apps/jose-ja/server/model | models.go, models_test.go
- internal/apps/jose-ja/server/repository | audit_repository.go, audit_repository_error_test.go, audit_repository_list_test.go, audit_repository_test.go, coverage, elastic_jwk_repository.go, elastic_jwk_repository_edge_test.go, elastic_jwk_repository_error_test.go, elastic_jwk_repository_test.go, material_jwk_repository.go, material_jwk_repository_edge_test.go, material_jwk_repository_error_test.go, material_jwk_repository_rotate_test.go, material_jwk_repository_test.go, migrations.go, migrations_test.go, testmain_test.go
- internal/apps/jose-ja/server/repository/migrations | 4001_elastic_jwks.down.sql, 4001_elastic_jwks.up.sql, 4002_material_jwks.down.sql, 4002_material_jwks.up.sql, 4003_audit_config.down.sql, 4003_audit_config.up.sql, 4004_audit_log.down.sql, 4004_audit_log.up.sql
- internal/apps/jose-ja/server/service | audit_log_service.go, audit_log_service_error_test.go, audit_log_service_test.go, coverage, elastic_jwk_service.go, elastic_jwk_service_error_test.go, elastic_jwk_service_test.go, jwe_service.go, jwe_service_error_test.go, jwe_service_test.go, jwks_service.go, jwks_service_error_test.go, jwks_service_test.go, jws_service.go, jws_service_error_test.go, jws_service_test.go, jwt_service.go, jwt_service_encrypted_test.go, jwt_service_error_test.go, jwt_service_error2_test.go, jwt_service_test.go, mapping_service_parse_test.go, mapping_service_test.go, material_rotation_service.go, material_rotation_service_error_test.go, material_rotation_service_test.go, testmain_test.go

### skeleton-template
- internal/apps/skeleton-template | template.go, template_cli_test.go, template_lifecycle_test.go, template_port_conflict_test.go, template_usage.go, testmain_test.go
- internal/apps/skeleton-template/client | client.go, package_test.go
- internal/apps/skeleton-template/domain |
- internal/apps/skeleton-template/e2e | e2e_test.go, e2e_tls_test.go, testmain_e2e_test.go
- internal/apps/skeleton-template/repository |
- internal/apps/skeleton-template/repository/migrations |
- internal/apps/skeleton-template/server | server.go, server_integration_test.go, server_test.go, swagger.go, swagger_test.go, testmain_test.go
- internal/apps/skeleton-template/server/apis | handler.go, handler_test.go
- internal/apps/skeleton-template/server/config | config.go, config_test.go, config_test_helper.go
- internal/apps/skeleton-template/server/handler |
- internal/apps/skeleton-template/server/model | model.go, model_test.go
- internal/apps/skeleton-template/server/repository | item_repository.go, item_repository_test.go, migrations.go, migrations_test.go
- internal/apps/skeleton-template/server/repository/migrations | 11001_template_items.down.sql, 11001_template_items.up.sql, 11002_template_items_add_fields.down.sql, 11002_template_items_add_fields.up.sql

```

### Group 2: pki-ca, identity-*

```text
### pki-ca
- internal/apps/pki-ca | ca.go, ca_cli_test.go, ca_lifecycle_test.go, ca_port_conflict_test.go, ca_usage.go, README.md, testmain_test.go
- internal/apps/pki-ca/api |
- internal/apps/pki-ca/api/handler | handler.go, handler_ca_nilissuer_test.go, handler_cert_ops_test.go, handler_certs.go, handler_endpoint_test.go, handler_enroll_test.go, handler_enrollment_chain_test.go, handler_enrollment_test.go, handler_error_paths_test.go, handler_est.go, handler_est_csrattrs_test.go, handler_est_keygen_test.go, handler_est_timestamp_test.go, handler_fuzz_test.go, handler_keygen_test.go, handler_list_pagination_test.go, handler_mapping_test.go, handler_newhandler_validation_test.go, handler_ocsp.go, handler_ocsp_test.go, handler_realissuer_test.go, handler_revoke_test.go, handler_services_test.go, handler_storage_error_test.go, handler_test.go, handler_tsa_test.go
- internal/apps/pki-ca/bootstrap | bootstrap.go, bootstrap_errors_test.go, bootstrap_test.go
- internal/apps/pki-ca/cli | cli.go, cli_test.go, cli_validate_test.go
- internal/apps/pki-ca/client | client.go, package_test.go
- internal/apps/pki-ca/compliance | compliance.go, compliance_checker.go, compliance_errorpaths_test.go, compliance_report_test.go, compliance_test.go
- internal/apps/pki-ca/config | config.go, config_test.go, config_validation_test.go
- internal/apps/pki-ca/crypto | provider.go, provider_operations_test.go, provider_test.go
- internal/apps/pki-ca/domain | certificate.go, certificate_test.go, repository.go
- internal/apps/pki-ca/domain-v2 | model.go, model_test.go
- internal/apps/pki-ca/e2e | ca_e2e_test.go, testmain_e2e_test.go
- internal/apps/pki-ca/intermediate | intermediate.go, intermediate_test.go, intermediate_validation_test.go
- internal/apps/pki-ca/observability | observability.go, observability_ca.go, observability_test.go
- internal/apps/pki-ca/profile |
- internal/apps/pki-ca/profile/certificate | certificate.go, certificate_test.go, certificate_validation_test.go
- internal/apps/pki-ca/profile/subject | subject.go, subject_errorpaths_test.go, subject_test.go
- internal/apps/pki-ca/repository-v2 | migrations.go, migrations_test.go
- internal/apps/pki-ca/repository-v2/migrations | 5001_ca_items.down.sql, 5001_ca_items.up.sql
- internal/apps/pki-ca/security | security.go, security_csr_validation_test.go, security_report_test.go, security_test.go, security_threat_model.go, security_validate_test.go, security_validation_test.go
- internal/apps/pki-ca/server | admin.go, public_server.go, public_server_test.go, server.go, server_integration_test.go, server_lifecycle_test.go, server_test.go, swagger.go, swagger_test.go, testmain_test.go
- internal/apps/pki-ca/server/cmd | commands.go, commands_test.go
- internal/apps/pki-ca/server/config | config.go, config_defaults_test.go, config_error_paths_test.go, config_test.go
- internal/apps/pki-ca/server/middleware | mtls.go, mtls_errorpaths_test.go, mtls_test.go
- internal/apps/pki-ca/service |
- internal/apps/pki-ca/service/issuer | issuer.go, issuer_bench_test.go, issuer_invalid_test.go, issuer_operations_test.go, issuer_test.go, issuer_validation_test.go
- internal/apps/pki-ca/service/ra | ra.go, ra_cancel_test.go, ra_test.go, ra_workflow.go, ra_workflow_test.go
- internal/apps/pki-ca/service/revocation | revocation.go, revocation_crl_test.go, revocation_errorpaths_test.go, revocation_ocsp_test.go, revocation_test.go
- internal/apps/pki-ca/service/timestamp | timestamp.go, timestamp_serial_test.go, timestamp_test.go, timestamp_tsa.go, timestamp_tsa_errorpaths_test.go
- internal/apps/pki-ca/storage | storage.go, storage_operations_test.go, storage_test.go

### identity-authz
- internal/apps/identity-authz | authz.go, authz_cli_test.go, authz_contract_test.go, authz_usage.go, handlers_recovery_codes_test.go.TODO
- internal/apps/identity-authz/client | client.go, package_test.go
- internal/apps/identity-authz/clientauth | basic.go, basic_test.go, certificate_validator.go, certificate_validator_test.go, client_secret_jwt.go, client_secret_jwt_test.go, clientauth_error_scenarios_test.go, clientauth_test.go, integration_test.go, interface.go, jwt_tls_test.go, jwt_validator.go, jwt_validator_secret_test.go, jwt_validator_test.go, pbkdf2_hasher.go, pbkdf2_hasher_test.go, post.go, post_test.go, private_key_jwt.go, private_key_jwt_test.go, registry.go, registry_test.go, revocation.go, revocation_test.go, secret_hash.go, secret_hash_test.go, secret_hasher.go, secret_hasher_test.go, self_signed_auth.go, self_signed_auth_test.go, test_helpers_cert.go, test_helpers_test.go, tls_client_auth.go, tls_client_auth_test.go
- internal/apps/identity-authz/dpop | dpop.go, dpop_test.go, dpop_validation_test.go
- internal/apps/identity-authz/e2e | authz_e2e_test.go, rotation_test.go, testmain_e2e_test.go
- internal/apps/identity-authz/pkce | generator.go, pkce_test.go, validator.go
- internal/apps/identity-authz/server | admin.go, admin_error_test.go, admin_test.go, authz_lifecycle_test.go, authz_port_conflict_test.go, public_server.go, server.go, server_integration_test.go, swagger.go, swagger_test.go, testmain_test.go
- internal/apps/identity-authz/server/apis | authorization_request.go, authorization_request_store_test.go, authorization_request_test.go, authz_test.go, cleanup.go, cleanup_migration_test.go, cleanup_test.go, client_authentication.go, client_authentication_flow_test.go, client_authentication_test.go, code_generator.go, code_generator_test.go, device_code_generator.go, device_code_generator_test.go, handlers_authorize.go, handlers_authorize_pkce_test.go, handlers_authorize_test.go, handlers_authorize_validation_test.go, handlers_authz_code_test.go, handlers_cleanup_test.go, handlers_client_credentials_test.go, handlers_client_rotation.go, handlers_client_rotation_test.go, handlers_device_authorization.go, handlers_device_authorization_flow_integration_test.go, handlers_device_authorization_test.go, handlers_discovery.go, handlers_discovery_test.go, handlers_email_otp.go, handlers_email_otp_test.go, handlers_health.go, handlers_health_test.go, handlers_introspect_revoke.go, handlers_introspect_revoke_error_paths_test.go, handlers_introspect_revoke_errors_test.go, handlers_introspect_revoke_flow_test.go, handlers_introspect_revoke_test.go, handlers_introspect_test.go, handlers_introspection_revocation_flow_test.go, handlers_mfa_admin.go, handlers_mfa_admin_delete_test.go, handlers_mfa_admin_test.go, handlers_middleware_test.go, handlers_multitenant_isolation_test.go, handlers_par.go, handlers_par_test.go, handlers_par_validation_test.go, handlers_recovery_codes.go, handlers_recovery_codes_test.go, handlers_refresh_token_grant_test.go, handlers_refresh_token_test.go, handlers_routes_test.go, handlers_service_test.go, handlers_token.go, handlers_token_device.go, handlers_token_flow_test.go, handlers_token_grant_test.go, handlers_token_test.go, handlers_token_validation_test.go, handlers_totp.go, handlers_totp_backup_test.go, handlers_totp_helpers_test.go, handlers_totp_integration_test.go, handlers_totp_test.go, handlers_uncovered_paths_test.go, handlers_webauthn.go, middleware.go, middleware_test.go, performance_bench_test.go, request_uri_generator.go, request_uri_generator_test.go, routes.go, routes_test.go, service.go, service_lifecycle_test.go, service_test.go, swagger.go, swagger_test.go, test_helpers_test.go
- internal/apps/identity-authz/server/config | config.go, config_test.go, config_test_helper.go
- internal/apps/identity-authz/server/model | model.go, package_test.go
- internal/apps/identity-authz/server/repository | migrations.go, package_test.go
- internal/apps/identity-authz/server/repository/migrations | 6001_identity_authz_init.down.sql, 6001_identity_authz_init.up.sql
- internal/apps/identity-authz/unified | authz.go

### identity-idp
- internal/apps/identity-idp | idp.go, idp_cli_test.go, idp_contract_test.go, idp_usage.go
- internal/apps/identity-idp/auth | email_password.go, mfa.go, mfa_otp.go, mfa_telemetry.go, mfa_test.go, otp.go, passkey.go, profiles.go, totp.go, username_password.go, username_password_email_test.go, username_password_otp_test.go, username_password_test.go, username_password_totp_test.go
- internal/apps/identity-idp/client | client.go, package_test.go
- internal/apps/identity-idp/e2e | idp_e2e_test.go, testmain_e2e_test.go
- internal/apps/identity-idp/server | admin.go, admin_error_test.go, admin_test.go, idp_lifecycle_test.go, idp_port_conflict_test.go, public_server.go, server.go, server_integration_test.go, swagger.go, swagger_test.go, testmain_test.go
- internal/apps/identity-idp/server/apis | backchannel_logout.go, backchannel_logout_test.go, client_secret.go, client_secret_test.go, handlers_consent.go, handlers_consent_errors_test.go, handlers_consent_internal_test.go, handlers_consent_test.go, handlers_discovery.go, handlers_discovery_test.go, handlers_endsession_test.go, handlers_health.go, handlers_health_test.go, handlers_jwks.go, handlers_jwks_test.go, handlers_login.go, handlers_login_test.go, handlers_logout.go, handlers_logout_test.go, handlers_oidc_e2e_test.go, handlers_openapi_validation_test.go, handlers_parallel_safety_test.go, handlers_postgres_test.go, handlers_security_attacks_test.go, handlers_security_validation_rate_test.go, handlers_security_validation_test.go, handlers_token_expiration_test.go, handlers_userinfo.go, handlers_userinfo_claims_test.go, handlers_userinfo_internal_test.go, handlers_userinfo_jwt_test.go, handlers_userinfo_test.go, magic_test_constants.go, middleware.go, middleware_register_test.go, middleware_test.go, random.go, routes.go, routes_test.go, service.go, service_lifecycle_test.go, service_rotate_test.go, service_test.go, swagger.go, swagger_test.go, test_helpers_external_test.go, test_helpers_test.go
- internal/apps/identity-idp/server/apis/templates | consent.html, login.html
- internal/apps/identity-idp/server/config | config.go, config_test.go, config_test_helper.go
- internal/apps/identity-idp/server/model | model.go, package_test.go
- internal/apps/identity-idp/server/repository | migrations.go, package_test.go
- internal/apps/identity-idp/server/repository/migrations | 7001_identity_idp_init.down.sql, 7001_identity_idp_init.up.sql
- internal/apps/identity-idp/unified | idp.go
- internal/apps/identity-idp/userauth | adaptive_e2e_test.go, audit.go, audit_test.go, context_analyzer.go, context_analyzer_test.go, hardware.go, hardware_error_validation.go, hardware_error_validation_test.go, hardware_stub_test.go, interface.go, magic_link.go, magic_link_initiate_test.go, magic_link_test.go, mock_delivery.go, phone_call_otp.go, phone_call_otp_test.go, policy_loader.go, policy_loader_cache_test.go, policy_loader_impl.go, policy_loader_stepup_test.go, policy_loader_test.go, push_notification.go, push_notification_test.go, rate_limiter.go, rate_limiter_test.go, risk_based_auth.go, risk_engine.go, risk_engine_test.go, risk_engine_velocity_test.go, risk_scenarios_high_test.go, risk_scenarios_test.go, sms_otp.go, sms_otp_test.go, step_up_auth.go, step_up_auth_test.go, storage.go, storage_test.go, telemetry.go, telemetry_test.go, token_hashing.go, token_hashing_test.go, totp_hotp_auth.go, totp_hotp_auth_test.go, userauth_otp_success_test.go, username_password.go, username_password_test.go, username_password_verify_test.go, webauthn_authenticator.go, webauthn_authenticator_auth.go, webauthn_authenticator_test.go, webauthn_basic_test.go, webauthn_integration_test.go
- internal/apps/identity-idp/userauth/mocks | delivery_mock_validation_test.go, delivery_service.go, delivery_service_test.go

### identity-rp
- internal/apps/identity-rp | rp.go, rp_cli_test.go, rp_usage.go
- internal/apps/identity-rp/client | client.go, package_test.go
- internal/apps/identity-rp/e2e | rp_e2e_test.go, testmain_e2e_test.go
- internal/apps/identity-rp/server | public_server.go, rp_lifecycle_test.go, rp_port_conflict_test.go, rp_test.go, server.go, server_integration_test.go, swagger.go, swagger_test.go, testmain_test.go
- internal/apps/identity-rp/server/apis | handler.go, package_test.go
- internal/apps/identity-rp/server/config | config.go, config_test.go, config_test_helper.go
- internal/apps/identity-rp/server/model | model.go, package_test.go
- internal/apps/identity-rp/server/repository | migrations.go, package_test.go
- internal/apps/identity-rp/server/repository/migrations | 9001_identity_rp_init.down.sql, 9001_identity_rp_init.up.sql
- internal/apps/identity-rp/unified | rp.go

### identity-rs
- internal/apps/identity-rs | rs.go, rs_cli_test.go, rs_contract_test.go, rs_usage.go
- internal/apps/identity-rs/client | client.go, package_test.go
- internal/apps/identity-rs/e2e | rs_e2e_test.go, testmain_e2e_test.go
- internal/apps/identity-rs/server | admin.go, admin_test.go, public_server.go, rs_lifecycle_test.go, rs_port_conflict_test.go, server.go, server_integration_test.go, service.go, service_admin_test.go, service_test.go, swagger.go, swagger_test.go, testmain_test.go, validator.go
- internal/apps/identity-rs/server/apis | handler.go, package_test.go
- internal/apps/identity-rs/server/config | config.go, config_test.go, config_test_helper.go
- internal/apps/identity-rs/server/model | model.go, package_test.go
- internal/apps/identity-rs/server/repository | migrations.go, package_test.go
- internal/apps/identity-rs/server/repository/migrations | 8001_identity_rs_init.down.sql, 8001_identity_rs_init.up.sql
- internal/apps/identity-rs/unified | rs.go

### identity-spa
- internal/apps/identity-spa | spa.go, spa_cli_test.go, spa_usage.go
- internal/apps/identity-spa/client | client.go, package_test.go
- internal/apps/identity-spa/e2e | spa_e2e_test.go, testmain_e2e_test.go
- internal/apps/identity-spa/server | public_server.go, server.go, server_integration_test.go, spa_lifecycle_test.go, spa_port_conflict_test.go, spa_test.go, swagger.go, swagger_test.go, testmain_test.go
- internal/apps/identity-spa/server/apis | handler.go, package_test.go
- internal/apps/identity-spa/server/config | config.go, config_test.go, config_test_helper.go
- internal/apps/identity-spa/server/model | model.go, package_test.go
- internal/apps/identity-spa/server/repository | migrations.go, package_test.go
- internal/apps/identity-spa/server/repository/migrations | 10001_identity_spa_init.down.sql, 10001_identity_spa_init.up.sql
- internal/apps/identity-spa/unified | spa.go

```

### pki-ca SQL Migration Evidence

- Current migration SQL files are in:
  - `internal/apps/pki-ca/repository-v2/migrations/5001_ca_items.up.sql`
  - `internal/apps/pki-ca/repository-v2/migrations/5001_ca_items.down.sql`

---

## Question 1: Canonical `server/**` recursive structure to enforce for all 10 PS-IDs

**Question**: Which policy should V21 adopt as the target canonical recursive `server/**` structure across all 10 PS-IDs (with linter/template enforcement)?

**A)** Strict immediate canonical set:
- Required everywhere: `server/apis`, `server/businesslogic`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`
- Forbidden everywhere: `server/handler`, `server/service`, `server/cmd`, `server/middleware`, `server/repository/orm`, `server/apis/templates`
- One-shot migration for all 10 in V21

**B)** Transitional canonical set with sunset (recommended):
- Required everywhere: `server/apis`, `server/businesslogic`, `server/config`, `server/model`, `server/repository`, `server/repository/migrations`
- Temporary allowlist (must be retired by scheduled phases): `server/handler`, `server/service`, `server/cmd`, `server/middleware`, `server/repository/orm`, `server/apis/templates`
- Linter enforces required-now plus time-boxed deprecation plan

**C)** Minimal convergence:
- Require only: `server/apis`, `server/model`, `server/repository`
- Keep service-specific subdirectories indefinitely (no sunset)

**D)** Keep current mixed structure and only ensure required dirs exist (no consolidation mandate)

**E)**

**Answer**:

**Rationale**: This decision controls the all-10 migration scope, linter invariants, and how aggressively sprawl (especially pki-ca) is reduced.

---

## Question 2: pki-ca consolidation strategy under the selected canonical policy

**Question**: For pki-ca package/subdirectory sprawl, which execution strategy should tasks implement?

**A)** Full consolidation in V21:
- Move/merge pki-ca subdirectories to canonical targets immediately
- Migrate domain packages that sit outside canonical paths
- Remove legacy directories in same phase

**B)** Two-stage consolidation (recommended):
- Stage 1 (V21): establish canonical `server/**` directories, introduce wrappers/adapters, migrate SQL paths from `repository-v2/migrations` to `server/repository/migrations`
- Stage 2 (next phase): move domain-heavy packages (`bootstrap`, `compliance`, `intermediate`, `profile`, `service`, `storage`, etc.) behind canonical boundaries and remove legacy paths after compatibility gates pass

**C)** Structural-only for V21:
- Create canonical dirs and linter checks
- Keep pki-ca legacy package sprawl untouched

**D)** pki-ca-specific exception:
- Exempt pki-ca from canonical structure and keep bespoke layout

**E)**

**Answer**:

**Rationale**: Determines whether V21 includes concrete pki-ca sprawl reduction tasks versus deferring most consolidation work.

---

## Question 3: Root-level PS-ID directory policy for all 10 services

**Question**: Should V21 enforce a canonical root-level PS-ID directory policy in addition to `server/**` policy?

**A)** Yes, strict required-only root set for all 10 (recommended):
- Required: `client`, `e2e`, `server`
- Optional (explicitly approved only): `testing`, `unified`, authn/authz-specific modules
- All other root-level directories must be migrated or explicitly sunset

**B)** Yes, but service-class based policy:
- Identity services may keep additional authn/authz roots
- pki-ca may keep additional PKI roots
- SM/JOSE services follow strict root set

**C)** No root-level policy in V21; enforce only `server/**`

**D)** Keep current root-level sprawl and rely on naming conventions only

**E)**

**Answer**:

**Rationale**: This controls whether V21 includes all-10 root-level cleanup tasks or limits scope to `server/**` only.
