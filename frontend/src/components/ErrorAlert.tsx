import { ApolloError } from "@apollo/client";
import { GraphQLErrors } from "@apollo/client/errors";
import { Alert, AlertTitle, Typography } from "@mui/material";

type GqlError = GraphQLErrors[0];

type Props = {
  error: ApolloError;
  dismiss?: () => void;
};

/**
 * A helper for displaying Apollo errors as an alert.
 */
export default function ErrorAlert({ error, dismiss }: Props) {
  return (
    <Alert severity="error" onClose={dismiss}>
      <AlertTitle>{error.message}</AlertTitle>
      {error.networkError && <NetworkError error={error.networkError} />}
      {error.graphQLErrors.map((e, i) => (
        <GraphQLError key={i} error={e} />
      ))}
      {error.clientErrors.map((e, i) => (
        <ClientError key={i} error={e} />
      ))}

      <details>
        <summary>Stack Trace</summary>
        <pre>{error.stack}</pre>
      </details>
    </Alert>
  );
}

function NetworkError({
  error,
}: {
  error: Exclude<ApolloError["networkError"], null>;
}) {
  if ("statusCode" in error) {
    return <Typography>Status: {error.statusCode}</Typography>;
  }
}

function GraphQLError({
  error: { message, originalError, path },
}: {
  error: GqlError;
}) {
  return (
    <Typography>
      {originalError?.message || message} {path && `(${formatPath(path)})`}
    </Typography>
  );
}

function formatPath(path: readonly (string | number)[]) {
  return path
    .map((elem, i) => {
      if (typeof elem == "string") {
        return i == 0 ? elem : "." + elem;
      } else {
        return `[${i}]`;
      }
    })
    .join();
}

function ClientError({ error }: { error: Error }) {
  return <Typography>{error.message}</Typography>;
}
