// Copyright (c) 2025-2026 Justin Cranford.

package listener

func ExportedSetAdminServerShutdownDelayFn(server *AdminServer, shutdownDelayFn func()) {
	server.shutdownDelayFn = shutdownDelayFn
}
