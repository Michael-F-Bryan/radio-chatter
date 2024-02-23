"use client";
import { useQuery } from "@apollo/client";
import { CircularProgress, Container, Typography } from "@mui/material";
import { notFound } from "next/navigation";

import { gql } from "@/__generated__";
import StreamTable from "@/components/StreamTable";
import ErrorAlert from "@/components/ErrorAlert";

const STREAM_QUERY = gql(/* GraphQL */ `
  query stream($id: ID!) {
    getStreamById(id: $id) {
      displayName
    }
  }
`);

type Props = {
  params: {
    id: string;
  };
};

export default function Stream(props: Props) {
  const streamID = decodeURIComponent(props.params.id);

  const {
    data,
    loading: streamIsLoading,
    error,
  } = useQuery(STREAM_QUERY, { variables: { id: streamID } });

  if (streamIsLoading) {
    return (
      <Container>
        <CircularProgress variant="indeterminate" />
      </Container>
    );
  } else if (error) {
    return (
      <Container>
        <ErrorAlert error={error} />
      </Container>
    );
  } else if (data?.getStreamById) {
    return (
      <Container>
        <Typography variant="h1" sx={{ textAlign: "center" }}>
          {data.getStreamById.displayName}
        </Typography>
        <StreamTable streamID={streamID} />
      </Container>
    );
  } else {
    // We've finished loading, but the stream doesn't exist
    return notFound();
  }
}
