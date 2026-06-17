import type { CodegenConfig } from '@graphql-codegen/cli';

// Generates typed graphql() + TypedDocumentNodes from the LOCAL schema SDL (no network).
// Output lives under src/__generated__/ — a path our guard/lint/prettier deliberately skip.
const config: CodegenConfig = {
  schema: 'docs/schema.graphqls',
  documents: ['src/graphql/operations/**/*.graphql'],
  ignoreNoDocuments: true,
  generates: {
    'src/__generated__/': {
      preset: 'client',
      config: {
        useTypeImports: true,
        fragmentMasking: false,
        // Map custom scalars so generated types stay precise (no `any`).
        scalars: {
          Time: 'string',
          Upload: 'File',
        },
      },
    },
  },
};

export default config;
