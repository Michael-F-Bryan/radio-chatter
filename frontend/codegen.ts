import { CodegenConfig } from "@graphql-codegen/cli";
import path from "path";

const config: CodegenConfig = {
  schema: path.join(__dirname, "..", "pkg", "graphql", "schema.graphql"),
  documents: ["src/**/*.{ts,tsx}"],
  generates: {
    "./src/__generated__/": {
      preset: "client",
      // plugins: [
      //   "typescript",
      //   "typescript-operations",
      //   "typescript-react-apollo",
      // ],
      presetConfig: {
        gqlTagName: "gql",
      },
    },
  },
  ignoreNoDocuments: true,
};

export default config;
