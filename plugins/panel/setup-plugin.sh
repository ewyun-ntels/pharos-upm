#!/bin/bash

# External Plugin Setup Script
# 6. externalPlugins.ts 확인
echo "🔍 Checking externalPlugins.ts..."
EXTERNAL_PLUGINS_FILE="${FRONTEND_DIR}/src/features/panels/plugins/registry/externalPlugins.ts"
if [ -f "${EXTERNAL_PLUGINS_FILE}" ]; then
    echo "✅ externalPlugins.ts exists"
else
    echo "❌ externalPlugins.ts not found"
    echo "   Expected at: ${EXTERNAL_PLUGINS_FILE}"
    exit 1
fi
echo ""

# 7. 빌드 결과 확인
echo "📊 Build artifacts:"
ls -lh "${PLUGIN_DIR}/dist/"
echo ""

# 8. public/plugins 확인
echo "📊 Copied to public/plugins:"
ls -lh "${FRONTEND_DIR}/public/plugins/" 2>/dev/null || echo "⚠️  public/plugins directory not found"
echo ""

# 9. 테스트 가이드 출력다.

set -e

PLUGIN_NAME="my-awesome-plugin"
PLUGIN_DIR="plugins/panel/${PLUGIN_NAME}"
FRONTEND_DIR="core/frontend"

echo "🚀 External Plugin Setup Script"
echo "================================"
echo ""

# 1. 플러그인 디렉토리 확인
echo "📁 Checking plugin directory..."
if [ ! -d "${PLUGIN_DIR}" ]; then
    echo "❌ Plugin directory not found: ${PLUGIN_DIR}"
    exit 1
fi
echo "✅ Plugin directory exists"
echo ""

# 2. 플러그인 의존성 설치
echo "📦 Installing plugin dependencies..."
cd "${PLUGIN_DIR}"
pnpm install
echo "✅ Plugin dependencies installed"
echo ""

# 3. 플러그인 빌드
echo "🔨 Building plugin..."
pnpm build
echo "✅ Plugin built successfully"
echo ""

# 4. 플러그인을 public 폴더로 복사 (Golang embed를 위해)
echo "📋 Copying plugin to frontend/public/plugins..."
pnpm run copy
echo "✅ Plugin copied to public/plugins/"
echo ""

# 5. Frontend tsconfig.json 확인
echo "🔍 Checking frontend tsconfig.json..."
cd - > /dev/null
if grep -q "@external-plugins/\*" "${FRONTEND_DIR}/tsconfig.json"; then
    echo "✅ @external-plugins/* alias already configured"
else
    echo "⚠️  @external-plugins/* alias not found in tsconfig.json"
    echo "   Please add it manually:"
    echo '   "@external-plugins/*": ["../../plugins/panel/*"]'
fi
echo ""

# 6. externalPlugins.ts 확인
echo "🔍 Checking externalPlugins.ts..."
EXTERNAL_PLUGINS_FILE="${FRONTEND_DIR}/src/features/panels/plugins/registry/externalPlugins.ts"
if [ -f "${EXTERNAL_PLUGINS_FILE}" ]; then
    echo "✅ externalPlugins.ts exists"
else
    echo "❌ externalPlugins.ts not found"
    echo "   Expected at: ${EXTERNAL_PLUGINS_FILE}"
    exit 1
fi
echo ""

# 6. 빌드 결과 확인
echo "📊 Build artifacts:"
ls -lh "${PLUGIN_DIR}/dist/"
echo ""

# 7. 테스트 가이드 출력
# 8. 테스트 가이드 출력
echo "✅ Setup complete!"
echo ""
echo "📝 Next steps:"
echo ""
echo "🔹 개발 모드 (Development):"
echo "   # Terminal 1: Frontend dev server"
echo "   cd ${FRONTEND_DIR}"
echo "   pnpm dev"
echo ""
echo "   # Terminal 2: Go server"
echo "   cd core"
echo "   go run cmd/pharos/main.go"
echo ""
echo "🔹 프로덕션 빌드 (Production):"
echo "   # 1. Frontend 빌드 (플러그인 포함)"
echo "   cd ${FRONTEND_DIR}"
echo "   pnpm run build"
echo ""
echo "   # 2. Go 빌드 (frontend embed)"
echo "   cd .."
echo "   go generate ./frontend"
echo "   go build -o bin/pharos cmd/pharos/main.go"
echo ""
echo "   # 3. 실행"
echo "   ./bin/pharos"
echo ""
echo "🔹 브라우저 확인:"
echo "   - Console: [ExternalPlugins] Registered: My Awesome Panel"
echo "   - Dashboard → Add Panel → 'My Awesome Panel' 선택"
echo ""
echo "📚 Documentation:"
echo "   - Golang Embed Guide: plugins/panel/GOLANG_EMBED_GUIDE.md"
echo "   - Integration Guide: plugins/panel/INTEGRATION_GUIDE.md"
echo "   - Plugin README: ${PLUGIN_DIR}/README.md"
