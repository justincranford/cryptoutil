// Copyright (c) 2025 Justin Cranford
//
//

package magic

// HTTP header constants.
const (
        // AuthorizationBearer - HTTP Bearer authentication scheme name (without trailing space).
        AuthorizationBearer = "Bearer"

        // AuthorizationBearerPrefix - HTTP Authorization Bearer scheme prefix (with trailing space).
        // Use this when constructing or parsing "Authorization: Bearer <token>" headers.
        AuthorizationBearerPrefix = "Bearer "
)
