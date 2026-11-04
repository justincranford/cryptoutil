# Protected prepublished PKI verification

The existing suite executable supports:

```text
/app/cryptoutil pki-verify --input=/run/cryptoutil/secrets/OWNED-SECRET/pki-verification.json
```

This is read-only validation of prepublished material, not generation,
Kubernetes publication, token use or a substitute for cluster acceptance.

## Strict version-1 manifest

- `version`: exactly `1`.
- `files`: nonempty map from exact clean
  `/run/cryptoutil/secrets/SECRET/KEY` paths to SHA256 hex digests of protected
  bytes. No traversal, alternate path spellings or manifest self-hash.
- `certificates`: nonempty array with explicit `certificate`, `key`, `trust`,
  `usage`, `dnsNames`, `commonName`. Usage is `server` or `client`; server
  bindings require explicit DNS names. Empty fields are supplied explicitly
  when applicable, not omitted.

Unknown/duplicate/case-mismatched fields and malformed values reject.
Each file is read into bounded memory, hash-checked, and validated under the
existing approved certificate/private-key/trust/validity/signature/EKU/identity
policy. Every requested certificate/private key must appear in a
certificate/key/trust binding; arbitrary extra material cannot hide behind a
valid leaf. Trust anchors and expected DNS/CommonName policy are explicit,
never derived from a token or fetched untrusted leaf.

The host lifecycle verifier additionally requires every restricted mounted
Secret key except this manifest itself to be hash-bound. The canonical Job
contract requires its exact manifest to be in an owned mounted source, with
no API token and no Secret writes. Manifest/material bytes are not logged or
persisted by the verifier.

## Completion barrier

Host `pki-publish` verifies-or-creates approved existing protected output before
Jobs. The later execution imports the exact same Secret type/key set/bytes.
The host validates all verification plans and runtime ordinal/admin/PG policy
before contacting Kubernetes. After each owned verification Job completes, it
uses read-only APIs to recheck Job UID/revision/spec/security, Secret
UID/type/full data/owner labels, and current certificate/runtime policy.
Only then can the dependent Job start.

`pki-init` remains a filesystem generator. Lifecycle execution rejects that
generation-only command because its ephemeral output is not the immutable
prepublished bundle. No create RBAC, exec/log transport, overwrite or hidden
publication is added. A protected regeneration/publication alternative needs
a separate owner-approved lifecycle topology.
