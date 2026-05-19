# 데이터소스 플러그인 + Panel 통합 완전 가이드

Pharos에 **HashiCorp 플러그인으로 외부 데이터소스를 추가**하고, Panel 플러그인에서 사용하는 완전한 가이드입니다.

---

## 🎯 현재 상황

### ✅ 이미 준비된 것

1. **Backend**: HashiCorp 플러그인 시스템으로 동적 로드 지원
2. **API**: REST API로 데이터소스 CRUD 가능
   ```
   GET    /plugins                  - 사용 가능한 플러그인 목록
   GET    /plugins/datasources      - 등록된 데이터소스 목록
   POST   /plugins/datasources      - 데이터소스 생성
   PUT    /plugins/datasources/:name - 데이터소스 수정
   DELETE /plugins/datasources/:name - 데이터소스 삭제
   POST   /plugins/ds/query         - 쿼리 실행
   ```
3. **JSON Schema**: React JSON Schema Form으로 동적 UI 생성 가능

### ❌ 없는 것

- **데이터소스 관리 UI**: 데이터소스를 생성/수정/삭제하는 프론트엔드 페이지가 없음

---

## 📋 목차

- [1단계: HashiCorp 데이터소스 플러그인 개발](#1단계-hashicorp-데이터소스-플러그인-개발)
- [2단계: 데이터소스 관리 UI 추가 (선택)](#2단계-데이터소스-관리-ui-추가-선택)
- [3단계: Panel 플러그인에서 사용](#3단계-panel-플러그인에서-사용)
- [실전 예제: MongoDB 플러그인](#실전-예제-mongodb-플러그인)

---

## 1단계: HashiCorp 데이터소스 플러그인 개발

### 1.1 디렉토리 구조

```
my-datasource-plugin/
├── go.mod
├── main.go                    # HashiCorp 플러그인 진입점
├── plugin.go                  # 데이터소스 구현
├── schema/
│   ├── json.json             # JSON Schema
│   ├── ui.json               # UI Schema
│   └── sample-form-data.json # 샘플 데이터
└── README.md
```

### 1.2 go.mod

```go
module github.com/myorg/pharos-mongodb-plugin

go 1.25

require (
    github.com/hashicorp/go-plugin v1.6.0
    go.mongodb.org/mongo-driver v1.12.0
    ntels.com/pharos/core v0.0.0
)

// Pharos core 로컬 개발
replace ntels.com/pharos/core => /path/to/pharos/core
```

### 1.3 main.go (HashiCorp 플러그인 진입점)

```go
package main

import (
    "github.com/hashicorp/go-plugin"
    "ntels.com/pharos/core/pkg/plugins/model"
    myplugin "github.com/myorg/pharos-mongodb-plugin"
)

func main() {
    // HashiCorp 플러그인 핸드셰이크
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: model.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "datasource": &model.DataPlugin{Impl: &myplugin.MongoDBPlugin{}},
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

### 1.4 plugin.go (데이터소스 구현)

```go
package myplugin

import (
    _ "embed"
    "context"
    "encoding/json"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "ntels.com/pharos/core/external/orm"
    "ntels.com/pharos/core/pkg/plugins/model"
)

//go:embed schema/json.json
var jsonSchema string

//go:embed schema/ui.json
var uiSchema string

//go:embed schema/sample-form-data.json
var sampleFormData string

type MongoDBPlugin struct {
    clients map[string]*mongo.Client
}

// Type은 플러그인 타입을 반환합니다 (Pharos에서 인식하는 ID)
func (p *MongoDBPlugin) Type() string {
    return "mongodb"
}

// Name은 플러그인 표시 이름을 반환합니다
func (p *MongoDBPlugin) Name() string {
    return "MongoDB"
}

// GetJSONSchema는 JSON Schema를 반환합니다
func (p *MongoDBPlugin) GetJSONSchema() string {
    return jsonSchema
}

// GetUISchema는 UI Schema를 반환합니다
func (p *MongoDBPlugin) GetUISchema() string {
    return uiSchema
}

// GetSampleFormData는 샘플 폼 데이터를 반환합니다
func (p *MongoDBPlugin) GetSampleFormData() string {
    return sampleFormData
}

// QueryData는 쿼리를 실행하고 결과를 반환합니다
func (p *MongoDBPlugin) QueryData(request *model.QueryDataRequest) *model.QueryDataResponse {
    response := &model.QueryDataResponse{Results: map[string]model.QueryDataResult{}}

    if p.clients == nil {
        p.clients = make(map[string]*mongo.Client)
    }

    // MongoDB 클라이언트 가져오기
    client, err := p.getClient(request.Datasource)
    if err != nil {
        for _, query := range request.Queries {
            response.Results[query.ID] = model.QueryDataResult{
                Error: fmt.Sprintf("Failed to connect: %v", err),
            }
        }
        return response
    }

    // 각 쿼리 실행
    for _, query := range request.Queries {
        result := p.executeQuery(client, request.Datasource, query)
        response.Results[query.ID] = result
    }

    return response
}

func (p *MongoDBPlugin) getClient(datasource model.Datasource) (*mongo.Client, error) {
    if client, exists := p.clients[datasource.Name]; exists {
        return client, nil
    }

    connectionString := datasource.Data["connection_string"].(string)
    timeout := 30 * time.Second

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
    if err != nil {
        return nil, err
    }

    if err := client.Ping(ctx, nil); err != nil {
        return nil, err
    }

    p.clients[datasource.Name] = client
    return client, nil
}

func (p *MongoDBPlugin) executeQuery(client *mongo.Client, datasource model.Datasource, query model.Query) model.QueryDataResult {
    database := datasource.Data["database"].(string)

    // MongoDB 쿼리 파싱
    var mongoQuery bson.M
    if err := json.Unmarshal([]byte(query.SQL), &mongoQuery); err != nil {
        return model.QueryDataResult{Error: fmt.Sprintf("Invalid query: %v", err)}
    }

    collectionName, ok := mongoQuery["collection"].(string)
    if !ok {
        return model.QueryDataResult{Error: "Missing 'collection' field"}
    }

    collection := client.Database(database).Collection(collectionName)
    findQuery, _ := mongoQuery["find"].(bson.M)

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    cursor, err := collection.Find(ctx, findQuery)
    if err != nil {
        return model.QueryDataResult{Error: fmt.Sprintf("Query failed: %v", err)}
    }
    defer cursor.Close(ctx)

    var results []map[string]interface{}
    if err := cursor.All(ctx, &results); err != nil {
        return model.QueryDataResult{Error: fmt.Sprintf("Failed to decode: %v", err)}
    }

    return model.QueryDataResult{Frame: convertToFrame(results)}
}

func convertToFrame(results []map[string]interface{}) orm.DatabaseResponse {
    if len(results) == 0 {
        return orm.DatabaseResponse{Meta: []orm.Meta{}, Data: []map[string]interface{}{}}
    }

    var meta []orm.Meta
    for key := range results[0] {
        meta = append(meta, orm.Meta{Name: key})
    }

    return orm.DatabaseResponse{Meta: meta, Data: results}
}

// RemoveDatasource는 데이터소스를 제거할 때 호출됩니다
func (p *MongoDBPlugin) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
    if client, exists := p.clients[request.Datasource.Name]; exists {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        if err := client.Disconnect(ctx); err != nil {
            return err
        }

        delete(p.clients, request.Datasource.Name)
    }
    return nil
}
```

### 1.5 schema/json.json

```json
{
  "type": "object",
  "required": ["type", "name", "connection_string", "database"],
  "properties": {
    "type": {
      "type": "string",
      "title": "Type",
      "default": "mongodb",
      "readOnly": true
    },
    "name": {
      "type": "string",
      "title": "Name",
      "default": "mongodb-01",
      "minLength": 1
    },
    "connection_string": {
      "type": "string",
      "title": "Connection String",
      "default": "mongodb://localhost:27017",
      "pattern": "^mongodb(\\+srv)?://.*"
    },
    "database": {
      "type": "string",
      "title": "Database",
      "minLength": 1
    },
    "timeout": {
      "type": "integer",
      "title": "Timeout (seconds)",
      "default": 30,
      "minimum": 1,
      "maximum": 300
    }
  }
}
```

### 1.6 schema/ui.json

```json
{
  "connection_string": {
    "ui:help": "MongoDB connection string (e.g., mongodb://user:pass@host:port)"
  },
  "timeout": {
    "ui:widget": "range"
  }
}
```

### 1.7 schema/sample-form-data.json

```json
{
  "type": "mongodb",
  "name": "mongodb-sample",
  "connection_string": "mongodb://localhost:27017",
  "database": "mydb",
  "timeout": 30
}
```

### 1.8 빌드

```bash
# 플러그인 빌드
go build -o mongodb-plugin

# 실행 테스트
./mongodb-plugin

# Pharos가 인식할 수 있도록 배포
mkdir -p /opt/pharos/plugins
mv mongodb-plugin /opt/pharos/plugins/
chmod +x /opt/pharos/plugins/mongodb-plugin
```

---

## 2단계: 데이터소스 관리 UI 추가 (선택)

Pharos Core에 데이터소스 관리 페이지를 추가합니다.

### 2.1 페이지 구조

```
core/frontend/src/app/datasources/
├── page.tsx                    # 메인 페이지
├── components/
│   ├── DatasourceList.tsx     # 데이터소스 목록
│   ├── DatasourceForm.tsx     # 생성/수정 폼
│   └── DatasourceCard.tsx     # 카드 UI
└── hooks/
    └── useDatasources.ts      # API 훅
```

### 2.2 app/datasources/page.tsx

```typescript
import {Suspense} from 'react';
import DatasourceList from './components/DatasourceList';

export default function DatasourcesPage() {
  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Datasources</h1>
      </div>
      
      <Suspense fallback={<div>Loading...</div>}>
        <DatasourceList />
      </Suspense>
    </div>
  );
}
```

### 2.3 components/DatasourceList.tsx

```typescript
'use client';

import React, {useState} from 'react';
import {Button} from '@components/ui/button';
import {Plus} from 'lucide-react';
import DatasourceCard from './DatasourceCard';
import DatasourceForm from './DatasourceForm';
import {useDatasources} from '../hooks/useDatasources';

export default function DatasourceList() {
  const {datasources, plugins, loading, createDatasource, updateDatasource, deleteDatasource} = useDatasources();
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingDatasource, setEditingDatasource] = useState(null);

  const handleCreate = () => {
    setEditingDatasource(null);
    setIsFormOpen(true);
  };

  const handleEdit = (datasource: any) => {
    setEditingDatasource(datasource);
    setIsFormOpen(true);
  };

  const handleSubmit = async (data: any) => {
    if (editingDatasource) {
      await updateDatasource(editingDatasource.name, data);
    } else {
      await createDatasource(data);
    }
    setIsFormOpen(false);
  };

  const handleDelete = async (name: string) => {
    if (confirm(`Delete datasource "${name}"?`)) {
      await deleteDatasource(name);
    }
  };

  if (loading) return <div>Loading...</div>;

  return (
    <div>
      <div className="mb-4">
        <Button onClick={handleCreate}>
          <Plus className="mr-2 h-4 w-4" />
          Add Datasource
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {datasources.map((ds) => (
          <DatasourceCard
            key={ds.name}
            datasource={ds}
            onEdit={() => handleEdit(ds)}
            onDelete={() => handleDelete(ds.name)}
          />
        ))}
      </div>

      {isFormOpen && (
        <DatasourceForm
          plugins={plugins}
          initialData={editingDatasource}
          onSubmit={handleSubmit}
          onCancel={() => setIsFormOpen(false)}
        />
      )}
    </div>
  );
}
```

### 2.4 components/DatasourceForm.tsx

```typescript
'use client';

import React, {useState} from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import {Dialog, DialogContent, DialogHeader, DialogTitle} from '@components/ui/dialog';

interface DatasourceFormProps {
  plugins: any[];
  initialData?: any;
  onSubmit: (data: any) => void;
  onCancel: () => void;
}

export default function DatasourceForm({plugins, initialData, onSubmit, onCancel}: DatasourceFormProps) {
  const [selectedPlugin, setSelectedPlugin] = useState(initialData?.type || '');
  const [formData, setFormData] = useState(initialData || {});

  const plugin = plugins.find((p) => p.id === selectedPlugin);

  const handlePluginChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const pluginId = e.target.value;
    setSelectedPlugin(pluginId);
    
    // 플러그인이 변경되면 기본값으로 리셋
    const newPlugin = plugins.find((p) => p.id === pluginId);
    if (newPlugin) {
      const defaultData = JSON.parse(newPlugin.settings.sampleFormData || '{}');
      setFormData({...defaultData, type: pluginId});
    }
  };

  return (
    <Dialog open onOpenChange={onCancel}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {initialData ? 'Edit Datasource' : 'Add Datasource'}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {!initialData && (
            <div>
              <label className="block text-sm font-medium mb-2">Datasource Type</label>
              <select
                className="w-full border rounded p-2"
                value={selectedPlugin}
                onChange={handlePluginChange}
              >
                <option value="">Select a datasource type...</option>
                {plugins.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {plugin && (
            <Form
              schema={plugin.settings.jsonSchema}
              uiSchema={plugin.settings.uiSchema}
              formData={formData}
              validator={validator}
              onChange={(e) => setFormData(e.formData)}
              onSubmit={(e) => onSubmit(e.formData)}
            >
              <div className="flex gap-2 mt-4">
                <button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded">
                  {initialData ? 'Update' : 'Create'}
                </button>
                <button type="button" onClick={onCancel} className="px-4 py-2 border rounded">
                  Cancel
                </button>
              </div>
            </Form>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

### 2.5 hooks/useDatasources.ts

```typescript
import {useState, useEffect} from 'react';
import axios from 'axios';

export function useDatasources() {
  const [datasources, setDatasources] = useState<any[]>([]);
  const [plugins, setPlugins] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [dsRes, pluginRes] = await Promise.all([
        axios.get('/plugins/datasources'),
        axios.get('/plugins'),
      ]);

      setDatasources(dsRes.data);
      setPlugins(pluginRes.data.filter((p: any) => p.type === 'datasource'));
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const createDatasource = async (data: any) => {
    await axios.post('/plugins/datasources', data);
    await loadData();
  };

  const updateDatasource = async (name: string, data: any) => {
    await axios.put(`/plugins/datasources/${name}`, data);
    await loadData();
  };

  const deleteDatasource = async (name: string) => {
    await axios.delete(`/plugins/datasources/${name}`);
    await loadData();
  };

  return {
    datasources,
    plugins,
    loading,
    createDatasource,
    updateDatasource,
    deleteDatasource,
  };
}
```

### 2.6 데이터소스 쿼리 테스트 UI (중요!)

데이터소스가 제대로 작동하는지 테스트하는 **쿼리 에디터 페이지**를 추가합니다.

#### app/datasources/query/page.tsx

```typescript
import {Suspense} from 'react';
import QueryEditor from './components/QueryEditor';

export default function DatasourceQueryPage() {
  return (
    <div className="container mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Datasource Query Editor</h1>
      
      <Suspense fallback={<div>Loading...</div>}>
        <QueryEditor />
      </Suspense>
    </div>
  );
}
```

#### app/datasources/query/components/QueryEditor.tsx

```typescript
'use client';

import React, {useState, useEffect} from 'react';
import axios from 'axios';
import AceEditor from 'react-ace';
import 'ace-builds/src-noconflict/mode-sql';
import 'ace-builds/src-noconflict/mode-json';
import 'ace-builds/src-noconflict/theme-github';
import 'ace-builds/src-noconflict/ext-language_tools';
import {Button} from '@components/ui/button';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@components/ui/select';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@components/ui/tabs';
import {Card, CardContent, CardHeader, CardTitle} from '@components/ui/card';

export default function QueryEditor() {
  const [datasources, setDatasources] = useState<any[]>([]);
  const [selectedDatasource, setSelectedDatasource] = useState('');
  const [query, setQuery] = useState('');
  const [queryMode, setQueryMode] = useState<'sql' | 'json'>('sql');
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    loadDatasources();
  }, []);

  const loadDatasources = async () => {
    try {
      const res = await axios.get('/plugins/datasources');
      setDatasources(res.data);
      if (res.data.length > 0) {
        setSelectedDatasource(res.data[0].name);
        
        // 데이터소스 타입에 따라 샘플 쿼리 설정
        const firstDs = res.data[0];
        if (firstDs.type === 'mongodb') {
          setQueryMode('json');
          setQuery(JSON.stringify({
            collection: 'users',
            find: {status: 'active'}
          }, null, 2));
        } else {
          setQueryMode('sql');
          setQuery('SELECT * FROM table_name LIMIT 10');
        }
      }
    } catch (error) {
      console.error('Failed to load datasources:', error);
    }
  };

  const handleDatasourceChange = (name: string) => {
    setSelectedDatasource(name);
    setResult(null);
    setError('');
    
    // 데이터소스 타입에 따라 쿼리 모드 변경
    const ds = datasources.find((d) => d.name === name);
    if (ds?.type === 'mongodb') {
      setQueryMode('json');
      setQuery(JSON.stringify({
        collection: 'users',
        find: {}
      }, null, 2));
    } else {
      setQueryMode('sql');
      setQuery('SELECT * FROM table_name LIMIT 10');
    }
  };

  const handleRunQuery = async () => {
    if (!selectedDatasource || !query) {
      setError('Please select a datasource and enter a query');
      return;
    }

    setLoading(true);
    setError('');
    setResult(null);

    try {
      const response = await axios.post('/plugins/ds/query', {
        queries: [
          {
            id: 'test-query',
            datasourceName: selectedDatasource,
            sql: query,
          },
        ],
      });

      const queryResult = response.data.Results['test-query'];
      
      if (queryResult.Error) {
        setError(queryResult.Error);
      } else {
        setResult(queryResult.Frame);
      }
    } catch (err: any) {
      setError(err.response?.data?.message || err.message || 'Query execution failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Query Configuration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-4 items-center">
            <div className="flex-1">
              <label className="block text-sm font-medium mb-2">Datasource</label>
              <Select value={selectedDatasource} onValueChange={handleDatasourceChange}>
                <SelectTrigger>
                  <SelectValue placeholder="Select datasource..." />
                </SelectTrigger>
                <SelectContent>
                  {datasources.map((ds) => (
                    <SelectItem key={ds.name} value={ds.name}>
                      {ds.name} ({ds.type})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex gap-2 items-end">
              <Button onClick={handleRunQuery} disabled={loading}>
                {loading ? 'Running...' : 'Run Query'}
              </Button>
            </div>
          </div>

          <div>
            <div className="flex justify-between items-center mb-2">
              <label className="block text-sm font-medium">Query</label>
              <Tabs value={queryMode} onValueChange={(v) => setQueryMode(v as 'sql' | 'json')}>
                <TabsList>
                  <TabsTrigger value="sql">SQL</TabsTrigger>
                  <TabsTrigger value="json">JSON</TabsTrigger>
                </TabsList>
              </Tabs>
            </div>
            <AceEditor
              mode={queryMode === 'sql' ? 'sql' : 'json'}
              theme="github"
              name="query-editor"
              value={query}
              onChange={setQuery}
              width="100%"
              height="200px"
              setOptions={{
                enableBasicAutocompletion: true,
                enableLiveAutocompletion: true,
                enableSnippets: true,
                showLineNumbers: true,
                tabSize: 2,
              }}
              className="border rounded"
            />
          </div>
        </CardContent>
      </Card>

      {error && (
        <Card className="border-red-500">
          <CardHeader>
            <CardTitle className="text-red-500">Error</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="text-red-500 text-sm overflow-auto">{error}</pre>
          </CardContent>
        </Card>
      )}

      {result && (
        <Card>
          <CardHeader>
            <CardTitle>Query Result</CardTitle>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="table">
              <TabsList>
                <TabsTrigger value="table">Table</TabsTrigger>
                <TabsTrigger value="json">JSON</TabsTrigger>
              </TabsList>

              <TabsContent value="table" className="mt-4">
                {result.data && result.data.length > 0 ? (
                  <div className="overflow-auto max-h-96">
                    <table className="w-full border-collapse">
                      <thead>
                        <tr className="bg-gray-100">
                          {result.meta?.map((col: any) => (
                            <th key={col.name} className="border p-2 text-left">
                              {col.name}
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {result.data.map((row: any, idx: number) => (
                          <tr key={idx} className="hover:bg-gray-50">
                            {result.meta?.map((col: any) => (
                              <td key={col.name} className="border p-2">
                                {JSON.stringify(row[col.name])}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                    <p className="text-sm text-gray-500 mt-2">
                      {result.data.length} rows returned
                    </p>
                  </div>
                ) : (
                  <p className="text-gray-500">No data returned</p>
                )}
              </TabsContent>

              <TabsContent value="json" className="mt-4">
                <AceEditor
                  mode="json"
                  theme="github"
                  name="result-json"
                  value={JSON.stringify(result, null, 2)}
                  readOnly
                  width="100%"
                  height="400px"
                  setOptions={{
                    showLineNumbers: true,
                    tabSize: 2,
                  }}
                  className="border rounded"
                />
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
```

#### 네비게이션 추가 예시

기존 네비게이션에 메뉴 추가:

```typescript
// app/layout.tsx 또는 navigation component
const navigation = [
  // ... 기존 메뉴들
  {
    name: 'Datasources',
    href: '/datasources',
    icon: DatabaseIcon,
    children: [
      {name: 'Manage', href: '/datasources'},
      {name: 'Query Editor', href: '/datasources/query'},  // 추가
    ],
  },
];
```

---

## 3단계: Panel 플러그인에서 사용

이제 외부 Panel 플러그인에서 새 데이터소스를 사용할 수 있습니다.

### 3.1 src/setParam.ts

```typescript
export function toPanelData(formData: any) {
  // MongoDB 쿼리 생성
  const mongoQuery = {
    collection: formData.collection || 'users',
    find: formData.query ? JSON.parse(formData.query) : {},
  };

  return {
    queries: [
      {
        id: 'mongodb-query',
        datasourceName: formData.datasourceName || 'mongodb-01',  // 생성한 데이터소스 이름
        sql: JSON.stringify(mongoQuery),  // MongoDB는 JSON 쿼리 사용
      },
    ],
  };
}
```

### 3.2 src/MyPanel.tsx

```typescript
import React, {useEffect, useState} from 'react';
import {useList} from '@refinedev/core';

export default function MyPanel({options}: any) {
  const [data, setData] = useState<any[]>([]);

  const {data: queryResult, isLoading} = useList({
    dataProviderName: 'datasource-provider',
    resource: 'rechart',
    queryOptions: {
      enabled: !!options?.queries,
    },
    meta: {
      variables: {
        value: options.queries,  // setParam에서 생성한 쿼리
      },
    },
  });

  useEffect(() => {
    if (queryResult?.data) {
      setData(queryResult.data);
    }
  }, [queryResult]);

  if (isLoading) return <div>Loading from MongoDB...</div>;

  return (
    <div>
      <h3>MongoDB Data</h3>
      <pre>{JSON.stringify(data, null, 2)}</pre>
    </div>
  );
}
```

### 3.3 src/Options.tsx

```typescript
import React, {useEffect, useState} from 'react';
import axios from 'axios';
import {Label} from '@pharos/ui/label';
import {Input} from '@pharos/ui/input';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@pharos/ui/select';

export default function Options({value, onChange}: any) {
  const [datasources, setDatasources] = useState<any[]>([]);

  useEffect(() => {
    // MongoDB 데이터소스 목록 가져오기
    axios.get('/plugins/datasources').then((res) => {
      const mongoDatasources = res.data.filter((ds: any) => ds.type === 'mongodb');
      setDatasources(mongoDatasources);
    });
  }, []);

  return (
    <div className="space-y-4">
      <div>
        <Label>MongoDB Datasource</Label>
        <Select
          value={value?.datasourceName || ''}
          onValueChange={(v) => onChange({...value, datasourceName: v})}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select datasource..." />
          </SelectTrigger>
          <SelectContent>
            {datasources.map((ds) => (
              <SelectItem key={ds.name} value={ds.name}>
                {ds.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div>
        <Label>Collection</Label>
        <Input
          value={value?.collection || ''}
          onChange={(e) => onChange({...value, collection: e.target.value})}
          placeholder="users"
        />
      </div>

      <div>
        <Label>Query (JSON)</Label>
        <Input
          value={value?.query || ''}
          onChange={(e) => onChange({...value, query: e.target.value})}
          placeholder='{"age": {"$gt": 18}}'
        />
      </div>
    </div>
  );
}
```

---

## 실전 예제: MongoDB 플러그인

### 전체 워크플로우

```bash
# 1. MongoDB 플러그인 개발
cd my-datasource-plugin
go build -o mongodb-plugin
sudo mv mongodb-plugin /opt/pharos/plugins/

# 2. Pharos 설정 (HashiCorp 플러그인 자동 로드)
# config.toml
[plugins]
directory = "/opt/pharos/plugins"
auto_discover = true

# 3. Pharos 재시작
systemctl restart pharos

# 4. 데이터소스 관리 UI에서 추가 (또는 API)
curl -X POST http://localhost:8080/plugins/datasources \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mongodb",
    "name": "mongodb-prod",
    "data": {
      "connection_string": "mongodb://prod-mongo:27017",
      "database": "myapp",
      "timeout": 30
    }
  }'

# 5. Panel 플러그인에서 사용
# Options에서 "mongodb-prod" 선택
# Collection: "users"
# Query: {"status": "active"}
```

### 쿼리 실행 흐름

```
1. Panel UI → setParam.ts
   ↓ (MongoDB 쿼리 JSON 생성)
   
2. datasource-provider → /plugins/ds/query
   ↓ (POST request)
   
3. Pharos Backend → HashiCorp Plugin
   ↓ (gRPC 호출)
   
4. MongoDB Plugin → MongoDB Server
   ↓ (쿼리 실행)
   
5. 결과 반환 → Panel 렌더링
```

---

## ✅ 체크리스트

### HashiCorp 플러그인 개발

- [ ] `go.mod` 설정
- [ ] `main.go` 진입점 작성
- [ ] `plugin.go` 데이터소스 로직 구현
- [ ] JSON/UI Schema 작성
- [ ] 빌드 및 배포 (`/opt/pharos/plugins/`)
- [ ] Pharos 재시작 및 플러그인 인식 확인

### 데이터소스 관리 UI (선택)

- [ ] `app/datasources/` 페이지 생성
- [ ] `DatasourceList` 컴포넌트
- [ ] `DatasourceForm` (React JSON Schema Form)
- [ ] `useDatasources` 훅
- [ ] 네비게이션에 메뉴 추가

### Panel 플러그인 통합

- [ ] `setParam.ts`에서 쿼리 생성
- [ ] `Options.tsx`에서 데이터소스 선택 UI
- [ ] `datasource-provider` 사용하여 데이터 조회
- [ ] 에러 핸들링 및 로딩 상태

---

## 🔗 참고

- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin)
- [React JSON Schema Form](https://rjsf-team.github.io/react-jsonschema-form/)
- [Pharos Plugins API](../core/pkg/plugins/README.md)
- [Panel Plugin 가이드](./INTEGRATION_GUIDE.md)

---

**완성!** 🎉

이제 외부 HashiCorp 플러그인으로 데이터소스를 추가하고, Panel 플러그인에서 자유롭게 사용할 수 있습니다!
