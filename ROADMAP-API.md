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
  - [ ] 3.2.1. Write tests for WebSocket message types
    - [ ] Test: Message encoding/decoding
    - [ ] Test: Handle different message types (test results, commands)
    - [ ] Test: Error handling for malformed messages
  - [ ] 3.2.2. Implement WebSocket message handling
    - [ ] Create message types and encoders/decoders
    - [ ] Implement message routing system
    - [ ] Add proper error handling for malformed messages

- [ ] **3.3. Real-Time Test Updates Broadcasting**
  - [ ] 3.3.1. Write tests for real-time broadcasting
    - [ ] Test: New test results broadcast to all connections
    - [ ] Test: Broadcast throttling for high volume updates
    - [ ] Test: Message ordering and delivery guarantees
  - [ ] 3.3.2. Implement real-time broadcasting system
    - [ ] Create broadcast mechanism for new test results
    - [ ] Implement throttling for high-frequency events
    - [ ] Add ability to send targeted messages to specific clients

## Phase 4: REST API Endpoints

- [ ] **4.1. Test Result History Endpoints**
  - [ ] 4.1.1. Write tests for test history endpoints
    - [ ] Test: Retrieve recent test runs
    - [ ] Test: Pagination and filtering
    - [ ] Test: Proper error responses
  - [ ] 4.1.2. Implement test history endpoints
    - [ ] Create endpoint for retrieving test history
    - [ ] Add pagination and filtering support
    - [ ] Implement proper error handling

- [ ] **4.2. Test Control Endpoints**
  - [ ] 4.2.1. Write tests for test control endpoints
    - [ ] Test: Trigger new test runs
    - [ ] Test: Filter tests to run
    - [ ] Test: Cancel running tests
  - [ ] 4.2.2. Implement test control endpoints
    - [ ] Create endpoints for triggering test runs
    - [ ] Add test filtering capabilities
    - [ ] Implement test run cancellation

- [ ] **4.3. Configuration Endpoints**
  - [ ] 4.3.1. Write tests for configuration endpoints
    - [ ] Test: Retrieve current configuration
    - [ ] Test: Update configuration
    - [ ] Test: Validate configuration changes
  - [ ] 4.3.2. Implement configuration endpoints
    - [ ] Create endpoints for config retrieval/updates
    - [ ] Add validation for configuration changes
    - [ ] Implement configuration persistence

## Phase 5: Integration With Core Engine

- [ ] **5.1. Core Engine Integration Tests**
  - [ ] 5.1.1. Write tests for integration with core Go Sentinel
    - [ ] Test: API receives updates from core test runner
    - [ ] Test: Commands from API propagate to core engine
    - [ ] Test: Proper error handling between components
  - [ ] 5.1.2. Implement core engine integration
    - [ ] Create adapter between core engine and API
    - [ ] Set up event channels between components
    - [ ] Implement error handling and recovery mechanisms

- [ ] **5.2. In-Memory Cache for Recent Results**
  - [ ] 5.2.1. Write tests for result caching
    - [ ] Test: Cache recent test results
    - [ ] Test: Cache eviction policies
    - [ ] Test: Thread-safety of cache access
  - [ ] 5.2.2. Implement result caching system
    - [ ] Create thread-safe cache for recent results
    - [ ] Implement eviction policies
    - [ ] Add performance metrics for cache operations

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
