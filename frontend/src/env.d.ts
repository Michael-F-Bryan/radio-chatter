declare namespace NodeJS {
  export interface ProcessEnv {
    /**
     * The GraphQL API for the backend.
     */
    NEXT_PUBLIC_GRAPHQL_ENDPOINT: string;
    GITHUB_ID: string;
    GITHUB_SECRET: string;
    NEXTAUTH_URL: string;
  }
}
