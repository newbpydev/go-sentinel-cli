# Go Sentinel API Roadmap

This roadmap outlines the development plan for the Go Sentinel API server which will power our web interface. Following our TDD approach, each implementation task is preceded by its corresponding test task.

## Phase 1: API Foundations & Core Structure

- [x] **1.1. API Project Structure Setup**
  - [x] 1.1.1. Create `api` package directory structure
    ```
    internal/api/            # Main API package
    ├── server/              # HTTP server implementation
    ├── handlers/            # Route handlers
    ├── middleware/          # HTTP middleware components
    ├── websocket/           # WebSocket implementation
    └── models/              # Data models for API
    ```
  - [x] 1.1.2. Set up Go module dependencies for API (`gorilla/websocket`, `chi` router, etc.)
  - [x] 1.1.3. Create configuration struct for API settings

- [ ] **1.2. Base HTTP Server**
  - [x] 1.2.1. Write tests for HTTP server setup and configuration
    - [x] Test: Server initializes with proper routes
    - [x] Test: Server handles graceful shutdown
    - [x] Test: Server applies correct middleware chain
  - [x] 1.2.2. Implement core HTTP server with proper middleware and routing
    - [x] Implement graceful startup/shutdown
    - [x] Configure CORS, logging, and security middleware
    - [x] Set up basic health endpoint for monitoring

## Phase 2: Test Result Data Models

- [ ] **2.1. Event Model Serialization**
  - [x] 2.1.1. Write tests for event serialization/deserialization
    - [x] Test: Convert internal test events to API models
    - [x] Test: Handle all test event types (pass, fail, run, output)
    - [x] Test: Properly serialize nested test structures
  - [x] 2.1.2. Implement API data models matching internal structures
    - [x] Create JSON-serializable test result models
    - [x] Implement conversion from internal models
    - [x] Add proper validation and error handling

- [ ] **2.2. Test Result Aggregation**
  - [x] 2.2.1. Write tests for test result aggregation
    - [x] Test: Aggregate individual test results into summary format
    - [x] Test: Calculate proper statistics (pass/fail counts, durations)
    - [x] Test: Handle edge cases (empty results, errors)
  - [x] 2.2.2. Implement test result aggregation
    - [x] Create aggregator service
    - [x] Implement statistics calculation
    - [x] Add filtering capabilities

## Phase 3: WebSocket Implementation

- [ ] **3.1. WebSocket Connection Manager**
  - [x] 3.1.1. Write tests for WebSocket connection manager
    - [x] Test: Connection creation and tracking
    - [x] Test: Connection cleanup on disconnect
    - [x] Test: Proper goroutine management
  - [x] 3.1.2. Implement WebSocket connection manager
    - [x] Create connection pool with thread-safe access
    - [x] Implement proper cleanup on disconnects
    - [x] Add connection metadata tracking

- [ ] **3.2. WebSocket Message Handling**
  - [x] 3.2.1. Write tests for WebSocket message types
    - [x] Test: Message encoding/decoding
    - [x] Test: Handle different message types (test results, commands)
    - [x] Test: Error handling for malformed messages
  - [x] 3.2.2. Implement WebSocket message handling
    - [x] Create message types and encoders/decoders
    - [x] Implement message routing system
    - [x] Add proper error handling for malformed messages

- [ ] **3.3. Real-Time Test Updates Broadcasting**
  - [x] 3.3.1. Write tests for real-time broadcasting
    - [x] Test: New test results broadcast to all connections
    - [x] Test: Broadcast throttling for high volume updates
    - [x] Test: Message ordering and delivery guarantees
  - [x] 3.3.2. Implement real-time broadcasting system
    - [x] Create broadcast mechanism for new test results
    - [x] Implement throttling for high-frequency events
    - [x] Add ability to send targeted messages to specific clients

## Phase 4: REST API Endpoints

- [ ] **4.1. Test Result History Endpoints**
  - [x] 4.1.1. Write tests for test history endpoints
    - [x] Test: Retrieve recent test runs
    - [x] Test: Pagination and filtering
    - [x] Test: Proper error responses
  - [x] 4.1.2. Implement test history endpoints
    - [x] Create endpoint for retrieving test history
    - [x] Add pagination and filtering support
    - [x] Implement proper error handling

- [ ] **4.2. Test Control Endpoints**
  - [x] 4.2.1. Write tests for test control endpoints
    - [x] Test: Trigger new test runs
    - [x] Test: Filter tests to run
    - [x] Test: Cancel running tests
  - [x] 4.2.2. Implement test control endpoints
    - [x] Create endpoints for triggering test runs
    - [x] Add test filtering capabilities
    - [x] Implement test run cancellation

- [ ] **4.3. Configuration Endpoints**
  - [x] 4.3.1. Write tests for configuration endpoints
    - [x] Test: Retrieve current configuration
    - [x] Test: Update configuration
    - [x] Test: Validate configuration changes
  - [x] 4.3.2. Implement configuration endpoints
    - [x] Create endpoints for config retrieval/updates
    - [x] Add validation for configuration changes
    - [x] Implement configuration persistence

## Phase 5: Integration With Core Engine

- [ ] **5.1. Core Engine Integration Tests**
  - [x] 5.1.1. Write tests for integration with core Go Sentinel
    - [x] Test: API receives updates from core test runner
    - [x] Test: Commands from API propagate to core engine
    - [x] Test: Proper error handling between components
  - [x] 5.1.2. Implement core engine integration
    - [x] Create adapter between core engine and API
    - [x] Set up event channels between components
    - [x] Implement error handling and recovery mechanisms

- [ ] **5.2. In-Memory Cache for Recent Results**
  - [x] 5.2.1. Write tests for result caching
    - [x] Test: Cache recent test results
    - [x] Test: Cache eviction policies
    - [x] Test: Thread-safety of cache access
  - [x] 5.2.2. Implement result caching system
    - [x] Create thread-safe cache for recent results
    - [x] Implement eviction policies
    - [x] Add performance metrics for cache operations

## Phase 6: API Security & Performance

- [ ] **6.1. Security Measures**
  - [ ] 6.1.1. Write tests for API security features
    - [ ] Test: Authentication mechanisms
    - [ ] Test: Rate limiting
    - [ ] Test: Input validation and sanitization
  - [ ] 6.1.2. Implement security features
    - [ ] Add basic authentication (if needed)
    - [ ] Implement rate limiting for all endpoints
    - [ ] Add comprehensive input validation

- [ ] **6.2. Performance Optimizations**
  - [ ] 6.2.1. Write tests for API performance
    - [ ] Test: Response time under load
    - [ ] Test: Memory usage during high activity
    - [ ] Test: Connection handling under stress
  - [ ] 6.2.2. Implement performance optimizations
    - [ ] Add response caching where appropriate
    - [ ] Optimize WebSocket message handling
    - [ ] Implement resource usage monitoring

## Phase 7: API Documentation & Developer Experience

- [ ] **7.1. OpenAPI/Swagger Documentation**
  - [ ] 7.1.1. Write tests for API documentation
    - [ ] Test: Documentation generation
    - [ ] Test: Documentation accuracy
  - [ ] 7.1.2. Implement API documentation
    - [ ] Create OpenAPI/Swagger specifications
    - [ ] Add documentation comments to all handlers
    - [ ] Generate interactive API documentation

- [ ] **7.2. Developer Tools**
  - [ ] 7.2.1. Write tests for developer tools
    - [ ] Test: Development mode features
    - [ ] Test: Debugging helpers
  - [ ] 7.2.2. Implement developer tools
    - [ ] Add development mode with detailed logging
    - [ ] Create debugging endpoints and tools
    - [ ] Implement request/response inspection capabilities
