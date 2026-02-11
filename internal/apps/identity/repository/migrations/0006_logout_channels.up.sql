-- Migration: 0006_logout_channels.up.sql
-- Description: Add front-channel and back-channel logout configuration columns to clients table.
-- Reference: OpenID Connect Front-Channel Logout 1.0, OpenID Connect Back-Channel Logout 1.0

ALTER TABLE clients ADD COLUMN frontchannel_logout_uri TEXT DEFAULT '';
ALTER TABLE clients ADD COLUMN frontchannel_logout_session_required BOOLEAN DEFAULT false;
ALTER TABLE clients ADD COLUMN backchannel_logout_uri TEXT DEFAULT '';
ALTER TABLE clients ADD COLUMN backchannel_logout_session_required BOOLEAN DEFAULT false;
