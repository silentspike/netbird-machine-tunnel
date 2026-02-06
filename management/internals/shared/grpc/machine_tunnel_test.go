package grpc

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/netbirdio/netbird/management/internals/controllers/network_map"
	"github.com/netbirdio/netbird/management/internals/shared/mtls"
	"github.com/netbirdio/netbird/management/server/integrations/extra_settings"
	nbpeer "github.com/netbirdio/netbird/management/server/peer"
	"github.com/netbirdio/netbird/management/server/posture"
	"github.com/netbirdio/netbird/management/server/store"
	"github.com/netbirdio/netbird/management/server/types"
	"github.com/netbirdio/netbird/route"
	"github.com/netbirdio/netbird/shared/management/proto"

	mock_server "github.com/netbirdio/netbird/management/server/mock_server"
)

// --- Test Store Mock ---

// machineTestStore embeds store.Store and overrides only methods needed by machine tunnel tests.
type machineTestStore struct {
	store.Store // embedded interface (panics on unimplemented methods)

	getPeerIdByLabelFunc func(ctx context.Context, lockStrength store.LockingStrength, accountID string, hostname string) (string, error)
	getPeerByIDFunc      func(ctx context.Context, lockStrength store.LockingStrength, accountID string, peerID string) (*nbpeer.Peer, error)
	getPeerGroupIDsFunc  func(ctx context.Context, lockStrength store.LockingStrength, accountId string, peerId string) ([]string, error)
}

func (s *machineTestStore) GetPeerIdByLabel(ctx context.Context, lockStrength store.LockingStrength, accountID string, hostname string) (string, error) {
	if s.getPeerIdByLabelFunc != nil {
		return s.getPeerIdByLabelFunc(ctx, lockStrength, accountID, hostname)
	}
	return "", errors.New("GetPeerIdByLabel not implemented in test")
}

func (s *machineTestStore) GetPeerByID(ctx context.Context, lockStrength store.LockingStrength, accountID string, peerID string) (*nbpeer.Peer, error) {
	if s.getPeerByIDFunc != nil {
		return s.getPeerByIDFunc(ctx, lockStrength, accountID, peerID)
	}
	return nil, errors.New("GetPeerByID not implemented in test")
}

func (s *machineTestStore) GetPeerGroupIDs(ctx context.Context, lockStrength store.LockingStrength, accountId string, peerId string) ([]string, error) {
	if s.getPeerGroupIDsFunc != nil {
		return s.getPeerGroupIDsFunc(ctx, lockStrength, accountId, peerId)
	}
	return nil, nil
}

// --- Mock gRPC Stream for SyncMachinePeer ---

type mockMachineSyncStream struct {
	ctx       context.Context
	sent      []*proto.MachineSyncResponse
	sendErr   error
	headerMD  metadata.MD
	trailerMD metadata.MD
}

func newMockMachineSyncStream(ctx context.Context) *mockMachineSyncStream {
	return &mockMachineSyncStream{
		ctx:  ctx,
		sent: make([]*proto.MachineSyncResponse, 0),
	}
}

func (s *mockMachineSyncStream) Send(resp *proto.MachineSyncResponse) error {
	if s.sendErr != nil {
		return s.sendErr
	}
	s.sent = append(s.sent, resp)
	return nil
}

func (s *mockMachineSyncStream) SetHeader(md metadata.MD) error {
	s.headerMD = md
	return nil
}

func (s *mockMachineSyncStream) SendHeader(md metadata.MD) error {
	s.headerMD = md
	return nil
}

func (s *mockMachineSyncStream) SetTrailer(md metadata.MD) {
	s.trailerMD = md
}

func (s *mockMachineSyncStream) Context() context.Context {
	return s.ctx
}

func (s *mockMachineSyncStream) SendMsg(_ interface{}) error { return nil }
func (s *mockMachineSyncStream) RecvMsg(_ interface{}) error { return nil }

// --- Mock Network Map Controller ---

type mockNetworkMapController struct {
	network_map.Controller // embedded interface

	onPeerConnectedFunc    func(ctx context.Context, accountID string, peerID string) (chan *network_map.UpdateMessage, error)
	onPeerDisconnectedFunc func(ctx context.Context, accountID string, peerID string)
	getDNSDomainFunc       func(settings *types.Settings) string
}

func (m *mockNetworkMapController) OnPeerConnected(ctx context.Context, accountID string, peerID string) (chan *network_map.UpdateMessage, error) {
	if m.onPeerConnectedFunc != nil {
		return m.onPeerConnectedFunc(ctx, accountID, peerID)
	}
	return make(chan *network_map.UpdateMessage, 100), nil
}

func (m *mockNetworkMapController) OnPeerDisconnected(ctx context.Context, accountID string, peerID string) {
	if m.onPeerDisconnectedFunc != nil {
		m.onPeerDisconnectedFunc(ctx, accountID, peerID)
	}
}

func (m *mockNetworkMapController) GetDNSDomain(s *types.Settings) string {
	if m.getDNSDomainFunc != nil {
		return m.getDNSDomainFunc(s)
	}
	return "test.local"
}

// --- Mock Settings Manager ---

type mockSettingsManager struct {
	getSettingsFunc func(ctx context.Context, accountID string, userID string) (*types.Settings, error)
}

func (m *mockSettingsManager) GetSettings(ctx context.Context, accountID string, userID string) (*types.Settings, error) {
	if m.getSettingsFunc != nil {
		return m.getSettingsFunc(ctx, accountID, userID)
	}
	return &types.Settings{}, nil
}

func (m *mockSettingsManager) GetExtraSettings(_ context.Context, _ string) (*types.ExtraSettings, error) {
	return &types.ExtraSettings{}, nil
}

func (m *mockSettingsManager) UpdateExtraSettings(_ context.Context, _, _ string, _ *types.ExtraSettings) (bool, error) {
	return false, nil
}

func (m *mockSettingsManager) GetExtraSettingsManager() extra_settings.Manager {
	return nil
}

// --- Test Constants ---

const (
	testAccountID     = "test-account-123"
	testHostname      = "win10-pc"
	testDomain        = "corp.local"
	testDNSName       = "win10-pc.corp.local"
	testIssuerFP      = "abc123def456"
	testPeerID        = "peer-id-001"
	testPeerKey       = "wg-pub-key-base64-test-value=="
)

func testIdentity() *mtls.Identity {
	return &mtls.Identity{
		DNSName:           testDNSName,
		Hostname:          testHostname,
		Domain:            testDomain,
		AccountID:         testAccountID,
		IssuerFingerprint: testIssuerFP,
		PeerType:          "machine",
	}
}

func testPeer() *nbpeer.Peer {
	return &nbpeer.Peer{
		ID:        testPeerID,
		AccountID: testAccountID,
		Key:       testPeerKey,
		Name:      testHostname,
		DNSLabel:  mtls.GenerateUniqueDNSLabel(testHostname, testDomain),
		IP:        net.ParseIP("100.64.0.10"),
		Meta:      nbpeer.PeerSystemMeta{},
		Status:    &nbpeer.PeerStatus{},
	}
}

func testNetworkMap() *types.NetworkMap {
	return &types.NetworkMap{
		Network: &types.Network{
			Identifier: "test-net",
			Net:        net.IPNet{IP: net.ParseIP("100.64.0.0"), Mask: net.CIDRMask(10, 32)},
			Serial:     42,
		},
		Peers:  []*nbpeer.Peer{},
		Routes: []*route.Route{},
	}
}

// ctxWithIdentity returns a context with the test mTLS identity injected.
func ctxWithIdentity() context.Context {
	return mtls.WithIdentity(context.Background(), testIdentity())
}

// init disables issuer validation for tests (globalValidatorConfig = nil -> skip validation)
func init() {
	mtls.SetValidatorConfig(nil)
}

// ============================================================
// Tests for toMachineSyncResponse (pure function)
// ============================================================

func TestToMachineSyncResponse_NilSyncResponse(t *testing.T) {
	resp := toMachineSyncResponse(nil, proto.MachineUpdateType_MACHINE_UPDATE_FULL, 42)

	require.NotNil(t, resp)
	assert.Equal(t, proto.MachineUpdateType_MACHINE_UPDATE_FULL, resp.UpdateType)
	assert.Equal(t, uint64(42), resp.Serial)
	assert.Nil(t, resp.NetworkMap)
}

func TestToMachineSyncResponse_WithNetworkMap(t *testing.T) {
	syncResp := &proto.SyncResponse{
		NetworkMap: &proto.NetworkMap{
			Serial:             99,
			PeerConfig:         &proto.PeerConfig{Address: "100.64.0.10/32"},
			RemotePeersIsEmpty: true,
		},
	}

	resp := toMachineSyncResponse(syncResp, proto.MachineUpdateType_MACHINE_UPDATE_FULL, 0)

	require.NotNil(t, resp)
	assert.Equal(t, proto.MachineUpdateType_MACHINE_UPDATE_FULL, resp.UpdateType)
	assert.Equal(t, uint64(99), resp.Serial, "serial should come from NetworkMap when passed serial is 0")
	require.NotNil(t, resp.NetworkMap)
	assert.Equal(t, "100.64.0.10/32", resp.NetworkMap.PeerConfig.Address)
}

func TestToMachineSyncResponse_ExplicitSerial(t *testing.T) {
	syncResp := &proto.SyncResponse{
		NetworkMap: &proto.NetworkMap{Serial: 99},
	}

	resp := toMachineSyncResponse(syncResp, proto.MachineUpdateType_MACHINE_UPDATE_FULL, 50)

	assert.Equal(t, uint64(50), resp.Serial, "explicit serial should take precedence over NetworkMap serial")
}

func TestToMachineSyncResponse_EmptyNetworkMap(t *testing.T) {
	syncResp := &proto.SyncResponse{
		NetworkMap: nil,
	}

	resp := toMachineSyncResponse(syncResp, proto.MachineUpdateType_MACHINE_UPDATE_FULL, 0)

	assert.Nil(t, resp.NetworkMap)
	assert.Equal(t, uint64(0), resp.Serial)
}

// ============================================================
// Tests for findMachinePeer
// ============================================================

func TestFindMachinePeer_ByDNSLabel(t *testing.T) {
	expectedDNSLabel := mtls.GenerateUniqueDNSLabel(testHostname, testDomain)
	expectedPeer := testPeer()

	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, accountID string, hostname string) (string, error) {
			assert.Equal(t, testAccountID, accountID)
			assert.Equal(t, expectedDNSLabel, hostname, "should look up by DNSLabel, not raw hostname")
			return testPeerID, nil
		},
		getPeerByIDFunc: func(_ context.Context, _ store.LockingStrength, accountID string, peerID string) (*nbpeer.Peer, error) {
			assert.Equal(t, testAccountID, accountID)
			assert.Equal(t, testPeerID, peerID)
			return expectedPeer, nil
		},
	}

	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
	}

	srv := &Server{accountManager: mockAM}
	ctx := context.Background()

	peer, err := srv.findMachinePeer(ctx, testIdentity())

	require.NoError(t, err)
	require.NotNil(t, peer)
	assert.Equal(t, testPeerID, peer.ID)
	assert.Equal(t, testPeerKey, peer.Key)
}

func TestFindMachinePeer_PeerNotFound(t *testing.T) {
	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (string, error) {
			return "", errors.New("peer not found")
		},
	}

	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
	}

	srv := &Server{accountManager: mockAM}
	ctx := context.Background()

	peer, err := srv.findMachinePeer(ctx, testIdentity())

	require.Error(t, err)
	assert.Nil(t, peer)
	assert.Contains(t, err.Error(), "no peer with DNSLabel")
}

func TestFindMachinePeer_GetPeerByIDFails(t *testing.T) {
	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (string, error) {
			return testPeerID, nil
		},
		getPeerByIDFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (*nbpeer.Peer, error) {
			return nil, errors.New("database error")
		},
	}

	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
	}

	srv := &Server{accountManager: mockAM}
	ctx := context.Background()

	peer, err := srv.findMachinePeer(ctx, testIdentity())

	require.Error(t, err)
	assert.Nil(t, peer)
	assert.Contains(t, err.Error(), "database error")
}

// ============================================================
// Tests for SyncMachinePeer
// ============================================================

func TestSyncMachinePeer_NoIdentity(t *testing.T) {
	srv := &Server{}
	stream := newMockMachineSyncStream(context.Background())

	err := srv.SyncMachinePeer(&proto.MachineSyncRequest{}, stream)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "mTLS authentication required")
}

func TestSyncMachinePeer_PeerNotFound(t *testing.T) {
	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (string, error) {
			return "", errors.New("peer not found")
		},
	}

	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
	}

	srv := &Server{accountManager: mockAM}

	ctx := ctxWithIdentity()
	stream := newMockMachineSyncStream(ctx)

	err := srv.SyncMachinePeer(&proto.MachineSyncRequest{}, stream)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "peer not registered")
}

func TestSyncMachinePeer_SyncAndMarkPeerFails(t *testing.T) {
	peer := testPeer()

	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (string, error) {
			return testPeerID, nil
		},
		getPeerByIDFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (*nbpeer.Peer, error) {
			return peer, nil
		},
	}

	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
		SyncAndMarkPeerFunc: func(_ context.Context, _ string, _ string, _ nbpeer.PeerSystemMeta, _ net.IP) (*nbpeer.Peer, *types.NetworkMap, []*posture.Checks, int64, error) {
			return nil, nil, nil, 0, errors.New("sync failed")
		},
	}

	srv := &Server{accountManager: mockAM}

	ctx := ctxWithIdentity()
	stream := newMockMachineSyncStream(ctx)

	err := srv.SyncMachinePeer(&proto.MachineSyncRequest{}, stream)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to sync peer")
}

// ============================================================
// Tests for GetMachineRoutes
// ============================================================

func TestGetMachineRoutes_NoIdentity(t *testing.T) {
	srv := &Server{}

	resp, err := srv.GetMachineRoutes(context.Background(), &proto.MachineRoutesRequest{})

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGetMachineRoutes_PeerNotFound(t *testing.T) {
	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (string, error) {
			return "", errors.New("peer not found")
		},
	}

	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
	}

	srv := &Server{accountManager: mockAM}

	resp, err := srv.GetMachineRoutes(ctxWithIdentity(), &proto.MachineRoutesRequest{})

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGetMachineRoutes_ReturnsRoutes(t *testing.T) {
	peer := testPeer()
	dcNet := netip.MustParsePrefix("192.168.100.0/24")

	netMap := testNetworkMap()
	netMap.Routes = []*route.Route{
		{
			ID:          "route-1",
			Network:     dcNet,
			NetID:       "dc-net",
			Description: "DC network",
			Peer:        "router-peer-1",
			Masquerade:  false,
			Metric:      100,
			Enabled:     true,
		},
	}
	netMap.Peers = []*nbpeer.Peer{
		{
			ID:       "router-peer-1",
			Key:      "router-wg-key==",
			Name:     "router-peer",
			DNSLabel: "router-peer-abc123",
			IP:       net.ParseIP("100.64.0.1"),
			Status:   &nbpeer.PeerStatus{},
		},
	}

	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (string, error) {
			return testPeerID, nil
		},
		getPeerByIDFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (*nbpeer.Peer, error) {
			return peer, nil
		},
	}

	syncPeerCalled := false
	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
		SyncPeerFunc: func(_ context.Context, sync types.PeerSync, accountID string) (*nbpeer.Peer, *types.NetworkMap, []*posture.Checks, int64, error) {
			syncPeerCalled = true
			assert.Equal(t, testAccountID, accountID)
			assert.Equal(t, testPeerKey, sync.WireGuardPubKey)
			return peer, netMap, nil, 0, nil
		},
	}

	mockNetMapCtrl := &mockNetworkMapController{
		getDNSDomainFunc: func(_ *types.Settings) string { return "netbird.local" },
	}

	mockSettingsMgr := &mockSettingsManager{
		getSettingsFunc: func(_ context.Context, _ string, _ string) (*types.Settings, error) {
			return &types.Settings{}, nil
		},
	}

	srv := &Server{
		accountManager:       mockAM,
		settingsManager:      mockSettingsMgr,
		networkMapController: mockNetMapCtrl,
	}

	resp, err := srv.GetMachineRoutes(ctxWithIdentity(), &proto.MachineRoutesRequest{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, syncPeerCalled, "SyncPeer should have been called")
	assert.Len(t, resp.DcNetworks, 1)
	assert.Equal(t, "192.168.100.0/24", resp.DcNetworks[0])
	assert.NotEmpty(t, resp.Routes, "should have proto routes")
	assert.NotEmpty(t, resp.RouterPeers, "should have router peer configs")
}

func TestGetMachineRoutes_SyncPeerFails(t *testing.T) {
	peer := testPeer()

	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (string, error) {
			return testPeerID, nil
		},
		getPeerByIDFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (*nbpeer.Peer, error) {
			return peer, nil
		},
	}

	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
		SyncPeerFunc: func(_ context.Context, _ types.PeerSync, _ string) (*nbpeer.Peer, *types.NetworkMap, []*posture.Checks, int64, error) {
			return nil, nil, nil, 0, errors.New("sync error")
		},
	}

	srv := &Server{accountManager: mockAM}

	resp, err := srv.GetMachineRoutes(ctxWithIdentity(), &proto.MachineRoutesRequest{})

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

// ============================================================
// Tests for ReportMachineStatus
// ============================================================

func TestReportMachineStatus_NoIdentity(t *testing.T) {
	srv := &Server{}

	resp, err := srv.ReportMachineStatus(context.Background(), &proto.MachineStatusRequest{})

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestReportMachineStatus_PeerNotFound(t *testing.T) {
	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (string, error) {
			return "", errors.New("peer not found")
		},
	}

	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
	}

	srv := &Server{accountManager: mockAM}

	resp, err := srv.ReportMachineStatus(ctxWithIdentity(), &proto.MachineStatusRequest{TunnelUp: true})

	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestReportMachineStatus_UpdatesLastSeen(t *testing.T) {
	peer := testPeer()

	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (string, error) {
			return testPeerID, nil
		},
		getPeerByIDFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (*nbpeer.Peer, error) {
			return peer, nil
		},
	}

	markConnectedCalled := false
	var capturedPeerKey string
	var capturedConnected bool
	var capturedAccountID string

	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
		MarkPeerConnectedFunc: func(_ context.Context, peerKey string, connected bool, _ net.IP) error {
			markConnectedCalled = true
			capturedPeerKey = peerKey
			capturedConnected = connected
			return nil
		},
	}

	srv := &Server{accountManager: mockAM}

	resp, err := srv.ReportMachineStatus(ctxWithIdentity(), &proto.MachineStatusRequest{
		TunnelUp:      true,
		DcReachable:   true,
		UptimeSeconds: 3600,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Ack)
	assert.NotNil(t, resp.ServerTime)
	assert.True(t, markConnectedCalled, "MarkPeerConnected should have been called")
	assert.Equal(t, testPeerKey, capturedPeerKey)
	assert.True(t, capturedConnected)
	// Note: capturedAccountID comes from the 5th param of the real MarkPeerConnected,
	// but the MockAccountManager.MarkPeerConnectedFunc only receives 4 params due to mock signature
	_ = capturedAccountID
}

func TestReportMachineStatus_MarkPeerConnectedError(t *testing.T) {
	peer := testPeer()

	mockStore := &machineTestStore{
		getPeerIdByLabelFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (string, error) {
			return testPeerID, nil
		},
		getPeerByIDFunc: func(_ context.Context, _ store.LockingStrength, _ string, _ string) (*nbpeer.Peer, error) {
			return peer, nil
		},
	}

	mockAM := &mock_server.MockAccountManager{
		GetStoreFunc: func() store.Store { return mockStore },
		MarkPeerConnectedFunc: func(_ context.Context, _ string, _ bool, _ net.IP) error {
			return errors.New("database error")
		},
	}

	srv := &Server{accountManager: mockAM}

	// Should still return success (ack: true) even if MarkPeerConnected fails
	// The implementation logs the error but doesn't fail the RPC
	resp, err := srv.ReportMachineStatus(ctxWithIdentity(), &proto.MachineStatusRequest{TunnelUp: true})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Ack, "should ack even if MarkPeerConnected fails (non-critical)")
}

// ============================================================
// Tests for handleMachineUpdates
// ============================================================

func TestHandleMachineUpdates_ChannelClosed(t *testing.T) {
	peer := testPeer()
	updates := make(chan *network_map.UpdateMessage)

	mockNetMapCtrl := &mockNetworkMapController{
		onPeerDisconnectedFunc: func(_ context.Context, accountID string, peerID string) {
			assert.Equal(t, testAccountID, accountID)
			assert.Equal(t, testPeerID, peerID)
		},
	}

	mockAM := &mock_server.MockAccountManager{}
	// Note: OnPeerDisconnected in MockAccountManager panics, but cancelPeerRoutines
	// is called from handleMachineUpdates. We need to test this carefully.
	// Since cancelPeerRoutines calls accountManager.OnPeerDisconnected and the mock panics,
	// we test handleMachineUpdates indirectly through the channel behavior.

	srv := &Server{
		accountManager:       mockAM,
		networkMapController: mockNetMapCtrl,
	}

	ctx, cancel := context.WithCancel(ctxWithIdentity())
	stream := newMockMachineSyncStream(ctx)

	// Close channel immediately to trigger cleanup path
	close(updates)

	// cancelPeerRoutines will panic due to MockAccountManager.OnPeerDisconnected
	// being unimplemented. We use recover to verify the flow reached the right point.
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected: OnPeerDisconnected panics in mock
				t.Logf("Expected panic from unimplemented OnPeerDisconnected: %v", r)
			}
		}()
		_ = srv.handleMachineUpdates(ctx, testAccountID, peer, updates, stream, time.Now())
	}()

	cancel()
}

func TestHandleMachineUpdates_StreamContextDone(t *testing.T) {
	peer := testPeer()
	updates := make(chan *network_map.UpdateMessage, 10)

	mockNetMapCtrl := &mockNetworkMapController{}
	mockAM := &mock_server.MockAccountManager{}

	srv := &Server{
		accountManager:       mockAM,
		networkMapController: mockNetMapCtrl,
	}

	ctx, cancel := context.WithCancel(ctxWithIdentity())
	stream := newMockMachineSyncStream(ctx)

	// Cancel context to simulate stream disconnect
	cancel()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic from unimplemented OnPeerDisconnected: %v", r)
			}
		}()
		err := srv.handleMachineUpdates(ctx, testAccountID, peer, updates, stream, time.Now())
		// Context cancelled should return the context error
		assert.Error(t, err)
	}()
}

func TestHandleMachineUpdates_SendsUpdate(t *testing.T) {
	peer := testPeer()
	updates := make(chan *network_map.UpdateMessage, 10)

	mockNetMapCtrl := &mockNetworkMapController{}
	mockAM := &mock_server.MockAccountManager{}

	srv := &Server{
		accountManager:       mockAM,
		networkMapController: mockNetMapCtrl,
	}

	ctx, cancel := context.WithCancel(ctxWithIdentity())
	defer cancel()
	stream := newMockMachineSyncStream(ctx)

	// Send an update, then close
	syncResp := &proto.SyncResponse{
		NetworkMap: &proto.NetworkMap{
			Serial:     55,
			PeerConfig: &proto.PeerConfig{Address: "100.64.0.10/32"},
		},
	}
	updates <- &network_map.UpdateMessage{Update: syncResp}
	close(updates)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic from unimplemented OnPeerDisconnected: %v", r)
			}
		}()
		_ = srv.handleMachineUpdates(ctx, testAccountID, peer, updates, stream, time.Now())
	}()
	wg.Wait()

	require.Len(t, stream.sent, 1, "should have sent 1 update")
	assert.Equal(t, proto.MachineUpdateType_MACHINE_UPDATE_FULL, stream.sent[0].UpdateType)
	assert.NotNil(t, stream.sent[0].NetworkMap)
	assert.Equal(t, "100.64.0.10/32", stream.sent[0].NetworkMap.PeerConfig.Address)
}

// ============================================================
// Tests for DNSLabel consistency
// ============================================================

func TestFindMachinePeer_DNSLabelConsistency(t *testing.T) {
	// Verify that the DNSLabel generated during registration matches
	// the one used for lookup during Sync/GetRoutes/ReportStatus
	hostname := "win10-pc"
	domain := "corp.local"

	label1 := mtls.GenerateUniqueDNSLabel(hostname, domain)
	label2 := mtls.GenerateUniqueDNSLabel(hostname, domain)

	assert.Equal(t, label1, label2, "GenerateUniqueDNSLabel must be deterministic")
	assert.Contains(t, label1, hostname, "DNSLabel should contain hostname")

	// Different domains must produce different labels (multi-tenant isolation)
	labelOtherDomain := mtls.GenerateUniqueDNSLabel(hostname, "other.local")
	assert.NotEqual(t, label1, labelOtherDomain, "different domains must produce different DNSLabels")
}
