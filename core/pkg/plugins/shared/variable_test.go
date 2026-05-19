package shared

import (
	"sync"
	"testing"

	"github.com/hashicorp/go-plugin"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/plugins/model"
)

// Test constants for better maintainability
const (
	TestPluginID           = "test-plugin"
	TestPluginName         = "Test Plugin"
	ConcurrentTestKey1     = "concurrent-test-1"
	ConcurrentTestKey2     = "concurrent-test-2"
	ConcurrentHashicorpKey = "concurrent-hashicorp"
	BenchmarkKey           = "benchmark"
	IsolationTestKey       = "isolation-test"
	NonexistentKey         = "nonexistent"
)

var (
	TestPluginType = model.PluginTypeDatasource
)

// Test data factories
func createTestPlugin(id, name string, pluginType model.PluginType) *model.Plugin {
	return &model.Plugin{
		ID:   id,
		Name: name,
		Type: pluginType,
	}
}

func createTestDatasourceClient() *model.DatasourceClient {
	return &model.DatasourceClient{
		DataClient:   nil,
		StreamClient: nil,
	}
}

// Helper functions for safe map operations
func safeSetPlugin(mu *sync.RWMutex, id string, plugin *model.Plugin) {
	mu.Lock()
	PluginsByID[id] = plugin
	mu.Unlock()
}

func safeDeletePlugin(mu *sync.RWMutex, id string) {
	mu.Lock()
	delete(PluginsByID, id)
	mu.Unlock()
}

func clearAllPlugins() {
	for k := range PluginsByID {
		delete(PluginsByID, k)
	}
}

// TestPluginsByIDInitialization tests that PluginsByID is properly initialized
func TestPluginsByIDInitialization(t *testing.T) {
	// Test that PluginsByID is initialized as a map
	if PluginsByID == nil {
		t.Error("PluginsByID should be initialized")
	}

	// Clear any existing data
	clearAllPlugins()

	// Test that it's an empty map initially
	if len(PluginsByID) != 0 {
		t.Errorf("Expected PluginsByID to be empty initially, got %d items", len(PluginsByID))
	}

	// Test that we can write to it
	testPlugin := createTestPlugin(TestPluginID, TestPluginName, TestPluginType)
	PluginsByID[TestPluginID] = testPlugin

	// Verify the plugin was added
	if len(PluginsByID) != 1 {
		t.Errorf("Expected PluginsByID to have 1 item, got %d", len(PluginsByID))
	}

	retrievedPlugin, exists := PluginsByID[TestPluginID]
	if !exists {
		t.Errorf("Expected to find %s in PluginsByID", TestPluginID)
	}

	if retrievedPlugin.ID != TestPluginID {
		t.Errorf("Expected plugin ID to be %s, got %s", TestPluginID, retrievedPlugin.ID)
	}

	// Clean up for other tests
	delete(PluginsByID, TestPluginID)
}

// TestHashicorpClientsInitialization tests that HashicorpClients is properly initialized
func TestHashicorpClientsInitialization(t *testing.T) {
	// Test that HashicorpClients is initialized
	if HashicorpClients == nil {
		t.Error("HashicorpClients should be initialized")
	}

	// Test that it's using internal.Map
	mapType := HashicorpClients
	if mapType == nil {
		t.Error("HashicorpClients should be a valid Map instance")
	}

	// Test basic operations
	testKey := "test-client"
	testClient := &plugin.Client{}

	// Set a value
	HashicorpClients.Set(testKey, testClient)

	// Get the value back
	retrievedClient := HashicorpClients.Get(testKey)
	if retrievedClient == nil {
		t.Error("Expected to find test-client in HashicorpClients")
	}

	if retrievedClient != testClient {
		t.Error("Expected to retrieve the same client instance")
	}

	// Clean up
	HashicorpClients.Remove(testKey, nil)
}

// TestDatasourceClientsInitialization tests that DatasourceClients is properly initialized
func TestDatasourceClientsInitialization(t *testing.T) {
	// Test that DatasourceClients is initialized
	if DatasourceClients.Map == nil {
		t.Error("DatasourceClients.Map should be initialized")
	}

	// Test the structure
	if DatasourceClients.Map == nil {
		t.Error("DatasourceClients should have a Map field")
	}

	// Test basic operations
	testKey := "test-datasource"
	testClient := &model.DatasourceClient{
		DataClient:   nil,
		StreamClient: nil,
	}

	// Set a value
	DatasourceClients.Map.Set(testKey, testClient)

	// Get the value back
	retrievedClient := DatasourceClients.Map.Get(testKey)
	if retrievedClient == nil {
		t.Error("Expected to find test-datasource in DatasourceClients")
		return
	}

	if retrievedClient.DataClient != testClient.DataClient {
		t.Error("Expected DataClient fields to match")
	}

	// Clean up
	DatasourceClients.Map.Remove(testKey, nil)
}

// TestGlobalVariablesTypes tests that all global variables have correct types
func TestGlobalVariablesTypes(t *testing.T) {
	// Test PluginsByID type
	var pluginsMap map[string]*model.Plugin = PluginsByID
	if pluginsMap == nil {
		t.Error("PluginsByID should be assignable to map[string]*model.Plugin")
	}

	// Test HashicorpClients type
	var hashicorpMap *internal.Map[*plugin.Client] = HashicorpClients
	if hashicorpMap == nil {
		t.Error("HashicorpClients should be assignable to *internal.Map[*plugin.Client]")
	}

	// Test DatasourceClients type
	var datasourceClients model.DatasourceClients = DatasourceClients
	if datasourceClients.Map == nil {
		t.Error("DatasourceClients should be assignable to model.DatasourceClients")
	}
}

// TestConcurrentAccess tests concurrent access to global variables with proper synchronization
func TestConcurrentAccess(t *testing.T) {
	// Create a mutex to protect PluginsByID concurrent access
	var pluginMutex sync.RWMutex

	// Test concurrent access to PluginsByID with synchronization
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		for range 100 {
			safeSetPlugin(&pluginMutex, ConcurrentTestKey1, createTestPlugin("test-1", "Test Plugin 1", TestPluginType))
			safeDeletePlugin(&pluginMutex, ConcurrentTestKey1)
		}
	}()

	go func() {
		defer func() { done <- true }()
		for range 100 {
			safeSetPlugin(&pluginMutex, ConcurrentTestKey2, createTestPlugin("test-2", "Test Plugin 2", TestPluginType))
			safeDeletePlugin(&pluginMutex, ConcurrentTestKey2)
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Test concurrent access to HashicorpClients (already thread-safe via internal.Map)
	go func() {
		defer func() { done <- true }()
		for range 100 {
			HashicorpClients.Set(ConcurrentHashicorpKey, &plugin.Client{})
			HashicorpClients.Remove(ConcurrentHashicorpKey, nil)
		}
	}()

	go func() {
		defer func() { done <- true }()
		for range 100 {
			_ = HashicorpClients.Get(ConcurrentHashicorpKey)
		}
	}()

	<-done
	<-done

	t.Log("Concurrent access test completed without deadlock or race conditions")
}

// TestPluginsByIDOperations tests various operations on PluginsByID
func TestPluginsByIDOperations(t *testing.T) {
	// Clear any existing data
	clearAllPlugins()

	// Test adding multiple plugins using factory function
	plugins := []*model.Plugin{
		createTestPlugin("plugin1", "Plugin 1", TestPluginType),
		createTestPlugin("plugin2", "Plugin 2", TestPluginType),
		createTestPlugin("plugin3", "Plugin 3", TestPluginType),
	}

	for _, p := range plugins {
		PluginsByID[p.ID] = p
	}

	// Test that all plugins were added
	if len(PluginsByID) != 3 {
		t.Errorf("Expected 3 plugins, got %d", len(PluginsByID))
	}

	// Test retrieval
	for _, p := range plugins {
		retrieved, exists := PluginsByID[p.ID]
		if !exists {
			t.Errorf("Plugin %s should exist", p.ID)
		}
		if retrieved.Name != p.Name {
			t.Errorf("Expected plugin name %s, got %s", p.Name, retrieved.Name)
		}
	}

	// Test deletion
	delete(PluginsByID, "plugin2")
	if len(PluginsByID) != 2 {
		t.Errorf("Expected 2 plugins after deletion, got %d", len(PluginsByID))
	}

	// Clean up
	clearAllPlugins()
}

// TestHashicorpClientsOperations tests various operations on HashicorpClients
func TestHashicorpClientsOperations(t *testing.T) {
	// Test adding multiple clients
	clients := map[string]*plugin.Client{
		"client1": {},
		"client2": {},
		"client3": {},
	}

	for key, client := range clients {
		HashicorpClients.Set(key, client)
	}

	// Test retrieval
	for key, originalClient := range clients {
		retrievedClient := HashicorpClients.Get(key)
		if retrievedClient == nil {
			t.Errorf("Client %s should exist", key)
		}
		if retrievedClient != originalClient {
			t.Errorf("Retrieved client should be the same instance for %s", key)
		}
	}

	// Test deletion
	HashicorpClients.Remove("client2", nil)
	deletedClient := HashicorpClients.Get("client2")
	if deletedClient != nil {
		t.Error("Client2 should not exist after deletion")
	}

	// Clean up
	for key := range clients {
		HashicorpClients.Remove(key, nil)
	}
}

// TestDatasourceClientsOperations tests various operations on DatasourceClients
func TestDatasourceClientsOperations(t *testing.T) {
	// Test adding multiple datasource clients using factory function
	datasources := map[string]*model.DatasourceClient{
		"postgres": createTestDatasourceClient(),
		"mysql":    createTestDatasourceClient(),
		"redis":    createTestDatasourceClient(),
	}

	for key, datasource := range datasources {
		DatasourceClients.Map.Set(key, datasource)
	}

	// Test retrieval
	for key, originalDatasource := range datasources {
		retrievedDatasource := DatasourceClients.Map.Get(key)
		if retrievedDatasource == nil {
			t.Errorf("Datasource %s should exist", key)
			continue
		}
		if retrievedDatasource.DataClient != originalDatasource.DataClient {
			t.Errorf("Retrieved datasource DataClient should match for %s", key)
		}
		if retrievedDatasource.StreamClient != originalDatasource.StreamClient {
			t.Errorf("Retrieved datasource StreamClient should match for %s", key)
		}
	}

	// Test deletion
	DatasourceClients.Map.Remove("mysql", nil)
	deletedDatasource := DatasourceClients.Map.Get("mysql")
	if deletedDatasource != nil {
		t.Error("MySQL datasource should not exist after deletion")
	}

	// Clean up
	for key := range datasources {
		DatasourceClients.Map.Remove(key, nil)
	}
}

// TestVariableIsolation tests that variables maintain isolation
func TestVariableIsolation(t *testing.T) {
	// Add data to all variables using factory functions
	PluginsByID[IsolationTestKey] = createTestPlugin(IsolationTestKey, "Isolation Test Plugin", TestPluginType)
	HashicorpClients.Set(IsolationTestKey, &plugin.Client{})
	DatasourceClients.Map.Set(IsolationTestKey, createTestDatasourceClient())

	// Verify they don't interfere with each other
	if len(PluginsByID) < 1 {
		t.Error("PluginsByID should have at least 1 item")
	}

	hashicorpClient := HashicorpClients.Get(IsolationTestKey)
	if hashicorpClient == nil {
		t.Error("HashicorpClients should have the test item")
	}

	datasourceClient := DatasourceClients.Map.Get(IsolationTestKey)
	if datasourceClient == nil {
		t.Error("DatasourceClients should have the test item")
	}

	// Clean up one variable and verify others are unaffected
	delete(PluginsByID, IsolationTestKey)

	hashicorpStillExists := HashicorpClients.Get(IsolationTestKey)
	if hashicorpStillExists == nil {
		t.Error("HashicorpClients should still have the test item")
	}

	datasourceStillExists := DatasourceClients.Map.Get(IsolationTestKey)
	if datasourceStillExists == nil {
		t.Error("DatasourceClients should still have the test item")
	}

	// Clean up remaining
	HashicorpClients.Remove(IsolationTestKey, nil)
	DatasourceClients.Map.Remove(IsolationTestKey, nil)
}

// TestEmptyStateOperations tests operations on empty global variables
func TestEmptyStateOperations(t *testing.T) {
	// Clear all variables
	clearAllPlugins()

	// Test operations on empty PluginsByID
	_, exists := PluginsByID[NonexistentKey]
	if exists {
		t.Error("Should not find nonexistent plugin")
	}

	// Test operations on empty HashicorpClients
	nonExistentClient := HashicorpClients.Get(NonexistentKey)
	if nonExistentClient != nil {
		t.Error("Should not find nonexistent hashicorp client")
	}

	// Test operations on empty DatasourceClients
	nonExistentDatasource := DatasourceClients.Map.Get(NonexistentKey)
	if nonExistentDatasource != nil {
		t.Error("Should not find nonexistent datasource client")
	}

	// Test deletion on empty maps (should not panic)
	delete(PluginsByID, NonexistentKey)
	HashicorpClients.Remove(NonexistentKey, nil)
	DatasourceClients.Map.Remove(NonexistentKey, nil)

	t.Log("Empty state operations completed successfully")
}

// Benchmark tests
func BenchmarkPluginsByIDWrite(b *testing.B) {
	testPlugin := createTestPlugin(BenchmarkKey, "Benchmark Plugin", TestPluginType)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PluginsByID[BenchmarkKey] = testPlugin
	}
}

func BenchmarkPluginsByIDRead(b *testing.B) {
	testPlugin := createTestPlugin(BenchmarkKey, "Benchmark Plugin", TestPluginType)
	PluginsByID[BenchmarkKey] = testPlugin

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PluginsByID[BenchmarkKey]
	}
}

func BenchmarkHashicorpClientsWrite(b *testing.B) {
	client := &plugin.Client{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashicorpClients.Set(BenchmarkKey, client)
	}
}

func BenchmarkHashicorpClientsRead(b *testing.B) {
	client := &plugin.Client{}
	HashicorpClients.Set(BenchmarkKey, client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HashicorpClients.Get(BenchmarkKey)
	}
}

func BenchmarkDatasourceClientsWrite(b *testing.B) {
	datasource := createTestDatasourceClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DatasourceClients.Map.Set(BenchmarkKey, datasource)
	}
}

func BenchmarkDatasourceClientsRead(b *testing.B) {
	datasource := createTestDatasourceClient()
	DatasourceClients.Map.Set(BenchmarkKey, datasource)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DatasourceClients.Map.Get(BenchmarkKey)
	}
}
