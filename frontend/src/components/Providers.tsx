"use client";

import { ApolloLink, HttpLink } from "@apollo/client";
import {
  NextSSRInMemoryCache,
  NextSSRApolloClient,
  SSRMultipartLink,
  ApolloNextAppProvider,
} from "@apollo/experimental-nextjs-app-support/ssr";
import { ThemeProvider } from "@mui/material";
import { SessionProvider } from "next-auth/react";

import { theme } from "../theme";
import { ExistingTransmissionsQuery } from "../__generated__/graphql";

/**
 * A wrapper which is used for providing global context objects.
 *
 * Providers:
 * - `SessionProvider` for Next Auth
 * - `ApolloNextAppProvider` for making Apollo's GraphQL client available
 */
export default function Providers({ children }: React.PropsWithChildren) {
  return (
    <ThemeProvider theme={theme}>
      <SessionProvider>
        <ApolloNextAppProvider makeClient={makeClient}>
          {children}
        </ApolloNextAppProvider>
      </SessionProvider>
    </ThemeProvider>
  );
}

function makeClient() {
  const httpLink = new HttpLink({
    uri: process.env.NEXT_PUBLIC_GRAPHQL_ENDPOINT,
  });

  const cache = new NextSSRInMemoryCache();

  return new NextSSRApolloClient({
    cache,
    link:
      typeof window === "undefined"
        ? ApolloLink.from([
            new SSRMultipartLink({
              stripDefer: true,
            }),
            httpLink,
          ])
        : httpLink,
  });
}
