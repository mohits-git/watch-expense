#!/bin/bash

set -e

SWAGGER_UI_VERSION="v5.11.0"
TARGET_DIR="./docs/swagger"

echo "Setting up Swagger UI..."

# Remove and recreate target dir
rm -r "$TARGET_DIR"
mkdir -p "$TARGET_DIR"

# Download Swagger UI dist
curl -L "https://github.com/swagger-api/swagger-ui/archive/refs/tags/${SWAGGER_UI_VERSION}.zip" -o swagger-ui.zip
unzip swagger-ui.zip
# mkdir -p "$TARGET_DIR"
cp -r "swagger-ui-${SWAGGER_UI_VERSION#v}/dist"/* "$TARGET_DIR"
chmod -R 755 "$TARGET_DIR"
rm -rf "swagger-ui-${SWAGGER_UI_VERSION#v}" swagger-ui.zip

# Create swagger-initializer.js
cat > "$TARGET_DIR/swagger-initializer.js" <<EOF
window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  window.ui = SwaggerUIBundle({
    url: "/docs/openapi.yaml",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl,
      () => ({
        components: {
          Topbar: () => null
        }
      })
    ],
    layout: "StandaloneLayout"
  });

  //</editor-fold>
};
EOF

echo "Swagger UI setup complete"
