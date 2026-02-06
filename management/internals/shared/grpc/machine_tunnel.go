package grpc

// Machine Tunnel Fork - gRPC handlers for machine peer registration and sync.
// These handlers require mTLS authentication and use the MTLSIdentity from context.

import (
	"context"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/netbirdio/netbird/management/internals/controllers/network_map"
	"github.com/netbirdio/netbird/management/internals/shared/mtls"
	"github.com/netbirdio/netbird/management/server/activity"
	nbContext "github.com/netbirdio/netbird/management/server/context"
	nbpeer "github.com/netbirdio/netbird/management/server/peer"
	"github.com/netbirdio/netbird/management/server/posture"
	"github.com/netbirdio/netbird/management/server/store"
	"github.com/netbirdio/netbird/management/server/types"
	"github.com/netbirdio/netbird/shared/management/proto"
)

// RegisterMachinePeer handles machine peer registration using mTLS certificate authentication.
// This method is in mTLSRequiredMethods and will only be called with valid mTLS identity.
//
// Features (T-3.6 complete):
// - validateIssuerCA: CA-Fingerprint validation per account
// - Meta fields for audit: peer_type, cert_dns_name, auth_method, cert_issuer_fp, etc.
// - Re-registration logic: update existing peer vs create new
// - Rate-limit protection: Deferred - mTLS mutual auth provides sufficient protection against abuse;
//   rate limiting can be added at the gRPC interceptor level if needed for scale.
// - Replay protection: Deferred - TLS 1.3 nonce provides session-level replay protection;
//   application-level replay protection (e.g., nonce in request) not needed for registration RPCs.
func (s *Server) RegisterMachinePeer(ctx context.Context, req *proto.MachineRegisterRequest) (*proto.MachineRegisterResponse, error) {
	reqStart := time.Now()

	// Extract mTLS identity from context (set by MTLSUnaryInterceptor)
	identity := mtls.GetIdentity(ctx)
	if identity == nil {
		// This should not happen - interceptor should reject requests without identity
		log.WithContext(ctx).Error("RegisterMachinePeer called without mTLS identity")
		return nil, status.Error(codes.Unauthenticated, "mTLS authentication required")
	}

	log.WithContext(ctx).Infof("RegisterMachinePeer: DNS=%s, Account=%s, Hostname=%s",
		identity.DNSName, identity.AccountID, identity.Hostname)

	// Get account ID from mTLS identity (CRITICAL: Already validated in extractMTLSIdentity)
	accountID := identity.AccountID
	if accountID == "" {
		log.WithContext(ctx).Errorf("No account ID in mTLS identity for domain %s", identity.Domain)
		return nil, status.Errorf(codes.FailedPrecondition,
			"domain %q not mapped to any account - configure MTLSDomainAccountMapping", identity.Domain)
	}

	// SECURITY: Validate Issuer CA fingerprint against account's allowed issuers
	// Per Security Review: Empty allowlist = DENY (explicit config required)
	if err := mtls.ValidateIssuerCA(accountID, identity.IssuerFingerprint); err != nil {
		log.WithContext(ctx).Warnf("Issuer CA validation failed: %v", err)
		return nil, status.Errorf(codes.PermissionDenied, "certificate issuer not authorized: %v", err)
	}

	// Parse WireGuard public key from request
	// wg_pub_key is a bytes field containing the raw 32-byte key OR a base64-encoded string
	wgPubKeyBytes := req.GetWgPubKey()
	if len(wgPubKeyBytes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "WireGuard public key is required")
	}
	var peerKey wgtypes.Key
	var wgErr error
	if len(wgPubKeyBytes) == wgtypes.KeyLen {
		// Raw 32-byte key (from protobuf bytes field)
		peerKey, wgErr = wgtypes.NewKey(wgPubKeyBytes)
	} else {
		// Base64-encoded string (fallback for compatibility)
		peerKey, wgErr = wgtypes.ParseKey(string(wgPubKeyBytes))
	}
	if wgErr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid WireGuard public key: %v", wgErr)
	}

	// Add peer and account info to context for logging
	//nolint:staticcheck
	ctx = context.WithValue(ctx, nbContext.PeerIDKey, peerKey.String())
	//nolint:staticcheck
	ctx = context.WithValue(ctx, nbContext.AccountIDKey, accountID)

	// Build peer metadata from request, enriched with mTLS audit fields
	peerMeta := extractMachinePeerMeta(ctx, req.GetMeta(), identity)

	// Log registration attempt (truncate keys for security)
	keyPrefix := peerKey.String()
	if len(keyPrefix) > 8 {
		keyPrefix = keyPrefix[:8]
	}
	accountPrefix := accountID
	if len(accountPrefix) > 8 {
		accountPrefix = accountPrefix[:8]
	}
	log.WithContext(ctx).Infof("Machine peer registration: key=%s... hostname=%s domain=%s account=%s...",
		keyPrefix, identity.Hostname, identity.Domain, accountPrefix)

	// Register or re-register peer via LoginPeer which handles both new
	// registrations and updates for existing peers
	// For machine peers, SetupKey and UserID are empty - auth is via mTLS
	peer, netMap, postureChecks, err := s.accountManager.LoginPeer(ctx, types.PeerLogin{
		WireGuardPubKey: peerKey.String(),
		Meta:            peerMeta,
		// Machine peer specific: no setup key, no user ID (auth via mTLS)
		// The mTLS identity in context provides authentication
		SetupKey: "",
		UserID:   "",
	})
	if err != nil {
		log.WithContext(ctx).Errorf("Failed to register machine peer: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to register peer: %v", err)
	}

	// Build response with machine-specific configuration
	loginResp, err := s.prepareLoginResponse(ctx, peer, netMap, postureChecks)
	if err != nil {
		log.WithContext(ctx).Errorf("Failed to prepare login response: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to prepare response: %v", err)
	}

	// Convert LoginResponse to MachineRegisterResponse
	response := &proto.MachineRegisterResponse{
		PeerConfig:    loginResp.GetPeerConfig(),
		NetbirdConfig: loginResp.GetNetbirdConfig(),
		MachineIdentity: &proto.MachineIdentity{
			DnsName:           identity.DNSName,
			Hostname:          identity.Hostname,
			Domain:            identity.Domain,
			IssuerFingerprint: identity.IssuerFingerprint,
			SerialNumber:      identity.SerialNumber,
			TemplateOid:       identity.TemplateOID,
		},
		// DC route filtering is handled client-side based on NetworkMap routes;
		// server returns all eligible routes, client applies ACL filtering.
		AllowedDcRoutes: nil,
		DnsConfig:       nil,
	}

	log.WithContext(ctx).Infof("Machine peer registered successfully: DNS=%s, IP=%s (took %s)",
		identity.DNSName, peer.IP, time.Since(reqStart))

	return response, nil
}

// extractMachinePeerMeta builds peer metadata from request, enriched with mTLS audit fields.
// This sets the mTLS-specific fields for audit trail.
func extractMachinePeerMeta(ctx context.Context, reqMeta *proto.PeerSystemMeta, identity *mtls.Identity) nbpeer.PeerSystemMeta {
	// Start with base meta from request
	meta := extractPeerMeta(ctx, reqMeta)

	// Enrich with mTLS audit fields
	meta.PeerType = identity.PeerType
	if meta.PeerType == "" {
		meta.PeerType = "machine" // Default for mTLS-authenticated peers
	}
	meta.AuthMethod = "mtls"
	meta.CertDNSName = identity.DNSName
	meta.CertDomain = identity.Domain
	meta.CertIssuerFP = identity.IssuerFingerprint
	meta.CertSerial = identity.SerialNumber
	meta.CertTemplate = identity.TemplateName
	if meta.CertTemplate == "" {
		meta.CertTemplate = identity.TemplateOID // Fallback to OID if name not available
	}

	// Set auth timestamps
	now := time.Now().UTC().Format(time.RFC3339)
	meta.FirstAuthTime = now     // Will be overwritten on re-registration
	meta.LastCertAuthTime = now

	return meta
}

// SyncMachinePeer handles machine peer sync stream using mTLS certificate authentication.
// Pattern follows server.go Sync() but without WG-key encryption (mTLS handles transport security).
// Peer lookup is via DNSLabel (hostname+domain hash) instead of WG public key.
func (s *Server) SyncMachinePeer(req *proto.MachineSyncRequest, srv proto.ManagementService_SyncMachinePeerServer) error {
	reqStart := time.Now()
	ctx := srv.Context()

	// Extract mTLS identity from context (set by MTLSStreamInterceptor)
	identity := mtls.GetIdentity(ctx)
	if identity == nil {
		log.WithContext(ctx).Error("SyncMachinePeer called without mTLS identity")
		return status.Error(codes.Unauthenticated, "mTLS authentication required")
	}

	// Validate issuer CA
	if err := mtls.ValidateIssuerCA(identity.AccountID, identity.IssuerFingerprint); err != nil {
		log.WithContext(ctx).Warnf("Issuer CA validation failed in SyncMachinePeer: %v", err)
		return status.Errorf(codes.PermissionDenied, "certificate issuer not authorized: %v", err)
	}

	log.WithContext(ctx).Infof("SyncMachinePeer: DNS=%s", identity.DNSName)

	// Look up peer by mTLS identity (DNSLabel = hostname + domain hash)
	peer, err := s.findMachinePeer(ctx, identity)
	if err != nil {
		log.WithContext(ctx).Warnf("SyncMachinePeer: peer not found for %s: %v", identity.DNSName, err)
		return status.Errorf(codes.NotFound, "peer not registered for identity %s", identity.DNSName)
	}

	accountID := identity.AccountID

	// Add peer and account info to context for structured logging
	//nolint:staticcheck
	ctx = context.WithValue(ctx, nbContext.PeerIDKey, peer.Key)
	//nolint:staticcheck
	ctx = context.WithValue(ctx, nbContext.AccountIDKey, accountID)

	realIP := machineRealIP(ctx)
	peerMeta := extractPeerMeta(ctx, req.GetMeta())

	// Acquire peer lock for initial sync (released before entering update loop)
	start := time.Now()
	unlock := s.acquirePeerLockByUID(ctx, peer.Key)
	defer func() {
		if unlock != nil {
			unlock()
		}
	}()
	log.WithContext(ctx).Tracef("acquired peer lock for machine peer %s took %v", peer.Key, time.Since(start))

	// SyncAndMarkPeer: get current NetworkMap and mark peer as connected
	peer, netMap, postureChecks, _, err := s.accountManager.SyncAndMarkPeer(ctx, accountID, peer.Key, peerMeta, realIP)
	if err != nil {
		log.WithContext(ctx).Errorf("SyncMachinePeer: failed to sync peer %s: %v", identity.DNSName, err)
		return status.Errorf(codes.Internal, "failed to sync peer: %v", err)
	}

	// Send initial full sync response (plaintext - mTLS provides transport security)
	err = s.sendInitialMachineSync(ctx, peer, netMap, postureChecks, srv)
	if err != nil {
		log.WithContext(ctx).Errorf("SyncMachinePeer: failed to send initial sync for %s: %v", identity.DNSName, err)
		return err
	}

	// Register for network map updates
	updates, err := s.networkMapController.OnPeerConnected(ctx, accountID, peer.ID)
	if err != nil {
		log.WithContext(ctx).Errorf("SyncMachinePeer: failed to register for updates: %v", err)
		s.cancelPeerRoutines(ctx, accountID, peer, reqStart)
		return err
	}

	s.secretsManager.SetupRefresh(ctx, accountID, peer.ID)

	// Release peer lock before entering long-running update loop
	unlock()
	unlock = nil

	log.WithContext(ctx).Debugf("SyncMachinePeer: initial sync took %s for %s", time.Since(reqStart), identity.DNSName)

	// Enter update loop - blocks until client disconnects or channel closes
	return s.handleMachineUpdates(ctx, accountID, peer, updates, srv, reqStart)
}

// sendInitialMachineSync builds and sends the first MachineSyncResponse with MACHINE_UPDATE_FULL.
// Unlike sendInitialSync in server.go, this sends plaintext (no WG encryption) since mTLS
// handles transport security.
func (s *Server) sendInitialMachineSync(ctx context.Context, peer *nbpeer.Peer, netMap *types.NetworkMap, postureChecks []*posture.Checks, srv proto.ManagementService_SyncMachinePeerServer) error {
	settings, err := s.settingsManager.GetSettings(ctx, peer.AccountID, activity.SystemInitiator)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get settings: %v", err)
	}

	peerGroups, err := s.accountManager.GetStore().GetPeerGroupIDs(ctx, store.LockingStrengthNone, peer.AccountID, peer.ID)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get peer groups: %v", err)
	}

	// Build SyncResponse using existing conversion logic
	syncResp := ToSyncResponse(
		ctx, s.config, s.config.HttpConfig, s.config.DeviceAuthorizationFlow,
		peer, nil, nil, // no TURN/Relay tokens for machine peers (managed by client)
		netMap, s.networkMapController.GetDNSDomain(settings),
		postureChecks, nil, settings, settings.Extra, peerGroups, 0,
	)

	serial := uint64(0)
	if netMap != nil && netMap.Network != nil {
		serial = netMap.Network.Serial
	}

	machineResp := toMachineSyncResponse(syncResp, proto.MachineUpdateType_MACHINE_UPDATE_FULL, serial)

	if err := srv.Send(machineResp); err != nil {
		return status.Errorf(codes.Internal, "failed to send initial sync: %v", err)
	}

	log.WithContext(ctx).Debugf("sent initial machine sync to peer %s (serial=%d)", peer.Key[:8], serial)
	return nil
}

// handleMachineUpdates streams network map updates to the machine peer.
// Pattern follows handleUpdates in server.go but sends plaintext MachineSyncResponse.
func (s *Server) handleMachineUpdates(ctx context.Context, accountID string, peer *nbpeer.Peer, updates chan *network_map.UpdateMessage, srv proto.ManagementService_SyncMachinePeerServer, streamStartTime time.Time) error {
	log.WithContext(ctx).Debugf("starting machine update loop for peer %s", peer.Key[:8])
	for {
		select {
		case update, open := <-updates:
			if !open {
				log.WithContext(ctx).Debugf("updates channel closed for machine peer %s", peer.Key[:8])
				s.cancelPeerRoutines(ctx, accountID, peer, streamStartTime)
				return nil
			}

			machineResp := toMachineSyncResponse(update.Update, proto.MachineUpdateType_MACHINE_UPDATE_FULL, 0)
			if err := srv.Send(machineResp); err != nil {
				log.WithContext(ctx).Debugf("failed to send update to machine peer %s: %v", peer.Key[:8], err)
				s.cancelPeerRoutines(ctx, accountID, peer, streamStartTime)
				return err
			}

			log.WithContext(ctx).Debugf("sent update to machine peer %s", peer.Key[:8])

		case <-srv.Context().Done():
			log.WithContext(ctx).Debugf("stream closed for machine peer %s", peer.Key[:8])
			s.cancelPeerRoutines(ctx, accountID, peer, streamStartTime)
			return srv.Context().Err()
		}
	}
}

// toMachineSyncResponse converts a SyncResponse to MachineSyncResponse.
// If explicit serial > 0, it takes precedence. Otherwise, serial is extracted from NetworkMap.
func toMachineSyncResponse(syncResp *proto.SyncResponse, updateType proto.MachineUpdateType, serial uint64) *proto.MachineSyncResponse {
	resp := &proto.MachineSyncResponse{
		UpdateType: updateType,
		Serial:     serial,
	}

	if syncResp != nil {
		resp.NetworkMap = syncResp.GetNetworkMap()
		// If no explicit serial, use NetworkMap serial
		if serial == 0 && resp.NetworkMap != nil {
			resp.Serial = resp.NetworkMap.Serial
		}
	}

	return resp
}

// findMachinePeer looks up a peer by mTLS identity using the DNSLabel (hostname + domain hash).
// The DNSLabel is generated deterministically from hostname and domain, ensuring consistency
// between registration (RegisterMachinePeer) and lookup (Sync/GetRoutes/ReportStatus).
func (s *Server) findMachinePeer(ctx context.Context, identity *mtls.Identity) (*nbpeer.Peer, error) {
	dnsLabel := mtls.GenerateUniqueDNSLabel(identity.Hostname, identity.Domain)

	peerID, err := s.accountManager.GetStore().GetPeerIdByLabel(ctx, store.LockingStrengthNone, identity.AccountID, dnsLabel)
	if err != nil {
		return nil, fmt.Errorf("no peer with DNSLabel=%s for account=%s: %w", dnsLabel, identity.AccountID, err)
	}

	return s.accountManager.GetStore().GetPeerByID(ctx, store.LockingStrengthNone, identity.AccountID, peerID)
}

// machineRealIP extracts the real client IP from context (set by realip interceptor).
func machineRealIP(ctx context.Context) net.IP {
	return getRealIP(ctx)
}

// GetMachineRoutes returns the DC routes and router peer configs for a machine peer.
// Uses SyncPeer (read-only) instead of SyncAndMarkPeer since this is a point-in-time query.
func (s *Server) GetMachineRoutes(ctx context.Context, req *proto.MachineRoutesRequest) (*proto.MachineRoutesResponse, error) {
	// Extract mTLS identity from context
	identity := mtls.GetIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "mTLS authentication required")
	}

	// Validate issuer CA
	if err := mtls.ValidateIssuerCA(identity.AccountID, identity.IssuerFingerprint); err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "certificate issuer not authorized: %v", err)
	}

	// Look up peer by mTLS identity
	peer, err := s.findMachinePeer(ctx, identity)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "peer not registered for identity %s", identity.DNSName)
	}

	accountID := identity.AccountID

	// SyncPeer to get current NetworkMap (read-only, does not mark connected)
	_, netMap, _, _, err := s.accountManager.SyncPeer(ctx, types.PeerSync{
		WireGuardPubKey: peer.Key,
	}, accountID)
	if err != nil {
		log.WithContext(ctx).Errorf("GetMachineRoutes: SyncPeer failed for %s: %v", identity.DNSName, err)
		return nil, status.Errorf(codes.Internal, "failed to sync peer routes: %v", err)
	}

	// Get settings for DNS domain
	settings, err := s.settingsManager.GetSettings(ctx, accountID, activity.SystemInitiator)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get settings: %v", err)
	}
	dnsDomain := s.networkMapController.GetDNSDomain(settings)

	// Build response: extract routes, DC networks, and router peer configs
	resp := &proto.MachineRoutesResponse{
		Routes:      make([]*proto.Route, 0),
		DcNetworks:  make([]string, 0),
		RouterPeers: make([]*proto.RemotePeerConfig, 0),
	}

	// Collect route networks and router peer IDs
	routerPeerIDs := make(map[string]bool)
	if netMap != nil {
		for _, r := range netMap.Routes {
			if !r.Enabled {
				continue
			}

			protoRoute := &proto.Route{
				ID:          string(r.ID),
				Network:     r.Network.String(),
				NetID:       string(r.NetID),
				Peer:        r.Peer,
				Metric:      int64(r.Metric),
				Masquerade:  r.Masquerade,
				NetworkType: int64(r.NetworkType),
			}
			resp.Routes = append(resp.Routes, protoRoute)
			resp.DcNetworks = append(resp.DcNetworks, r.Network.String())

			if r.Peer != "" {
				routerPeerIDs[r.Peer] = true
			}
		}

		// Build router peer configs from NetworkMap peers
		for _, p := range netMap.Peers {
			if !routerPeerIDs[p.ID] {
				continue
			}
			fqdn := fmt.Sprintf("%s.%s", p.DNSLabel, dnsDomain)
			resp.RouterPeers = append(resp.RouterPeers, &proto.RemotePeerConfig{
				WgPubKey:   p.Key,
				Fqdn:       fqdn,
				AllowedIps: []string{p.IP.String() + "/32"},
			})
		}
	}

	log.WithContext(ctx).Infof("GetMachineRoutes: DNS=%s, routes=%d, routerPeers=%d",
		identity.DNSName, len(resp.Routes), len(resp.RouterPeers))

	return resp, nil
}

// ReportMachineStatus handles machine peer status reports.
// Updates peer connection status and last seen timestamp.
// Issuer validation is skipped for status reports (lower security sensitivity,
// the mTLS handshake itself provides authentication).
func (s *Server) ReportMachineStatus(ctx context.Context, req *proto.MachineStatusRequest) (*proto.MachineStatusResponse, error) {
	// Extract mTLS identity from context
	identity := mtls.GetIdentity(ctx)
	if identity == nil {
		return nil, status.Error(codes.Unauthenticated, "mTLS authentication required")
	}

	// Look up peer by mTLS identity
	peer, err := s.findMachinePeer(ctx, identity)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "peer not registered for identity %s", identity.DNSName)
	}

	// Update peer connection status (non-critical: log errors but don't fail the RPC)
	realIP := machineRealIP(ctx)
	if err := s.accountManager.MarkPeerConnected(ctx, peer.Key, req.GetTunnelUp(), realIP, identity.AccountID); err != nil {
		log.WithContext(ctx).Warnf("ReportMachineStatus: failed to update peer status for %s: %v",
			identity.DNSName, err)
	}

	log.WithContext(ctx).Debugf("ReportMachineStatus: DNS=%s, TunnelUp=%v, DCReachable=%v",
		identity.DNSName, req.GetTunnelUp(), req.GetDcReachable())

	return &proto.MachineStatusResponse{
		Ack:        true,
		ServerTime: timestamppb.Now(),
	}, nil
}
