package app // import "github.com/newbpydev/go-sentinel/internal/app"

Package app provides argument parsing implementation

# Package app provides configuration loading implementation

# Package app provides dependency injection container implementation

# Package app provides the main application controller implementation

# Package app provides display rendering bridging to the modular UI system

# Package app provides application event handling implementation

# Package app provides application orchestration and lifecycle management

# Package app provides application lifecycle management

Package app provides test execution bridging to the modular test system

TYPES

type ApplicationController interface {
	// Run executes the main application flow with the given arguments
	Run(args []string) error

	// Initialize sets up the application with dependencies
	Initialize() error

	// Shutdown gracefully shuts down the application
	Shutdown(ctx context.Context) error
}
    ApplicationController orchestrates the main application flow

func NewController(
	argParser ArgumentParser,
	configLoader ConfigurationLoader,
	lifecycle LifecycleManager,
	container DependencyContainer,
	eventHandler ApplicationEventHandler,
) ApplicationController
    NewController creates a new application controller

type ApplicationEventHandler interface {
	// OnStartup is called when the application starts
	OnStartup(ctx context.Context) error

	// OnShutdown is called when the application shuts down
	OnShutdown(ctx context.Context) error

	// OnError is called when an error occurs
	OnError(err error)

	// OnConfigChanged is called when configuration changes
	OnConfigChanged(config *Configuration)
}
    ApplicationEventHandler handles application-level events

func NewApplicationEventHandler() ApplicationEventHandler
    NewApplicationEventHandler creates a new application event handler

type ArgumentParser interface {
	// Parse parses command-line arguments into a structured format
	Parse(args []string) (*Arguments, error)

	// Help returns help text for the application
	Help() string

	// Version returns version information
	Version() string
}
    ArgumentParser handles command-line argument parsing

func NewArgumentParser() ArgumentParser
    NewArgumentParser creates a new argument parser

type Arguments struct {
	// Packages to test
	Packages []string

	// Watch mode enabled
	Watch bool

	// Verbose output
	Verbose bool

	// Colors enabled
	Colors bool

	// Optimization enabled
	Optimized bool

	// Optimization mode
	OptimizationMode string

	// Output writer
	Writer io.Writer
}
    Arguments represents parsed command-line arguments

type Cleaner interface {
	Cleanup() error
}
    Cleaner interface for components that need cleanup

type ComponentFactory func() (interface{}, error)
    ComponentFactory is a function that creates a component instance

type Configuration struct {
	// Watch configuration
	Watch WatchConfig

	// Paths configuration
	Paths PathsConfig

	// Visual configuration
	Visual VisualConfig

	// Test configuration
	Test TestConfig

	// Colors enabled
	Colors bool

	// Verbosity level
	Verbosity int
}
    Configuration represents application configuration

type ConfigurationLoader interface {
	// LoadFromFile loads configuration from a file
	LoadFromFile(path string) (*Configuration, error)

	// LoadFromDefaults returns default configuration
	LoadFromDefaults() *Configuration

	// Merge merges CLI arguments with configuration
	Merge(config *Configuration, args *Arguments) *Configuration

	// Validate validates the final configuration
	Validate(config *Configuration) error
}
    ConfigurationLoader handles application configuration loading

func NewConfigurationLoader() ConfigurationLoader
    NewConfigurationLoader creates a new configuration loader

type Controller struct {
	// Has unexported fields.
}
    Controller implements the ApplicationController interface

func (c *Controller) Initialize() error
    Initialize implements the ApplicationController interface

func (c *Controller) Run(args []string) error
    Run implements the ApplicationController interface

func (c *Controller) Shutdown(ctx context.Context) error
    Shutdown implements the ApplicationController interface

type DefaultApplicationEventHandler struct {
	// Has unexported fields.
}
    DefaultApplicationEventHandler implements the ApplicationEventHandler
    interface

func (h *DefaultApplicationEventHandler) GetLogger() *log.Logger
    GetLogger returns the current logger

func (h *DefaultApplicationEventHandler) LogDebug(format string, args ...interface{})
    LogDebug logs a debug message if verbosity is high enough

func (h *DefaultApplicationEventHandler) LogError(format string, args ...interface{})
    LogError logs an error message

func (h *DefaultApplicationEventHandler) LogInfo(format string, args ...interface{})
    LogInfo logs an info message

func (h *DefaultApplicationEventHandler) LogWarning(format string, args ...interface{})
    LogWarning logs a warning message

func (h *DefaultApplicationEventHandler) OnConfigChanged(config *Configuration)
    OnConfigChanged is called when configuration changes

func (h *DefaultApplicationEventHandler) OnError(err error)
    OnError is called when an error occurs

func (h *DefaultApplicationEventHandler) OnShutdown(ctx context.Context) error
    OnShutdown is called when the application shuts down

func (h *DefaultApplicationEventHandler) OnStartup(ctx context.Context) error
    OnStartup is called when the application starts

func (h *DefaultApplicationEventHandler) OnTestComplete(testName string, success bool)
    OnTestComplete is called when a test completes (optional extension)

func (h *DefaultApplicationEventHandler) OnTestStart(testName string)
    OnTestStart is called when a test starts (optional extension)

func (h *DefaultApplicationEventHandler) OnWatchEvent(filePath string, eventType string)
    OnWatchEvent is called when a file watch event occurs (optional extension)

func (h *DefaultApplicationEventHandler) SetLogger(logger *log.Logger)
    SetLogger sets a custom logger

func (h *DefaultApplicationEventHandler) SetVerbosity(level int)
    SetVerbosity sets the verbosity level for logging

type DefaultArgumentParser struct {
	// Has unexported fields.
}
    DefaultArgumentParser implements the ArgumentParser interface

func (p *DefaultArgumentParser) Help() string
    Help returns help text for the application

func (p *DefaultArgumentParser) Parse(args []string) (*Arguments, error)
    Parse parses command-line arguments into a structured format

func (p *DefaultArgumentParser) Version() string
    Version returns version information

type DefaultConfigurationLoader struct {
	// Has unexported fields.
}
    DefaultConfigurationLoader implements the ConfigurationLoader interface

func (l *DefaultConfigurationLoader) LoadFromDefaults() *Configuration
    LoadFromDefaults returns default configuration

func (l *DefaultConfigurationLoader) LoadFromFile(path string) (*Configuration, error)
    LoadFromFile loads configuration from a file

func (l *DefaultConfigurationLoader) Merge(config *Configuration, args *Arguments) *Configuration
    Merge merges CLI arguments with configuration

func (l *DefaultConfigurationLoader) Validate(config *Configuration) error
    Validate validates the final configuration

type DefaultContainer struct {
	// Has unexported fields.
}
    DefaultContainer implements the DependencyContainer interface

func (c *DefaultContainer) Cleanup() error
    Cleanup implements the DependencyContainer interface

func (c *DefaultContainer) HasComponent(name string) bool
    HasComponent checks if a component is registered

func (c *DefaultContainer) Initialize() error
    Initialize implements the DependencyContainer interface

func (c *DefaultContainer) ListComponents() []string
    ListComponents returns a list of all registered component names

func (c *DefaultContainer) Register(name string, component interface{}) error
    Register implements the DependencyContainer interface

func (c *DefaultContainer) RegisterSingleton(name string, factory ComponentFactory) error
    RegisterSingleton registers a component as a singleton

func (c *DefaultContainer) Resolve(name string) (interface{}, error)
    Resolve implements the DependencyContainer interface

func (c *DefaultContainer) ResolveAs(name string, target interface{}) error
    ResolveAs implements the DependencyContainer interface

type DefaultDisplayRenderer struct {
	// Has unexported fields.
}
    DefaultDisplayRenderer implements the DisplayRenderer interface using
    modular UI components

func (r *DefaultDisplayRenderer) GetWriter() io.Writer
    GetWriter returns the current output writer

func (r *DefaultDisplayRenderer) RenderFailedTests(ctx context.Context, failedTests []*models.TestResult) error
    RenderFailedTests renders failed test results with detailed error
    information

func (r *DefaultDisplayRenderer) RenderIncrementalResults(ctx context.Context, results interface{}) error
    RenderIncrementalResults renders results incrementally for watch mode

func (r *DefaultDisplayRenderer) RenderResults(ctx context.Context) error
    RenderResults renders the test results using the modular UI components

func (r *DefaultDisplayRenderer) RenderTestResults(ctx context.Context, results []*models.TestResult) error
    RenderTestResults renders individual test results

func (r *DefaultDisplayRenderer) SetConfiguration(config *Configuration) error
    SetConfiguration configures the display renderer with the application
    configuration

func (r *DefaultDisplayRenderer) SetWriter(writer io.Writer)
    SetWriter sets the output writer

type DefaultLifecycleManager struct {
	// Has unexported fields.
}
    DefaultLifecycleManager implements the LifecycleManager interface

func (lm *DefaultLifecycleManager) Context() context.Context
    Context returns the lifecycle context

func (lm *DefaultLifecycleManager) IsRunning() bool
    IsRunning implements the LifecycleManager interface

func (lm *DefaultLifecycleManager) RegisterShutdownHook(hook func() error)
    RegisterShutdownHook implements the LifecycleManager interface

func (lm *DefaultLifecycleManager) Shutdown(ctx context.Context) error
    Shutdown implements the LifecycleManager interface

func (lm *DefaultLifecycleManager) ShutdownChannel() <-chan struct{}
    ShutdownChannel returns a channel that closes when shutdown is initiated

func (lm *DefaultLifecycleManager) Startup(ctx context.Context) error
    Startup implements the LifecycleManager interface

type DefaultTestExecutor struct {
	// Has unexported fields.
}
    DefaultTestExecutor implements the TestExecutor interface using modular
    components

func (e *DefaultTestExecutor) ExecuteSingle(ctx context.Context, packages []string, config *Configuration) error
    ExecuteSingle executes tests once for the specified packages

func (e *DefaultTestExecutor) ExecuteWatch(ctx context.Context, config *Configuration) error
    ExecuteWatch executes tests in watch mode with file monitoring

func (e *DefaultTestExecutor) SetConfiguration(config *Configuration) error
    SetConfiguration configures the test executor with the application
    configuration

type DependencyContainer interface {
	// Register registers a component with the container
	Register(name string, component interface{}) error

	// Resolve retrieves a component from the container
	Resolve(name string) (interface{}, error)

	// ResolveAs retrieves a component and casts it to the specified type
	ResolveAs(name string, target interface{}) error

	// Initialize initializes all registered components
	Initialize() error

	// Cleanup cleans up all registered components
	Cleanup() error
}
    DependencyContainer manages component dependencies and injection

func NewContainer() DependencyContainer
    NewContainer creates a new dependency injection container

type DisplayRenderer interface {
	RenderResults(ctx context.Context) error
	SetConfiguration(config *Configuration) error
}
    DisplayRenderer interface for result display (will be defined in ui package)

func NewDisplayRenderer() DisplayRenderer
    NewDisplayRenderer creates a new display renderer with modular UI components

type Initializer interface {
	Initialize() error
}
    Initializer interface for components that need initialization

type LifecycleManager interface {
	// Startup initializes all application components
	Startup(ctx context.Context) error

	// Shutdown gracefully stops all application components
	Shutdown(ctx context.Context) error

	// IsRunning returns whether the application is currently running
	IsRunning() bool

	// RegisterShutdownHook adds a function to be called during shutdown
	RegisterShutdownHook(hook func() error)
}
    LifecycleManager manages application startup and shutdown

func NewLifecycleManager() LifecycleManager
    NewLifecycleManager creates a new lifecycle manager

type PathsConfig struct {
	// IncludePatterns lists patterns to include
	IncludePatterns []string

	// ExcludePatterns lists patterns to exclude
	ExcludePatterns []string
}
    PathsConfig represents path-specific configuration

type SimpleController struct{}
    SimpleController provides basic functionality for CLI migration

func NewLegacyAppController() *SimpleController
    NewLegacyAppController creates a simple controller for backwards
    compatibility

func (s *SimpleController) Run(args []string) error
    Run executes the simple controller

type TestConfig struct {
	// Timeout for test execution
	Timeout string

	// Parallel execution settings
	Parallel int

	// Coverage settings
	Coverage bool
}
    TestConfig represents test execution configuration

type TestExecutor interface {
	ExecuteSingle(ctx context.Context, packages []string, config *Configuration) error
	ExecuteWatch(ctx context.Context, config *Configuration) error
}
    TestExecutor interface for test execution (will be defined in test package)

func NewTestExecutor() TestExecutor
    NewTestExecutor creates a new test executor with modular components

type VisualConfig struct {
	// Icons setting (none, simple, rich)
	Icons string

	// Theme setting
	Theme string

	// TerminalWidth for display formatting
	TerminalWidth int
}
    VisualConfig represents visual/UI configuration

type WatchConfig struct {
	// Enabled indicates if watch mode is enabled
	Enabled bool

	// IgnorePatterns lists patterns to ignore
	IgnorePatterns []string

	// Debounce duration for file events
	Debounce string

	// RunOnStart runs tests on startup
	RunOnStart bool

	// ClearOnRerun clears screen between runs
	ClearOnRerun bool
}
    WatchConfig represents watch-specific configuration

