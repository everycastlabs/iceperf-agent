# Pion WebRTC v5 Migration Plan for ICEPerf

## Executive Summary

ICEPerf currently uses Pion WebRTC v4 (beta.29) and needs to migrate to v5 to leverage improvements in:
- **ICE candidate handling** (especially TCP ICE)
- **Data channel throughput** (SCTP optimizations)
- **Overall performance** for STUN/TURN testing

This document outlines a structured migration plan without making code changes yet.

## Current State Analysis

### Current Dependencies
- **Pion WebRTC**: `v4.0.0-beta.29`
- **Go Version**: 1.21.6
- **Key Pion modules used**:
  - `github.com/pion/webrtc/v4`
  - `github.com/pion/stun/v2`
  - Related: `pion/ice`, `pion/sctp`, `pion/datachannel`

### Critical Code Areas Using WebRTC

1. **ICE Candidate Handling** (`client/client.go`)
   - `OnICECandidate` handlers for both offerer and answerer
   - Candidate type filtering (`ICECandidateTypeSrflx`, `ICECandidateTypeRelay`, `ICECandidateTypeHost`)
   - Candidate exchange via `AddICECandidate(i.ToJSON())`
   - Candidate timing metrics

2. **Data Channel Throughput** (`client/dataChannel.go`)
   - Data channel creation with custom options (`Ordered`, `MaxRetransmits`)
   - Throughput testing loop using `dc.Send(buf)`
   - `BufferedAmountLowThreshold` management (4-8 MiB for throughput tests)
   - `OnBufferedAmountLow` callback for flow control
   - Statistics collection via `GetStats()` and `GetDataChannelStats()`

3. **PeerConnection Setup** (`client/dataChannel.go`)
   - `SettingEngine` configuration with ICE timeouts
   - `NewAPI()` with custom settings
   - `NewPeerConnection()` for both offerer and answerer
   - ICE transport policy configuration (`ICETransportPolicyRelay`, `ICETransportPolicyAll`)

4. **Configuration** (`config/config.go`, `client/ice-servers.go`)
   - `webrtc.Configuration` struct usage
   - `ICEServer` struct population across multiple adapters
   - `ICETransportPolicy` configuration

5. **Statistics** (`client/dataChannel.go`)
   - `GetStats()` API usage
   - `TransportStats` type assertions
   - `DataChannelStats` extraction

## Migration Strategy

### Phase 1: Dependency Updates

#### 1.1 Update go.mod
```bash
go get github.com/pion/webrtc/v5@latest
go mod tidy
```

#### 1.2 Update All Import Statements
Files requiring import updates:
- `client/client.go`
- `client/dataChannel.go`
- `client/ice-servers.go`
- `config/config.go`
- `cmd/iceperf/main.go`
- All adapter files in `adapters/`:
  - `adapters/api/driver.go`
  - `adapters/cloudflare/driver.go`
  - `adapters/elixir/driver.go`
  - `adapters/expressturn/driver.go`
  - `adapters/google/driver.go`
  - `adapters/metered/driver.go`
  - `adapters/serverconnect/driver.go`
  - `adapters/stunner/driver.go`
  - `adapters/twilio/driver.go`
  - `adapters/webrtcpeerconnect/driver.go`
  - `adapters/xirsys/driver.go`
- `acceptance-tests/server_connect_test.go`

**Change**: `github.com/pion/webrtc/v4` → `github.com/pion/webrtc/v5`

#### 1.3 Verify Compatible Sub-modules
Check and update if needed:
- `github.com/pion/ice/v4` (or v5 if available)
- `github.com/pion/sctp` (latest compatible version)
- `github.com/pion/datachannel` (latest compatible version)
- `github.com/pion/stun/v2` (verify compatibility)

### Phase 2: API Compatibility Review

#### 2.1 ICE Candidate API Changes
**Location**: `client/client.go` lines 91-121

**Potential Changes to Verify**:
- `ICECandidate` struct fields (may have new fields or changed types)
- `ICECandidate.ToJSON()` method signature
- `AddICECandidate()` method signature
- `OnICECandidate` callback signature
- Candidate type constants (`ICECandidateTypeSrflx`, etc.)

**Action Items**:
1. Check if `i.ToJSON()` still returns the same format
2. Verify candidate type enum values haven't changed
3. Test candidate exchange between offerer/answerer still works
4. Verify timing metrics collection still functions

#### 2.2 Data Channel API Changes
**Location**: `client/dataChannel.go` lines 108-176

**Potential Changes to Verify**:
- `CreateDataChannel()` method signature
- `DataChannelInit` struct fields
- `DataChannel.Send()` method signature
- `BufferedAmount()` method
- `SetBufferedAmountLowThreshold()` method
- `OnBufferedAmountLow()` callback signature
- `OnOpen()`, `OnClose()`, `OnMessage()` callbacks
- `DataChannelMessage` type

**Action Items**:
1. Verify data channel creation options still work
2. Test throughput loop with new SCTP optimizations
3. Verify buffered amount threshold behavior
4. Check if message sending API changed

#### 2.3 PeerConnection API Changes
**Location**: `client/dataChannel.go` lines 93-199

**Potential Changes to Verify**:
- `NewPeerConnection()` constructor
- `NewAPI()` constructor
- `SettingEngine` struct and methods
- `SetICETimeouts()` method signature
- `CreateOffer()`, `CreateAnswer()` methods
- `SetLocalDescription()`, `SetRemoteDescription()` methods
- `OnConnectionStateChange()` callback
- `ConnectionState()` method
- `Close()` method

**Action Items**:
1. Verify `SettingEngine` configuration still works
2. Test ICE timeout settings
3. Verify connection state callbacks
4. Check peer connection lifecycle management

#### 2.4 Statistics API Changes
**Location**: `client/dataChannel.go` lines 292-313

**Potential Changes to Verify**:
- `GetStats()` return type and structure
- `GetDataChannelStats()` method signature
- `TransportStats` struct fields
- `DataChannelStats` struct fields
- Stats map key names (`"iceTransport"`)

**Action Items**:
1. Verify stats collection still works
2. Check if stats struct field names changed
3. Test type assertions for stats extraction
4. Verify bytes sent/received metrics accuracy

#### 2.5 Configuration API Changes
**Location**: `config/config.go`, `client/ice-servers.go`

**Potential Changes to Verify**:
- `webrtc.Configuration` struct fields
- `ICEServer` struct fields
- `ICETransportPolicy` enum values
- `SDPSemantics` enum values
- URL parsing and validation

**Action Items**:
1. Verify ICE server configuration still works
2. Test TURN server URL parsing
3. Check credential handling
4. Verify transport policy settings

### Phase 3: Leverage v5 Improvements

#### 3.1 TCP ICE Enhancements
**Current State**: ICEPerf tests both UDP and TCP TURN servers

**v5 Improvements to Leverage**:
- Better TCP candidate gathering
- Improved TCP ICE connection establishment
- Enhanced TCP muxing support

**Action Items**:
1. Review `SettingEngine` for new TCP-specific options
2. Test TCP TURN throughput improvements
3. Verify TCP candidate gathering timing
4. Check if explicit TCP configuration needed

**Code to Review**:
- `client/dataChannel.go:95-97` (SettingEngine setup)
- `client/client.go:91-121` (ICE candidate handling)

#### 3.2 Data Channel Throughput Optimizations
**Current State**: Uses buffered amount thresholds (4-8 MiB) for throughput testing

**v5 Improvements to Leverage**:
- SCTP stack optimizations
- Better memory management
- Improved flow control

**Action Items**:
1. Test if higher throughput achievable with same buffer sizes
2. Review if buffer thresholds need adjustment
3. Verify `OnBufferedAmountLow` callback performance
4. Test sustained high-speed transfers

**Code to Review**:
- `client/dataChannel.go:143-157` (throughput sending loop)
- `client/dataChannel.go:162-176` (buffered amount handling)
- `client/dataChannel.go:29-30` (buffer threshold constants)

#### 3.3 ICE Candidate Gathering Improvements
**Current State**: Filters candidates by type and exchanges them immediately

**v5 Improvements to Leverage**:
- Faster candidate gathering
- Better candidate prioritization
- Improved trickle ICE support

**Action Items**:
1. Verify candidate gathering timing improvements
2. Test if candidate filtering logic needs updates
3. Check candidate exchange performance
4. Review timing metrics accuracy

**Code to Review**:
- `client/client.go:91-121` (candidate handlers)
- `client/client.go:94,110` (candidate type filtering)

### Phase 4: Testing Strategy

#### 4.1 Unit Testing
1. **Compile Check**: Ensure all code compiles with v5
2. **Import Verification**: Verify all imports resolve correctly
3. **Type Checking**: Ensure no type mismatches

#### 4.2 Functional Testing
1. **STUN Tests**: Verify STUN server connectivity still works
2. **TURN Tests**: Verify TURN server connectivity still works
3. **Candidate Exchange**: Test offerer/answerer candidate exchange
4. **Connection Establishment**: Verify peer connections establish correctly
5. **Data Channel**: Test data channel creation and messaging
6. **Throughput**: Verify throughput testing still functions

#### 4.3 Performance Testing
1. **Baseline Comparison**: Run tests with v4, then v5
2. **Metrics to Compare**:
   - Time to receive first candidate
   - Time to connected state
   - Data channel throughput (Mbps)
   - First packet latency
   - CPU usage during throughput tests
   - Memory usage during throughput tests

#### 4.4 Regression Testing
1. **All Adapters**: Test each ICE server adapter
2. **All Protocols**: Test UDP, TCP, TLS, DTLS
3. **All Schemes**: Test STUN, STUNS, TURN, TURNS
4. **Edge Cases**: Test with various network conditions

### Phase 5: Migration Steps

#### Step 1: Preparation
1. Create a new git branch: `migration/pion-v5`
2. Document current performance baseline
3. Run full test suite with v4 to establish baseline

#### Step 2: Dependency Update
1. Update `go.mod` to v5
2. Run `go mod tidy`
3. Fix any immediate dependency conflicts

#### Step 3: Import Updates
1. Update all import statements (use find/replace)
2. Compile and fix import errors
3. Verify no missing imports

#### Step 4: API Compatibility
1. Review each API usage against v5 documentation
2. Update any deprecated methods
3. Fix any breaking API changes
4. Compile and fix errors iteratively

#### Step 5: Functional Verification
1. Run unit tests
2. Run acceptance tests
3. Manual testing of core flows
4. Fix any functional issues

#### Step 6: Performance Validation
1. Run performance benchmarks
2. Compare against v4 baseline
3. Verify improvements in:
   - ICE candidate gathering speed
   - Data channel throughput
   - Connection establishment time
4. Document performance changes

#### Step 7: Code Review
1. Review all changes
2. Verify no regressions
3. Ensure code quality maintained
4. Update documentation if needed

#### Step 8: Final Testing
1. Full test suite execution
2. Extended testing with various providers
3. Stress testing with high throughput
4. Network condition testing

## Risk Assessment

### High Risk Areas
1. **Statistics API**: Type assertions may break if stats structure changed
2. **ICE Candidate Handling**: Core functionality, any breakage affects all tests
3. **Data Channel Throughput**: Critical for TURN testing, must maintain accuracy

### Medium Risk Areas
1. **Configuration**: May have subtle changes in validation
2. **Connection State**: Callback signatures may have changed
3. **Error Handling**: Error types may have changed

### Low Risk Areas
1. **Import Paths**: Simple find/replace
2. **Basic API Calls**: Most should remain compatible
3. **Adapter Code**: Mostly configuration, less API-dependent

## Rollback Plan

If migration encounters critical issues:

1. **Git Revert**: Revert to previous commit
2. **Dependency Rollback**: `go get github.com/pion/webrtc/v4@v4.0.0-beta.29`
3. **Restore Imports**: Revert import changes
4. **Verify**: Ensure v4 still works correctly

## Success Criteria

Migration is successful when:

1. ✅ All code compiles without errors
2. ✅ All tests pass
3. ✅ Core functionality works (STUN/TURN connectivity)
4. ✅ Data channel throughput testing works
5. ✅ Performance metrics are accurate
6. ✅ No regressions in existing functionality
7. ✅ Performance improvements are measurable (if applicable)

## Documentation Updates Needed

After successful migration:

1. Update README.md with v5 requirement
2. Update any setup/installation docs
3. Document any new features or capabilities
4. Update example configurations if needed

## Timeline Estimate

- **Phase 1 (Dependencies)**: 1-2 hours
- **Phase 2 (API Review)**: 2-4 hours
- **Phase 3 (Improvements)**: 2-3 hours
- **Phase 4 (Testing)**: 4-6 hours
- **Phase 5 (Migration)**: 4-8 hours
- **Total**: 13-23 hours

## Notes

- This is a major version upgrade, expect some breaking changes
- Focus on maintaining functionality first, optimizations second
- Test thoroughly before merging
- Keep v4 branch available for comparison
- Monitor Pion v5 release notes and changelog for specific breaking changes
- Consider reaching out to Pion community if encountering issues

## References

- [Pion WebRTC v5 Documentation](https://pkg.go.dev/github.com/pion/webrtc/v5)
- [Pion WebRTC GitHub Repository](https://github.com/pion/webrtc)
- [Pion WebRTC v4 Documentation](https://pkg.go.dev/github.com/pion/webrtc/v4) (for comparison)
- [Pion ICE Documentation](https://pkg.go.dev/github.com/pion/ice/v4)
- [Pion DataChannel Documentation](https://pkg.go.dev/github.com/pion/datachannel)
