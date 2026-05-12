import path from "path";
import swaggerJsdoc from "swagger-jsdoc";

const options: swaggerJsdoc.Options = {
  definition: {
    openapi: "3.0.0",
    info: {
      title: "Jobs API",
      version: "1.0.0",
    },
  },
  apis: [path.resolve("src/**/*.ts")],
};

const swaggerSpec = swaggerJsdoc(options);

export default swaggerSpec;
