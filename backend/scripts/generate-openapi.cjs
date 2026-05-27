const swaggerJSDoc = require("swagger-jsdoc");
const path = require("path");
const fs = require("fs");

try {
  const options = {
    definition: {
      openapi: "3.0.0",
      info: {
        title: "Jobs API",
        version: "1.0.0",
      },
    },
    apis: [path.resolve(__dirname, "..", "src", "**", "*.ts")],
  };

  const swaggerSpec = swaggerJSDoc(options);

  const outJson = path.resolve(__dirname, "..", "openapi.json");
  fs.writeFileSync(outJson, JSON.stringify(swaggerSpec, null, 2));
  console.log("Wrote", outJson);

  try {
    const yaml = require("js-yaml");
    const outYaml = path.resolve(__dirname, "..", "openapi.yaml");
    fs.writeFileSync(outYaml, yaml.dump(swaggerSpec));
    console.log("Wrote", outYaml);
  } catch (err) {
    console.warn("js-yaml not available — skipping YAML output");
  }
} catch (err) {
  console.error("Failed to generate OpenAPI spec:", err.message || err);
  process.exit(1);
}
