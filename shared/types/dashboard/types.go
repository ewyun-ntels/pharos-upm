// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    annotation, err := UnmarshalAnnotation(bytes)
//    bytes, err = annotation.Marshal()
//
//    chartQuery, err := UnmarshalChartQuery(bytes)
//    bytes, err = chartQuery.Marshal()
//
//    dashboardConfig, err := UnmarshalDashboardConfig(bytes)
//    bytes, err = dashboardConfig.Marshal()
//
//    dashboardCreateResponse, err := UnmarshalDashboardCreateResponse(bytes)
//    bytes, err = dashboardCreateResponse.Marshal()
//
//    dashboardData, err := UnmarshalDashboardData(bytes)
//    bytes, err = dashboardData.Marshal()
//
//    dashboardFolder, err := UnmarshalDashboardFolder(bytes)
//    bytes, err = dashboardFolder.Marshal()
//
//    dashboardPermission, err := UnmarshalDashboardPermission(bytes)
//    bytes, err = dashboardPermission.Marshal()
//
//    dataProvider, err := UnmarshalDataProvider(bytes)
//    bytes, err = dataProvider.Marshal()
//
//    filterConfig, err := UnmarshalFilterConfig(bytes)
//    bytes, err = filterConfig.Marshal()
//
//    kind, err := UnmarshalKind(bytes)
//    bytes, err = kind.Marshal()
//
//    panel, err := UnmarshalPanel(bytes)
//    bytes, err = panel.Marshal()
//
//    panelLayout, err := UnmarshalPanelLayout(bytes)
//    bytes, err = panelLayout.Marshal()
//
//    queryInspectResponse, err := UnmarshalQueryInspectResponse(bytes)
//    bytes, err = queryInspectResponse.Marshal()
//
//    queryResponse, err := UnmarshalQueryResponse(bytes)
//    bytes, err = queryResponse.Marshal()

package dashboard

import "time"

import "encoding/json"

func UnmarshalAnnotation(data []byte) (Annotation, error) {
	var r Annotation
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Annotation) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalChartQuery(data []byte) (ChartQuery, error) {
	var r ChartQuery
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ChartQuery) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalDashboardConfig(data []byte) (DashboardConfig, error) {
	var r DashboardConfig
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *DashboardConfig) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalDashboardCreateResponse(data []byte) (DashboardCreateResponse, error) {
	var r DashboardCreateResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *DashboardCreateResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalDashboardData(data []byte) (DashboardData, error) {
	var r DashboardData
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *DashboardData) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalDashboardFolder(data []byte) (DashboardFolder, error) {
	var r DashboardFolder
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *DashboardFolder) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalDashboardPermission(data []byte) (DashboardPermission, error) {
	var r DashboardPermission
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *DashboardPermission) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalDataProvider(data []byte) (DataProvider, error) {
	var r DataProvider
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *DataProvider) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalFilterConfig(data []byte) (FilterConfig, error) {
	var r FilterConfig
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *FilterConfig) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalKind(data []byte) (Kind, error) {
	var r Kind
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Kind) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalPanel(data []byte) (Panel, error) {
	var r Panel
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Panel) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalPanelLayout(data []byte) (PanelLayout, error) {
	var r PanelLayout
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PanelLayout) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalQueryInspectResponse(data []byte) (QueryInspectResponse, error) {
	var r QueryInspectResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *QueryInspectResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalQueryResponse(data []byte) (QueryResponse, error) {
	var r QueryResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *QueryResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// API response type for dashboard POST requests
type DashboardCreateResponse struct {
	// Created dashboard ID
	ID string `json:"id"`
}

// API response type for dashboard GET requests
type DashboardData struct {
	// Annotations stored separately from config
	Annotations []Annotation `json:"annotations,omitempty"`
	// Dashboard configuration
	Config DashboardConfig `json:"config"`
	// Folder this dashboard belongs to (optional)
	FolderID *string `json:"folderId,omitempty"`
	// Unique identifier for the dashboard
	ID string `json:"id"`
	// User permission level for this dashboard
	Permission Action `json:"permission"`
}

// Annotation for marking specific timestamps on timeseries charts. scope='global' means
// visible on all timeseries panels; scope='panel' means visible only on the panel that owns
// it.
type Annotation struct {
	// Line and label color (CSS color string). Defaults to red if omitted.
	Color *string `json:"color,omitempty"`
	// Unique identifier for the annotation
	ID string `json:"id"`
	// For panel-scoped annotations: the panel ID this annotation belongs to. Null/omitted for
	// global annotations.
	PanelID *string `json:"panel_id,omitempty"`
	// global: stored in DashboardConfig.annotations, shown on all timeseries panels. panel:
	// stored in panel.options.annotations, shown only on that panel.
	Scope Scope `json:"scope"`
	// Annotation label text
	Text string `json:"text"`
	// Unix timestamp in milliseconds
	Time float64 `json:"time"`
}

// Dashboard configuration
type DashboardConfig struct {
	// Global annotations visible on all timeseries panels in this dashboard
	Annotations []Annotation `json:"annotations,omitempty"`
	Description string       `json:"description"`
	DisplayName string       `json:"displayName"`
	Favorite    bool         `json:"favorite"`
	// Unified filter configuration (replaces headerL, headerR, left)
	Filters   []FilterConfig `json:"filters"`
	GroupName *string        `json:"groupName,omitempty"`
	// Array of panels in the dashboard
	Panels []Panel `json:"panels,omitempty"`
	Title  string  `json:"title"`
	Type   string  `json:"type"`
}

// Unified filter configuration (replaces headerL, headerR, left). Follows Panel pattern
// with common fields + type-specific options.
type FilterConfig struct {
	// Data source name for dynamic filters
	DatasourceName *string `json:"datasourceName,omitempty"`
	// Unique identifier for the filter
	ID   string `json:"id"`
	Kind Kind   `json:"kind"`
	// Display label for the filter
	Label *string `json:"label,omitempty"`
	// Type-specific options. Structure varies by type (select, datetime, etc). For select type,
	// use isMulti field to enable multi-selection. Stored as flexible JSON object.
	Options map[string]interface{} `json:"options,omitempty"`
	// Query string for data fetching
	Query *string `json:"query,omitempty"`
	// Filter type: select, datetime, step, refreshInterval, custom, tableSearch
	Type FilterConfigType `json:"type"`
}

// Panel: Dashboard visualization unit. Contains content data and grid layout information.
type Panel struct {
	// Whether panel background is transparent
	BgTransparent *bool `json:"bgTransparent,omitempty"`
	// Data source configuration
	DataProvider *DataProvider `json:"dataProvider,omitempty"`
	// Optional description for the panel
	Description *string `json:"description,omitempty"`
	// Unique identifier for the panel
	ID string `json:"id"`
	// Grid layout configuration for react-grid-layout
	Layout PanelLayout `json:"layout"`
	// Plugin-specific options. Structure varies by renderType (pie, table, line, etc). Stored
	// as flexible JSON object.
	Options map[string]interface{} `json:"options,omitempty"`
	// Visualization type: pie, table, timeSeries, barGauge, etc.
	RenderType *string `json:"renderType,omitempty"`
	// Panel title displayed in UI
	Title string `json:"title"`
}

// Data source configuration
//
// Data source configuration for panels
type DataProvider struct {
	// Array of chart queries
	ChartQuery []ChartQuery `json:"chartQuery,omitempty"`
	// Name of the data provider
	DataProviderName *string `json:"dataProviderName,omitempty"`
	// Resource identifier
	Resource *string `json:"resource,omitempty"`
}

// Query configuration for chart data source
type ChartQuery struct {
	// Name of the datasource
	DatasourceName string `json:"datasourceName"`
	// Timestamp to force cache invalidation
	ForceRefresh *float64 `json:"forceRefresh,omitempty"`
	// Label for the query
	Label string `json:"label"`
	// Query string
	Query string `json:"query"`
	// Template interpolation style
	TemplateStyle *string `json:"templateStyle,omitempty"`
}

// Grid layout configuration for react-grid-layout
type PanelLayout struct {
	// Whether row is collapsed (only for type='row')
	Collapsed *bool `json:"collapsed,omitempty"`
	// Grid height (rows)
	H float64 `json:"h"`
	// Whether panel can be dragged
	IsDraggable *bool `json:"isDraggable,omitempty"`
	// Whether panel can be resized
	IsResizable *bool `json:"isResizable,omitempty"`
	// Maximum height constraint
	MaxH *float64 `json:"maxH,omitempty"`
	// Maximum width constraint
	MaxW *float64 `json:"maxW,omitempty"`
	// Minimum height constraint
	MinH *float64 `json:"minH,omitempty"`
	// Minimum width constraint
	MinW *float64 `json:"minW,omitempty"`
	// Whether panel is fixed (no drag/resize)
	Static *bool `json:"static,omitempty"`
	// Layout type: card (normal panel) or row (container for nested panels)
	Type PanelLayoutType `json:"type"`
	// Grid width (columns)
	W float64 `json:"w"`
	// Grid X position
	X float64 `json:"x"`
	// Grid Y position
	Y float64 `json:"y"`
}

// Dashboard folder for grouping dashboards
type DashboardFolder struct {
	// Creation timestamp
	CreatedAt time.Time `json:"createdAt"`
	// Unique identifier for the folder
	ID string `json:"id"`
	// Folder name
	Name string `json:"name"`
	// Last update timestamp
	UpdatedAt time.Time `json:"updatedAt"`
}

type DashboardPermission struct {
	Action  Action `json:"action"`
	Kind    string `json:"kind"`
	Object  string `json:"object"`
	Subject string `json:"subject"`
}

// Extended query response with debug information (when ?inspect=true)
type QueryInspectResponse struct {
	// Query result rows
	Data    []map[string]interface{} `json:"data"`
	Inspect InspectInfo              `json:"inspect"`
	// Column metadata
	Meta []map[string]string `json:"meta"`
	// Row count
	Rows       int64                          `json:"rows"`
	Statistics QueryInspectResponseStatistics `json:"statistics"`
}

// Additional debug information for query inspection
type InspectInfo struct {
	// Datasource name
	DatasourceName string `json:"datasourceName"`
	// Error message if query failed
	Error *string `json:"error,omitempty"`
	// Actual executed query (variables substituted)
	ExecutedQuery string `json:"executedQuery"`
	// Template query (with {{variables}})
	RawQuery string `json:"rawQuery"`
	// Query execution timestamp (ISO 8601)
	Timestamp time.Time `json:"timestamp"`
	// Variable values used in substitution
	Variables map[string]interface{} `json:"variables"`
}

// Query execution statistics
type QueryInspectResponseStatistics struct {
	// Execution time in seconds
	Elapsed float64 `json:"elapsed"`
}

// Dashboard query execution response
type QueryResponse struct {
	// Query result rows
	Data []map[string]interface{} `json:"data"`
	// Column metadata
	Meta []map[string]string `json:"meta"`
	// Row count
	Rows       int64                   `json:"rows"`
	Statistics QueryResponseStatistics `json:"statistics"`
}

// Query execution statistics
type QueryResponseStatistics struct {
	// Execution time in seconds
	Elapsed float64 `json:"elapsed"`
}

// global: stored in DashboardConfig.annotations, shown on all timeseries panels. panel:
// stored in panel.options.annotations, shown only on that panel.
type Scope string

const (
	Global     Scope = "global"
	ScopePanel Scope = "panel"
)

// Panel/filter kind used for backend query routing. Determines where a panel or filter is
// located: headerL (top-left), headerR (top-right), left (sidebar), panels (chart panel
// grid).
type Kind string

const (
	HeaderL Kind = "headerL"
	HeaderR Kind = "headerR"
	Left    Kind = "left"
	Panels  Kind = "panels"
)

// Filter type: select, datetime, step, refreshInterval, custom, tableSearch
type FilterConfigType string

const (
	Custom          FilterConfigType = "custom"
	Datetime        FilterConfigType = "datetime"
	RefreshInterval FilterConfigType = "refreshInterval"
	Select          FilterConfigType = "select"
	Step            FilterConfigType = "step"
	TableSearch     FilterConfigType = "tableSearch"
)

// Layout type: card (normal panel) or row (container for nested panels)
type PanelLayoutType string

const (
	Card PanelLayoutType = "card"
	Row  PanelLayoutType = "row"
)

// User permission level for this dashboard
type Action string

const (
	Editor Action = "editor"
	Owner  Action = "owner"
	Viewer Action = "viewer"
)
