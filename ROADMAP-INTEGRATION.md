Go Sentinel Backend-Frontend Integration Roadmap
[ ] Phase 1: Assessment & Setup
    [x] 1.1. Project Structure Analysis
        [x] Review existing Go server implementation in internal/web/server/server.go
            - Server uses chi, robust template rendering, static and API routes, minimal changes needed to rendering system.
        [x] Document current template rendering system (layouts → partials → pages)
            - Three-tier hierarchy, explicit block definitions, strict load order.
        [x] Analyze WebSocket implementation in frontend
            - HTMX WebSocket extension, reconnection, real-time updates, selection mode, connection status indicator.
        [x] Map API endpoints to corresponding frontend components
            - Mapped all major API endpoints to frontend dashboard, test list, history, coverage, and settings.
    [ ] 1.2. Development Environment Setup
        [x] Configure Air for hot reloading
            - Air is already configured and working; no changes made.
        [x] Create .air.toml configuration
            - .air.toml is present and correctly set up for the project.
        [x] Test auto-restart functionality
            - Verified Air triggers rebuilds and restarts on code/template changes.
        [x] Set up consistent testing environment for backend-frontend integration
            - Added integration_test.go for HTTP server and dashboard route.
        [x] Create test fixtures for WebSocket communication tests
            - Added websocket_integration_test.go for WebSocket connection checks.

[ ] Phase 2: Core WebSocket Integration
    [x] 2.1. WebSocket Handler Enhancement
        [x] Write tests for WebSocket message encoding/decoding
        [x] Test: Verify test event serialization to WebSocket messages
        [x] Test: Handle different message types (test results, status updates)
        [x] Test: Test reconnection logic
        [x] Implement WebSocket message serialization
        [x] Create standardized message format for test events
        [x] Implement JSON marshaling/unmarshaling for WebSocket messages
        [x] Add proper error handling for malformed messages
    [ ] 2.2. Real-Time Test Result Broadcasting
        [x] Write tests for test result broadcasting
        [x] Test: Verify new test results are broadcast to all connections
        [x] Test: Test proper client message routing
        [x] Test: Verify message delivery during connection hiccups
        [x] Implement test result broadcasting
        [x] Create broadcaster service in Go backend
        [x] Implement message routing by type
        [ ] Add connection tracking and cleanup
    [ ] 2.3. Frontend Message Handling
        [ ] Write tests for frontend message processing
        [ ] Test: Verify HTMX properly processes incoming WebSocket messages
        [ ] Test: Test UI updates based on message types
        [ ] Test: Verify error handling for unexpected messages
        [ ] Implement frontend message handlers
        [ ] Create HTMX swap targets for different message types
        [ ] Add animation and styling for real-time updates
        [ ] Implement client-side error handling

[ ] Phase 3: API Integration
    [ ] 3.1. Test Results API
        [ ] Write tests for test results API
        [ ] Test: Verify test results endpoint returns correct data
        [ ] Test: Test filtering and pagination
        [ ] Test: Verify error handling
        [ ] Implement test results API
        [ ] Create endpoint for retrieving test results
        [ ] Add filtering and pagination
        [ ] Implement error handling and response formatting
    [ ] 3.2. Test Control API
        [ ] Write tests for test control API
        [ ] Test: Verify test run triggers
        [ ] Test: Test filtering for specific tests
        [ ] Test: Verify test cancellation
        [ ] Implement test control API
        [ ] Create endpoints for triggering test runs
        [ ] Add test filtering capabilities
        [ ] Implement test run cancellation
    [ ] 3.3. Settings & Configuration API
        [ ] Write tests for settings API
        [ ] Test: Verify settings retrieval
        [ ] Test: Test settings updates
        [ ] Test: Verify validation
        [ ] Implement settings API
        [ ] Create endpoints for retrieving/updating settings
        [ ] Add validation for settings changes
        [ ] Implement settings persistence

[ ] Phase 4: Frontend Component Integration
    [ ] 4.1. Dashboard Components
        [ ] Write tests for dashboard components
        [ ] Test: Verify test statistics display
        [ ] Test: Test real-time updates via WebSocket
        [ ] Test: Verify responsive layout
        [ ] Implement dashboard components
        [ ] Create statistics tiles with real data
        [ ] Add WebSocket bindings for real-time updates
        [ ] Implement responsive layout
    [ ] 4.2. Test Results List
        [ ] Write tests for test results list
        [ ] Test: Verify test results rendering
        [ ] Test: Test sorting and filtering
        [ ] Test: Verify real-time updates
        [ ] Implement test results list
        [ ] Create expandable test result rows
        [ ] Add sorting and filtering
        [ ] Implement WebSocket updates
    [ ] 4.3. Test Detail View
        [ ] Write tests for test detail view
        [ ] Test: Verify test details display
        [ ] Test: Test error output formatting
        [ ] Test: Verify source context display
        [ ] Implement test detail view
        [ ] Create detailed view for test output
        [ ] Add source code context
        [ ] Implement expandable sections

[ ] Phase 5: User Interaction Features
    [ ] 5.1. Test Selection & Management
        [ ] Write tests for test selection
        [ ] Test: Verify selection state management
        [ ] Test: Test keyboard shortcuts
        [ ] Test: Verify clipboard operations
        [ ] Implement test selection
        [ ] Create selection UI
        [ ] Add keyboard shortcuts
        [ ] Implement clipboard operations
    [ ] 5.2. Test Run Controls
        [ ] Write tests for test run controls
        [ ] Test: Verify run all tests button
        [ ] Test: Test run selected tests
        [ ] Test: Verify cancel running tests
        [ ] Implement test run controls
        [ ] Create run controls UI
        [ ] Add WebSocket command sending
        [ ] Implement loading states and feedback
    [ ] 5.3. Notifications System
        [ ] Write tests for notifications
        [ ] Test: Verify notification display
        [ ] Test: Test notification dismissal
        [ ] Test: Verify different notification types
        [ ] Implement notifications system
        [ ] Create toast notification component
        [ ] Add notification triggers
        [ ] Implement notification preferences

[ ] Phase 6: Core Engine Integration
    [ ] 6.1. Test Runner Integration
        [ ] Write tests for test runner integration
        [ ] Test: Verify backend can trigger test runs
        [ ] Test: Test result capture and processing
        [ ] Test: Verify error handling
        [ ] Implement test runner integration
        [ ] Create adapter between test runner and web server
        [ ] Add event channels for test results
        [ ] Implement error handling
    [ ] 6.2. Real-time Status Updates
        [ ] Write tests for status updates
        [ ] Test: Verify status updates are sent via WebSocket
        [ ] Test: Test UI updates based on status
        [ ] Test: Verify error recovery
        [ ] Implement status updates
        [ ] Create status update broadcaster
        [ ] Add UI components for status display
        [ ] Implement error recovery

[ ] Phase 7: Performance & Polish
    [ ] 7.1. Performance Optimization
        [ ] Write tests for performance
        [ ] Test: Measure response times under load
        [ ] Test: Test WebSocket message throughput
        [ ] Test: Verify memory usage
        [ ] Implement performance optimizations
        [ ] Add response caching where appropriate
        [ ] Optimize WebSocket message handling
        [ ] Implement resource monitoring
    [ ] 7.2. UI Refinement
        [ ] Write tests for UI components
        [ ] Test: Verify responsive behavior
        [ ] Test: Test accessibility
        [ ] Test: Verify cross-browser compatibility
        [ ] Implement UI refinements
        [ ] Add loading states and animations
        [ ] Improve accessibility
        [ ] Fix cross-browser issues
    [ ] 7.3. Documentation & Deployment
        [ ] Update project documentation
        [ ] Document WebSocket message format
        [ ] Create API documentation
        [ ] Update development guides
        [ ] Prepare for deployment
        [ ] Create Docker configuration
        [ ] Add environment configuration
        [ ] Test deployment process
